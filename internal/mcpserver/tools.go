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
		mcp.WithDescription(
			`Analyze carbon emissions by Azure resource type.

			This tool scans carbon emissions data across your Azure subscriptions and provides:
			- Period-over-period emission tracking by resource type
			- Latest month emissions
			- Previous month emissions
			- Month-over-month change ratio and absolute value
			- Emission units

			Results are saved to Excel/JSON files and returned with resource URIs for download.`),
		mcp.WithArray("services",
			mcp.Items(map[string]any{"type": "string"}),
			mcp.Description("Optional array of service type abbreviations to include in the scan (e.g., ['st', 'vm', 'sql']). Leave empty to scan all resource types."),
		),
		mcp.WithBoolean("mask",
			mcp.DefaultBool(true),
			mcp.Description("Mask sensitive data in output (default: true)."),
		),
	)
	s.AddTool(carbonEmissionsTool, mcp.NewTypedToolHandler(scanPluginHandler("carbon-emissions")))

	// Plugin Tools: OpenAI Throttling
	openAIThrottlingTool := mcp.NewTool("scan-openai-throttling",
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
		mcp.WithArray("services",
			mcp.Items(map[string]any{"type": "string"}),
			mcp.Description("Optional array of service type abbreviations to include in the scan (e.g., ['st', 'vm', 'sql']). Leave empty to scan all resource types."),
		),
		mcp.WithBoolean("mask",
			mcp.DefaultBool(true),
			mcp.Description("Mask sensitive data in output (default: true)."),
		),
	)
	s.AddTool(openAIThrottlingTool, mcp.NewTypedToolHandler(scanPluginHandler("openai-throttling")))

	// Plugin Tools: Zone Mapping
	zoneMappingTool := mcp.NewTool("scan-zone-mapping",
		mcp.WithDescription(
			`Retrieve logical-to-physical availability zone mappings for all Azure regions.

			This tool provides availability zone mapping information across subscriptions:
			- Maps logical zones to physical zones for each region
			- Helps understand cross-subscription zone alignment
			- Supports disaster recovery planning
			- Assists with multi-region architecture design

			Important: Physical zone mappings are subscription-specific and may differ between subscriptions.
			Results are saved to Excel/JSON files and returned with resource URIs for download.`),
		mcp.WithArray("services",
			mcp.Items(map[string]any{"type": "string"}),
			mcp.Description("Optional array of service type abbreviations to include in the scan (e.g., ['st', 'vm', 'sql']). Leave empty to scan all resource types."),
		),
		mcp.WithBoolean("mask",
			mcp.DefaultBool(true),
			mcp.Description("Mask sensitive data in output (default: true)."),
		),
	)
	s.AddTool(zoneMappingTool, mcp.NewTypedToolHandler(scanPluginHandler("zone-mapping")))

	scan := mcp.NewTool("scan",
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
			- {"services": ["aks", "vm", "sql"], "cost": false} -> Scan AKS/VM/SQL without cost analysis
			- {"defender": false, "advisor": false} -> Scan all resources but skip Defender and Advisor

			Use get-supported-services tool to see all available service abbreviations.`),
		mcp.WithArray("services",
			mcp.Items(map[string]any{"type": "string"}),
			mcp.Description("Optional array of service type abbreviations to scan (e.g., ['aks', 'st', 'sql']). Leave empty or omit to scan all supported resource types. Use get-supported-services tool to see available abbreviations."),
		),
		mcp.WithBoolean("defender",
			mcp.Description("Include Microsoft Defender for Cloud scanning (default: true). Set to false to skip Defender scanning."),
		),
		mcp.WithBoolean("advisor",
			mcp.Description("Include Azure Advisor recommendations (default: true). Set to false to skip Advisor scanning."),
		),
		mcp.WithBoolean("cost",
			mcp.Description("Include cost analysis (default: true). Set to false to skip cost analysis. Useful if you have permission issues."),
		),
		mcp.WithBoolean("policy",
			mcp.Description("Include Azure Policy compliance scanning (default: false). Set to true to enable policy scanning."),
		),
		mcp.WithBoolean("arc",
			mcp.Description("Include Arc-enabled SQL Server scanning (default: false). Set to true to enable Arc SQL scanning."),
		),
		mcp.WithBoolean("mask",
			mcp.DefaultBool(true),
			mcp.Description("Mask sensitive data in output (default: true)."),
		),
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
