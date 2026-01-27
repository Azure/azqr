// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners"
	"github.com/rs/zerolog/log"
)

// AzurePolicyStage executes Azure Policy scan.
type AzurePolicyStage struct {
	*BaseStage
}

func NewAzurePolicyStage() *AzurePolicyStage {
	return &AzurePolicyStage{
		BaseStage: NewBaseStage("Azure Policy Scan", false),
	}
}

func (s *AzurePolicyStage) Skip(ctx *ScanContext) bool {
	return !ctx.Params.Stages.IsStageEnabled(models.StageNamePolicy)
}

func (s *AzurePolicyStage) Execute(ctx *ScanContext) error {
	azurePolicyScanner := scanners.AzurePolicyScanner{}

	ctx.ReportData.AzurePolicy = azurePolicyScanner.Scan(
		ctx.Ctx,
		ctx.Cred,
		ctx.Subscriptions,
		ctx.Params.Filters,
	)

	log.Info().Msg("Azure Policy scan completed")

	return nil
}
