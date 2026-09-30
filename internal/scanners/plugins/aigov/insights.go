// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/azqr/internal/scanners/plugins/aigov/metrics"
)

// insightFinding is one evidence-grounded finding for an account. The rules and
// thresholds are adapted from TokenLens-for-Azure's aggregate-evidence rules;
// trace-level rules are out of scope because only Azure Monitor aggregates are
// collected.
type insightFinding struct {
	Priority       int
	ID             string
	Category       string
	Title          string
	Observation    string
	Evidence       string
	Impact         string
	Recommendation string
	Verification   string
	Confidence     string
	Applicability  string
	Missing        string
	Caveats        string
}

// accountTotals are the account-wide aggregates the insight rules evaluate.
type accountTotals struct {
	Requests, RateLimited, Failed float64
	Input, Output                 float64
	LatencyN                      int
	RequestsComplete              bool
	Elapsed, Observed             int
	// Gaps are collection problems and metric-plan notes from Azure Monitor.
	Gaps  []string
	Grain string
}

func totalsFor(d *metrics.AccountData) accountTotals {
	t := accountTotals{RequestsComplete: d.RequestsComplete, Elapsed: d.Elapsed, Observed: d.Observed(), Gaps: d.Gaps, Grain: d.Grain}
	for _, buckets := range d.Buckets {
		for _, b := range buckets {
			t.Requests += b.Requests
			t.RateLimited += b.RateLimited
			t.Failed += b.Failed
			t.Input += b.Input
			t.Output += b.Output
			t.LatencyN += b.LatencyN
		}
	}
	return t
}

// joinGaps renders at most three collection notes for a table cell.
func joinGaps(gaps []string) string {
	const maxShown = 3
	shown := gaps
	if len(shown) > maxShown {
		shown = shown[:maxShown]
	}
	out := strings.Join(shown, "; ")
	if len(gaps) > maxShown {
		out += fmt.Sprintf("; and %d more", len(gaps)-maxShown)
	}
	return out
}

// severity grades a percentage: 10% and above is material, otherwise advisory.
func severity(pct float64) string {
	if pct >= 10 {
		return "material"
	}
	return "advisory"
}

// evaluateInsights returns findings sorted by (priority, id).
func evaluateInsights(t accountTotals) []insightFinding {
	var out []insightFinding
	tokens := t.Input + t.Output

	if len(t.Gaps) > 0 {
		out = append(out, insightFinding{
			Priority: 0, ID: "AGG-COLLECTION-GAP", Category: "data_quality",
			Title:          "Metric collection had gaps or limitations",
			Observation:    fmt.Sprintf("%d collection note(s) were recorded for this account: %s", len(t.Gaps), joinGaps(t.Gaps)),
			Evidence:       fmt.Sprintf("gaps=%d grain=%s", len(t.Gaps), t.Grain),
			Impact:         "Rates, averages and PTU verdicts may be computed from incomplete evidence.",
			Recommendation: "Treat conclusions as provisional; resolve the listed gaps and re-run.",
			Verification:   "Re-run and confirm no collection notes remain.",
			Confidence:     "high", Applicability: "ACTIONABLE_FINDING",
		})
	}

	if t.Elapsed > 0 && t.Observed < t.Elapsed {
		out = append(out, insightFinding{
			Priority: 3, ID: "AGG-COVERAGE", Category: "data_quality",
			Title:          "Activity coverage of the metric window",
			Observation:    fmt.Sprintf("Traffic was observed in %d of %d time buckets in the window (%.1f%%).", t.Observed, t.Elapsed, float64(t.Observed)/float64(t.Elapsed)*100),
			Evidence:       fmt.Sprintf("observed=%d elapsed=%d grain=%s", t.Observed, t.Elapsed, t.Grain),
			Impact:         "Informational: low coverage usually means the account is idle most of the time; averages cover only the active span.",
			Recommendation: "No action required unless continuous traffic was expected.",
			Verification:   "Compare with the account's expected usage pattern.",
			Confidence:     "medium", Applicability: "EVALUATED_NO_ISSUE",
			Caveats: "Buckets with no traffic legitimately return no data points; this is not a data-quality defect by itself. Genuine collection failures are reported by AGG-COLLECTION-GAP.",
		})
	}

	if t.RequestsComplete && t.Requests > 0 {
		rate := t.RateLimited / t.Requests * 100
		if rate >= 5 {
			sev := severity(rate)
			out = append(out, insightFinding{
				Priority: 1, ID: "AGG-RATE-LIMIT", Category: "capacity",
				Title:          fmt.Sprintf("Rate limiting is %s", sev),
				Observation:    fmt.Sprintf("%.1f%% of requests were rate limited (HTTP 429).", rate),
				Evidence:       fmt.Sprintf("rate_limited=%.0f requests=%.0f", t.RateLimited, t.Requests),
				Impact:         "Callers are being throttled, which adds latency and retries.",
				Recommendation: "Review quota, deployment capacity and client retry/backoff before choosing a capacity model.",
				Verification:   "Compare the 429 share after the change.",
				Confidence:     "high", Applicability: "ACTIONABLE_FINDING",
				Caveats: "Rate-limit pressure alone does not imply PTU is the right answer.",
			})
		}
		// Failed already excludes 429s (see classifyStatus).
		nonThrottle := t.Failed
		if nonThrottle > 0 {
			frate := nonThrottle / t.Requests * 100
			sev := severity(frate)
			out = append(out, insightFinding{
				Priority: 1, ID: "AGG-FAILURE-RATE", Category: "reliability",
				Title:          fmt.Sprintf("Non-throttling failures are %s", sev),
				Observation:    fmt.Sprintf("%.1f%% of requests failed with a status other than 429.", frate),
				Evidence:       fmt.Sprintf("failed_non_429=%.0f requests=%.0f", nonThrottle, t.Requests),
				Impact:         "Failed requests waste capacity and hide application errors.",
				Recommendation: "Fix invalid requests and deployment references before adding retries.",
				Verification:   "Compare the Failed Requests (non-429) and Rate-Limited Requests columns on the AI Gov Usage sheet.",
				Confidence:     "high", Applicability: "ACTIONABLE_FINDING",
			})
		}
	}

	if t.Requests > 0 && tokens > 0 {
		out = append(out, insightFinding{
			Priority: 2, ID: "AGG-TOKENS-PER-REQUEST", Category: "efficiency",
			Title:          "Average token footprint per request",
			Observation:    fmt.Sprintf("%.0f tokens per request on average.", tokens/t.Requests),
			Evidence:       fmt.Sprintf("tokens=%.0f requests=%.0f", tokens, t.Requests),
			Impact:         "Informational.",
			Recommendation: "No action required.",
			Verification:   "Track the figure over time.",
			Confidence:     "medium", Applicability: "EVALUATED_NO_ISSUE",
			Caveats: "Pools every deployment and request type in the account (chat, embeddings, batch, provisioned); see per-deployment AI Gov Usage rows. A large footprint is an investigation trigger, not proof of waste.",
		})
	}

	if t.Requests > 0 && tokens == 0 {
		out = append(out, insightFinding{
			Priority: 1, ID: "AGG-NATIVE-USAGE-GAP", Category: "data_quality",
			Title:          "Requests observed without native token usage",
			Observation:    "Request counts exist but no input or output token metric returned data.",
			Evidence:       fmt.Sprintf("requests=%.0f tokens=0", t.Requests),
			Impact:         "Cost and PTU analysis cannot be evaluated.",
			Recommendation: "Confirm the account exposes token usage metrics for the deployed models.",
			Verification:   "Re-run after token metrics appear in Azure Monitor.",
			Confidence:     "high", Applicability: "NOT_EVALUATED_MISSING_EVIDENCE",
			Missing: "input and output token metrics",
		})
	}

	if tokens > 0 {
		out = append(out, insightFinding{
			Priority: 2, ID: "AGG-CACHED-TOKEN-GAP", Category: "pricing_readiness",
			Title:          "Cached-token counts are unavailable",
			Observation:    "Azure Monitor does not expose cached-token counts for Azure OpenAI Standard deployments; Anthropic deployments expose cache read/write metrics that this report does not use.",
			Evidence:       "no cached-token metric",
			Impact:         "Cached input is priced as normal input, so cost may be overstated for cache reads and understated for cache writes.",
			Recommendation: "Reconcile estimated cost against the invoice for models that bill cached input differently.",
			Verification:   "Compare against Azure Cost Management.",
			Confidence:     "high", Applicability: "NOT_EVALUATED_MISSING_EVIDENCE",
			Missing: "cached input tokens",
		})
	}

	if !t.RequestsComplete && t.Requests > 0 {
		out = append(out, insightFinding{
			Priority: 1, ID: "AGG-OUTCOME-GAP", Category: "data_quality",
			Title:          "Request outcome coverage is incomplete",
			Observation:    "The request-count metric is missing or failed for part of the window, so the throttle rate is withheld.",
			Evidence:       fmt.Sprintf("requests=%.0f outcome_complete=false", t.Requests),
			Impact:         "Capacity-pressure judgements are unavailable.",
			Recommendation: "Resolve the metric collection gap and re-run.",
			Verification:   "The throttle rate column should show a value.",
			Confidence:     "medium", Applicability: "NOT_EVALUATED_MISSING_EVIDENCE",
			Missing: "complete request-outcome data",
		})
	}

	if t.Requests > 0 && t.LatencyN == 0 {
		out = append(out, insightFinding{
			Priority: 1, ID: "DQ-LATENCY-GAP", Category: "data_quality",
			Title:          "No latency evidence",
			Observation:    "No latency metric returned data although requests were observed.",
			Evidence:       fmt.Sprintf("requests=%.0f latency_points=0", t.Requests),
			Impact:         "The PTU latency-sensitivity dimension cannot be scored.",
			Recommendation: "Confirm the account exposes a latency metric and re-run.",
			Verification:   "The PTU sheet latency verdict should no longer be insufficient.",
			Confidence:     "high", Applicability: "NOT_EVALUATED_MISSING_EVIDENCE",
			Missing: "request latency",
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority != out[j].Priority {
			return out[i].Priority < out[j].Priority
		}
		return out[i].ID < out[j].ID
	})
	return out
}
