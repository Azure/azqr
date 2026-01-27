// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners"
)

// DefenderStatusStage executes Defender status scan.
type DefenderStatusStage struct {
	*BaseStage
}

func NewDefenderStatusStage() *DefenderStatusStage {
	return &DefenderStatusStage{
		BaseStage: NewBaseStage("Defender Status Scan", false),
	}
}

func (s *DefenderStatusStage) Skip(ctx *ScanContext) bool {
	return !ctx.Params.Stages.IsStageEnabled(models.StageNameDefender)
}

func (s *DefenderStatusStage) Execute(ctx *ScanContext) error {
	defenderScanner := scanners.DefenderScanner{}

	ctx.ReportData.Defender = defenderScanner.Scan(
		ctx.Ctx,
		ctx.Cred,
		ctx.Subscriptions,
		ctx.Params.Filters,
	)

	return nil
}
