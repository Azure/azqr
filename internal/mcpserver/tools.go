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

				Results are saved to Excel/JSON files and returned with resource URIs for download.`),
		)...,
	)
	s.AddTool(carbonEmissionsTool, mcp.NewTypedToolHandler(scanPluginHandler("carbon-emissions")))

	// Plugin Tools: OpenAI Throttling
	openAIThrottlingTool := mcp.NewTool("scan-openai-throttling",
		withBasicOptions(
			mcp.WithDescription(
				`Check OpenAI/Cognitive Services accounts for 429 throttling errors.

				This tool monitors OpenAI and Azure Cognitive Services accounts for rate limiting issues:
				- Detects HTTP 429 (Too Many Requests) responses
				- Tracks throttling by account, deployment, and model
				- Monitors spillover configuration
				- Provides hourly throttling metrics
				- Shows SKU and kind information

				Useful for identifying capacity planning issues and optimizing API usage patterns.
				Results are saved to Excel/JSON files and returned with resource URIs for download.`),
		)...,
	)
	s.AddTool(openAIThrottlingTool, mcp.NewTypedToolHandler(scanPluginHandler("openai-throttling")))

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
				Results are saved to Excel/JSON files and returned with resource URIs for download.`),
		)...,
	)
	s.AddTool(zoneMappingTool, mcp.NewTypedToolHandler(scanPluginHandler("zone-mapping")))

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

				Use get-supported-services tool to see all available service abbreviations.`),
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
	}
	return append(opts, basicOpts...)
}
