// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hpc

import (
	"github.com/Azure/azqr/internal/scanners"
)

// HighPerformanceComputingScanner - Scanner for HPC
type HighPerformanceComputingScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the HPC Scanner
func (a *HighPerformanceComputingScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all HPC in a Resource Group
func (a *HighPerformanceComputingScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *HighPerformanceComputingScanner) ResourceTypes() []string {
	return []string{"Specialized.Workload/HPC"}
}

func (a *HighPerformanceComputingScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}