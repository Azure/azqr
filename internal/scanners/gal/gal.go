// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package gal

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["gal"] = []models.IAzureScanner{&GalleryScanner{}}
}

// GalleryScanner - Scanner for Galleries
type GalleryScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Galleries Scanner
func (a *GalleryScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Galleries in a Resource Group
func (a *GalleryScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []*models.AzqrServiceResult{}, nil
}

func (a *GalleryScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/galleries"}
}

func (a *GalleryScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
