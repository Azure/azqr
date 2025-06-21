// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterPrompts(s *server.MCPServer) {
	scanPrompt := mcp.NewPrompt(
		"scan_subscription",
		mcp.WithPromptDescription("Comprehensive azqr scan for an Azure subscription"),
		mcp.WithArgument("subscription_id", mcp.RequiredArgument()),
	)

	s.AddPrompt(scanPrompt, handleScanPrompt())

	// Carbon Emissions Plugin Prompt
	carbonEmissionsPrompt := mcp.NewPrompt(
		"analyze_carbon_emissions",
		mcp.WithPromptDescription("Analyze carbon emissions by Azure resource type across subscriptions"),
	)

	s.AddPrompt(carbonEmissionsPrompt, handleCarbonEmissionsPrompt())

	// OpenAI Throttling Plugin Prompt
	openAIThrottlingPrompt := mcp.NewPrompt(
		"check_openai_throttling",
		mcp.WithPromptDescription("Check OpenAI/Cognitive Services accounts for 429 throttling errors"),
	)

	s.AddPrompt(openAIThrottlingPrompt, handleOpenAIThrottlingPrompt())

	// Zone Mapping Plugin Prompt
	zoneMappingPrompt := mcp.NewPrompt(
		"get_zone_mappings",
		mcp.WithPromptDescription("Retrieve logical-to-physical availability zone mappings for all Azure regions"),
	)

	s.AddPrompt(zoneMappingPrompt, handleZoneMappingPrompt())
}
