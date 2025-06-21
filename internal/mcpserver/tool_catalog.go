// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/mark3labs/mcp-go/mcp"
)

func handleCatalogTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	output := renderers.GetAllRecommendations(true)

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
