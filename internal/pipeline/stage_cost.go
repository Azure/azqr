// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners"
)

// CostStage executes Cost analysis scan.
type CostStage struct {
	*BaseStage
}

func NewCostStage() *CostStage {
	return &CostStage{
		BaseStage: NewBaseStage("Cost Analysis Scan", false),
	}
}

func (s *CostStage) Skip(ctx *ScanContext) bool {
	return !ctx.Params.Stages.IsStageEnabled(models.StageNameCost)
}

func (s *CostStage) Execute(ctx *ScanContext) error {
	costScanner := scanners.CostScanner{}

	// Get cost stage options and extract previousMonth flag (default false)
	costOpts := ctx.Params.Stages.GetStageOptions(models.StageNameCost)
	previousMonth := false
	if costOpts != nil {
		if val, ok := costOpts["previousMonth"]; ok {
			if b, ok := val.(bool); ok {
				previousMonth = b
			}
		}
	}

	// Scan costs for all subscriptions
	var allCosts []*models.CostResult
	for subid := range ctx.Subscriptions {
		scannerConfig := &models.ScannerConfig{
			Ctx:            ctx.Ctx,
			Cred:           ctx.Cred,
			ClientOptions:  ctx.ClientOptions,
			SubscriptionID: subid,
		}
		result := costScanner.Scan(scannerConfig, previousMonth)
		if len(result) > 0 {
			allCosts = append(allCosts, result...)
		}
	}

	// Aggregate all cost items into report data
	ctx.ReportData.Cost = allCosts

	return nil
}
