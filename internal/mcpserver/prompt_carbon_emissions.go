// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func handleCarbonEmissionsPrompt() func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		prompt := `Analyze carbon emissions for Azure resources.

Please:
1. Use the scan-carbon-emissions tool to retrieve carbon emissions data
2. Analyze the results focusing on:
   - Resource types with highest emissions
   - Month-over-month trends (increases or decreases)
   - Significant emission changes that may require attention
3. Provide actionable insights for reducing carbon footprint
`

		promptMessage := mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompt))

		return mcp.NewGetPromptResult(
			"analyze carbon emissions",
			[]mcp.PromptMessage{promptMessage},
		), nil
	}
}
