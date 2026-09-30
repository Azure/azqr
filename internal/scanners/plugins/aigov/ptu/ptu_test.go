// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ptu

import (
	"strings"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/scanners/plugins/aigov/metrics"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/modelinfo"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/pricing"
)

// steadyBuckets builds n consecutive buckets of constant volume.
func steadyBuckets(n, minutes int, input, output, latency, requests float64) map[time.Time]*metrics.Bucket {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := map[time.Time]*metrics.Bucket{}
	for i := 0; i < n; i++ {
		m[start.Add(time.Duration(i*minutes)*time.Minute)] = &metrics.Bucket{
			Input: input, Output: output, Requests: requests,
			LatencySum: latency, LatencyN: 1, Minutes: minutes,
		}
	}
	return m
}

func TestAssessPTU_NotApplicable(t *testing.T) {
	cases := map[string]Input{
		"claude":      {Model: "claude-sonnet-4-5", Mode: modelinfo.ModeGlobal},
		"provisioned": {Model: "gpt-4o", Mode: modelinfo.ModeGlobalProvisioned},
		"batch":       {Model: "gpt-4o", Mode: "batch", Buckets: steadyBuckets(800, 5, 100000, 10000, 800, 10), GrainMinutes: 5},
	}
	for name, in := range cases {
		got := Assess(in)
		if got.Recommendation != recNotApplicable {
			t.Errorf("%s: expected %q, got %q", name, recNotApplicable, got.Recommendation)
		}
		if got.RecommendedPTUs != 0 || got.PTUMonthlyUSD != 0 {
			t.Errorf("%s: expected no PTU figures, got %+v", name, got)
		}
	}
}

func TestAssessPTU_NoteStatesSizingBasis(t *testing.T) {
	got := Assess(Input{Model: "gpt-4o", Mode: modelinfo.ModeGlobal, Buckets: steadyBuckets(800, 5, 100000, 10000, 800, 10), GrainMinutes: 5})
	if !strings.Contains(got.Note, "5-minute averages") {
		t.Errorf("expected sizing basis note, got %q", got.Note)
	}
}

func TestAssessPTU_UnknownModel_RequiresCapacityData(t *testing.T) {
	got := Assess(Input{Model: "gpt-9-unlisted", Mode: modelinfo.ModeGlobal})
	if got.Recommendation != RecCapacityNeeded {
		t.Errorf("expected %q, got %q", RecCapacityNeeded, got.Recommendation)
	}
}

func TestAssessPTU_InsufficientActivity(t *testing.T) {
	got := Assess(Input{Model: "gpt-4o", Mode: modelinfo.ModeGlobal, Buckets: steadyBuckets(20, 5, 1000, 100, 800, 10), GrainMinutes: 5})
	if got.Recommendation != RecInsufficient || got.Shape.Verdict != verdictInsufficient {
		t.Errorf("expected insufficient, got %+v", got)
	}
}

func TestPTUMinimumBuckets_ScalesWithGrain(t *testing.T) {
	if minimumBuckets(5) != 100 || minimumBuckets(15) != 34 || minimumBuckets(60) != 9 {
		t.Errorf("unexpected thresholds: %d %d %d", minimumBuckets(5), minimumBuckets(15), minimumBuckets(60))
	}
}

func TestAssessPTU_StablePressuredLatentRecommendsPTU(t *testing.T) {
	in := Input{
		Model: "gpt-4o", Mode: modelinfo.ModeGlobal,
		Buckets:      steadyBuckets(400, 5, 500_000, 50_000, 9000, 100),
		GrainMinutes: 5, ThrottleRate: 8, ThrottleKnown: true,
		Price: pricing.Resolution{InputPerMTokUSD: 2.5, OutputPerMTokUSD: 10, Source: "exact_retail"},
	}
	got := Assess(in)
	if got.Recommendation != RecPTURecommended {
		t.Fatalf("expected %q, got %q (%+v)", RecPTURecommended, got.Recommendation, got)
	}
	if got.RecommendedPTUs < 15 || got.PAYGMonthlyUSD <= 0 || got.PTUMonthlyUSD <= 0 {
		t.Errorf("expected sizing and cost estimates, got %+v", got)
	}
}

func TestAssessPTU_WithheldThrottleIsInsufficientPressure(t *testing.T) {
	in := Input{Model: "gpt-4o", Mode: modelinfo.ModeGlobal, Buckets: steadyBuckets(400, 5, 500_000, 50_000, 800, 100), GrainMinutes: 5}
	got := Assess(in)
	if got.Pressure.Verdict != verdictInsufficient {
		t.Errorf("expected insufficient pressure, got %q", got.Pressure.Verdict)
	}
}

func TestAssessPTU_NoLatencyIsInsufficient(t *testing.T) {
	b := steadyBuckets(400, 5, 500_000, 50_000, 0, 100)
	for _, v := range b {
		v.LatencyN = 0
	}
	got := Assess(Input{Model: "gpt-4o", Mode: modelinfo.ModeGlobal, Buckets: b, GrainMinutes: 5, ThrottleKnown: true})
	if got.Latency.Verdict != verdictInsufficient {
		t.Errorf("expected insufficient latency, got %q", got.Latency.Verdict)
	}
}

func TestPTURoundUp(t *testing.T) {
	cases := []struct {
		need           float64
		min, inc, want int
	}{
		{3, 15, 5, 15}, {16, 15, 5, 20}, {20, 15, 5, 20}, {51, 50, 50, 100},
	}
	for _, c := range cases {
		if got := roundUp(c.need, c.min, c.inc); got != c.want {
			t.Errorf("roundUp(%v,%d,%d)=%d want %d", c.need, c.min, c.inc, got, c.want)
		}
	}
}

func TestPTUCapacityTable_RowsAreComplete(t *testing.T) {
	if len(capacityTable) == 0 {
		t.Fatal("PTU capacity table is empty")
	}
	for model, c := range capacityTable {
		if c.InputTPMPerPTU <= 0 || c.OutputRatio <= 0 || c.GlobalMinPTU <= 0 || c.GlobalIncrement <= 0 {
			t.Errorf("%s: incomplete capacity row %+v", model, c)
		}
		if (c.RegionalMinPTU == 0) != (c.RegionalIncrem == 0) {
			t.Errorf("%s: regional min and increment must both be set or both be 0: %+v", model, c)
		}
		if model != modelinfo.CanonicalModelName(model) {
			t.Errorf("%s: key is not a canonical model name", model)
		}
	}
}

func TestScorePressure_UnknownLimitIsNotZeroPercent(t *testing.T) {
	got := scorePressure(Input{ThrottleKnown: true, ThrottleRate: 0.2}, capacity{}, 1000)
	if !strings.Contains(got.Metric, "limit unknown") || strings.Contains(got.Metric, "0.0%") {
		t.Errorf("expected unknown limit text, got %q", got.Metric)
	}
	if got.Verdict != verdictNegative && got.Verdict != verdictNeutral {
		t.Errorf("unexpected verdict %q", got.Verdict)
	}
}

func TestAssessPTU_DeploymentLimitDrivesPressureAndNotes(t *testing.T) {
	in := Input{
		Model: "gpt-4o", Mode: modelinfo.ModeGlobal, DeploymentLimitTPM: 600_000,
		Buckets:      steadyBuckets(400, 5, 500_000, 50_000, 800, 100),
		GrainMinutes: 5, ThrottleKnown: true, PTUHourlyUSD: 1.5,
		Price: pricing.Resolution{InputPerMTokUSD: 2.5, OutputPerMTokUSD: 10, Source: "exact_retail"},
	}
	got := Assess(in)
	if !strings.Contains(got.Pressure.Metric, "deployment limit") {
		t.Errorf("expected deployment limit basis, got %q", got.Pressure.Metric)
	}
	if !strings.Contains(got.Note, "extrapolate") || strings.Contains(got.Note, "fallback") {
		t.Errorf("unexpected note %q", got.Note)
	}
}
