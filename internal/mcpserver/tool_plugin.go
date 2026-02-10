// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"fmt"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

// scanPluginHandler creates a handler for plugin-specific scans
func scanPluginHandler(pluginName string) func(context.Context, mcp.CallToolRequest, models.PluginScanArgs) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest, args models.PluginScanArgs) (*mcp.CallToolResult, error) {
		currentDir, err := getCurrentFolder(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get current working directory")
		}

		params := models.NewScanParamsForPlugins(args)
		params.Xlsx = true
		params.Json = true
		params.OutputName = fmt.Sprintf("%s/azqr_%s_results", currentDir, pluginName)
		params.EnabledInternalPlugins = map[string]bool{pluginName: true}

		scanner := pipeline.Scanner{}
		r := scanner.ScanPlugins(params)

		fileName := r.OutputFileName + ".xlsx"
		uri := fmt.Sprintf("file://%s", fileName)
		uriJSON := fmt.Sprintf("file://%s.json", r.OutputFileName)

		resultName := fmt.Sprintf("Azure Quick Review %s Scan Results", pluginName)
		registerScanResources(r.OutputFileName, resultName, uriJSON, uri)

		resultText := fmt.Sprintf("Scan results saved to:\n- Excel: %s\n- JSON: %s", uri, uriJSON)
		return mcp.NewToolResultStructured(r, resultText), nil
	}
}

// ExecutePluginScanTool is a public wrapper for scanPluginHandler that can be called from other packages
func ExecutePluginScanTool(ctx context.Context, request mcp.CallToolRequest, args models.PluginScanArgs, pluginName string) (*mcp.CallToolResult, error) {
	return scanPluginHandler(pluginName)(ctx, request, args)
}
