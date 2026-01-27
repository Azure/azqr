// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog/log"
)

// ArcSQLStage executes Arc-enabled SQL Server scan.
type ArcSQLStage struct {
	*BaseStage
}

func NewArcSQLStage() *ArcSQLStage {
	return &ArcSQLStage{
		BaseStage: NewBaseStage("Arc-enabled SQL Server Scan", false),
	}
}

func (s *ArcSQLStage) Skip(ctx *ScanContext) bool {
	return !ctx.Params.Stages.IsStageEnabled(models.StageNameArc)
}

func (s *ArcSQLStage) Execute(ctx *ScanContext) error {
	arcSQLScanner := scanners.ArcSQLScanner{}

	ctx.ReportData.ArcSQL = arcSQLScanner.Scan(
		ctx.Ctx,
		ctx.Cred,
		ctx.Subscriptions,
		ctx.Params.Filters,
	)

	log.Info().Msg("Arc-enabled SQL Server scan completed")

	return nil
}
