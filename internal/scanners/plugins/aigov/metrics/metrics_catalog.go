// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package metrics

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
)

// Canonical Azure Monitor dimension names.
const (
	dimDeployment = "ModelDeploymentName"
	dimModel      = "ModelName"
	dimStatusCode = "StatusCode"
	dimTier       = "ServiceTierRequest"
	dimContext    = "ContextLength"
	dimStream     = "StreamType"
)

// Metric field priority lists, highest priority first. They mirror TokenLens's
// foundry/metrics_catalog.py so accounts that expose only a fallback name
// (non-Azure-OpenAI-branded models, Foundry partner models) are still covered.
var (
	inputMetricCandidates     = []string{"InputTokens", "ProcessedPromptTokens", "PromptTokenCount"}
	outputMetricCandidates    = []string{"OutputTokens", "GeneratedTokens", "CompletionTokenCount"}
	requestMetricCandidates   = []string{"ModelRequests", "AzureOpenAIRequests", "TotalCalls"}
	throttledMetricCandidates = []string{"ThrottledCalls", "AzureOpenAIThrottledRequests"}
	latencyMetricCandidates   = []string{"AzureOpenAITimeToResponse", "TimeToResponse", "AzureOpenAINormalizedTTFTInMS", "NormalizedTimeToFirstByte", "Latency"}
)

// Definition is the subset of an Azure Monitor metric definition the
// catalog needs: the metric name and the dimensions it can be split by.
type Definition struct {
	Name       string
	Dimensions []string
}

// Plan is the set of metrics (and optional dimensions) that will be
// queried for one account, resolved from its own metric definitions.
type Plan struct {
	Input, Output, Requests, Throttled, Latency string
	// LatencyFallbacks are lower-priority latency metrics the account also
	// exposes. A higher-priority metric can exist yet emit no data (for
	// example AzureOpenAITimeToResponse on Foundry partner models), so each is
	// queried and the first one that returns points wins.
	LatencyFallbacks []string
	// StatusSplit is true when the requests metric exposes StatusCode.
	StatusSplit bool
	// Optional tier/context/stream dimensions shared by both token metrics.
	TierDim, ContextDim, StreamDim bool
	// Notes records coverage gaps found while resolving (missing metrics,
	// metrics excluded because they cannot be split per deployment).
	Notes []string
}

// Signature identifies plans that can share one batch query set.
func (p Plan) Signature() string {
	return strings.Join([]string{
		p.Input, p.Output, p.Requests, p.Throttled, p.Latency, strings.Join(p.LatencyFallbacks, ","),
		strconv.FormatBool(p.StatusSplit), strconv.FormatBool(p.TierDim),
		strconv.FormatBool(p.ContextDim), strconv.FormatBool(p.StreamDim),
	}, "|")
}

// DefaultPlan is used when metric definitions cannot be listed. It
// relies on the documented primary Azure OpenAI metric names; queries that the
// account does not support fail individually and are reported as gaps.
func DefaultPlan() Plan {
	return Plan{
		Input:       "InputTokens",
		Output:      "OutputTokens",
		Requests:    "AzureOpenAIRequests",
		StatusSplit: true,
		Latency:     "AzureOpenAITimeToResponse",
		Notes:       []string{"metric definitions unavailable; documented default metric names were used"},
	}
}

// canonicalDimension maps case-insensitive Azure dimension spellings to their
// canonical name, preserving unknown dimensions verbatim.
func canonicalDimension(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "modeldeploymentname", "deploymentname":
		return dimDeployment
	case "modelname", "model":
		return dimModel
	case "statuscode":
		return dimStatusCode
	case "servicetierrequest":
		return dimTier
	case "contextlength":
		return dimContext
	case "streamtype":
		return dimStream
	default:
		return strings.TrimSpace(name)
	}
}

type definitionIndex map[string]map[string]bool

func indexDefinitions(defs []Definition) definitionIndex {
	idx := make(definitionIndex, len(defs))
	for _, d := range defs {
		dims := map[string]bool{}
		for _, dim := range d.Dimensions {
			dims[canonicalDimension(dim)] = true
		}
		idx[strings.ToLower(d.Name)] = dims
	}
	return idx
}

// pick returns the first candidate present in the account's definitions that
// can be split per deployment and model, appending a note to notes for each
// skipped candidate or when none is usable.
func (idx definitionIndex) pick(field string, candidates []string, notes *[]string) string {
	for _, name := range candidates {
		dims, ok := idx[strings.ToLower(name)]
		if !ok {
			continue
		}
		if !dims[dimDeployment] || !dims[dimModel] {
			*notes = append(*notes, field+": "+name+" skipped (not filterable by deployment and model)")
			continue
		}
		return name
	}
	*notes = append(*notes, field+": no supported metric exposed by this account")
	return ""
}

// pickTierCapable returns the first candidate splittable by deployment, model
// and service tier, or "" when none is.
func (idx definitionIndex) pickTierCapable(candidates []string) string {
	for _, name := range candidates {
		if idx.has(name, dimDeployment) && idx.has(name, dimModel) && idx.has(name, dimTier) {
			return name
		}
	}
	return ""
}

func (idx definitionIndex) has(metric, dim string) bool {
	dims, ok := idx[strings.ToLower(metric)]
	return ok && dims[dim]
}

// ResolvePlan selects, for each canonical field, the highest-priority
// metric the account actually exposes and can split per deployment.
func ResolvePlan(defs []Definition) Plan {
	idx := indexDefinitions(defs)
	var plan Plan
	var notes []string
	plan.Input = idx.pick("input tokens", inputMetricCandidates, &notes)
	plan.Output = idx.pick("output tokens", outputMetricCandidates, &notes)
	// InputTokens/OutputTokens carry no ServiceTierRequest dimension; prefer a
	// tier-capable pair (ProcessedPromptTokens/GeneratedTokens) when exposed.
	if !idx.has(plan.Input, dimTier) || !idx.has(plan.Output, dimTier) {
		in, out := idx.pickTierCapable(inputMetricCandidates), idx.pickTierCapable(outputMetricCandidates)
		if in != "" && out != "" {
			plan.Input, plan.Output = in, out
		}
	}
	plan.Requests = idx.pick("requests", requestMetricCandidates, &notes)
	plan.Latency = idx.pick("latency", latencyMetricCandidates, &notes)
	for _, name := range latencyMetricCandidates {
		if name != plan.Latency && idx.has(name, dimDeployment) && idx.has(name, dimModel) {
			plan.LatencyFallbacks = append(plan.LatencyFallbacks, name)
		}
	}

	if plan.Requests != "" {
		plan.StatusSplit = idx.has(plan.Requests, dimStatusCode)
	}
	if plan.Requests != "" && !plan.StatusSplit {
		plan.Throttled = idx.pick("throttled requests", throttledMetricCandidates, &notes)
	}

	if plan.Input != "" && plan.Output != "" {
		both := func(dim string) bool { return idx.has(plan.Input, dim) && idx.has(plan.Output, dim) }
		plan.TierDim = both(dimTier)
		plan.ContextDim = both(dimContext)
		plan.StreamDim = both(dimStream)
	}

	plan.Notes = notes
	return plan
}

// FetchDefinitions lists the metric definitions of one account.
func FetchDefinitions(ctx context.Context, cred azcore.TokenCredential, subscriptionID, resourceID string) ([]Definition, error) {
	client, err := armmonitor.NewMetricDefinitionsClient(subscriptionID, cred, az.NewDefaultClientOptions())
	if err != nil {
		return nil, err
	}
	pager := client.NewListPager(resourceID, nil)
	var defs []Definition
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, d := range page.Value {
			if d == nil || d.Name == nil || d.Name.Value == nil {
				continue
			}
			def := Definition{Name: *d.Name.Value}
			for _, dim := range d.Dimensions {
				if dim != nil && dim.Value != nil {
					def.Dimensions = append(def.Dimensions, *dim.Value)
				}
			}
			defs = append(defs, def)
		}
	}
	return defs, nil
}

// Status classes used by the throttling and insight logic.
const (
	statusSuccess     = "success"
	statusRateLimited = "rate_limited"
	statusFailed      = "failed"
	StatusUnknown     = "unknown"
)

// ClassifyStatus maps an HTTP status code dimension value to a class: 429 is
// rate limited, 2xx success, anything else failed. Missing values stay
// unknown rather than being counted as success.
func ClassifyStatus(code string) string {
	code = strings.TrimSpace(code)
	if code == "" || strings.EqualFold(code, "unknown") {
		return StatusUnknown
	}
	n, err := strconv.Atoi(code)
	if err != nil {
		return statusFailed
	}
	switch {
	case n == 429:
		return statusRateLimited
	case n >= 200 && n < 300:
		return statusSuccess
	default:
		return statusFailed
	}
}

// longContextThresholdTokens is the documented prompt-token boundary where
// the long-context meter takes over for models that publish a split.
const longContextThresholdTokens = 128_000

// contextBand normalises a ContextLength dimension value into short, long or
// unknown. An unparsable value is preserved as unknown, never assumed short.
func contextBand(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return StatusUnknown
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return StatusUnknown
	}
	if n >= longContextThresholdTokens {
		return "long"
	}
	return "short"
}

// normalizeTier lowercases a service tier value, returning "unknown" for an
// empty one so tier-less series are not mistaken for Standard.
func normalizeTier(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return StatusUnknown
	}
	return v
}

// SortedKeys returns map keys in deterministic order.
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
