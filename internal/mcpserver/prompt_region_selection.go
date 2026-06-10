// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func handleRegionSelectionPrompt() func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		targetRegions := request.Params.Arguments["target_regions"]
		prompt := `Analyze optimal Azure region selection for your workloads.

Please:
1. Use the scan-region-selection tool with options: {"target-regions": "%s"} to retrieve region analysis data
2. Analyze the results focusing on:
   - Service and SKU availability across target regions
   - Network latency between source and target regions
   - Cost differences between regions
   - Recommendation scores and quality indicators
3. Provide actionable recommendations for:
   - Best target regions based on availability, latency, and cost
   - Resources or SKUs that may not be available in target regions
   - Multi-region architecture considerations
   - Trade-offs between latency, cost, and service coverage
`
		promptText := fmt.Sprintf(prompt, targetRegions)
		promptMessage := mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(promptText))

		return mcp.NewGetPromptResult(
			"analyze region selection",
			[]mcp.PromptMessage{promptMessage},
		), nil
	}
}
