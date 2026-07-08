// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package latency

import (
	"testing"

	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

// setTestLatencyMatrix replaces the package-level latency matrix and cluster averages
// for testing and returns a restore function to be deferred.
func setTestLatencyMatrix(m map[string]map[string]float64) func() {
	origMatrix := azureRegionLatency
	origClusterAvgs := clusterPairAverages
	azureRegionLatency = m
	clusterPairAverages = computeClusterAverages(m)
	return func() {
		azureRegionLatency = origMatrix
		clusterPairAverages = origClusterAvgs
	}
}

func TestEnrichWithLatencyData_KnownPair(t *testing.T) {
	defer setTestLatencyMatrix(map[string]map[string]float64{
		"eastus": {"westeurope": 85.0},
	})()

	results := []types.RegionComparison{
		{SourceRegion: "eastus", TargetRegion: "westeurope"},
	}
	EnrichWithLatencyData(results)
	if results[0].AvgLatencyMs != 85.0 {
		t.Errorf("expected 85.0 ms, got %.1f", results[0].AvgLatencyMs)
	}
}

func TestEnrichWithLatencyData_SymmetricFallback(t *testing.T) {
	// Only reverse direction is present in the matrix
	defer setTestLatencyMatrix(map[string]map[string]float64{
		"westeurope": {"eastus": 85.0},
	})()

	results := []types.RegionComparison{
		{SourceRegion: "eastus", TargetRegion: "westeurope"},
	}
	EnrichWithLatencyData(results)
	if results[0].AvgLatencyMs != 85.0 {
		t.Errorf("expected symmetric fallback 85.0 ms, got %.1f", results[0].AvgLatencyMs)
	}
}

func TestEnrichWithLatencyData_SameRegion(t *testing.T) {
	defer setTestLatencyMatrix(map[string]map[string]float64{})()

	results := []types.RegionComparison{
		{SourceRegion: "eastus", TargetRegion: "eastus"},
	}
	EnrichWithLatencyData(results)
	if results[0].AvgLatencyMs != 0 {
		t.Errorf("expected 0 ms for same region, got %.1f", results[0].AvgLatencyMs)
	}
}

func TestEnrichWithLatencyData_UnknownPair(t *testing.T) {
	defer setTestLatencyMatrix(map[string]map[string]float64{})() // empty — no cluster averages either

	results := []types.RegionComparison{
		{SourceRegion: "eastus", TargetRegion: "brazilsouth"},
	}
	EnrichWithLatencyData(results)
	if results[0].AvgLatencyMs != 0 {
		t.Errorf("expected 0 ms (N/A) for unknown pair, got %.1f", results[0].AvgLatencyMs)
	}
	if results[0].LatencyEstimated {
		t.Error("expected LatencyEstimated=false for truly unknown pair")
	}
}

func TestEnrichWithLatencyData_MultipleResults(t *testing.T) {
	defer setTestLatencyMatrix(map[string]map[string]float64{
		"eastus": {
			"westeurope":    85.0,
			"swedencentral": 110.0,
		},
	})()

	results := []types.RegionComparison{
		{SourceRegion: "eastus", TargetRegion: "westeurope"},
		{SourceRegion: "eastus", TargetRegion: "swedencentral"},
		{SourceRegion: "eastus", TargetRegion: "unknown"},
	}
	EnrichWithLatencyData(results)

	if results[0].AvgLatencyMs != 85.0 {
		t.Errorf("row 0: expected 85.0, got %.1f", results[0].AvgLatencyMs)
	}
	if results[1].AvgLatencyMs != 110.0 {
		t.Errorf("row 1: expected 110.0, got %.1f", results[1].AvgLatencyMs)
	}
	if results[2].AvgLatencyMs != 0 {
		t.Errorf("row 2: expected 0 (unknown), got %.1f", results[2].AvgLatencyMs)
	}
}

func TestEnrichWithLatencyData_RegionNamesNormalized(t *testing.T) {
	// Matrix uses lowercase no-space names; input has mixed case / spaces
	defer setTestLatencyMatrix(map[string]map[string]float64{
		"eastus": {"westeurope": 85.0},
	})()

	results := []types.RegionComparison{
		{SourceRegion: "East US", TargetRegion: "West Europe"},
	}
	EnrichWithLatencyData(results)
	if results[0].AvgLatencyMs != 85.0 {
		t.Errorf("expected normalization to yield 85.0 ms, got %.1f", results[0].AvgLatencyMs)
	}
}

func TestGetRegionLatency(t *testing.T) {
	defer setTestLatencyMatrix(map[string]map[string]float64{
		"eastus":     {"westeurope": 85.0},
		"westeurope": {"swedencentral": 30.0},
	})()

	// Cluster averages derived from the test matrix:
	//   americas:europe = 85  (eastus→westeurope)
	//   europe:europe   = 30  (westeurope→swedencentral)

	tests := []struct {
		name    string
		source  string
		target  string
		wantMs  float64
		wantEst bool
	}{
		{name: "same region is zero", source: "eastus", target: "eastus", wantMs: 0, wantEst: false},
		{name: "direct lookup", source: "eastus", target: "westeurope", wantMs: 85.0, wantEst: false},
		{name: "symmetric reverse lookup", source: "westeurope", target: "eastus", wantMs: 85.0, wantEst: false},
		{name: "cross-cluster estimate", source: "westus", target: "francesouth", wantMs: 85.0, wantEst: true},
		{name: "intra-cluster estimate", source: "northeurope", target: "spaincentral", wantMs: 30.0, wantEst: true},
		{name: "unknown pair no cluster data", source: "eastus", target: "brazilsouth", wantMs: 0, wantEst: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMs, gotEst := getRegionLatency(tt.source, tt.target)
			if gotMs != tt.wantMs {
				t.Errorf("getRegionLatency(%q, %q) ms = %.1f, want %.1f", tt.source, tt.target, gotMs, tt.wantMs)
			}
			if gotEst != tt.wantEst {
				t.Errorf("getRegionLatency(%q, %q) estimated = %v, want %v", tt.source, tt.target, gotEst, tt.wantEst)
			}
		})
	}
}

func TestEnrichWithLatencyData_ClusterEstimate(t *testing.T) {
	// Only one measured Americas→Europe pair; asking for a different pair in the same clusters.
	defer setTestLatencyMatrix(map[string]map[string]float64{
		"eastus": {"westeurope": 100.0},
	})()
	// Cluster averages: americas:europe = 100

	results := []types.RegionComparison{
		{SourceRegion: "westus", TargetRegion: "francesouth"}, // americas→europe, not directly measured
	}
	EnrichWithLatencyData(results)

	if results[0].AvgLatencyMs != 100.0 {
		t.Errorf("expected cluster estimate 100.0 ms, got %.1f", results[0].AvgLatencyMs)
	}
	if !results[0].LatencyEstimated {
		t.Error("expected latencyEstimated=true for cluster-based estimate")
	}
}

func TestNormalizeRegionName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "East US", want: "eastus"},
		{in: "eastus", want: "eastus"},
		{in: "West Europe", want: "westeurope"},
		{in: "AUSTRALIA CENTRAL", want: "australiacentral"},
		{in: "", want: ""},
		{in: "  East  US  ", want: "eastus"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := types.NormalizeRegionName(tt.in); got != tt.want {
				t.Errorf("normalizeRegionName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
