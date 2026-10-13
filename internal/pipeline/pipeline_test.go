// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

// MockStage is a test helper stage.
type MockStage struct {
	*BaseStage
	executed  bool
	shouldErr bool
}

func NewMockStage(name string, required bool, shouldErr bool) *MockStage {
	return &MockStage{
		BaseStage: NewBaseStage(name, required),
		shouldErr: shouldErr,
	}
}

func (s *MockStage) Execute(ctx *ScanContext) error {
	s.executed = true
	if s.shouldErr {
		return errors.New("mock error")
	}
	return nil
}

func TestPipeline_Execute_Success(t *testing.T) {
	// Arrange
	stage1 := NewMockStage("stage1", true, false)
	stage2 := NewMockStage("stage2", true, false)
	stage3 := NewMockStage("stage3", true, false)

	pipeline := NewPipeline(stage1, stage2, stage3)

	ctx := &ScanContext{
		Ctx:        context.Background(),
		Params:     &models.ScanParams{},
		ReportData: &renderers.ReportData{},
	}

	// Act
	err := pipeline.Execute(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !stage1.executed {
		t.Error("Expected stage1 to be executed")
	}
	if !stage2.executed {
		t.Error("Expected stage2 to be executed")
	}
	if !stage3.executed {
		t.Error("Expected stage3 to be executed")
	}
	if pipeline.metrics.StagesExecuted != 3 {
		t.Errorf("Expected 3 stages executed, got %d", pipeline.metrics.StagesExecuted)
	}
	if pipeline.metrics.StagesSkipped != 0 {
		t.Errorf("Expected 0 stages skipped, got %d", pipeline.metrics.StagesSkipped)
	}
}

func TestPipeline_Execute_StageFailure(t *testing.T) {
	// Arrange
	stage1 := NewMockStage("stage1", true, false)
	stage2 := NewMockStage("stage2", true, true) // This one fails
	stage3 := NewMockStage("stage3", true, false)

	pipeline := NewPipeline(stage1, stage2, stage3)

	ctx := &ScanContext{
		Ctx:        context.Background(),
		Params:     &models.ScanParams{},
		ReportData: &renderers.ReportData{},
	}

	// Act
	err := pipeline.Execute(ctx)

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !stage1.executed {
		t.Error("Expected stage1 to be executed")
	}
	if !stage2.executed {
		t.Error("Expected stage2 to be executed")
	}
	if stage3.executed {
		t.Error("Expected stage3 NOT to be executed after failure")
	}
	if pipeline.metrics.StagesExecuted != 2 {
		t.Errorf("Expected 2 stages executed, got %d", pipeline.metrics.StagesExecuted)
	}
}

func TestPipeline_Execute_SkipOptionalStage(t *testing.T) {
	// Arrange
	stage1 := NewMockStage("stage1", true, false)
	stage2 := NewMockStage("stage2", false, false) // Optional stage
	stage3 := NewMockStage("stage3", true, false)

	pipeline := NewPipeline(stage1, stage2, stage3)

	ctx := &ScanContext{
		Ctx:        context.Background(),
		Params:     &models.ScanParams{},
		ReportData: &renderers.ReportData{},
	}

	// Act
	err := pipeline.Execute(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !stage1.executed {
		t.Error("Expected stage1 to be executed")
	}
	if stage2.executed {
		t.Error("Expected stage2 NOT to be executed (optional stage should be skipped)")
	}
	if !stage3.executed {
		t.Error("Expected stage3 to be executed")
	}
	if pipeline.metrics.StagesExecuted != 2 {
		t.Errorf("Expected 2 stages executed, got %d", pipeline.metrics.StagesExecuted)
	}
	if pipeline.metrics.StagesSkipped != 1 {
		t.Errorf("Expected 1 stage skipped, got %d", pipeline.metrics.StagesSkipped)
	}
}

func TestBaseStage_CanSkip(t *testing.T) {
	tests := []struct {
		name     string
		required bool
		expected bool
	}{
		{"Required stage cannot skip", true, false},
		{"Optional stage can skip", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage := NewBaseStage("test", tt.required)
			ctx := &ScanContext{}
			result := stage.Skip(ctx)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPluginExecutionStage_CanSkip(t *testing.T) {
	tests := []struct {
		name           string
		enabledPlugins map[string]bool
		expected       bool
	}{
		{"Skip when no plugins", map[string]bool{}, true},
		{"Skip when nil plugins", nil, true},
		{"Execute when plugins enabled", map[string]bool{"plugin1": true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage := NewPluginExecutionStage()
			ctx := &ScanContext{
				Params: &models.ScanParams{
					EnabledInternalPlugins: tt.enabledPlugins,
				},
			}
			result := stage.Skip(ctx)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPipeline_Integration(t *testing.T) {
	// This tests a realistic pipeline flow
	// Note: This will fail in actual execution without Azure credentials,
	// but demonstrates the pipeline structure

	// Arrange - Create a typical scan pipeline
	pipeline := NewPipeline(
		NewInitializationStage(),
		NewSubscriptionDiscoveryStage(),
		NewResourceDiscoveryStage(),
		NewGraphScanStage(),
		NewPluginExecutionStage(),
		NewAdvisorStage(),
		NewDefenderStatusStage(),
		NewDefenderRecommendationsStage(),
		NewAzurePolicyStage(),
		NewArcSQLStage(),
		NewCostStage(),
		NewFindingsSummaryStage(),
		NewReportRenderingStage(),
	)

	// Just verify pipeline structure is valid
	if len(pipeline.stages) != 13 {
		t.Errorf("Expected 13 stages, got %d", len(pipeline.stages))
	}

	t.Logf("Pipeline created with %d stages", len(pipeline.stages))
}

// recordingObserver is a test StageObserver that records every StageCompleted call it receives.
type recordingObserver struct {
	calls []struct {
		stageName string
		position  int
		status    StageStatus
		err       error
	}
}

func (o *recordingObserver) StageCompleted(ctx *ScanContext, stageName string, position int, status StageStatus, duration time.Duration, err error) {
	o.calls = append(o.calls, struct {
		stageName string
		position  int
		status    StageStatus
		err       error
	}{stageName, position, status, err})
}

func TestPipeline_Execute_NotifiesObserversOnSuccessSkipAndFailure(t *testing.T) {
	stage1 := NewMockStage("stage1", true, false)
	stage2 := NewMockStage("stage2", false, false) // optional, will be skipped
	stage3 := NewMockStage("stage3", true, true)   // fails

	p := NewPipeline(stage1, stage2, stage3)
	obs := &recordingObserver{}
	if returned := p.AddObserver(obs); returned != p {
		t.Error("AddObserver() should return the same *Pipeline for chaining")
	}

	ctx := &ScanContext{
		Ctx:        context.Background(),
		Params:     &models.ScanParams{},
		ReportData: &renderers.ReportData{},
	}

	err := p.Execute(ctx)
	if err == nil {
		t.Fatal("Expected error from failing stage3")
	}

	if len(obs.calls) != 3 {
		t.Fatalf("Expected 3 observer calls, got %d", len(obs.calls))
	}
	if obs.calls[0].stageName != "stage1" || obs.calls[0].status != StageStatusExecuted || obs.calls[0].err != nil {
		t.Errorf("call[0] = %+v, want executed stage1 with no error", obs.calls[0])
	}
	if obs.calls[1].stageName != "stage2" || obs.calls[1].status != StageStatusSkipped {
		t.Errorf("call[1] = %+v, want skipped stage2", obs.calls[1])
	}
	if obs.calls[2].stageName != "stage3" || obs.calls[2].status != StageStatusFailed || obs.calls[2].err == nil {
		t.Errorf("call[2] = %+v, want failed stage3 with an error", obs.calls[2])
	}
}

func TestPipeline_Execute_NoObserversIsNoOp(t *testing.T) {
	stage1 := NewMockStage("stage1", true, false)
	p := NewPipeline(stage1)

	ctx := &ScanContext{
		Ctx:        context.Background(),
		Params:     &models.ScanParams{},
		ReportData: &renderers.ReportData{},
	}

	if err := p.Execute(ctx); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestPipeline_AddObserver_IgnoresNil(t *testing.T) {
	p := NewPipeline()
	p.AddObserver(nil)
	if len(p.observers) != 0 {
		t.Errorf("Expected 0 observers after adding nil, got %d", len(p.observers))
	}
}
