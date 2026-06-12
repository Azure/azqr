// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/scanners"
)

// simpleStage is a generic pipeline stage that covers the common pattern of:
//  1. Checking whether a named stage flag is enabled (Skip)
//  2. Calling a scan function (run)
//  3. Assigning the result to a ReportData field (assign)
//
// This removes the need for separate, near-identical stage structs for each
// scanner that follows this pattern.
type simpleStage[T any] struct {
	*BaseStage
	stageName string
	run       func(*ScanContext) T
	assign    func(*renderers.ReportData, T)
}

// Skip returns true when the stage's flag is not enabled in the scan params.
func (s *simpleStage[T]) Skip(ctx *ScanContext) bool {
	return !ctx.Params.Stages.IsStageEnabled(s.stageName)
}

// Execute calls the scan function and stores the result in ReportData.
func (s *simpleStage[T]) Execute(ctx *ScanContext) error {
	s.assign(ctx.ReportData, s.run(ctx))
	return nil
}

// NewAdvisorStage creates the Advisor scan stage.
func NewAdvisorStage() Stage {
	return &simpleStage[[]*models.AdvisorResult]{
		BaseStage: NewBaseStage("Advisor Scan", false),
		stageName: models.StageNameAdvisor,
		run: func(ctx *ScanContext) []*models.AdvisorResult {
			return (&scanners.AdvisorScanner{}).Scan(ctx.Ctx, ctx.Cred, ctx.Subscriptions, ctx.Params.Filters)
		},
		assign: func(rd *renderers.ReportData, r []*models.AdvisorResult) { rd.Advisor = r },
	}
}

// NewArcSQLStage creates the Arc-enabled SQL Server scan stage.
func NewArcSQLStage() Stage {
	return &simpleStage[[]*models.ArcSQLResult]{
		BaseStage: NewBaseStage("Arc-enabled SQL Server Scan", false),
		stageName: models.StageNameArc,
		run: func(ctx *ScanContext) []*models.ArcSQLResult {
			return (&scanners.ArcSQLScanner{}).Scan(ctx.Ctx, ctx.Cred, ctx.Subscriptions, ctx.Params.Filters)
		},
		assign: func(rd *renderers.ReportData, r []*models.ArcSQLResult) { rd.ArcSQL = r },
	}
}

// NewAzurePolicyStage creates the Azure Policy scan stage.
func NewAzurePolicyStage() Stage {
	return &simpleStage[[]*models.AzurePolicyResult]{
		BaseStage: NewBaseStage("Azure Policy Scan", false),
		stageName: models.StageNamePolicy,
		run: func(ctx *ScanContext) []*models.AzurePolicyResult {
			return (&scanners.AzurePolicyScanner{}).Scan(ctx.Ctx, ctx.Cred, ctx.Subscriptions, ctx.Params.Filters)
		},
		assign: func(rd *renderers.ReportData, r []*models.AzurePolicyResult) { rd.AzurePolicy = r },
	}
}

// NewDefenderStatusStage creates the Defender status scan stage.
func NewDefenderStatusStage() Stage {
	return &simpleStage[[]*models.DefenderResult]{
		BaseStage: NewBaseStage("Defender Status Scan", false),
		stageName: models.StageNameDefender,
		run: func(ctx *ScanContext) []*models.DefenderResult {
			return (&scanners.DefenderScanner{}).Scan(ctx.Ctx, ctx.Cred, ctx.Subscriptions, ctx.Params.Filters)
		},
		assign: func(rd *renderers.ReportData, r []*models.DefenderResult) { rd.Defender = r },
	}
}

// NewDefenderRecommendationsStage creates the Defender recommendations scan stage.
func NewDefenderRecommendationsStage() Stage {
	return &simpleStage[[]*models.DefenderRecommendation]{
		BaseStage: NewBaseStage("Defender Recommendations Scan", false),
		stageName: models.StageNameDefenderRecommendations,
		run: func(ctx *ScanContext) []*models.DefenderRecommendation {
			return (&scanners.DefenderScanner{}).GetRecommendations(ctx.Ctx, ctx.Cred, ctx.Subscriptions, ctx.Params.Filters)
		},
		assign: func(rd *renderers.ReportData, r []*models.DefenderRecommendation) {
			rd.DefenderRecommendations = r
		},
	}
}
