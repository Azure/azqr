// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package internal

import (
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/Azure/azqr/internal/renderers/json"
	"github.com/rs/zerolog/log"
)

type Scanner struct{}

// Scan performs a full scan using the default pipeline
func (sc *Scanner) Scan(params *models.ScanParams) string {
	return sc.scan(params, true)
}

// ScanPlugins performs a scan using only the plugin execution stage
func (sc Scanner) ScanPlugins(params *models.ScanParams) string {
	return sc.scan(params, false)
}

// scan executes the scan using the composable pipeline pattern
func (sc *Scanner) scan(params *models.ScanParams, defaultPipeline bool) string {
	// Import pipeline package
	builder := pipeline.NewScanPipelineBuilder()

	// Create scan context
	scanCtx := pipeline.NewScanContext(params)

	var pipeline *pipeline.Pipeline
	if defaultPipeline {
		// Ensure graph stage is enabled for regular scans
		if err := params.Stages.ValidateGraphStageEnabled(); err != nil {
			log.Fatal().Err(err).Msg("Configuration error")
		}
		pipeline = builder.BuildDefault()
	} else {
		pipeline = builder.BuildPluginOnly()
	}

	err := pipeline.Execute(scanCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("Scan failed")
	}

	// Log metrics in debug mode
	if params.Debug {
		pipeline.LogMetrics()
	}

	// Log final timing
	elapsedTime := time.Since(scanCtx.StartTime)
	hours := int(elapsedTime.Hours())
	minutes := int(elapsedTime.Minutes()) % 60
	seconds := int(elapsedTime.Seconds()) % 60
	log.Info().Msgf("Scan completed in %02d:%02d:%02d", hours, minutes, seconds)

	outputJson := json.CreateJsonOutput(scanCtx.ReportData)
	return outputJson
}
