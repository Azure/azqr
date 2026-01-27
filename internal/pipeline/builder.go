// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"context"
	"time"

	"github.com/Azure/azqr/internal/models"
)

// ScanPipelineBuilder provides a fluent interface for building scan pipelines.
type ScanPipelineBuilder struct {
	stages []Stage
}

// NewScanPipelineBuilder creates a new pipeline builder.
func NewScanPipelineBuilder() *ScanPipelineBuilder {
	return &ScanPipelineBuilder{
		stages: []Stage{},
	}
}

// WithInitialization adds the initialization stage.
func (b *ScanPipelineBuilder) WithInitialization() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewInitializationStage())
	return b
}

// WithSubscriptionDiscovery adds the subscription discovery stage.
func (b *ScanPipelineBuilder) WithSubscriptionDiscovery() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewSubscriptionDiscoveryStage())
	return b
}

// WithResourceDiscovery adds the resource discovery stage.
func (b *ScanPipelineBuilder) WithResourceDiscovery() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewResourceDiscoveryStage())
	return b
}

// WithAprlScan adds the APRL scanning stage.
func (b *ScanPipelineBuilder) WithAprlScan() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewAprlScanStage())
	return b
}

// WithPluginExecution adds the plugin execution stage.
func (b *ScanPipelineBuilder) WithPluginExecution() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewPluginExecutionStage())
	return b
}

// WithAzqrScan adds the AZQR service scan stage.
func (b *ScanPipelineBuilder) WithAzqrScan() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewAzqrScanStage())
	return b
}

// WithAdvisorDefender adds the advisor and defender scan stage.
func (b *ScanPipelineBuilder) WithAdvisorDefender() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewAdvisorDefenderStage())
	return b
}

// WithReportRendering adds the report rendering stage.
func (b *ScanPipelineBuilder) WithReportRendering() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewReportRenderingStage())
	return b
}

// WithProfiling adds the profiling setup stage (only effective with debug builds).
func (b *ScanPipelineBuilder) WithProfiling() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewProfilingStage())
	return b
}

// WithProfilingCleanup adds the profiling cleanup stage (only effective with debug builds).
func (b *ScanPipelineBuilder) WithProfilingCleanup() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewProfilingCleanupStage())
	return b
}

// WithCustomStage adds a custom stage to the pipeline.
func (b *ScanPipelineBuilder) WithCustomStage(stage Stage) *ScanPipelineBuilder {
	b.stages = append(b.stages, stage)
	return b
}

// Build creates the pipeline with all configured stages.
func (b *ScanPipelineBuilder) Build() *Pipeline {
	return NewPipeline(b.stages...)
}

// BuildDefault creates a pipeline with all standard stages.
func (b *ScanPipelineBuilder) BuildDefault() *Pipeline {
	return b.
		WithProfiling().
		WithInitialization().
		WithSubscriptionDiscovery().
		WithResourceDiscovery().
		WithAprlScan().
		WithPluginExecution().
		WithAzqrScan().
		WithAdvisorDefender().
		WithReportRendering().
		WithProfilingCleanup().
		Build()
}

// BuildPluginOnly creates a pipeline for plugin-only scans.
func (b *ScanPipelineBuilder) BuildPluginOnly() *Pipeline {
	return b.
		WithProfiling().
		WithInitialization().
		WithSubscriptionDiscovery().
		WithPluginExecution().
		WithReportRendering().
		WithProfilingCleanup().
		Build()
}

// NewScanContext creates a scan context from ScanParams.
func NewScanContext(params *models.ScanParams) *ScanContext {
	ctx, cancel := context.WithCancel(context.Background())

	return &ScanContext{
		Ctx:       ctx,
		Cancel:    cancel,
		StartTime: time.Now(),
		Params:    params,
	}
}

// NewScanContextWithContext creates a scan context with a custom context.
func NewScanContextWithContext(ctx context.Context, cancel context.CancelFunc, params *models.ScanParams) *ScanContext {
	return &ScanContext{
		Ctx:       ctx,
		Cancel:    cancel,
		StartTime: time.Now(),
		Params:    params,
	}
}
