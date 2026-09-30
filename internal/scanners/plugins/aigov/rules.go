// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

// This file adapts the handful of TokenLens rules.py checks that apply to
// azqr's aggregate Azure Monitor evidence (no per-request trace data is
// collected, so trace-level rules such as prompt-prefix caching or retrieval
// detection are out of scope). Per the migration decision, this is a small
// set of plain functions reused by the Cost/PTU sheets rather than a new
// rule-engine abstraction.

// costRemediation returns a short remediation note for a cost row, or "" when
// no remediation is needed.
func costRemediation(priceResolved bool, source string) string {
	if !priceResolved {
		return "No verified pricing available for this model/region; confirm the exact SKU/meter before budgeting."
	}
	if source == "claude_ccu_equivalent" {
		return "Estimated from published Anthropic rates converted to Azure CCU billing; reconcile against your invoice."
	}
	return ""
}

// ptuRemediation returns a short remediation note for a PTU assessment row.
func ptuRemediation(assessment ptuAssessment) string {
	switch assessment.Eligibility {
	case "capacity_data_required":
		return "Model is not in the verified PTU capacity table; request capacity data from Azure before purchasing PTUs."
	case "insufficient_evidence":
		return "Collect a longer observation window before sizing a PTU purchase."
	case "eligible":
		if assessment.PAYGMonthlyUSD > 0 && assessment.PTUMonthlyUSD > 0 && assessment.PTUMonthlyUSD < assessment.PAYGMonthlyUSD {
			return "Estimated pay-as-you-go cost exceeds the estimated PTU cost at observed usage; evaluate a PTU purchase."
		}
		return ""
	default:
		return ""
	}
}
