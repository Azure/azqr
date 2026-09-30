// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import "github.com/Azure/azqr/internal/plugins"

// usageSheetMetadata describes the "AI Gov Usage" sheet: raw hourly token
// counts per deployment/model, the evidence base for the Cost and PTU sheets.
func usageSheetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "ai-gov-usage",
		Version:     "1.0.0",
		Description: "Hourly Azure OpenAI / Foundry token usage by deployment and model",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription"},
			{Name: "Resource Group"},
			{Name: "Account Name"},
			{Name: "Deployment Name"},
			{Name: "Model Name"},
			{Name: "Hour"},
			{Name: "Input Tokens"},
			{Name: "Output Tokens"},
		},
	}
}

// costSheetMetadata describes the "AI Gov Cost" sheet: estimated USD cost per
// deployment/model using verified pricing only (Azure Retail Prices API or
// the embedded Claude CCU-equivalent snapshot); no rate is ever guessed.
func costSheetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "ai-gov-cost",
		Version:     "1.0.0",
		Description: "Estimated cost per deployment/model based on verified retail pricing",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription"},
			{Name: "Resource Group"},
			{Name: "Account Name"},
			{Name: "Deployment Name"},
			{Name: "Model Name"},
			{Name: "Input Tokens"},
			{Name: "Output Tokens"},
			{Name: "Pricing Source"},
			{Name: "Estimated Cost (USD)"},
			{Name: "Remediation"},
		},
	}
}

// ptuSheetMetadata describes the "AI Gov PTU" sheet: PTU suitability and
// evidence quality per deployment. Capacity table and formulas are adapted
// from msftse-org/ptu-advisor (see ptuCapacitySource); the dashboard/chart
// rendering from that project's report layer is intentionally not ported.
func ptuSheetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "ai-gov-ptu",
		Version:     "1.0.0",
		Description: "PTU (Provisioned Throughput Unit) suitability and evidence quality per deployment",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription"},
			{Name: "Resource Group"},
			{Name: "Account Name"},
			{Name: "Deployment Name"},
			{Name: "Model Name"},
			{Name: "Observed Avg TPM"},
			{Name: "Observed P95 TPM"},
			{Name: "Eligibility"},
			{Name: "Recommended PTUs"},
			{Name: "Estimated PAYG Monthly (USD)"},
			{Name: "Estimated PTU Monthly (USD)"},
			{Name: "Break-even TPM"},
			{Name: "Remediation"},
		},
	}
}

// workloadsSheetMetadata describes the "AI Gov Workloads" sheet: cost rolled
// up per workload. Absent business-provided workload metadata, every
// deployment defaults to being its own technical workload (see workloads.go).
func workloadsSheetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "ai-gov-workloads",
		Version:     "1.0.0",
		Description: "Estimated cost rolled up per technical workload (account/deployment)",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Workload"},
			{Name: "Estimated Cost (USD)"},
			{Name: "Active Hour Buckets"},
		},
	}
}
