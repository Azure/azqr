// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
)

type Scanner struct{}

// Scan performs a full scan using the default pipeline
func (sc *Scanner) Scan(params *models.ScanParams) *renderers.ReportData {
	return sc.scan(params, true)
}

// ScanPlugins performs a scan using only the plugin execution stage
func (sc *Scanner) ScanPlugins(params *models.ScanParams) *renderers.ReportData {
	return sc.scan(params, false)
}

// scan executes the scan using the composable pipeline pattern
func (sc *Scanner) scan(params *models.ScanParams, defaultPipeline bool) *renderers.ReportData {
	// Import pipeline package
	builder := NewScanPipelineBuilder()

	// Create scan context
	scanCtx := NewScanContext(params)

	var pipe *Pipeline
	if defaultPipeline {
		// Ensure graph stage is enabled for regular scans
		if err := params.Stages.ValidateGraphStageEnabled(); err != nil {
			log.Fatal().Err(err).Msg("Configuration error")
		}
		pipe = builder.BuildDefault()
	} else {
		pipe = builder.BuildPluginOnly()
	}

	err := pipe.Execute(scanCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("Scan failed")
	}

	// Log metrics in debug mode
	if params.Debug {
		pipe.LogMetrics()
	}

	// Log final timing
	elapsedTime := time.Since(scanCtx.StartTime)
	hours := int(elapsedTime.Hours())
	minutes := int(elapsedTime.Minutes()) % 60
	seconds := int(elapsedTime.Seconds()) % 60
	log.Info().Msgf("Scan completed in %02d:%02d:%02d", hours, minutes, seconds)

	return scanCtx.ReportData
}
