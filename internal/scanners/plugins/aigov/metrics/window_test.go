// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package metrics

import (
	"context"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/internal/testutil"
	"testing"
	"time"
)

func testPlan() Plan {
	return Plan{Input: "InputTokens", Output: "OutputTokens", Requests: "AzureOpenAIRequests", StatusSplit: true}
}

func TestCollectBatch_ChunksDedupeAndCoverage(t *testing.T) {
	q := &testutil.FakeQuerier{}
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	got := CollectBatch(context.Background(), q, "sub", []string{"r1"}, testPlan(), end, 60)["r1"]

	// 60 days = 2 chunks; each chunk contributes only its start point (the end
	// point is ignored), so there is no double counting at the boundary.
	if len(got.Buckets[DepModel{"dep", "gpt-4o"}]) != 2 {
		t.Fatalf("buckets = %d, want 2", len(got.Buckets[DepModel{"dep", "gpt-4o"}]))
	}
	if got.Grain != "PT5M" {
		t.Errorf("grain = %s, want PT5M", got.Grain)
	}
	if got.Elapsed != 60*24*12 {
		t.Errorf("elapsed = %d, want %d", got.Elapsed, 60*24*12)
	}
	rate, ok := got.WindowThrottleRate(DepModel{"dep", "gpt-4o"})
	if !ok || rate != 100 {
		t.Errorf("throttle rate = %v ok=%v, want 100 true", rate, ok)
	}
}

func TestCollectBatch_GrainFallback(t *testing.T) {
	q := &testutil.FakeQuerier{FailGrains: map[string]bool{"PT5M": true}}
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	got := CollectBatch(context.Background(), q, "sub", []string{"r1"}, testPlan(), end, 30)["r1"]
	if got.Grain != "PT15M" {
		t.Errorf("grain = %s, want PT15M", got.Grain)
	}
	if len(got.Gaps) != 0 {
		t.Errorf("unexpected gaps: %v", got.Gaps)
	}
}

func TestCollectBatch_AllGrainsFailReportsGap(t *testing.T) {
	q := &testutil.FakeQuerier{FailGrains: map[string]bool{"PT5M": true, "PT15M": true, "PT1H": true}}
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	got := CollectBatch(context.Background(), q, "sub", []string{"r1"}, testPlan(), end, 30)["r1"]
	if len(got.Gaps) == 0 || got.RequestsComplete {
		t.Fatalf("expected a gap and incomplete requests, got gaps=%v complete=%v", got.Gaps, got.RequestsComplete)
	}
	if _, ok := got.WindowThrottleRate(DepModel{"dep", "gpt-4o"}); ok {
		t.Error("throttle rate must be withheld when the denominator is incomplete")
	}
}

func TestWindowChunks(t *testing.T) {
	end := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	chunks := windowChunks(end, 90)
	if len(chunks) != 3 || !chunks[2].End.Equal(end) || !chunks[0].End.Equal(chunks[1].Start) {
		t.Errorf("unexpected chunks: %+v", chunks)
	}
}
