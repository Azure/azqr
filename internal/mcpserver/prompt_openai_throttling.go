// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func handleOpenAIThrottlingPrompt() func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		prompt := `Check OpenAI/Cognitive Services throttling for Azure resources.

Please:
1. Use the scan-openai-throttling tool to detect 429 throttling errors
2. Calculate throttling rate for each Deployment/Model/Environment combination:
   - Throttling Rate (%) = (429 status count / total requests) × 100
   - HIGHLIGHT combinations where throttling rate > 1% as CRITICAL
   - Flag combinations with 0.1-1% as WARNING
   - Mark combinations with < 0.1% as HEALTHY
3. Analyze the results focusing on:
   - Instances experiencing throttling (group by account, deployment, model)
   - Deployments and models affected (calculate rates per combination)
   - Time patterns of throttling (identify peak hours and recurring patterns)
   - Spillover configuration status (check if spillover is enabled)
4. Provide prioritized recommendations for:
   - Capacity planning and scaling (focus on CRITICAL combinations first)
   - Load distribution strategies using Azure APIM
   - Spillover configuration optimization
   - Immediate actions vs. long-term improvements
`

		promptMessage := mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompt))

		return mcp.NewGetPromptResult(
			"check OpenAI throttling",
			[]mcp.PromptMessage{promptMessage},
		), nil
	}
}
