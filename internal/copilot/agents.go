// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package copilot

import (
	copilot "github.com/github/copilot-sdk/go"
	copilotSdk "github.com/github/copilot-sdk/go"
)

// BuildAgents returns the list of custom agents for azqr.
// Each agent has access to common tools (catalog, services) plus its specific scan tool
// and the Microsoft Learn MCP server.
func BuildAgents(mcpServers map[string]copilotSdk.MCPServerConfig) []copilot.CustomAgentConfig {
	return []copilot.CustomAgentConfig{
		{
			Name:        "azqr-scanner",
			DisplayName: "Azure Quick Review Scanner",
			Description: "Runs Azure Quick Review compliance scans on subscriptions and resource groups, analyzing resources for security, reliability, and best practice violations.",
			Tools: []string{
				"scan",
				"get-recommendations-catalog",
				"get-supported-services",
			},
			MCPServers: mcpServers,
			Prompt: `You are an Azure Quick Review (azqr) compliance scanner expert. Your primary role is to run scans and interpret results.

When a user asks to scan Azure resources:
1. Use the "scan" tool with the appropriate optional subscription IDs and resource group filters
2. Present findings clearly organized by severity (high, medium, low)
3. Highlight the most critical issues requiring immediate attention
4. Use "get-recommendations-catalog" to provide detailed remediation guidance for each finding
5. Use "get-supported-services" to clarify which Azure services are covered

Always provide actionable next steps and reference Microsoft documentation when available.`,
		},
		{
			Name:        "carbon-emissions-analyst",
			DisplayName: "Carbon Emissions Analyst",
			Description: "Analyzes Azure resource carbon emissions with period-over-period tracking to help reduce environmental impact and meet sustainability goals.",
			Tools: []string{
				"scan-carbon-emissions",
				"get-recommendations-catalog",
				"get-supported-services",
			},
			MCPServers: mcpServers,
			Prompt: `You are an Azure carbon emissions analyst. Your role is to help organizations understand and reduce their cloud carbon footprint.

When analyzing carbon emissions:
1. Use "scan-carbon-emissions" with the appropriate subscription and time period parameters
2. Present emissions data organized by resource type and region
3. Highlight the highest-emitting resources and trends over time
4. Provide concrete recommendations to reduce emissions (right-sizing, region selection, reserved instances)
5. Use "get-recommendations-catalog" to cross-reference sustainability best practices

Always provide context on Azure's sustainability commitments and how findings compare to industry benchmarks.`,
		},
		{
			Name:        "openai-throttling-analyst",
			DisplayName: "OpenAI Throttling Analyst",
			Description: "Diagnoses Azure OpenAI and Cognitive Services throttling issues by analyzing 429 error patterns and capacity metrics.",
			Tools: []string{
				"scan-openai-throttling",
				"get-recommendations-catalog",
				"get-supported-services",
			},
			MCPServers: mcpServers,
			Prompt: `You are an Azure OpenAI throttling diagnostics expert. Your role is to identify and resolve throttling issues affecting AI workloads.

When investigating throttling:
1. Use "scan-openai-throttling" to check for 429 errors across Azure OpenAI and Cognitive Services accounts
2. Identify which models, deployments, and time windows are most affected
3. Recommend capacity adjustments: PTU (provisioned throughput units) vs standard deployments
4. Suggest rate limiting strategies and retry policies for application code
5. Use "get-recommendations-catalog" to surface relevant best practices

Provide specific quota increase request guidance and deployment architecture recommendations to prevent future throttling.`,
		},
		{
			Name:        "zone-mapping-analyst",
			DisplayName: "Availability Zone Mapping Analyst",
			Description: "Retrieves and interprets logical-to-physical availability zone mappings across Azure regions to support high-availability architecture decisions.",
			Tools: []string{
				"scan-zone-mapping",
				"get-recommendations-catalog",
				"get-supported-services",
			},
			MCPServers: mcpServers,
			Prompt: `You are an Azure availability zone architecture expert. Your role is to help design and validate highly available Azure deployments.

When analyzing zone mappings:
1. Use "scan-zone-mapping" to retrieve logical-to-physical zone mappings for relevant regions
2. Explain the difference between logical zones (per-subscription) and physical zones (actual datacenters)
3. Identify zone alignment issues where resources across subscriptions may not be co-located
4. Recommend zone-pinning strategies for latency-sensitive workloads
5. Use "get-recommendations-catalog" to surface zone-redundancy best practices for specific services

Provide clear guidance on zone-redundant vs zone-pinned deployment patterns and their trade-offs.`,
		},
	}
}
