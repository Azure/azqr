// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners"
)

// AdvisorStage executes Advisor scan.
type AdvisorStage struct {
	*BaseStage
}

func NewAdvisorStage() *AdvisorStage {
	return &AdvisorStage{
		BaseStage: NewBaseStage("Advisor Scan", false),
	}
}

func (s *AdvisorStage) Skip(ctx *ScanContext) bool {
	return !ctx.Params.Stages.IsStageEnabled(models.StageNameAdvisor)
}

func (s *AdvisorStage) Execute(ctx *ScanContext) error {
	advisorScanner := scanners.AdvisorScanner{}

	ctx.ReportData.Advisor = advisorScanner.Scan(
		ctx.Ctx,
		ctx.Cred,
		ctx.Subscriptions,
		ctx.Params.Filters,
	)

	return nil
}
