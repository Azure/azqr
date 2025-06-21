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
2. Analyze the results focusing on:
   - Instances experiencing throttling
   - Deployments and models affected
   - Time patterns of throttling (peak hours)
   - Spillover configuration status
3. Provide recommendations for:
   - Capacity planning and scaling
   - Load distribution strategies using Azure APIM
   - Spillover configuration optimization
`

		promptText := prompt
		promptMessage := mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(promptText))

		return mcp.NewGetPromptResult(
			"check OpenAI throttling",
			[]mcp.PromptMessage{promptMessage},
		), nil
	}
}
