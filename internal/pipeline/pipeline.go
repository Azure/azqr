// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package pipeline provides a composable pipeline pattern for the scan execution flow.
// The pipeline breaks down the monolithic Scan() method into discrete, testable stages.
package pipeline

import (
	"context"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/rs/zerolog/log"
)

type ScanContext struct {
	Ctx           context.Context
	Cancel        context.CancelFunc
	Cred          azcore.TokenCredential
	ClientOptions *arm.ClientOptions
	Subscriptions map[string]string
	StartTime     time.Time
	Params        *models.ScanParams
	// Accumulated data through pipeline stages
	ReportData *renderers.ReportData
	// Profiler instance (if profiling is enabled)
	Profiler interface {
		Cleanup()
	}
}

// GetParams returns the ScanParams. This is a helper to avoid exposing internal package.
// Stages should type assert to *models.ScanParams.
func (sc *ScanContext) GetParams() interface{} {
	return sc.Params
}

// Stage represents a single stage in the scan pipeline.
// Each stage is a composable unit that processes the scan context.
type Stage interface {
	// Name returns the stage name for logging and metrics.
	Name() string

	// Execute runs the stage logic, modifying the scan context.
	// Returns error if the stage fails critically.
	Execute(ctx *ScanContext) error

	// Skip determines if this stage can be skipped based on context.
	// For example, skip Graph stage if UseGraphRecommendations is false.
	Skip(ctx *ScanContext) bool
}

// Pipeline orchestrates the execution of multiple stages in sequence.
type Pipeline struct {
	stages  []Stage
	metrics *PipelineMetrics
}

// PipelineMetrics tracks performance of each pipeline stage.
type PipelineMetrics struct {
	TotalDuration  time.Duration
	StageDurations map[string]time.Duration
	StageErrors    map[string]error
	StagesExecuted int
	StagesSkipped  int
}

// NewPipeline creates a new scan pipeline with the given stages.
func NewPipeline(stages ...Stage) *Pipeline {
	return &Pipeline{
		stages: stages,
		metrics: &PipelineMetrics{
			StageDurations: make(map[string]time.Duration),
			StageErrors:    make(map[string]error),
		},
	}
}

// Execute runs all pipeline stages in sequence.
func (p *Pipeline) Execute(ctx *ScanContext) error {
	startTime := time.Now()
	log.Info().
		Int("stages", len(p.stages)).
		Msg("Scan started")

	for i, stage := range p.stages {
		stageName := stage.Name()

		// Check if stage can be skipped
		if stage.Skip(ctx) {
			log.Debug().
				Str("stage", stageName).
				Int("position", i+1).
				Msg("Skipping stage")
			p.metrics.StagesSkipped++
			continue
		}

		// Execute stage
		log.Debug().
			Str("stage", stageName).
			Int("position", i+1).
			Int("total", len(p.stages)).
			Msg("Executing stage")

		stageStart := time.Now()
		err := stage.Execute(ctx)
		stageDuration := time.Since(stageStart)

		p.metrics.StageDurations[stageName] = stageDuration
		p.metrics.StagesExecuted++

		if err != nil {
			log.Error().
				Err(err).
				Str("stage", stageName).
				Dur("duration", stageDuration).
				Msg("Stage failed")
			p.metrics.StageErrors[stageName] = err
			return err
		}

		log.Debug().
			Str("stage", stageName).
			Dur("duration", stageDuration).
			Msg("Stage completed")
	}

	p.metrics.TotalDuration = time.Since(startTime)

	log.Debug().
		Dur("total_duration", p.metrics.TotalDuration).
		Int("executed", p.metrics.StagesExecuted).
		Int("skipped", p.metrics.StagesSkipped).
		Msg("Scan completed")

	return nil
}

// GetMetrics returns the pipeline execution metrics.
func (p *Pipeline) GetMetrics() *PipelineMetrics {
	return p.metrics
}

// LogMetrics logs detailed pipeline metrics (for debug mode).
func (p *Pipeline) LogMetrics() {
	log.Debug().Msg("=== Scan Performance Metrics ===")
	for i, stage := range p.stages {
		stageName := stage.Name()
		if duration, ok := p.metrics.StageDurations[stageName]; ok {
			percentage := float64(duration) / float64(p.metrics.TotalDuration) * 100
			log.Debug().
				Int("position", i+1).
				Str("stage", stageName).
				Dur("duration", duration).
				Float64("percentage", percentage).
				Msg("Stage metrics")
		}
	}
	log.Debug().
		Dur("total", p.metrics.TotalDuration).
		Int("executed", p.metrics.StagesExecuted).
		Int("skipped", p.metrics.StagesSkipped).
		Msg("=== End Scan Metrics ===")
}

// BaseStage provides default implementations for Stage interface.
// Stages can embed this to inherit default behavior.
type BaseStage struct {
	name     string
	required bool
}

// NewBaseStage creates a base stage with name and required flag.
func NewBaseStage(name string, required bool) *BaseStage {
	return &BaseStage{
		name:     name,
		required: required,
	}
}

// Name implements Stage.Name().
func (s *BaseStage) Name() string {
	return s.name
}

// CanSkip implements Stage.CanSkip().
// By default, required stages cannot be skipped.
func (s *BaseStage) Skip(ctx *ScanContext) bool {
	return !s.required
}
