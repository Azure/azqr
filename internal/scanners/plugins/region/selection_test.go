// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"math"
	"testing"

	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

// approxEqual returns true if a and b are within tolerance of each other.
func approxEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

// TestCalculateScores_AllNeutral: all components at their neutral/best values → score == 100.
func TestCalculateScores_AllNeutral(t *testing.T) {
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    0, // all unknown → SKU score neutral (100)
			AvgCostDifference:   0, // neutral
			AvgLatencyMs:        0, // neutral (no data)
			SourceZoneCount:     0,
			TargetZoneCount:     0,
		},
	}
	NewScanner().calculateScores(results)
	if !approxEqual(results[0].Score, 100.0, 0.01) {
		t.Errorf("expected score 100.0, got %.4f", results[0].Score)
	}
}

// TestCalculateScores_WeightedFormula: each component at a known value, verify weighted sum.
func TestCalculateScores_WeightedFormula(t *testing.T) {
	// resourceAvail=80, skuAvail=60 (6 avail / 10 confirmed), costDiff=10% → costScore=80 (100-10×2), latency=125ms → latencyScore=50
	// expected = 80*0.35 + 60*0.30 + 80*0.15 + 50*0.20
	//          = 28 + 18 + 12 + 10 = 68.0
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 80.0,
			TotalSKUsChecked:    10,
			AvailableSKUs:       6,
			UnknownSKUs:         0,
			AvgCostDifference:   10.0,
			HasCostData:         true,
			AvgLatencyMs:        125.0,
		},
	}
	NewScanner().calculateScores(results)
	expected := 68.0
	if !approxEqual(results[0].Score, expected, 0.01) {
		t.Errorf("expected score %.2f, got %.4f", expected, results[0].Score)
	}
}

// TestCalculateScores_NoCostData: missing cost data is treated as neutral (cost component = 100).
func TestCalculateScores_NoCostData(t *testing.T) {
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    0,
			AvgCostDifference:   0,
			AvgLatencyMs:        25.0, // <50ms → latencyScore=100
		},
	}
	NewScanner().calculateScores(results)
	if !approxEqual(results[0].Score, 100.0, 0.01) {
		t.Errorf("expected 100 when cost data absent, got %.4f", results[0].Score)
	}
}

// TestCalculateScores_NoLatencyData: avgLatencyMs==0 treated as neutral (latency component = 100).
func TestCalculateScores_NoLatencyData(t *testing.T) {
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    0,
			AvgCostDifference:   0,
			AvgLatencyMs:        0, // neutral sentinel
		},
	}
	NewScanner().calculateScores(results)
	if !approxEqual(results[0].Score, 100.0, 0.01) {
		t.Errorf("expected 100 when latency data absent, got %.4f", results[0].Score)
	}
}

// TestCalculateScores_HighLatency: latency > 200ms → latencyScore = 0.
func TestCalculateScores_HighLatency(t *testing.T) {
	// resourceAvail=100, skuAvail=100(neutral), cost=100(neutral), latency=0
	// expected = 100*0.35 + 100*0.30 + 100*0.15 + 0*0.20 = 80.0
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    0,
			AvgCostDifference:   0,
			AvgLatencyMs:        250.0,
		},
	}
	NewScanner().calculateScores(results)
	expected := 80.0
	if !approxEqual(results[0].Score, expected, 0.01) {
		t.Errorf("expected %.2f for high latency, got %.4f", expected, results[0].Score)
	}
}

// TestCalculateScores_LowLatency: latency < 50ms → latencyScore = 100.
func TestCalculateScores_LowLatency(t *testing.T) {
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    0,
			AvgCostDifference:   0,
			AvgLatencyMs:        20.0,
		},
	}
	NewScanner().calculateScores(results)
	if !approxEqual(results[0].Score, 100.0, 0.01) {
		t.Errorf("expected 100.0 for low latency, got %.4f", results[0].Score)
	}
}

// TestCalculateScores_MidLatencyInterpolation: latency at midpoint (125ms) → latencyScore = 50.
func TestCalculateScores_MidLatencyInterpolation(t *testing.T) {
	// latencyScore = 100 - ((125 - 50) / 150 * 100) = 100 - 50 = 50
	// all others neutral: resourceAvail=100, sku=100(neutral), cost=100(neutral)
	// expected = 100*0.35 + 100*0.30 + 100*0.15 + 50*0.20 = 90.0
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    0,
			AvgCostDifference:   0,
			AvgLatencyMs:        125.0,
		},
	}
	NewScanner().calculateScores(results)
	expected := 90.0
	if !approxEqual(results[0].Score, expected, 0.01) {
		t.Errorf("expected %.2f for mid-range latency, got %.4f", expected, results[0].Score)
	}
}

// TestCalculateScores_ZoneLossPenalty: zone loss applies multiplicative reduction.
func TestCalculateScores_ZoneLossPenalty(t *testing.T) {
	// Perfect base score = 100, src=3, tgt=0 → full zone loss → factor = 0.90
	// expected = 100 * (1 - 1.0 * 0.10) = 90.0
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    0,
			AvgCostDifference:   0,
			AvgLatencyMs:        0,
			SourceZoneCount:     3,
			TargetZoneCount:     0,
		},
	}
	NewScanner().calculateScores(results)
	expected := 90.0
	if !approxEqual(results[0].Score, expected, 0.01) {
		t.Errorf("expected %.2f for full zone loss, got %.4f", expected, results[0].Score)
	}
}

// TestCalculateScores_ZoneGainNopenalty: zone gain (tgt > src) is not penalized.
func TestCalculateScores_ZoneGainNoPenalty(t *testing.T) {
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    0,
			AvgCostDifference:   0,
			AvgLatencyMs:        0,
			SourceZoneCount:     1,
			TargetZoneCount:     3,
		},
	}
	NewScanner().calculateScores(results)
	if !approxEqual(results[0].Score, 100.0, 0.01) {
		t.Errorf("expected 100.0 for zone gain (no penalty), got %.4f", results[0].Score)
	}
}

// TestCalculateScores_NoSourceZones: source has no zones → no penalty regardless of target.
func TestCalculateScores_NoSourceZones(t *testing.T) {
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    0,
			AvgCostDifference:   0,
			AvgLatencyMs:        0,
			SourceZoneCount:     0,
			TargetZoneCount:     0,
		},
	}
	NewScanner().calculateScores(results)
	if !approxEqual(results[0].Score, 100.0, 0.01) {
		t.Errorf("expected 100.0 when source has no zones, got %.4f", results[0].Score)
	}
}

// TestCalculateScores_RestrictedSKUsCreditedHalf: restricted SKUs count as 50% in SKU score.
func TestCalculateScores_RestrictedSKUsCreditedHalf(t *testing.T) {
	// 4 avail + 2 restricted (×0.5 = 1.0 credit) out of 6 confirmed = 5.0/6 ≈ 83.33% SKU score
	// resourceAvail=100, cost=100(neutral), latency=100(neutral)
	// expected = 100*0.35 + 83.33*0.30 + 100*0.15 + 100*0.20 = 35 + 25 + 15 + 20 = 95.0
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    6,
			AvailableSKUs:       4,
			RestrictedSKUs:      []string{"sku-a", "sku-b"},
			UnknownSKUs:         0,
			AvgCostDifference:   0,
			AvgLatencyMs:        0,
		},
	}
	NewScanner().calculateScores(results)
	expected := 100*0.35 + (5.0/6.0)*100*0.30 + 100*0.15 + 100*0.20
	if !approxEqual(results[0].Score, expected, 0.01) {
		t.Errorf("expected %.4f for restricted SKUs, got %.4f", expected, results[0].Score)
	}
}

// TestCalculateScores_UnknownSKUsExcludedFromDenominator: unknown SKUs excluded from SKU score denominator.
func TestCalculateScores_UnknownSKUsExcludedFromDenominator(t *testing.T) {
	// 5 total, 3 unknown → confirmedChecked = 2; 2 avail → SKU score = 100
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    5,
			AvailableSKUs:       2,
			UnknownSKUs:         3,
			AvgCostDifference:   0,
			AvgLatencyMs:        0,
		},
	}
	NewScanner().calculateScores(results)
	if !approxEqual(results[0].Score, 100.0, 0.01) {
		t.Errorf("expected 100.0 when all confirmed SKUs available, got %.4f", results[0].Score)
	}
}

// TestCalculateScores_AllUnknownSKUsNeutral: all SKU checks unknown → SKU score stays 100.
func TestCalculateScores_AllUnknownSKUsNeutral(t *testing.T) {
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    5,
			AvailableSKUs:       0,
			UnknownSKUs:         5,
			AvgCostDifference:   0,
			AvgLatencyMs:        0,
		},
	}
	NewScanner().calculateScores(results)
	if !approxEqual(results[0].Score, 100.0, 0.01) {
		t.Errorf("expected neutral 100.0 when all SKUs unknown, got %.4f", results[0].Score)
	}
}

// TestCalculateScores_CostScoreFloorAtZero: extreme cost difference never goes below 0.
func TestCalculateScores_CostScoreFloorAtZero(t *testing.T) {
	// costDiff = 100% → costScore = 100 - 500 = floored to 0
	// expected = 100*0.35 + 100*0.30 + 0*0.15 + 100*0.20 = 85.0
	results := []types.RegionComparison{
		{
			AvailabilityPercent: 100.0,
			TotalSKUsChecked:    0,
			AvgCostDifference:   100.0,
			HasCostData:         true,
			AvgLatencyMs:        0,
		},
	}
	NewScanner().calculateScores(results)
	expected := 85.0
	if !approxEqual(results[0].Score, expected, 0.01) {
		t.Errorf("expected %.2f with cost score floored at 0, got %.4f", expected, results[0].Score)
	}
}
