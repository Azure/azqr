// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import "testing"

func TestAssessPTU_UnknownModel_RequiresCapacityData(t *testing.T) {
	assessment := assessPTU("some-unlisted-model", 1000, 1200, priceResolution{}, 20)
	if assessment.Eligibility != "capacity_data_required" {
		t.Errorf("expected capacity_data_required, got %q", assessment.Eligibility)
	}
}

func TestAssessPTU_InsufficientEvidence(t *testing.T) {
	assessment := assessPTU("gpt-4o", 1000, 1200, priceResolution{}, 5)
	if assessment.Eligibility != "insufficient_evidence" {
		t.Errorf("expected insufficient_evidence, got %q", assessment.Eligibility)
	}
}

func TestAssessPTU_EligibleWithPricing(t *testing.T) {
	price := priceResolution{InputPerMTokUSD: 2.5, OutputPerMTokUSD: 10.0, Source: "exact_retail"}
	assessment := assessPTU("gpt-4o", 5000, 8000, price, 50)
	if assessment.Eligibility != "eligible" {
		t.Errorf("expected eligible, got %q", assessment.Eligibility)
	}
	if assessment.RecommendedPTUs < 15 {
		t.Errorf("expected recommended PTUs to respect the global minimum of 15, got %d", assessment.RecommendedPTUs)
	}
	if assessment.PAYGMonthlyUSD <= 0 {
		t.Errorf("expected positive PAYG monthly estimate, got %v", assessment.PAYGMonthlyUSD)
	}
}

func TestAssessPTU_EligibleWithoutPricing(t *testing.T) {
	assessment := assessPTU("gpt-4o", 5000, 8000, priceResolution{}, 50)
	if assessment.Eligibility != "eligible" {
		t.Errorf("expected eligible, got %q", assessment.Eligibility)
	}
	if assessment.PAYGMonthlyUSD != 0 {
		t.Errorf("expected zero PAYG estimate when pricing is unavailable, got %v", assessment.PAYGMonthlyUSD)
	}
}

func TestTpmStats(t *testing.T) {
	samples := []hourlyTokenUsage{
		{InputTokens: 6000, OutputTokens: 6000}, // 200 tpm
		{InputTokens: 0, OutputTokens: 0},       // inactive bucket, excluded
		{InputTokens: 12000, OutputTokens: 0},   // 200 tpm
	}
	avg, p95, active := tpmStats(samples)
	if active != 2 {
		t.Fatalf("expected 2 active buckets, got %d", active)
	}
	if avg != 200 {
		t.Errorf("expected avg 200, got %v", avg)
	}
	if p95 < avg {
		t.Errorf("expected p95 >= avg, got p95=%v avg=%v", p95, avg)
	}
}

func TestCostRemediation(t *testing.T) {
	if got := costRemediation(false, ""); got == "" {
		t.Error("expected remediation text when price is unresolved")
	}
	if got := costRemediation(true, "claude_ccu_equivalent"); got == "" {
		t.Error("expected remediation text for claude_ccu_equivalent source")
	}
	if got := costRemediation(true, "exact_retail"); got != "" {
		t.Errorf("expected no remediation for exact_retail, got %q", got)
	}
}
