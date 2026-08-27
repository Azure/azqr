// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package skus

import "testing"

func TestFindAlternatives_ReturnsCapped(t *testing.T) {
	target, ok := Lookup("Standard_D4s_v5")
	if !ok {
		t.Fatal("Standard_D4s_v5 not found in known_skus.yaml")
	}

	results := FindAlternatives(target, 5)
	if len(results) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(results))
	}
	if len(results) == 0 {
		t.Error("expected at least one alternative")
	}
}

func TestFindAlternatives_SortedDescending(t *testing.T) {
	target, ok := Lookup("Standard_D4s_v5")
	if !ok {
		t.Fatal("Standard_D4s_v5 not found in known_skus.yaml")
	}

	results := FindAlternatives(target, 10)
	for i := 1; i < len(results); i++ {
		if results[i].CompatibilityScore > results[i-1].CompatibilityScore {
			t.Errorf("results not sorted: index %d (%.3f) > index %d (%.3f)",
				i, results[i].CompatibilityScore, i-1, results[i-1].CompatibilityScore)
		}
	}
}

func TestFindAlternatives_ExcludesTarget(t *testing.T) {
	target, ok := Lookup("Standard_D4s_v5")
	if !ok {
		t.Fatal("Standard_D4s_v5 not found in known_skus.yaml")
	}

	results := FindAlternatives(target, 0)
	for _, r := range results {
		if r.SKU.Name == target.Name {
			t.Errorf("target SKU %q must not appear in alternatives", target.Name)
		}
	}
}

func TestFindAlternatives_UnknownSKU(t *testing.T) {
	target := SKU{Name: "does-not-exist", VCPUs: 4, MemoryGB: 16}
	results := FindAlternatives(target, 5)
	if len(results) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(results))
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"Standard_D4s_v5", 5},
		{"Standard_D4_v2", 2},
		{"Standard_D11_v2_Promo", 2},
		{"Standard_B4ms", 0},
		{"Standard_D1", 0},
		{"Standard_D128ds_v6", 6},
	}
	for _, tt := range tests {
		got := extractVersion(tt.name)
		if got != tt.want {
			t.Errorf("extractVersion(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestScoreVersion(t *testing.T) {
	tests := []struct {
		targetVer    int
		candidateVer int
		wantScore    float64
	}{
		{5, 6, 0.05},
		{5, 5, 0.05},
		{5, 4, 0.03},
		{5, 3, 0.01},
		{5, 2, 0.00},
		{0, 5, 0.025},
		{5, 0, 0.025},
	}
	for _, tt := range tests {
		score := scoreVersion(tt.targetVer, tt.candidateVer)
		if score != tt.wantScore {
			t.Errorf("scoreVersion(%d, %d) = %.3f, want %.3f",
				tt.targetVer, tt.candidateVer, score, tt.wantScore)
		}
	}
}

func TestExtractBaseName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Standard_D16s_v5", "Standard_D16s"},
		{"Standard_D4s_v3", "Standard_D4s"},
		{"Standard_D11_v2_Promo", "Standard_D11"},
		{"Standard_B4ms", "Standard_B4ms"}, // no version suffix → unchanged
		{"Standard_D1", "Standard_D1"},
	}
	for _, tt := range tests {
		got := extractBaseName(tt.name)
		if got != tt.want {
			t.Errorf("extractBaseName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestExtractGPUCategory verifies workload class extraction from N-series SKU names.
func TestExtractGPUCategory(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Standard_NV36ads_A10_v5", "V"},
		{"Standard_NV4ads_V710_v5", "V"},
		{"Standard_NC64as_T4_v3", "C"},
		{"Standard_NC48ads_A100_v4", "C"},
		{"Standard_ND96asr_v4", "D"},
		{"Standard_D4s_v5", ""},  // not N-series
		{"Standard_B4ms", ""},
	}
	for _, tt := range tests {
		got := extractGPUCategory(tt.name)
		if got != tt.want {
			t.Errorf("extractGPUCategory(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestFindAlternatives_V710AppearsForA10 verifies that NVads V710 v5 SKUs surface
// when searching for alternatives to an NVads A10 v5 SKU of comparable size.
func TestFindAlternatives_V710AppearsForA10(t *testing.T) {
	// Use NV12ads_A10_v5: the V710 equivalent (NV12ads_V710_v5) has a comparable
	// memory footprint (64 GB vs 110 GB), so it should rank in the top alternatives.
	target, ok := Lookup("Standard_NV12ads_A10_v5")
	if !ok {
		t.Skip("Standard_NV12ads_A10_v5 not found in known_skus.yaml")
	}

	results := FindAlternatives(target, 10)

	found := false
	for _, r := range results {
		if r.SKU.Family == "StandardNVadsV710v5Family" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected at least one NVads V710 v5 SKU in top-10 alternatives for Standard_NV12ads_A10_v5")
	}
}

// TestFindAlternatives_NVAlternativesRankAboveNC verifies that NV (visualization)
// alternatives rank above NC (compute) alternatives for an NV source SKU.
func TestFindAlternatives_NVAlternativesRankAboveNC(t *testing.T) {
	target, ok := Lookup("Standard_NV36ads_A10_v5")
	if !ok {
		t.Skip("Standard_NV36ads_A10_v5 not found in known_skus.yaml")
	}

	results := FindAlternatives(target, 0) // all results

	nvPos, ncPos := -1, -1
	for i, r := range results {
		cat := extractGPUCategory(r.SKU.Name)
		if cat == "V" && nvPos == -1 {
			nvPos = i
		}
		if cat == "C" && ncPos == -1 {
			ncPos = i
		}
		if nvPos != -1 && ncPos != -1 {
			break
		}
	}

	if nvPos == -1 || ncPos == -1 {
		t.Skip("need both NV and NC alternatives to compare ranking")
	}
	if nvPos > ncPos {
		t.Errorf("first NV alternative (pos %d) should rank above first NC alternative (pos %d)", nvPos, ncPos)
	}
}

func TestFindAlternatives_SameModelVersionsRankFirst(t *testing.T) {
	target, ok := Lookup("Standard_D16s_v5")
	if !ok {
		t.Fatal("Standard_D16s_v5 not found in known_skus.yaml")
	}

	results := FindAlternatives(target, 10)

	// Find position of Standard_D16s_v4 and a known non-same-model alternative
	v4Pos := -1
	otherPos := -1
	for i, r := range results {
		if r.SKU.Name == "Standard_D16s_v4" {
			v4Pos = i
		}
		if r.SKU.Name == "Standard_D16ads_v5" {
			otherPos = i
		}
	}

	if v4Pos == -1 {
		t.Error("Standard_D16s_v4 should appear in the top 10 alternatives")
	}
	if otherPos != -1 && v4Pos > otherPos {
		t.Errorf("Standard_D16s_v4 (pos %d) should rank above Standard_D16ads_v5 (pos %d)", v4Pos, otherPos)
	}
}
