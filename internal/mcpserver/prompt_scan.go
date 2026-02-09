// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func handleScanPrompt() func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		subID := request.Params.Arguments["subscription_id"]
		prompt := `Perform a comprehensive azqr scan for Azure subscription %s.

Please:
1. Use the scan tool to analyze Azure resources against Well-Architected Framework
2. Analyze the results focusing on:
   - Critical and high-severity recommendations
   - Security posture from Microsoft Defender for Cloud
   - Cost optimization opportunities from Azure Advisor
   - Performance, reliability and observability concerns
   - Operational excellence improvements
3. Provide actionable recommendations for:
   - Highest priority issues to address
   - Quick wins for immediate improvement
   - Long-term architectural enhancements
   - Cost savings opportunities
`
		promptText := fmt.Sprintf(prompt, subID)
		promptMessage := mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(promptText))

		return mcp.NewGetPromptResult(
			"azqr scan subscription",
			[]mcp.PromptMessage{promptMessage},
		), nil
	}
}
