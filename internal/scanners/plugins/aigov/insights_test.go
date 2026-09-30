// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import "testing"

func TestEvaluateInsights(t *testing.T) {
	got := evaluateInsights(accountTotals{
		Requests: 1000, RateLimited: 120, Failed: 150, Input: 1e6, Output: 1e5,
		RequestsComplete: true, Elapsed: 100, Observed: 50,
		Gaps: []string{"no data for 2026-01-01"},
	})
	ids := map[string]bool{}
	for _, f := range got {
		ids[f.ID] = true
	}
	for _, id := range []string{"AGG-COLLECTION-GAP", "AGG-COVERAGE", "AGG-RATE-LIMIT", "AGG-FAILURE-RATE", "AGG-TOKENS-PER-REQUEST", "AGG-CACHED-TOKEN-GAP", "DQ-LATENCY-GAP"} {
		if !ids[id] {
			t.Errorf("missing finding %s", id)
		}
	}
	if got[0].ID != "AGG-COLLECTION-GAP" {
		t.Errorf("expected collection gap first, got %s", got[0].ID)
	}
	if ids["AGG-NATIVE-USAGE-GAP"] || ids["AGG-OUTCOME-GAP"] {
		t.Errorf("unexpected gap findings: %v", ids)
	}
}

func TestEvaluateInsights_FailureRateExcludesRateLimited(t *testing.T) {
	find := func(t accountTotals) *insightFinding {
		for _, f := range evaluateInsights(t) {
			if f.ID == "AGG-FAILURE-RATE" {
				return &f
			}
		}
		return nil
	}
	// Failed never includes 429s, so 4 failures stay visible next to 25 throttled requests.
	f := find(accountTotals{Requests: 100, RateLimited: 25, Failed: 4, RequestsComplete: true})
	if f == nil {
		t.Fatal("expected AGG-FAILURE-RATE with 4 failed requests")
	}
	if want := "failed_non_429=4 requests=100"; f.Evidence != want {
		t.Errorf("evidence = %q, want %q", f.Evidence, want)
	}
	if find(accountTotals{Requests: 100, RateLimited: 25, RequestsComplete: true}) != nil {
		t.Error("unexpected AGG-FAILURE-RATE when only 429s occurred")
	}
}

func TestEvaluateInsights_NativeUsageGap(t *testing.T) {
	got := evaluateInsights(accountTotals{Requests: 10, RequestsComplete: true})
	found := false
	for _, f := range got {
		found = found || f.ID == "AGG-NATIVE-USAGE-GAP"
	}
	if !found {
		t.Error("expected AGG-NATIVE-USAGE-GAP")
	}
}
