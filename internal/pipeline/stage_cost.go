// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"sync"

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
	subCount := len(ctx.Subscriptions)
	if subCount == 0 {
		ctx.ReportData.Cost = nil
		return nil
	}

	// Worker pool to limit concurrent cost scanner goroutines
	const numCostWorkers = 10
	workerCount := numCostWorkers
	if subCount < workerCount {
		workerCount = subCount
	}

	jobs := make(chan string, subCount)
	results := make(chan []*models.CostResult, subCount)

	// Start worker pool
	var workerWg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			// Create a new CostScanner per worker to avoid race conditions
			// since CostScanner stores state in struct fields during Scan()
			workerScanner := scanners.CostScanner{}
			for subID := range jobs {
				scannerConfig := &models.ScannerConfig{
					Ctx:            ctx.Ctx,
					Cred:           ctx.Cred,
					ClientOptions:  ctx.ClientOptions,
					SubscriptionID: subID,
				}
				result := workerScanner.Scan(scannerConfig)
				if len(result) > 0 {
					results <- result
				}
			}
		}()
	}

	// Send subscription jobs to workers
	for subID := range ctx.Subscriptions {
		jobs <- subID
	}
	close(jobs)

	// Wait for workers to finish and close results channel
	go func() {
		workerWg.Wait()
		close(results)
	}()

	// Collect results from all workers
	var allCosts []*models.CostResult
	for result := range results {
		allCosts = append(allCosts, result...)
	}

	// Aggregate all cost items into report data
	ctx.ReportData.Cost = allCosts

	return nil
}
