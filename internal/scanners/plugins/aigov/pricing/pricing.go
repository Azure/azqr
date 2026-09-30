// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pricing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/metrics"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/modelinfo"
	"github.com/rs/zerolog/log"
)

// retailPricesBaseURL is the Azure Retail Prices API endpoint, reused from the
// same public API the region plugin already queries for cross-region cost
// comparison (internal/scanners/plugins/region/cost).
const retailPricesBaseURL = "https://prices.azure.com/api/retail/prices"

// Resolution is the outcome of resolving a per-token rate for a model,
// regardless of whether it came from the Azure Retail Prices API or the
// embedded Claude snapshot. Source mirrors TokenLens's billing-basis labels
// so downstream sheets can show *why* a rate was or wasn't found.
type Resolution struct {
	InputPerMTokUSD  float64
	OutputPerMTokUSD float64
	// Source is one of "exact_retail", "claude_ccu_equivalent", or "" when unresolved.
	Source string
	// Tier and Context describe the meter that was matched ("global",
	// "regional", "data_zone" or "untagged"; "short", "long" or "untagged").
	Tier    string
	Context string
	// Note explains an unresolved price, for example when several model
	// versions are priced differently and the deployment's version cannot
	// choose between them.
	Note string
}

// RetailItem mirrors the fields azqr's region/cost plugin already
// consumes from the Retail Prices API response.
type RetailItem struct {
	RetailPrice   float64 `json:"retailPrice"`
	UnitOfMeasure string  `json:"unitOfMeasure"`
	MeterName     string  `json:"meterName"`
	ProductName   string  `json:"productName"`
	ServiceName   string  `json:"serviceName"`
	ArmRegionName string  `json:"armRegionName"`
}

type retailPriceResponse struct {
	Items        []RetailItem `json:"Items"`
	NextPageLink string       `json:"NextPageLink"`
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
var aigovInputMeterTokens = map[string]bool{"inp": true, "inpt": true, "in": true, "input": true, "inputs": true}
var aigovOutputMeterTokens = map[string]bool{"opt": true, "out": true, "outp": true, "outpt": true, "output": true, "outputs": true}

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

// aigovAllowedLeadingMeterTokens are the only tokens that may precede a model
// name inside a meter name. Anything else (for example "grok" before "4.1")
// means the meter belongs to a different model that merely ends the same way.
var aigovAllowedLeadingMeterTokens = map[string]bool{"azure": true, "openai": true, "oai": true, "gpt": true}

// aigovAllowedTrailingMeterTokens are the tokens that may directly follow a
// model name inside a meter name: dimension, tier, context and unit markers.
// A model variant word ("mini", "pro", "aud", "tts", ...) is not in this set,
// so "gpt-5" does not match the "gpt-5 pro" meters.
var aigovAllowedTrailingMeterTokens = map[string]bool{
	"tokens": true, "token": true, "1m": true, "1k": true,
	"cd": true, "cache": true, "cached": true, "caching": true, "wr": true, "write": true, "writes": true,
	"gl": true, "gbl": true, "glbl": true, "global": true,
	"reg": true, "regnl": true, "regional": true,
	"dz": true, "dzone": true, "datazone": true, "dzn": true, "data": true, "zone": true,
	"shortco": true, "short": true, "longco": true, "long": true,
	"inp": true, "inpt": true, "in": true, "input": true, "inputs": true,
	"opt": true, "out": true, "outp": true, "outpt": true, "output": true, "outputs": true,
}

func isNumericToken(t string) bool {
	if t == "" {
		return false
	}
	for _, r := range t {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// containsTokenSubsequence reports whether needle appears in haystack as a
// whole model name: a contiguous run that is not preceded by another model's
// name and not followed by a model-variant word. Numeric tokens directly
// after the name are treated as a version or date stamp.
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
		if !match {
			continue
		}
		if i > 0 && !aigovAllowedLeadingMeterTokens[haystack[i-1]] {
			continue
		}
		end := i + len(needle)
		if end < len(haystack) {
			next := haystack[end]
			if !aigovAllowedTrailingMeterTokens[next] && !isNumericToken(next) {
				continue
			}
		}
		return true
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
	// version is the digits of the model version in the meter name
	// ("0806", "20241120"), or "" when the meter carries none.
	version string
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
func resolveAzureOpenAIPrice(ctx context.Context, httpClient *az.HttpClient, modelName, region string) (Resolution, bool) {
	return resolveAzureOpenAIPriceFor(ctx, httpClient, modelName, region, modelinfo.ModeUnknown, metrics.StatusUnknown)
}

// resolveAzureOpenAIPriceFor resolves the rate for an observed deployment
// mode and context band. A known mode only accepts a meter of the same
// geography (or an untagged meter): a Global rate is never applied to a Data
// Zone deployment. The long-context band only accepts a long-context meter,
// so long-context usage is withheld rather than priced at the short rate.
// An unknown mode falls back to TokenLens's Global default.
func resolveAzureOpenAIPriceFor(ctx context.Context, httpClient *az.HttpClient, modelName, region, mode, band string) (Resolution, bool) {
	if len(modelSearchTokenSequences(modelName)) == 0 {
		return Resolution{}, false
	}
	items, ok := FetchRetailItems(ctx, httpClient, region)
	if !ok {
		return Resolution{}, false
	}
	return resolvePriceFromItems(items, modelName, mode, band, "")
}

// FetchRetailItems pages through every Foundry Models consumption meter for
// region. Callers that price several models, modes or bands in one region
// should fetch once and reuse the items with resolvePriceFromItems.
func FetchRetailItems(ctx context.Context, httpClient *az.HttpClient, region string) ([]RetailItem, bool) {
	filter := fmt.Sprintf("serviceName eq 'Foundry Models' and armRegionName eq '%s' and priceType eq 'Consumption'", odataEscapeAigov(region))
	pageURL := retailPricesBaseURL + "?$filter=" + url.QueryEscape(filter)

	var items []RetailItem
	for pageURL != "" {
		body, err := httpClient.Do(ctx, pageURL)
		if err != nil {
			log.Debug().Err(err).Str("region", region).Msg("failed to query retail prices for AI Gov cost")
			return nil, false
		}
		var resp retailPriceResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			log.Debug().Err(err).Str("region", region).Msg("failed to decode retail prices response for AI Gov cost")
			return nil, false
		}
		items = append(items, resp.Items...)
		pageURL = resp.NextPageLink
	}
	return items, true
}

// resolvePriceFromItems selects the input and output meters for modelName from
// already-fetched regional retail items.
func resolvePriceFromItems(items []RetailItem, modelName, mode, band, version string) (Resolution, bool) {
	searchSequences := modelSearchTokenSequences(modelName)
	if len(searchSequences) == 0 {
		return Resolution{}, false
	}
	// Embedding models bill only on input tokens, so their meters often
	// omit an explicit "Inp" marker (just "... Tokens"); treat those bare
	// meters as input meters instead of skipping them as ambiguous.
	isEmbeddingModel := strings.Contains(modelinfo.CanonicalModelName(modelName), "embedding")

	var inputs, outputs []aigovMeterCandidate

	for _, item := range items {
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
			version:         meterVersion(tokens),
		}

		if isInput {
			inputs = append(inputs, candidate)
		}
		if isOutput {
			outputs = append(outputs, candidate)
		}
	}

	bestInput, inRange := selectCandidate(inputs, mode, band, version)
	bestOutput, outRange := selectCandidate(outputs, mode, band, version)
	if (bestInput == nil && inRange != "") || (bestOutput == nil && outRange != "") {
		return Resolution{Note: ambiguityNote(inRange, outRange, version)}, false
	}
	switch {
	case bestInput != nil && bestOutput != nil:
		return Resolution{
			InputPerMTokUSD:  bestInput.pricePerMTokUSD,
			OutputPerMTokUSD: bestOutput.pricePerMTokUSD,
			Source:           "exact_retail",
			Tier:             tierLabel(bestInput.tier),
			Context:          contextLabel(bestInput.context),
		}, true
	case bestInput != nil && strings.Contains(modelinfo.CanonicalModelName(modelName), "embedding"):
		// Embedding models bill only on input tokens; there is no output
		// meter to find, so zero output cost is a fact, not a guess.
		return Resolution{
			InputPerMTokUSD: bestInput.pricePerMTokUSD,
			Source:          "exact_retail",
			Tier:            tierLabel(bestInput.tier),
			Context:         contextLabel(bestInput.context),
		}, true
	default:
		return Resolution{}, false
	}
}

func tierLabel(rank int) string {
	switch rank {
	case 0:
		return "global"
	case 1:
		return "regional"
	case 2:
		return "data_zone"
	default:
		return "untagged"
	}
}

func contextLabel(rank int) string {
	switch rank {
	case 0:
		return "short"
	case 2:
		return "long"
	default:
		return "untagged"
	}
}

// wantedTierRank maps a deployment mode to the meter tier rank it may use;
// ok is false for modes with no geography preference (unknown, batch).
func wantedTierRank(mode string) (int, bool) {
	switch mode {
	case modelinfo.ModeGlobal, modelinfo.ModeGlobalProvisioned:
		return 0, true
	case modelinfo.ModeRegional, modelinfo.ModeRegionalProvisioned:
		return 1, true
	case modelinfo.ModeDataZone, modelinfo.ModeDataZoneProvisioned:
		return 2, true
	default:
		return 0, false
	}
}

// meterVersion extracts the model version a meter name carries: a four-digit
// MMDD token ("0806") or a yyyy mm dd token run ("2024 08 06"). "" if none.
func meterVersion(tokens []string) string {
	for i, t := range tokens {
		if !isNumericToken(t) {
			continue
		}
		if len(t) == 8 {
			return t
		}
		if len(t) == 4 {
			if strings.HasPrefix(t, "20") && i+2 < len(tokens) && len(tokens[i+1]) == 2 && len(tokens[i+2]) == 2 &&
				isNumericToken(tokens[i+1]) && isNumericToken(tokens[i+2]) {
				return t + tokens[i+1] + tokens[i+2]
			}
			return t
		}
	}
	return ""
}

// versionMatches compares a deployment version ("2024-08-06") with a meter
// version ("0806" or "20240806").
func versionMatches(deployment, meter string) bool {
	d := strings.Join(tokenizeMeterLabel(deployment), "")
	return d != "" && meter != "" && (d == meter || strings.HasSuffix(d, meter))
}

// knownVersion returns "" for placeholder versions such as "N/A" so they
// are treated as no version rather than matched against meters.
func knownVersion(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "n/a", "na", "unknown", "none", "-":
		return ""
	}
	return strings.TrimSpace(v)
}

// ambiguityNote describes the version-dependent price range for the remediation text.
func ambiguityNote(ranges ...string) string {
	for _, r := range ranges {
		if r != "" {
			return r
		}
	}
	return ""
}

// selectCandidate picks the best meter for the mode, context band and model
// version, or nil when none is acceptable. When several meters of equal rank
// are priced differently and the deployment version cannot choose between
// them, it returns nil with a description of the price range.
func selectCandidate(cands []aigovMeterCandidate, mode, band, version string) (*aigovMeterCandidate, string) {
	version = knownVersion(version)
	wantTier, strict := wantedTierRank(mode)
	var eligible []aigovMeterCandidate
	for _, c := range cands {
		if strict && c.tier != wantTier && c.tier != 3 {
			continue
		}
		if band == "long" {
			if c.context != 2 {
				continue
			}
		} else if c.context == 2 {
			continue
		}
		eligible = append(eligible, c)
	}

	// A known deployment version keeps only meters of that version (or
	// meters that carry no version).
	if version != "" {
		var matched []aigovMeterCandidate
		for _, c := range eligible {
			if c.version == "" || versionMatches(version, c.version) {
				matched = append(matched, c)
			}
		}
		eligible = matched
	}

	var best *aigovMeterCandidate
	ambiguous := false
	lo, hi := 0.0, 0.0
	for i, c := range eligible {
		if i == 0 || c.pricePerMTokUSD < lo {
			lo = c.pricePerMTokUSD
		}
		if c.pricePerMTokUSD > hi {
			hi = c.pricePerMTokUSD
		}
		if best != nil && c.tier == best.tier && c.context == best.context && c.pricePerMTokUSD != best.pricePerMTokUSD {
			ambiguous = true
		}
		if betterAigovCandidate(best, c) {
			cc := c
			best = &cc
			ambiguous = false
		}
	}
	if ambiguous {
		why := "the deployment's model version is unknown"
		if version != "" {
			why = fmt.Sprintf("no meter matches the deployment's model version %s", version)
		}
		return nil, fmt.Sprintf("Several model versions are priced between $%.2f and $%.2f per 1M tokens and %s; set or confirm the model version to resolve.", lo, hi, why)
	}
	return best, ""
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

// ResolveFromRegionItems is resolvePrice over already-fetched regional
// retail items, so one scan pages each region's meter list only once.
func ResolveFromRegionItems(items []RetailItem, modelName, mode, band, version string) (Resolution, bool) {
	price, ok := resolvePriceFromItems(items, modelName, mode, band, version)
	if ok {
		return price, true
	}
	if modelinfo.InferPublisher(modelName) == "anthropic" {
		return resolveClaudePrice(modelName)
	}
	return price, false
}

// EstimatedCostUSD computes the estimated USD cost of the given token counts
// under the resolved price. Claude rates are already expressed as USD per
// MTok; the CCU indirection is Azure's internal billing mechanism and does
// not change the dollar amount actually charged.
func EstimatedCostUSD(price Resolution, inputTokens, outputTokens float64) float64 {
	return (inputTokens/1_000_000)*price.InputPerMTokUSD + (outputTokens/1_000_000)*price.OutputPerMTokUSD
}

// ResolvePTUHourlyUSD returns the Retail price of one provisioned throughput
// unit hour for the deployment mode ("Provisioned Managed Global|Data Zone|
// Regional Unit", billed 1/Hour). Several non-OpenAI products reuse these
// meter names, so only the "Azure OpenAI" product is considered.
func ResolvePTUHourlyUSD(items []RetailItem, mode string) (float64, bool) {
	var want string
	switch mode {
	case modelinfo.ModeGlobal:
		want = "provisioned managed global unit"
	case modelinfo.ModeDataZone:
		want = "provisioned managed data zone unit"
	case modelinfo.ModeRegional:
		want = "provisioned managed regional unit"
	default:
		return 0, false
	}
	for _, it := range items {
		if strings.EqualFold(strings.TrimSpace(it.MeterName), want) &&
			strings.EqualFold(strings.TrimSpace(it.ProductName), "Azure OpenAI") &&
			strings.EqualFold(strings.TrimSpace(it.UnitOfMeasure), "1/Hour") && it.RetailPrice > 0 {
			return it.RetailPrice, true
		}
	}
	return 0, false
}
