// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package gal

import (
	"github.com/Azure/azqr/internal/azqr"
)

// GalleryScanner - Scanner for Galleries
type GalleryScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Galleries Scanner
func (a *GalleryScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Galleries in a Resource Group
func (a *GalleryScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *GalleryScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/galleries"}
}

func (a *GalleryScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
