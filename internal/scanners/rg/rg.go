// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rg

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["rg"] = []scanners.IAzureScanner{&ResourceGroupScanner{}}
}

// ResourceGroupScanner - Scanner for Resource Groups
type ResourceGroupScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Resource Groups Scanner
func (a *ResourceGroupScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Resource Groups
func (a *ResourceGroupScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	return []scanners.AzqrServiceResult{}, nil
}

func (a *ResourceGroupScanner) ResourceTypes() []string {
	return []string{"Microsoft.Resources/resourceGroups"}
}

func (a *ResourceGroupScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
