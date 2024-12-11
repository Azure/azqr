// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hpc

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["hpc"] = []models.IAzureScanner{&HighPerformanceComputingScanner{}}
}

// HighPerformanceComputingScanner - Scanner for HPC
type HighPerformanceComputingScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the HPC Scanner
func (a *HighPerformanceComputingScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all HPC in a Resource Group
func (a *HighPerformanceComputingScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *HighPerformanceComputingScanner) ResourceTypes() []string {
	return []string{"Specialized.Workload/HPC"}
}

func (a *HighPerformanceComputingScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
