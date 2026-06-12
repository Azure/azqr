// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"context"
	"time"

	"github.com/Azure/azqr/internal/az"
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

// With appends a stage to the pipeline and returns the builder for chaining.
func (b *ScanPipelineBuilder) With(s Stage) *ScanPipelineBuilder {
	b.stages = append(b.stages, s)
	return b
}

// Build creates the pipeline with all configured stages.
func (b *ScanPipelineBuilder) Build() *Pipeline {
	return NewPipeline(b.stages...)
}

// BuildDefault creates a pipeline with all standard stages.
// Note: The graph stage is mandatory for regular scans and cannot be disabled.
func (b *ScanPipelineBuilder) BuildDefault() *Pipeline {
	return b.
		With(NewProfilingStage()).
		With(NewInitializationStage()).
		With(NewSubscriptionDiscoveryStage()).
		With(NewResourceDiscoveryStage()).
		With(NewGraphScanStage()).
		With(NewDiagnosticsScanStage()).
		With(NewAdvisorStage()).
		With(NewDefenderStatusStage()).
		With(NewDefenderRecommendationsStage()).
		With(NewAzurePolicyStage()).
		With(NewArcSQLStage()).
		With(NewCostStage()).
		With(NewPluginExecutionStage()).
		With(NewReportRenderingStage()).
		With(NewProfilingCleanupStage()).
		Build()
}

// BuildPluginOnly creates a pipeline for plugin-only scans.
func (b *ScanPipelineBuilder) BuildPluginOnly() *Pipeline {
	return b.
		With(NewProfilingStage()).
		With(NewInitializationStage()).
		With(NewSubscriptionDiscoveryStage()).
		With(NewPluginExecutionStage()).
		With(NewReportRenderingStage()).
		With(NewProfilingCleanupStage()).
		Build()
}

// NewScanContext creates a scan context from ScanParams.
func NewScanContext(params *models.ScanParams) *ScanContext {
	ctx, cancel := context.WithCancel(context.Background())

	return &ScanContext{
		Ctx:           ctx,
		Cancel:        cancel,
		StartTime:     time.Now(),
		Params:        params,
		ClientOptions: az.NewDefaultClientOptions(),
	}
}
