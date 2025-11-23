// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"testing"
)

// setTestLatencyMatrix replaces the package-level latency matrix for testing and returns
// a restore function to be deferred.
func setTestLatencyMatrix(m map[string]map[string]float64) func() {
	orig := azureRegionLatency
	azureRegionLatency = m
	return func() { azureRegionLatency = orig }
}

func newScanner() *RegionSelectorScanner {
	return &RegionSelectorScanner{}
}

func TestEnrichWithLatencyData_KnownPair(t *testing.T) {
	defer setTestLatencyMatrix(map[string]map[string]float64{
		"eastus": {"westeurope": 85.0},
	})()

	results := []regionComparison{
		{sourceRegion: "eastus", targetRegion: "westeurope"},
	}
	newScanner().enrichWithLatencyData(results)
	if results[0].avgLatencyMs != 85.0 {
		t.Errorf("expected 85.0 ms, got %.1f", results[0].avgLatencyMs)
	}
}

func TestEnrichWithLatencyData_SymmetricFallback(t *testing.T) {
	// Only reverse direction is present in the matrix
	defer setTestLatencyMatrix(map[string]map[string]float64{
		"westeurope": {"eastus": 85.0},
	})()

	results := []regionComparison{
		{sourceRegion: "eastus", targetRegion: "westeurope"},
	}
	newScanner().enrichWithLatencyData(results)
	if results[0].avgLatencyMs != 85.0 {
		t.Errorf("expected symmetric fallback 85.0 ms, got %.1f", results[0].avgLatencyMs)
	}
}

func TestEnrichWithLatencyData_SameRegion(t *testing.T) {
	defer setTestLatencyMatrix(map[string]map[string]float64{})()

	results := []regionComparison{
		{sourceRegion: "eastus", targetRegion: "eastus"},
	}
	newScanner().enrichWithLatencyData(results)
	if results[0].avgLatencyMs != 0 {
		t.Errorf("expected 0 ms for same region, got %.1f", results[0].avgLatencyMs)
	}
}

func TestEnrichWithLatencyData_UnknownPair(t *testing.T) {
	defer setTestLatencyMatrix(map[string]map[string]float64{})()

	results := []regionComparison{
		{sourceRegion: "eastus", targetRegion: "brazilsouth"},
	}
	newScanner().enrichWithLatencyData(results)
	if results[0].avgLatencyMs != 0 {
		t.Errorf("expected 0 ms (N/A) for unknown pair, got %.1f", results[0].avgLatencyMs)
	}
}

func TestEnrichWithLatencyData_MultipleResults(t *testing.T) {
	defer setTestLatencyMatrix(map[string]map[string]float64{
		"eastus": {
			"westeurope":    85.0,
			"swedencentral": 110.0,
		},
	})()

	results := []regionComparison{
		{sourceRegion: "eastus", targetRegion: "westeurope"},
		{sourceRegion: "eastus", targetRegion: "swedencentral"},
		{sourceRegion: "eastus", targetRegion: "unknown"},
	}
	newScanner().enrichWithLatencyData(results)

	if results[0].avgLatencyMs != 85.0 {
		t.Errorf("row 0: expected 85.0, got %.1f", results[0].avgLatencyMs)
	}
	if results[1].avgLatencyMs != 110.0 {
		t.Errorf("row 1: expected 110.0, got %.1f", results[1].avgLatencyMs)
	}
	if results[2].avgLatencyMs != 0 {
		t.Errorf("row 2: expected 0 (unknown), got %.1f", results[2].avgLatencyMs)
	}
}

func TestEnrichWithLatencyData_RegionNamesNormalized(t *testing.T) {
	// Matrix uses lowercase no-space names; input has mixed case / spaces
	defer setTestLatencyMatrix(map[string]map[string]float64{
		"eastus": {"westeurope": 85.0},
	})()

	results := []regionComparison{
		{sourceRegion: "East US", targetRegion: "West Europe"},
	}
	newScanner().enrichWithLatencyData(results)
	if results[0].avgLatencyMs != 85.0 {
		t.Errorf("expected normalization to yield 85.0 ms, got %.1f", results[0].avgLatencyMs)
	}
}
