// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hpc

import (
	"github.com/Azure/azqr/internal/azqr"
)

// HighPerformanceComputingScanner - Scanner for HPC
type HighPerformanceComputingScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the HPC Scanner
func (a *HighPerformanceComputingScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all HPC in a Resource Group
func (a *HighPerformanceComputingScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *HighPerformanceComputingScanner) ResourceTypes() []string {
	return []string{"Specialized.Workload/HPC"}
}

func (a *HighPerformanceComputingScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
