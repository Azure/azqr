// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/modelinfo"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/ptu"
)

// This file adapts the handful of TokenLens rules.py checks that apply to
// azqr's aggregate Azure Monitor evidence (no per-request trace data is
// collected, so trace-level rules such as prompt-prefix caching or retrieval
// detection are out of scope). Per the migration decision, this is a small
// set of plain functions reused by the Cost/PTU sheets rather than a new
// rule-engine abstraction.

// costRemediation returns a short remediation note for a cost row, or "" when
// no remediation is needed.
func costRemediation(priceResolved bool, source, mode, note string) string {
	if !priceResolved && modelinfo.IsProvisionedMode(mode) {
		return "Provisioned deployments are billed per reserved PTU hour, not per token; this report does not estimate the reserved-capacity cost."
	}
	if !priceResolved && modelinfo.IsBatchMode(mode) {
		return "Batch deployments are billed at batch rates that this report does not resolve; check the Azure pricing page for the model."
	}
	if !priceResolved && note != "" {
		return note
	}
	if !priceResolved {
		return "No verified pricing available for this model/region; confirm the exact SKU/meter before budgeting."
	}
	if source == "claude_ccu_equivalent" {
		return "Estimated from published Anthropic rates converted to Azure CCU billing; reconcile against your invoice."
	}
	return ""
}

// ptuRemediation returns a short remediation note for a PTU assessment row.
func ptuRemediation(a ptu.Assessment) string {
	switch a.Recommendation {
	case ptu.RecCapacityNeeded:
		return "Model is not in the verified PTU capacity table; request capacity data from Azure before purchasing PTUs."
	case ptu.RecInsufficient:
		return "Collect more activity (or a longer window) before sizing a PTU purchase."
	case ptu.RecPTURecommended, ptu.RecBorderline:
		if a.PAYGMonthlyUSD > 0 && a.PTUMonthlyUSD > 0 && a.PTUMonthlyUSD < a.PAYGMonthlyUSD {
			return "Estimated pay-as-you-go cost exceeds the estimated PTU cost at observed usage; evaluate a PTU purchase."
		}
		return ""
	default:
		return ""
	}
}
