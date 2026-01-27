// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog/log"
)

// DefenderRecommendationsStage executes Defender recommendations scan.
type DefenderRecommendationsStage struct {
	*BaseStage
}

func NewDefenderRecommendationsStage() *DefenderRecommendationsStage {
	return &DefenderRecommendationsStage{
		BaseStage: NewBaseStage("Defender Recommendations Scan", false),
	}
}

func (s *DefenderRecommendationsStage) Skip(ctx *ScanContext) bool {
	return !ctx.Params.Stages.IsStageEnabled(models.StageNameDefenderRecommendations)
}

func (s *DefenderRecommendationsStage) Execute(ctx *ScanContext) error {
	defenderScanner := scanners.DefenderScanner{}

	ctx.ReportData.DefenderRecommendations = defenderScanner.GetRecommendations(
		ctx.Ctx,
		ctx.Cred,
		ctx.Subscriptions,
		ctx.Params.Filters,
	)

	log.Info().Msg("Defender recommendations scan completed")

	return nil
}
