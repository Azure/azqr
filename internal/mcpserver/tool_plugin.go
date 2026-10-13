// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"fmt"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/Azure/azqr/internal/store"
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
		params.OutputName = fmt.Sprintf("%s/azqr_%s_results", currentDir, pluginName)

		// Apply plugin stage options if provided (e.g. target-regions for region-selection)
		if len(args.Options) > 0 {
			if err := params.Stages.SetStageOptions(models.StageNamePlugin, args.Options); err != nil {
				log.Fatal().Err(err).Msg("failed applying plugin stage options")
			}
		}

		// Enable the specific plugin for execution
		params.EnabledInternalPlugins = map[string]bool{
			pluginName: true,
		}

		var sess *store.Session
		r, err := runPersistedScanMCP(store.SessionKindPluginOnly, params, func(sc *pipeline.Scanner) (*findings.Summary, error) {
			return sc.ScanPlugins(params)
		}, func(s *store.Session) { sess = s })
		if err != nil {
			return nil, err
		}

		resultText := scanResultText(sess, params.OutputName, params.Xlsx, params.Json)
		result := mcp.NewToolResultStructured(scanSummaryEnvelope(sess, r), resultText)
		return result, nil
	}
}
