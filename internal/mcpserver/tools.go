// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTools(s *server.MCPServer) {
	// Plugin Tools: Carbon Emissions
	carbonEmissionsTool := mcp.NewTool("scan-carbon-emissions",
		withBasicOptions(
			mcp.WithDescription(
				`Analyze carbon emissions by Azure resource type.

				This tool scans carbon emissions data across your Azure subscriptions and provides:
				- Period-over-period emission tracking by resource type
				- Latest month emissions
				- Previous month emissions
				- Month-over-month change ratio and absolute value
				- Emission units

				Excel/JSON file export is opt-in and off by default: pass xlsx/json=true to also write a file and get its resource URI; otherwise use the returned summary and the azqr-query-results tool.`),
		)...,
	)
	s.AddTool(carbonEmissionsTool, mcp.NewTypedToolHandler(scanPluginHandler("carbon-emissions")))

	// Plugin Tools: AI Governance Throttling
	aiGovTool := mcp.NewTool("scan-ai-gov",
		withBasicOptions(
			mcp.WithDescription(
				`Check AI/Cognitive Services accounts for 429 throttling errors.

				This tool monitors OpenAI and Azure Cognitive Services accounts for rate limiting issues:
				- Detects HTTP 429 (Too Many Requests) responses
				- Tracks throttling by account, deployment, and model
				- Monitors spillover configuration
				- Provides hourly throttling metrics
				- Shows SKU and kind information

				Useful for identifying capacity planning issues and optimizing API usage patterns.
				Excel/JSON file export is opt-in and off by default: pass xlsx/json=true to also write a file and get its resource URI; otherwise use the returned summary and the azqr-query-results tool.`),
		)...,
	)
	s.AddTool(aiGovTool, mcp.NewTypedToolHandler(scanPluginHandler("ai-gov")))

	// Plugin Tools: Zone Mapping
	zoneMappingTool := mcp.NewTool("scan-zone-mapping",
		withBasicOptions(
			mcp.WithDescription(
				`Retrieve logical-to-physical availability zone mappings for all Azure regions.

				This tool provides availability zone mapping information across subscriptions:
				- Maps logical zones to physical zones for each region
				- Helps understand cross-subscription zone alignment
				- Supports disaster recovery planning
				- Assists with multi-region architecture design

				Important: Physical zone mappings are subscription-specific and may differ between subscriptions.
				Excel/JSON file export is opt-in and off by default: pass xlsx/json=true to also write a file and get its resource URI; otherwise use the returned summary and the azqr-query-results tool.`),
		)...,
	)
	s.AddTool(zoneMappingTool, mcp.NewTypedToolHandler(scanPluginHandler("zone-mapping")))

	// Plugin Tools: SQL EOL
	sqlEOLTool := mcp.NewTool("scan-sql-eol",
		withBasicOptions(
			mcp.WithDescription(
				`Analyze SQL Server End-of-Life and Extended Security Update (ESU) status.

				This tool scans SQL Server instances across your Azure subscriptions and provides:
				- SQL Server version and edition details
				- End-of-life status and dates
				- ESU coverage period and costs per core
				- Estimated monthly, annual, and 3-year ESU costs
				- Patch operations cost comparison
				- SQL Managed Instance migration cost estimates and potential savings

				Useful for planning SQL Server modernization and understanding ESU financial impact.
				Excel/JSON file export is opt-in and off by default: pass xlsx/json=true to also write a file and get its resource URI; otherwise use the returned summary and the azqr-query-results tool.`),
		)...,
	)
	s.AddTool(sqlEOLTool, mcp.NewTypedToolHandler(scanPluginHandler("sql-eol")))

	// Plugin Tools: Service Health
	serviceHealthTool := mcp.NewTool("scan-service-health",
		withBasicOptions(
			mcp.WithDescription(
				`Analyze Azure service health availability over the last 90 days.

				This tool scans service health events across your Azure subscriptions and provides:
				- Percentage of time each resource type was unaffected by service issues
				- Service issue event counts per subscription, region, and resource type
				- Number of affected resources per bucket
				- Results sorted worst-first (lowest availability at the top)

				Covers resolved Service Issues events from Azure Resource Health.
				Excel/JSON file export is opt-in and off by default: pass xlsx/json=true to also write a file and get its resource URI; otherwise use the returned summary and the azqr-query-results tool.`),
		)...,
	)
	s.AddTool(serviceHealthTool, mcp.NewTypedToolHandler(scanPluginHandler("service-health")))

	// Plugin Tools: Region Selection
	regionSelectionTool := mcp.NewTool("scan-region-selection",
		withBasicOptions(
			mcp.WithDescription(
				`Analyze optimal Azure region selection based on service availability, network latency, and cost.

				This tool evaluates target regions for your workloads and provides:
				- Resource type availability comparison between source and target regions
				- SKU availability analysis (available, unavailable, restricted)
				- Network latency estimates between regions
				- Cost difference percentages
				- Recommendation scores with quality indicators
				- Detailed lists of missing resource types and unavailable SKUs

				Useful for multi-region planning, disaster recovery site selection, and workload migration.
				Excel/JSON file export is opt-in and off by default: pass xlsx/json=true to also write a file and get its resource URI; otherwise use the returned summary and the azqr-query-results tool.`),
			mcp.WithObject("options",
				mcp.Required(),
				mcp.Description("Plugin options. Use 'target-regions' (comma-separated string) to specify which Azure regions to analyze (e.g., {\"target-regions\": \"eastus,westeurope\"})."),
			),
		)...,
	)
	s.AddTool(regionSelectionTool, mcp.NewTypedToolHandler(scanPluginHandler("region-selection")))

	scan := mcp.NewTool("scan",
		withBasicOptions(
			mcp.WithDescription(
				`Run an Azure Quick Review (azqr) scan to analyze Azure resources and identify recommendations for improvement.

				WHAT GETS SCANNED (by default):
				- Azure resources and their configurations (Well-Architected Framework recommendations)
				- Microsoft Defender for Cloud security posture and recommendations
				- Azure Advisor recommendations (cost optimization, performance, reliability, operational excellence)
				- Cost analysis (last 3 months of costs by resource)
				- Resource inventory and metadata
				- Best practice violations and configuration issues

				SERVICE SCOPE:
				- If services parameter is NOT provided or is EMPTY [] -> Comprehensive scan of ALL supported Azure resource types
				- If services parameter contains specific service abbreviations -> Only those service types will be scanned

				Examples:
				- {} or {"services": []} -> Full comprehensive scan (all resources, costs, advisor, defender)
				- {"services": ["st"]} -> Scan only Storage Accounts (plus costs, advisor, defender)
				- {"services": ["aks", "vm", "sql"], "stages": ["-diagnostics"]} -> Scan AKS/VM/SQL without diagnostic settings analysis
				- {"stages": ["-defender", "-advisor"]} -> Scan all resources but skip Defender and Advisor

				Use get-supported-services tool to see all available service abbreviations.

				Excel/JSON file export is opt-in and off by default: pass xlsx/json=true to also write a file and get its resource URI; otherwise use the returned summary and the azqr-query-results tool.`),
			mcp.WithArray("stages",
				mcp.Items(map[string]any{"type": "string"}),
				mcp.Description("Optional array of scan stages to execute. Available stages: diagnostics, advisor, defender (enabled by default), cost, arc, policy, defender-recommendations (disabled by default). To disable a stage, prefix it with '-' (e.g., ['-diagnostics', '-defender']). Leave empty or omit to run default stages."),
			),
			mcp.WithArray("stageParams",
				mcp.Items(map[string]any{"type": "string"}),
				mcp.Description("Optional stage parameters in the form 'stage.key=value' (repeatable). Example: ['cost.previousMonth=true']"),
			),
			mcp.WithArray("services",
				mcp.Items(map[string]any{"type": "string"}),
				mcp.Description("Optional array of service type abbreviations to scan (e.g., ['aks', 'st', 'sql']). Leave empty or omit to scan all supported resource types. Use get-supported-services tool to see available abbreviations."),
			),
		)...,
	)

	// Tool: Scan
	s.AddTool(scan, mcp.NewTypedToolHandler(scanHandler))

	// Tool: Get recommendation catalog
	catalogTool := mcp.NewTool("get-recommendations-catalog",
		mcp.WithDescription("Get the complete catalog of AZQR recommendations"),
	)

	s.AddTool(catalogTool, handleCatalogTool)

	// Tool: Get supported services
	servicesTool := mcp.NewTool("get-supported-services",
		mcp.WithDescription("Get the list of Azure services supported by AZQR"),
	)

	s.AddTool(servicesTool, handleServiceTypeTool)

	// Tool: Query paginated scan results from a past session
	queryResultsTool := mcp.NewTool("query-results",
		mcp.WithDescription(
			`Query a bounded, paginated page of a past scan session's data directly from the session store.

			This is the only MCP-level way to retrieve scan data beyond the small summary envelope
			returned by the scan tools: full report files are no longer transferred over MCP, so use
			this tool to read resources, recommendation/graph results, or a session's persisted
			summary in small pages.

			Result types:
			- "resources" (default): one page of the session's scanned resources.
			- "recommendations": one page of the session's recommendation/graph results.
			- "summary": the session's persisted findings.Summary (not paginated).

			If sessionId is omitted, the most recently *completed* session is used (an in-progress or
			failed session is never silently chosen). Use the sessionId returned by a scan tool to
			query that specific scan's results later.`),
		mcp.WithString("sessionId",
			mcp.Description("Optional session ID to query. Defaults to the most recently completed session when omitted."),
		),
		mcp.WithString("resultType",
			mcp.Enum(queryResultTypeResources, queryResultTypeRecommendations, queryResultTypeSummary),
			mcp.Description("Which result category to return. Defaults to 'resources'."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Max items per page for paginated result types (default 100, max 1000). Ignored for 'summary'."),
		),
		mcp.WithNumber("offset",
			mcp.Description("Zero-based offset into the result set for paginated result types. Ignored for 'summary'."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	s.AddTool(queryResultsTool, mcp.NewTypedToolHandler(queryResultsHandler))
}

func withBasicOptions(opts ...mcp.ToolOption) []mcp.ToolOption {
	basicOpts := []mcp.ToolOption{
		mcp.WithArray("subscriptions",
			mcp.Items(map[string]any{"type": "string"}),
			mcp.Description("Optional array of subscription IDs to limit the scan to specific subscriptions. Leave empty or omit to scan all accessible subscriptions."),
		),
		mcp.WithArray("resourceGroups",
			mcp.Items(map[string]any{"type": "string"}),
			mcp.Description("Optional array of resource group names to limit the scan to specific resource groups. Leave empty or omit to scan all resource groups."),
		),
		mcp.WithBoolean("mask",
			mcp.DefaultBool(true),
			mcp.Description("Mask sensitive data in output (default: true)."),
		),
		mcp.WithBoolean("xlsx",
			mcp.DefaultBool(false),
			mcp.Description("Opt in to writing an Excel (.xlsx) report file for this scan. Off by default; when true, the file is written to <output>.xlsx and its file:// URI is included in the tool result."),
		),
		mcp.WithBoolean("json",
			mcp.DefaultBool(false),
			mcp.Description("Opt in to writing a JSON (.json) report file for this scan. Off by default; when true, the file is written to <output>.json and its file:// URI is included in the tool result."),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
	}
	return append(opts, basicOpts...)
}
