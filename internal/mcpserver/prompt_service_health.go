// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func handleServiceHealthPrompt() func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		prompt := `Analyze Azure service health availability over the last 90 days.

Please:
1. Use the scan-service-health tool to retrieve service health data
2. Analyze the results focusing on:
   - Resource types and regions with the lowest availability (highest impact)
   - Number of service health events per subscription, region, and resource type
   - Patterns indicating recurring reliability problems
3. Provide actionable recommendations for:
   - Mitigating risk in regions or resource types with low availability
   - Designing resilient architectures to withstand future service events
`

		promptMessage := mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompt))

		return mcp.NewGetPromptResult(
			"analyze service health availability",
			[]mcp.PromptMessage{promptMessage},
		), nil
	}
}
