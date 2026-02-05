// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
)

// PluginExecutionStage executes internal plugin scanners.
type PluginExecutionStage struct {
	*BaseStage
}

func NewPluginExecutionStage() *PluginExecutionStage {
	return &PluginExecutionStage{
		BaseStage: NewBaseStage("Plugin Execution", false),
	}
}

func (s *PluginExecutionStage) Skip(ctx *ScanContext) bool {
	return len(ctx.Params.EnabledInternalPlugins) == 0
}

func (s *PluginExecutionStage) Execute(ctx *ScanContext) error {
	pluginRegistry := plugins.GetRegistry()
	registeredPlugins := pluginRegistry.List()

	var internalPluginScanners []plugins.InternalPluginScanner
	for _, plugin := range registeredPlugins {
		if plugin.InternalScanner != nil {
			if enabled, ok := ctx.Params.EnabledInternalPlugins[plugin.Metadata.Name]; ok && enabled {
				log.Info().
					Str("plugin", plugin.Metadata.Name).
					Str("version", plugin.Metadata.Version).
					Msg("Executing internal plugin")
				internalPluginScanners = append(internalPluginScanners, plugin.InternalScanner)
			}
		}
	}

	// Execute plugins and collect results
	results := []*renderers.PluginResult{}
	for _, pluginScanner := range internalPluginScanners {
		result, err := pluginScanner.Scan(ctx.Ctx, ctx.Cred, ctx.Subscriptions, ctx.Params.Filters)
		if err != nil {
			log.Error().Err(err).Str("plugin", result.Metadata.Name).Msg("Plugin scan failed")
			continue
		}
		// Convert ExternalPluginOutput to PluginResult
		pluginResult := renderers.PluginResult{
			PluginName:  result.Metadata.Name,
			SheetName:   result.SheetName,
			Description: result.Description,
			Table:       result.Table,
		}
		results = append(results, &pluginResult)
	}

	ctx.ReportData.PluginResults = results

	log.Info().
		Int("plugins", len(internalPluginScanners)).
		Msg("Plugin execution completed")

	return nil
}
