// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avs

import (
	"github.com/Azure/azqr/internal/azqr"
)

// AVSScanner - Scanner for AVS
type AVSScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the AVS Scanner
func (a *AVSScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all AVS in a Resource Group
func (a *AVSScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *AVSScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.AVS/privateClouds",
		"Specialized.Workload/AVS",
	}
}

func (a *AVSScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
