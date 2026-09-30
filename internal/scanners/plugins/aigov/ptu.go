// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import "fmt"

// ptuCapacitySource attributes the PTU capacity table below. The decision
// shape, thresholds, and capacity numbers are adapted from the MIT-licensed
// https://github.com/msftse-org/ptu-advisor project (as ported through
// TokenLens-for-Azure, https://github.com/zakarel/tokenlens-for-azure,
// revision eb0558cd4c6d3794be76d9caa2e87129d1f8221c). Capacity is matched on
// the exact canonical model key only; an unlisted model is reported as
// "capacity data required" rather than borrowing a related model's numbers.
const ptuCapacitySource = "msftse-org/ptu-advisor @ eb0558cd4c6d3794be76d9caa2e87129d1f8221c (via tokenlens-for-azure)"

// ptuCapacity describes the PTU purchasing shape for one canonical model.
type ptuCapacity struct {
	// InputTPMPerPTU is the sustained input tokens-per-minute one PTU buys.
	InputTPMPerPTU int
	// OutputRatio is how many input-token-equivalents one output token costs
	// against the same PTU capacity (output is more expensive to serve).
	OutputRatio     int
	GlobalMinPTU    int
	GlobalIncrement int
}

// ptuCapacityTable is a deliberately small, explicit subset of documented
// models. See ptuCapacitySource for provenance.
var ptuCapacityTable = map[string]ptuCapacity{
	"gpt-5.5":      {InputTPMPerPTU: 1200, OutputRatio: 6, GlobalMinPTU: 15, GlobalIncrement: 5},
	"gpt-5.4-mini": {InputTPMPerPTU: 7900, OutputRatio: 6, GlobalMinPTU: 15, GlobalIncrement: 5},
	"gpt-5.4":      {InputTPMPerPTU: 2400, OutputRatio: 6, GlobalMinPTU: 15, GlobalIncrement: 5},
	"gpt-5.1":      {InputTPMPerPTU: 4750, OutputRatio: 8, GlobalMinPTU: 15, GlobalIncrement: 5},
	"gpt-5-mini":   {InputTPMPerPTU: 23750, OutputRatio: 8, GlobalMinPTU: 15, GlobalIncrement: 5},
	"gpt-5":        {InputTPMPerPTU: 4750, OutputRatio: 8, GlobalMinPTU: 15, GlobalIncrement: 5},
	"gpt-4.1":      {InputTPMPerPTU: 3000, OutputRatio: 4, GlobalMinPTU: 15, GlobalIncrement: 5},
	"gpt-4.1-mini": {InputTPMPerPTU: 14900, OutputRatio: 4, GlobalMinPTU: 15, GlobalIncrement: 5},
	"gpt-4.1-nano": {InputTPMPerPTU: 59400, OutputRatio: 4, GlobalMinPTU: 15, GlobalIncrement: 5},
	"gpt-4o":       {InputTPMPerPTU: 2500, OutputRatio: 4, GlobalMinPTU: 15, GlobalIncrement: 5},
	"gpt-4o-mini":  {InputTPMPerPTU: 37000, OutputRatio: 4, GlobalMinPTU: 15, GlobalIncrement: 5},
	"o1":           {InputTPMPerPTU: 230, OutputRatio: 4, GlobalMinPTU: 15, GlobalIncrement: 5},
	"o3-mini":      {InputTPMPerPTU: 2500, OutputRatio: 4, GlobalMinPTU: 15, GlobalIncrement: 5},
	"o4-mini":      {InputTPMPerPTU: 5400, OutputRatio: 4, GlobalMinPTU: 15, GlobalIncrement: 5},
}

// ptuAssessment is one row of the "AI Gov PTU" sheet.
type ptuAssessment struct {
	Eligibility     string // "eligible", "capacity_data_required", "insufficient_evidence"
	RecommendedPTUs int
	PAYGMonthlyUSD  float64
	PTUMonthlyUSD   float64
	BreakEvenTPM    float64
	Note            string
}

// ptuHourlyPricePerUnit is a placeholder illustrative global PTU hourly rate
// used only to produce a directional monthly cost comparison. Real pricing
// should be resolved the same way as token pricing once Cognitive Services
// PTU meters are confirmed on the Retail Prices API; until then this constant
// is deliberately conservative and clearly labelled as an estimate.
const ptuHourlyPricePerUnitUSD = 1.0

// assessPTU evaluates whether a deployment is a PTU candidate from its
// observed average and p95 total tokens-per-minute (input+output) and its
// pay-as-you-go price. It mirrors TokenLens/ptu-advisor's shape (capacity
// lookup, weighted TPM, min/increment rounding, break-even) without the
// dashboard/chart rendering, which is intentionally excluded from this port.
func assessPTU(modelName string, avgTPM, p95TPM float64, price priceResolution, activeBuckets int) ptuAssessment {
	const minimumActiveBuckets = 12 // at least an hour of 5-minute buckets of real evidence

	capacity, ok := ptuCapacityTable[canonicalModelName(modelName)]
	if !ok {
		return ptuAssessment{
			Eligibility: "capacity_data_required",
			Note:        fmt.Sprintf("model %q is not in the verified PTU capacity table (%s)", modelName, ptuCapacitySource),
		}
	}

	if activeBuckets < minimumActiveBuckets {
		return ptuAssessment{
			Eligibility: "insufficient_evidence",
			Note:        "fewer than an hour of active usage buckets observed; PTU sizing needs more evidence",
		}
	}

	weightedTPM := p95TPM // p95 is used as the sizing reference, matching ptu-advisor's conservative default
	ptusNeeded := weightedTPM / float64(capacity.InputTPMPerPTU)
	recommended := int(ptusNeeded) + 1
	if recommended < capacity.GlobalMinPTU {
		recommended = capacity.GlobalMinPTU
	} else {
		// round up to the nearest purchasing increment above the minimum
		steps := (recommended - capacity.GlobalMinPTU + capacity.GlobalIncrement - 1) / capacity.GlobalIncrement
		recommended = capacity.GlobalMinPTU + steps*capacity.GlobalIncrement
	}

	ptuCapacityTPM := float64(recommended * capacity.InputTPMPerPTU)
	breakEvenTPM := ptuCapacityTPM // simplified: capacity itself is the break-even reference point

	note := "sufficient evidence for a directional PTU recommendation"
	eligibility := "eligible"
	var paygMonthly, ptuMonthly float64
	if price.Source != "" {
		// Approximate monthly PAYG cost from the observed average TPM sustained for a month.
		avgTokensPerMonth := avgTPM * 60 * 24 * 30
		paygMonthly = estimatedCostUSD(price, avgTokensPerMonth*0.6, avgTokensPerMonth*0.4)
		ptuMonthly = float64(recommended) * ptuHourlyPricePerUnitUSD * 24 * 30
	} else {
		eligibility = "eligible"
		note = "eligible, but pay-as-you-go pricing is unavailable so no cost comparison could be computed"
	}

	return ptuAssessment{
		Eligibility:     eligibility,
		RecommendedPTUs: recommended,
		PAYGMonthlyUSD:  paygMonthly,
		PTUMonthlyUSD:   ptuMonthly,
		BreakEvenTPM:    breakEvenTPM,
		Note:            note,
	}
}
