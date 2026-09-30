// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import "github.com/Azure/azqr/internal/plugins"

// usageSheetMetadata describes the "AI Gov Usage" sheet: raw hourly token
// counts per deployment/model, the evidence base for the Cost and PTU sheets.
func usageSheetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "ai-gov-usage",
		Version:     "1.2.0",
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
			{Name: "Hour (UTC)"},
			{Name: "Input Tokens"},
			{Name: "Output Tokens"},
			{Name: "Request Count"},
			{Name: "Rate-Limited Requests"},
			{Name: "Failed Requests (non-429)"},
			{Name: "Avg Latency (ms, unweighted)"},
			{Name: "Active Buckets"},
			{Name: "Service Tiers"},
			{Name: "Context Bands"},
		},
	}
}

// costSheetMetadata describes the "AI Gov Cost" sheet: estimated USD cost per
// deployment/model using verified pricing only (Azure Retail Prices API or
// the embedded Claude CCU-equivalent snapshot); no rate is ever guessed.
func costSheetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "ai-gov-cost",
		Version:     "1.2.0",
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
			{Name: "Deployment Mode"},
			{Name: "Service Tier"},
			{Name: "Context Band"},
			{Name: "Input Tokens"},
			{Name: "Output Tokens"},
			{Name: "Pricing Source"},
			{Name: "Estimated Cost (USD)"},
			{Name: "Assumptions"},
			{Name: "Remediation"},
		},
	}
}

// ptuSheetMetadata describes the "AI Gov PTU" sheet: PTU suitability and
// evidence quality per deployment. Capacity table comes from Microsoft Learn
// (see ptu.CapacitySource); the formulas are adapted from
// msftse-org/ptu-advisor; the dashboard/chart
// rendering from that project's report layer is intentionally not ported.
func ptuSheetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "ai-gov-ptu",
		Version:     "1.2.0",
		Description: "PTU (Provisioned Throughput Unit) suitability and evidence quality per deployment",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription"},
			{Name: "Resource Group"},
			{Name: "Account Name"},
			{Name: "Kind"},
			{Name: "SKU"},
			{Name: "Deployment Name"},
			{Name: "Model Name"},
			{Name: "Model Version"},
			{Name: "Model Format"},
			{Name: "SKU Capacity"},
			{Name: "Version Upgrade Option"},
			{Name: "Spillover Enabled"},
			{Name: "Spillover Deployment"},
			{Name: "Deployment SKU"},
			{Name: "Deployment Mode"},
			{Name: "Provider Family"},
			{Name: "Recommendation"},
			{Name: "Workload Shape"},
			{Name: "Capacity Pressure"},
			{Name: "Latency Sensitivity"},
			{Name: "Load Predictability"},
			{Name: "Observed Avg TPM (PTU-weighted)"},
			{Name: "Observed P95 TPM (PTU-weighted)"},
			{Name: "Active Buckets"},
			{Name: "Elapsed Buckets"},
			{Name: "Grain"},
			{Name: "Window Throttle Rate (%)"},
			{Name: "Recommended PTUs"},
			{Name: "Estimated PAYG Monthly (USD)"},
			{Name: "Estimated PTU Monthly (USD)"},
			{Name: "Break-even TPM (PTU-weighted)"},
			{Name: "Remediation"},
			{Name: "Observed Span (days)"},
			{Name: "Notes"},
		},
	}
}

// workloadsSheetMetadata describes the "AI Gov Workloads" sheet: cost rolled
// up per workload. Absent business-provided workload metadata, every
// deployment defaults to being its own technical workload (see workloads.go).
func workloadsSheetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "ai-gov-workloads",
		Version:     "1.2.0",
		Description: "Estimated cost rolled up per technical workload (account/deployment)",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Workload"},
			{Name: "Estimated Cost (USD)"},
			{Name: "Active Hour Buckets"},
			{Name: "Priced Token Share (%)"},
			{Name: "Cost Status"},
		},
	}
}

// insightsSheetMetadata describes the "AI Gov Insights" sheet: evidence-grounded
// findings derived from aggregate Azure Monitor data.
func insightsSheetMetadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "ai-gov-insights",
		Version:     "1.2.0",
		Description: "Evidence-grounded findings from aggregate Azure OpenAI / Foundry metrics",
		Author:      "Azure Quick Review Team",
		License:     "MIT",
		Type:        plugins.PluginTypeInternal,
		ColumnMetadata: []plugins.ColumnMetadata{
			{Name: "Subscription"},
			{Name: "Resource Group"},
			{Name: "Account Name"},
			{Name: "Priority"},
			{Name: "Rule ID"},
			{Name: "Category"},
			{Name: "Title"},
			{Name: "Observation"},
			{Name: "Evidence"},
			{Name: "Impact"},
			{Name: "Recommendation"},
			{Name: "Verification"},
			{Name: "Confidence"},
			{Name: "Applicability"},
			{Name: "Missing Evidence"},
			{Name: "Caveats"},
		},
	}
}
