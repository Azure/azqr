// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"context"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

// --- helpers ----------------------------------------------------------------

// stageEnabledCtx returns a ScanContext with exactly one stage enabled.
func stageEnabledCtx(stageName string) *ScanContext {
	stages := models.NewStageConfigs()
	_ = stages.EnableStage(stageName)
	return &ScanContext{
		Ctx:        context.Background(),
		Params:     &models.ScanParams{Stages: stages},
		ReportData: &renderers.ReportData{},
	}
}

// stageDisabledCtx returns a ScanContext where no stage is enabled.
func stageDisabledCtx() *ScanContext {
	return &ScanContext{
		Ctx:        context.Background(),
		Params:     &models.ScanParams{Stages: models.NewStageConfigs()},
		ReportData: &renderers.ReportData{},
	}
}

// --- TestSimpleStage --------------------------------------------------------

func TestSimpleStage_Skip_WhenStageDisabled(t *testing.T) {
	s := &simpleStage[string]{
		BaseStage: NewBaseStage("test", false),
		stageName: models.StageNameAdvisor,
		run:       func(*ScanContext) string { return "" },
		assign:    func(*renderers.ReportData, string) {},
	}
	if !s.Skip(stageDisabledCtx()) {
		t.Error("Skip should return true when stage is disabled")
	}
}

func TestSimpleStage_Skip_WhenStageEnabled(t *testing.T) {
	s := &simpleStage[string]{
		BaseStage: NewBaseStage("test", false),
		stageName: models.StageNameAdvisor,
		run:       func(*ScanContext) string { return "" },
		assign:    func(*renderers.ReportData, string) {},
	}
	if s.Skip(stageEnabledCtx(models.StageNameAdvisor)) {
		t.Error("Skip should return false when stage is enabled")
	}
}

func TestSimpleStage_Execute_CallsRunAndAssign(t *testing.T) {
	runCalled := false
	assignCalled := false
	const want = "scan-result"

	s := &simpleStage[string]{
		BaseStage: NewBaseStage("test", false),
		stageName: models.StageNameAdvisor,
		run: func(*ScanContext) string {
			runCalled = true
			return want
		},
		assign: func(_ *renderers.ReportData, got string) {
			assignCalled = true
			if got != want {
				t.Errorf("assign received %q, want %q", got, want)
			}
		},
	}

	ctx := stageEnabledCtx(models.StageNameAdvisor)
	if err := s.Execute(ctx); err != nil {
		t.Fatalf("Execute returned unexpected error: %v", err)
	}
	if !runCalled {
		t.Error("run function was not called")
	}
	if !assignCalled {
		t.Error("assign function was not called")
	}
}

func TestSimpleStage_Execute_ReturnsNilError(t *testing.T) {
	s := &simpleStage[int]{
		BaseStage: NewBaseStage("test", false),
		stageName: models.StageNameAdvisor,
		run:       func(*ScanContext) int { return 42 },
		assign:    func(*renderers.ReportData, int) {},
	}
	if err := s.Execute(stageEnabledCtx(models.StageNameAdvisor)); err != nil {
		t.Errorf("Execute should always return nil, got %v", err)
	}
}

// --- Per-constructor tests --------------------------------------------------
// Each test verifies: correct stage name, correct stage flag, and that
// Skip responds correctly to the matching stage being enabled/disabled.

func TestNewAdvisorStage(t *testing.T) {
	s := NewAdvisorStage()

	t.Run("name", func(t *testing.T) {
		if got := s.Name(); got != "Advisor Scan" {
			t.Errorf("Name() = %q, want %q", got, "Advisor Scan")
		}
	})
	t.Run("skip_when_disabled", func(t *testing.T) {
		if !s.Skip(stageDisabledCtx()) {
			t.Error("Skip should return true when advisor stage is disabled")
		}
	})
	t.Run("run_when_enabled", func(t *testing.T) {
		if s.Skip(stageEnabledCtx(models.StageNameAdvisor)) {
			t.Error("Skip should return false when advisor stage is enabled")
		}
	})
}

func TestNewArcSQLStage(t *testing.T) {
	s := NewArcSQLStage()

	t.Run("name", func(t *testing.T) {
		if got := s.Name(); got != "Arc-enabled SQL Server Scan" {
			t.Errorf("Name() = %q, want %q", got, "Arc-enabled SQL Server Scan")
		}
	})
	t.Run("skip_when_disabled", func(t *testing.T) {
		if !s.Skip(stageDisabledCtx()) {
			t.Error("Skip should return true when arc stage is disabled")
		}
	})
	t.Run("run_when_enabled", func(t *testing.T) {
		if s.Skip(stageEnabledCtx(models.StageNameArc)) {
			t.Error("Skip should return false when arc stage is enabled")
		}
	})
	t.Run("not_triggered_by_wrong_stage", func(t *testing.T) {
		if !s.Skip(stageEnabledCtx(models.StageNameAdvisor)) {
			t.Error("Skip should return true when a different stage (advisor) is enabled, not arc")
		}
	})
}

func TestNewAzurePolicyStage(t *testing.T) {
	s := NewAzurePolicyStage()

	t.Run("name", func(t *testing.T) {
		if got := s.Name(); got != "Azure Policy Scan" {
			t.Errorf("Name() = %q, want %q", got, "Azure Policy Scan")
		}
	})
	t.Run("skip_when_disabled", func(t *testing.T) {
		if !s.Skip(stageDisabledCtx()) {
			t.Error("Skip should return true when policy stage is disabled")
		}
	})
	t.Run("run_when_enabled", func(t *testing.T) {
		if s.Skip(stageEnabledCtx(models.StageNamePolicy)) {
			t.Error("Skip should return false when policy stage is enabled")
		}
	})
	t.Run("not_triggered_by_wrong_stage", func(t *testing.T) {
		if !s.Skip(stageEnabledCtx(models.StageNameAdvisor)) {
			t.Error("Skip should return true when a different stage (advisor) is enabled, not policy")
		}
	})
}

func TestNewDefenderStatusStage(t *testing.T) {
	s := NewDefenderStatusStage()

	t.Run("name", func(t *testing.T) {
		if got := s.Name(); got != "Defender Status Scan" {
			t.Errorf("Name() = %q, want %q", got, "Defender Status Scan")
		}
	})
	t.Run("skip_when_disabled", func(t *testing.T) {
		if !s.Skip(stageDisabledCtx()) {
			t.Error("Skip should return true when defender stage is disabled")
		}
	})
	t.Run("run_when_enabled", func(t *testing.T) {
		if s.Skip(stageEnabledCtx(models.StageNameDefender)) {
			t.Error("Skip should return false when defender stage is enabled")
		}
	})
	t.Run("not_triggered_by_wrong_stage", func(t *testing.T) {
		if !s.Skip(stageEnabledCtx(models.StageNameAdvisor)) {
			t.Error("Skip should return true when a different stage (advisor) is enabled, not defender")
		}
	})
}

func TestNewDefenderRecommendationsStage(t *testing.T) {
	s := NewDefenderRecommendationsStage()

	t.Run("name", func(t *testing.T) {
		if got := s.Name(); got != "Defender Recommendations Scan" {
			t.Errorf("Name() = %q, want %q", got, "Defender Recommendations Scan")
		}
	})
	t.Run("skip_when_disabled", func(t *testing.T) {
		if !s.Skip(stageDisabledCtx()) {
			t.Error("Skip should return true when defender-recommendations stage is disabled")
		}
	})
	t.Run("run_when_enabled", func(t *testing.T) {
		if s.Skip(stageEnabledCtx(models.StageNameDefenderRecommendations)) {
			t.Error("Skip should return false when defender-recommendations stage is enabled")
		}
	})
	t.Run("not_triggered_by_wrong_stage", func(t *testing.T) {
		if !s.Skip(stageEnabledCtx(models.StageNameDefender)) {
			t.Error("Skip should return true when defender (not defender-recommendations) is enabled")
		}
	})
}

// --- Builder tests ----------------------------------------------------------

func TestScanPipelineBuilder_With_AppendsAndChains(t *testing.T) {
	mock1 := NewMockStage("s1", true, false)
	mock2 := NewMockStage("s2", true, false)

	builder := NewScanPipelineBuilder()
	result := builder.With(mock1).With(mock2)

	// Chainable: same builder instance returned
	if result != builder {
		t.Error("With() should return the same builder instance for chaining")
	}

	pipe := builder.Build()
	if len(pipe.stages) != 2 {
		t.Errorf("pipeline has %d stages, want 2", len(pipe.stages))
	}
	if pipe.stages[0].Name() != "s1" {
		t.Errorf("stage[0].Name() = %q, want s1", pipe.stages[0].Name())
	}
	if pipe.stages[1].Name() != "s2" {
		t.Errorf("stage[1].Name() = %q, want s2", pipe.stages[1].Name())
	}
}

func TestBuildDefault_StageCountAndOrder(t *testing.T) {
	pipe := NewScanPipelineBuilder().BuildDefault()

	wantNames := []string{
		"Profiling Setup",
		"Initialization",
		"Subscription Discovery",
		"Resource Discovery",
		"Graph Scan",
		"Diagnostics Settings Scan",
		"Advisor Scan",
		"Defender Status Scan",
		"Defender Recommendations Scan",
		"Azure Policy Scan",
		"Arc-enabled SQL Server Scan",
		"Cost Analysis Scan",
		"Plugin Execution",
		"Report Rendering",
		"Profiling Cleanup",
	}

	if got := len(pipe.stages); got != len(wantNames) {
		t.Fatalf("BuildDefault produced %d stages, want %d", got, len(wantNames))
	}

	for i, want := range wantNames {
		if got := pipe.stages[i].Name(); got != want {
			t.Errorf("stage[%d].Name() = %q, want %q", i, got, want)
		}
	}
}

func TestBuildPluginOnly_StageCountAndOrder(t *testing.T) {
	pipe := NewScanPipelineBuilder().BuildPluginOnly()

	wantNames := []string{
		"Profiling Setup",
		"Initialization",
		"Subscription Discovery",
		"Plugin Execution",
		"Report Rendering",
		"Profiling Cleanup",
	}

	if got := len(pipe.stages); got != len(wantNames) {
		t.Fatalf("BuildPluginOnly produced %d stages, want %d", got, len(wantNames))
	}

	for i, want := range wantNames {
		if got := pipe.stages[i].Name(); got != want {
			t.Errorf("stage[%d].Name() = %q, want %q", i, got, want)
		}
	}
}

// TestBuildDefault_NoDuplicateStageNames ensures every stage in the default
// pipeline has a unique name (duplicate names break metrics tracking).
func TestBuildDefault_NoDuplicateStageNames(t *testing.T) {
	pipe := NewScanPipelineBuilder().BuildDefault()
	seen := make(map[string]int)
	for i, s := range pipe.stages {
		seen[s.Name()]++
		if seen[s.Name()] > 1 {
			t.Errorf("stage[%d] name %q appears more than once in BuildDefault", i, s.Name())
		}
	}
}

// TestWith_EmptyBuilderProducesEmptyPipeline verifies the zero-value behaviour.
func TestWith_EmptyBuilderProducesEmptyPipeline(t *testing.T) {
	pipe := NewScanPipelineBuilder().Build()
	if len(pipe.stages) != 0 {
		t.Errorf("empty builder: want 0 stages, got %d", len(pipe.stages))
	}
}
