// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package gal

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["gal"] = []scanners.IAzureScanner{&GalleryScanner{}}
}

// GalleryScanner - Scanner for Galleries
type GalleryScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Galleries Scanner
func (a *GalleryScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Galleries in a Resource Group
func (a *GalleryScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *GalleryScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/galleries"}
}

func (a *GalleryScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
