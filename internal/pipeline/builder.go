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

// WithGraphScan adds the APRL scanning stage.
func (b *ScanPipelineBuilder) WithGraphScan() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewGraphScanStage())
	return b
}

// WithDiagnosticsScan adds the diagnostics scanning stage.
func (b *ScanPipelineBuilder) WithDiagnosticsScan() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewDiagnosticsScanStage())
	return b
}

// WithPluginExecution adds the plugin execution stage.
func (b *ScanPipelineBuilder) WithPluginExecution() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewPluginExecutionStage())
	return b
}

// WithAdvisor adds the advisor scan stage.
func (b *ScanPipelineBuilder) WithAdvisor() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewAdvisorStage())
	return b
}

// WithDefenderStatus adds the defender status scan stage.
func (b *ScanPipelineBuilder) WithDefenderStatus() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewDefenderStatusStage())
	return b
}

// WithDefenderRecommendations adds the defender recommendations scan stage.
func (b *ScanPipelineBuilder) WithDefenderRecommendations() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewDefenderRecommendationsStage())
	return b
}

// WithAzurePolicy adds the Azure Policy scan stage.
func (b *ScanPipelineBuilder) WithAzurePolicy() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewAzurePolicyStage())
	return b
}

// WithArcSQL adds the Arc-enabled SQL Server scan stage.
func (b *ScanPipelineBuilder) WithArcSQL() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewArcSQLStage())
	return b
}

// WithCost adds the Cost analysis stage.
func (b *ScanPipelineBuilder) WithCost() *ScanPipelineBuilder {
	b.stages = append(b.stages, NewCostStage())
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

// Build creates the pipeline with all configured stages.
func (b *ScanPipelineBuilder) Build() *Pipeline {
	return NewPipeline(b.stages...)
}

// BuildDefault creates a pipeline with all standard stages.
// Note: The graph stage is mandatory for regular scans and cannot be disabled.
func (b *ScanPipelineBuilder) BuildDefault() *Pipeline {
	return b.
		WithProfiling().
		WithInitialization().
		WithSubscriptionDiscovery().
		WithResourceDiscovery().
		WithGraphScan().
		WithDiagnosticsScan().
		WithAdvisor().
		WithDefenderStatus().
		WithDefenderRecommendations().
		WithAzurePolicy().
		WithArcSQL().
		WithCost().
		WithPluginExecution().
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
		Ctx:           ctx,
		Cancel:        cancel,
		StartTime:     time.Now(),
		Params:        params,
		ClientOptions: az.NewDefaultClientOptions(),
	}
}
