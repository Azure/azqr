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

	// Scan costs for all subscriptions
	var allCostItems []*models.CostResultItem
	for subid := range ctx.Subscriptions {
		scannerConfig := &models.ScannerConfig{
			Ctx:            ctx.Ctx,
			Cred:           ctx.Cred,
			ClientOptions:  ctx.ClientOptions,
			SubscriptionID: subid,
		}
		result := costScanner.Scan(scannerConfig)
		if result != nil && result.Items != nil {
			allCostItems = append(allCostItems, result.Items...)
		}
	}

	// Aggregate all cost items into report data
	ctx.ReportData.Cost = &models.CostResult{
		Items: allCostItems,
	}

	return nil
}
