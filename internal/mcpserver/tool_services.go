// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/mark3labs/mcp-go/mcp"
)

func handleServiceTypeTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	st := renderers.SupportedTypes{}
	output := st.GetAll()
	return mcp.NewToolResultText(output), nil
}

// ExecuteServicesTool is a public wrapper for handleServiceTypeTool that can be called from other packages
func ExecuteServicesTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return handleServiceTypeTool(ctx, request)
}
