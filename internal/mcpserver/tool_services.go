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
