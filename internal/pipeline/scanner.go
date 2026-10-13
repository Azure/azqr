// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/rs/zerolog/log"
)

type Scanner struct {
	// Observers, if set, are registered on the pipeline before it executes, so each can react
	// to stage completion (for example, internal/store.Recorder persisting partial results).
	Observers []StageObserver

	lastMetrics PipelineMetrics
}

// LastMetrics returns the PipelineMetrics captured by the most recently completed Scan/ScanPlugins
// call on this Scanner. Zero-valued until the first scan completes.
func (sc *Scanner) LastMetrics() PipelineMetrics {
	return sc.lastMetrics
}

// Scan performs a full scan using the default pipeline
func (sc *Scanner) Scan(params *models.ScanParams) (*findings.Summary, error) {
	return sc.scan(params, true)
}

// ScanPlugins performs a scan using only the plugin execution stage
func (sc *Scanner) ScanPlugins(params *models.ScanParams) (*findings.Summary, error) {
	return sc.scan(params, false)
}

// scan executes the scan using the composable pipeline pattern
func (sc *Scanner) scan(params *models.ScanParams, defaultPipeline bool) (*findings.Summary, error) {
	builder := NewScanPipelineBuilder()

	// Create scan context
	scanCtx := NewScanContext(params)

	var pipe *Pipeline
	if defaultPipeline {
		// Ensure graph stage is enabled for regular scans
		if err := params.Stages.ValidateGraphStageEnabled(); err != nil {
			return nil, fmt.Errorf("invalid scan configuration: %w", err)
		}
		pipe = builder.BuildDefault()
	} else {
		pipe = builder.BuildPluginOnly()
	}

	for _, o := range sc.Observers {
		pipe.AddObserver(o)
	}

	err := pipe.Execute(scanCtx)
	sc.lastMetrics = pipe.Metrics()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
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

	return scanCtx.ReportData.Summary, nil
}
