// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package copilot

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/mcpserver"
	"github.com/Azure/azqr/internal/models"
	copilot "github.com/github/copilot-sdk/go"
	"github.com/mark3labs/mcp-go/mcp"
)

type emptyParams struct{}

// BuildTools returns the list of Copilot tools for azqr.
// Tool state is tracked via SDK events (tool.execution_start/complete) in the TUI.
func BuildTools() []copilot.Tool {
	return []copilot.Tool{
		copilot.DefineTool("scan", "Run Azure Quick Review scan to analyze resources for compliance, security, and best practices",
			func(params models.ScanArgs, _ copilot.ToolInvocation) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				return executeScanTool(ctx, params)
			}),

		copilot.DefineTool("get-recommendations-catalog", "Get the complete catalog of azqr recommendations",
			func(_ emptyParams, _ copilot.ToolInvocation) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				return getRecommendationsCatalog(ctx)
			}),

		copilot.DefineTool("get-supported-services", "List all Azure services supported by azqr",
			func(_ emptyParams, _ copilot.ToolInvocation) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				return getSupportedServices(ctx)
			}),

		copilot.DefineTool("scan-carbon-emissions", "Analyze carbon emissions by Azure resource type with period-over-period tracking",
			func(params models.PluginScanArgs, _ copilot.ToolInvocation) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				return executePluginScanTool(ctx, "carbon-emissions", params)
			}),

		copilot.DefineTool("scan-openai-throttling", "Check OpenAI/Cognitive Services accounts for 429 throttling errors",
			func(params models.PluginScanArgs, _ copilot.ToolInvocation) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				return executePluginScanTool(ctx, "openai-throttling", params)
			}),

		copilot.DefineTool("scan-zone-mapping", "Retrieve logical-to-physical availability zone mappings for all Azure regions",
			func(params models.PluginScanArgs, _ copilot.ToolInvocation) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				return executePluginScanTool(ctx, "zone-mapping", params)
			}),
	}
}

func executeScanTool(ctx context.Context, params models.ScanArgs) (string, error) {
	result, err := mcpserver.ExecuteScanTool(ctx, mcp.CallToolRequest{}, params)
	if err != nil {
		return "", fmt.Errorf("scan failed: %w", err)
	}

	return extractMCPResultText(result), nil
}

func getRecommendationsCatalog(ctx context.Context) (string, error) {
	result, err := mcpserver.ExecuteCatalogTool(ctx, mcp.CallToolRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to get catalog: %w", err)
	}
	return extractMCPResultText(result), nil
}

func getSupportedServices(ctx context.Context) (string, error) {
	result, err := mcpserver.ExecuteServicesTool(ctx, mcp.CallToolRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to get services: %w", err)
	}
	return extractMCPResultText(result), nil
}

func executePluginScanTool(ctx context.Context, pluginName string, params models.PluginScanArgs) (string, error) {
	result, err := mcpserver.ExecutePluginScanTool(ctx, mcp.CallToolRequest{}, params, pluginName)
	if err != nil {
		return "", fmt.Errorf("%s scan failed: %w", pluginName, err)
	}

	return extractMCPResultText(result), nil
}

func extractMCPResultText(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return "✓ Operation completed successfully."
	}
	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		return textContent.Text
	}
	return "✓ Operation completed successfully."
}
