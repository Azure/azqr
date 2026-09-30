// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/rs/zerolog/log"
)

// retailPricesBaseURL is the Azure Retail Prices API endpoint, reused from the
// same public API the region plugin already queries for cross-region cost
// comparison (internal/scanners/plugins/region/cost).
const retailPricesBaseURL = "https://prices.azure.com/api/retail/prices"

// priceResolution is the outcome of resolving a per-token rate for a model,
// regardless of whether it came from the Azure Retail Prices API or the
// embedded Claude snapshot. Source mirrors TokenLens's billing-basis labels
// so downstream sheets can show *why* a rate was or wasn't found.
type priceResolution struct {
	InputPerMTokUSD  float64
	OutputPerMTokUSD float64
	// Source is one of "exact_retail", "claude_ccu_equivalent", or "" when unresolved.
	Source string
}

// retailPriceItem mirrors the fields azqr's region/cost plugin already
// consumes from the Retail Prices API response.
type retailPriceItem struct {
	RetailPrice   float64 `json:"retailPrice"`
	UnitOfMeasure string  `json:"unitOfMeasure"`
	MeterName     string  `json:"meterName"`
	ProductName   string  `json:"productName"`
	ServiceName   string  `json:"serviceName"`
	ArmRegionName string  `json:"armRegionName"`
}

type retailPriceResponse struct {
	Items        []retailPriceItem `json:"Items"`
	NextPageLink string            `json:"NextPageLink"`
}

// meterTokenRe splits a meter or model name into lowercase alphanumeric
// (plus dot) tokens, mirroring TokenLens's own meter-label tokenizer so real
// Azure Retail Prices meter names (which mix hyphens, spaces, and fused
// text) can be matched reliably.
var meterTokenRe = regexp.MustCompile(`[^0-9a-z.]+`)

// tokenizeMeterLabel lowercases s and splits it into non-empty tokens on any
// run of characters other than digits, lowercase letters, and dots.
func tokenizeMeterLabel(s string) []string {
	fields := meterTokenRe.Split(strings.ToLower(s), -1)
	tokens := make([]string, 0, len(fields))
	for _, f := range fields {
		if f != "" {
			tokens = append(tokens, f)
		}
	}
	return tokens
}

// aigovDenylistedMeterTokens excludes meters that price something other than
// standard, non-cached, pay-as-you-go token consumption: batch discounts,
// priority processing/flex tiers, provisioned throughput, fine-tuning,
// non-text modalities, and non-token units of measure.
var aigovDenylistedMeterTokens = map[string]bool{
	"batch": true, "pp": true, "fl": true, "flex": true,
	"ptu": true, "provisioned": true, "reserved": true,
	"finetune": true, "finetuned": true, "ft": true,
	"tool": true, "session": true, "realtime": true,
	"image": true, "audio": true, "speech": true, "video": true,
	"character": true, "minute": true, "hour": true, "page": true, "unit": true,
	"grader": true,
}

var aigovCacheMeterTokens = map[string]bool{"cd": true, "cache": true, "cached": true, "caching": true}
var aigovWriteMeterTokens = map[string]bool{"wr": true, "write": true, "writes": true}
var aigovInputMeterTokens = map[string]bool{"inp": true, "in": true, "input": true, "inputs": true}
var aigovOutputMeterTokens = map[string]bool{"opt": true, "out": true, "outp": true, "output": true, "outputs": true}

// meterTier ranks the geographic/deployment tier implied by a meter's
// tokens; lower is preferred, matching Azure OpenAI's most common
// (cheapest, most broadly available) default deployment type.
func meterTier(tokens []string) int {
	for _, t := range tokens {
		switch t {
		case "gl", "gbl", "glbl", "global":
			return 0
		case "reg", "regnl", "regional":
			return 1
		case "dz", "dzone", "datazone", "dzn":
			return 2
		}
	}
	return 3
}

// meterContext ranks the context-window variant implied by a meter's
// tokens; lower is preferred (short/default context is the common base
// rate; long-context is a separate, pricier meter on some models).
func meterContext(tokens []string) int {
	for _, t := range tokens {
		switch t {
		case "shortco", "short":
			return 0
		case "longco", "long":
			return 2
		}
	}
	return 1
}

// unitsPerMillionMultiplier converts a Retail Prices API unitOfMeasure into
// the multiplier that turns retailPrice into a USD-per-million-tokens rate.
// It only recognizes the token-billing units Azure OpenAI actually uses
// ("1M" and "1K"); anything else is left unresolved rather than guessed.
func unitsPerMillionMultiplier(unitOfMeasure string) (float64, bool) {
	switch strings.ToLower(strings.TrimSpace(unitOfMeasure)) {
	case "1m":
		return 1, true
	case "1k":
		return 1000, true
	default:
		return 0, false
	}
}

// modelSearchTokenSequences returns the token sequences to look for inside a
// meter's tokens, trying the full model name and, since many meters drop the
// publisher prefix (e.g. "gpt-6-sol" appears as just "6-sol"), the same
// sequence with a leading "gpt" token removed.
func modelSearchTokenSequences(modelName string) [][]string {
	tokens := tokenizeMeterLabel(modelName)
	if len(tokens) == 0 {
		return nil
	}
	sequences := [][]string{tokens}
	if tokens[0] == "gpt" && len(tokens) > 1 {
		sequences = append(sequences, tokens[1:])
	}
	return sequences
}

// containsTokenSubsequence reports whether needle appears as a contiguous
// run within haystack.
func containsTokenSubsequence(haystack, needle []string) bool {
	if len(needle) == 0 || len(needle) > len(haystack) {
		return false
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j, n := range needle {
			if haystack[i+j] != n {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// aigovMeterCandidate is one exact, unambiguous meter match for a model's
// input or output token dimension, ranked so the cheapest/most-common tier
// and context variant can be preferred when several exact matches exist.
type aigovMeterCandidate struct {
	pricePerMTokUSD float64
	tier            int
	context         int
}

// betterAigovCandidate reports whether candidate b should replace the
// current best candidate a (lower tier/context rank wins; nil a always
// loses).
func betterAigovCandidate(a *aigovMeterCandidate, b aigovMeterCandidate) bool {
	if a == nil {
		return true
	}
	if b.tier != a.tier {
		return b.tier < a.tier
	}
	return b.context < a.context
}

// resolveAzureOpenAIPrice queries the Azure Retail Prices API for the input
// and output token meters of an Azure OpenAI / AI Services model deployed in
// region. It never guesses from a similar model or region: ok is false when
// no unambiguous, non-cached, pay-as-you-go meter pair is found.
//
// Azure rebranded the Retail Prices API's serviceName for these meters from
// "Cognitive Services" to "Foundry Models". Meter names are inconsistently
// formatted (hyphenated, space-separated, or fused; with or without a "gpt"
// prefix) and can encode several axes at once (input/output, cached vs.
// standard, global/regional/data-zone, short/long context, standard vs.
// priority-processing/flex/batch). Matching tokenizes both the model name
// and each meter name, requires a contiguous token match of the model name,
// excludes non-standard billing axes (batch, priority processing, flex,
// provisioned/PTU, fine-tuning, non-text modalities, cached/cache-write
// rates), and prefers the global tier / short-context variant when more
// than one exact match remains — mirroring the common default Azure OpenAI
// deployment rather than guessing an unrelated model or rate.
func resolveAzureOpenAIPrice(ctx context.Context, httpClient *az.HttpClient, modelName, region string) (priceResolution, bool) {
	filter := fmt.Sprintf("serviceName eq 'Foundry Models' and armRegionName eq '%s' and priceType eq 'Consumption'", odataEscapeAigov(region))
	pageURL := retailPricesBaseURL + "?$filter=" + url.QueryEscape(filter)

	searchSequences := modelSearchTokenSequences(modelName)
	if len(searchSequences) == 0 {
		return priceResolution{}, false
	}
	// Embedding models bill only on input tokens, so their meters often
	// omit an explicit "Inp" marker (just "... Tokens"); treat those bare
	// meters as input meters instead of skipping them as ambiguous.
	isEmbeddingModel := strings.Contains(canonicalModelName(modelName), "embedding")

	var bestInput, bestOutput *aigovMeterCandidate

	for pageURL != "" {
		body, err := httpClient.Do(ctx, pageURL)
		if err != nil {
			log.Debug().Err(err).Str("model", modelName).Str("region", region).Msg("failed to query retail prices for AI Gov cost")
			return priceResolution{}, false
		}

		var resp retailPriceResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return priceResolution{}, false
		}

		for _, item := range resp.Items {
			tokens := tokenizeMeterLabel(item.MeterName)
			if !containsAny(tokens, "tokens") {
				continue
			}

			matched := false
			for _, seq := range searchSequences {
				if containsTokenSubsequence(tokens, seq) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}

			if hasAnyDenylistedToken(tokens) || hasAnyToken(tokens, aigovCacheMeterTokens) || hasAnyToken(tokens, aigovWriteMeterTokens) {
				continue
			}

			isInput := hasAnyToken(tokens, aigovInputMeterTokens)
			isOutput := hasAnyToken(tokens, aigovOutputMeterTokens)
			if !isInput && !isOutput {
				if !isEmbeddingModel {
					// Neither dimension marker present - ambiguous, skip rather than guess.
					continue
				}
				isInput = true
			} else if isInput && isOutput {
				// Both markers present - ambiguous, skip rather than guess.
				continue
			}

			multiplier, ok := unitsPerMillionMultiplier(item.UnitOfMeasure)
			if !ok {
				continue
			}

			candidate := aigovMeterCandidate{
				pricePerMTokUSD: item.RetailPrice * multiplier,
				tier:            meterTier(tokens),
				context:         meterContext(tokens),
			}

			if isInput && betterAigovCandidate(bestInput, candidate) {
				bestInput = &candidate
			}
			if isOutput && betterAigovCandidate(bestOutput, candidate) {
				bestOutput = &candidate
			}
		}

		pageURL = resp.NextPageLink
	}

	switch {
	case bestInput != nil && bestOutput != nil:
		return priceResolution{
			InputPerMTokUSD:  bestInput.pricePerMTokUSD,
			OutputPerMTokUSD: bestOutput.pricePerMTokUSD,
			Source:           "exact_retail",
		}, true
	case bestInput != nil && strings.Contains(canonicalModelName(modelName), "embedding"):
		// Embedding models bill only on input tokens; there is no output
		// meter to find, so zero output cost is a fact, not a guess.
		return priceResolution{
			InputPerMTokUSD: bestInput.pricePerMTokUSD,
			Source:          "exact_retail",
		}, true
	default:
		return priceResolution{}, false
	}
}

// containsAny reports whether any of tokens equals want.
func containsAny(tokens []string, want string) bool {
	for _, t := range tokens {
		if t == want {
			return true
		}
	}
	return false
}

// hasAnyToken reports whether tokens contains any key present in set.
func hasAnyToken(tokens []string, set map[string]bool) bool {
	for _, t := range tokens {
		if set[t] {
			return true
		}
	}
	return false
}

// hasAnyDenylistedToken reports whether tokens contains any excluded axis
// (batch, priority processing, provisioned/PTU, fine-tuning, non-text
// modalities, or non-token billing units).
func hasAnyDenylistedToken(tokens []string) bool {
	return hasAnyToken(tokens, aigovDenylistedMeterTokens)
}

// odataEscapeAigov escapes a string for use in an OData $filter expression by
// doubling single quotes. Named distinctly from the region plugin's
// odataEscape to avoid any cross-package confusion; logic is identical.
func odataEscapeAigov(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// resolvePrice resolves a per-token price for modelName in region, trying the
// Azure Retail Prices API first and falling back to the embedded Claude
// snapshot for Anthropic models. It never invents a rate: ok is false when
// neither source has an exact match.
func resolvePrice(ctx context.Context, httpClient *az.HttpClient, modelName, region string) (priceResolution, bool) {
	if price, ok := resolveAzureOpenAIPrice(ctx, httpClient, modelName, region); ok {
		return price, true
	}
	if inferPublisher(modelName) == "anthropic" {
		return resolveClaudePrice(modelName)
	}
	return priceResolution{}, false
}

// estimatedCostUSD computes the estimated USD cost of the given token counts
// under the resolved price. Claude rates are already expressed as USD per
// MTok; the CCU indirection is Azure's internal billing mechanism and does
// not change the dollar amount actually charged.
func estimatedCostUSD(price priceResolution, inputTokens, outputTokens float64) float64 {
	return (inputTokens/1_000_000)*price.InputPerMTokUSD + (outputTokens/1_000_000)*price.OutputPerMTokUSD
}
