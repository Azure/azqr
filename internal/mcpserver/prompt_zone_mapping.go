// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func handleZoneMappingPrompt() func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		prompt := `Retrieve availability zone mappings for all Azure regions.

Please:
1. Use the scan-zone-mapping tool to get logical-to-physical zone mappings
2. Present the results showing:
   - Zone mappings per subscription
3. Explain the importance of zone mappings for:
   - High availability architecture
   - Disaster recovery planning
   - Cross-subscription zone alignment considerations
`

		promptText := prompt
		promptMessage := mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(promptText))

		return mcp.NewGetPromptResult(
			"get availability zone mappings",
			[]mcp.PromptMessage{promptMessage},
		), nil
	}
}
