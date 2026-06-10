// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func handleSQLESUPrompt() func(context.Context, mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		prompt := `Analyze SQL Server End-of-Life and Extended Security Update (ESU) status.

Please:
1. Use the scan-sql-esu tool to retrieve SQL Server EOL/ESU data
2. Analyze the results focusing on:
   - SQL Server instances approaching or past end-of-life
   - ESU coverage and cost implications
   - Estimated costs for ESU vs migration to SQL Managed Instance
   - Potential savings from modernization
3. Provide actionable recommendations for:
   - Prioritizing migrations based on cost and risk
   - Optimizing ESU spend where migration is not immediately feasible
   - Planning timeline for end-of-support transitions
`

		promptMessage := mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(prompt))

		return mcp.NewGetPromptResult(
			"analyze SQL ESU status",
			[]mcp.PromptMessage{promptMessage},
		), nil
	}
}
