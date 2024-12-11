// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rg

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["rg"] = []models.IAzureScanner{&ResourceGroupScanner{}}
}

// ResourceGroupScanner - Scanner for Resource Groups
type ResourceGroupScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Resource Groups Scanner
func (a *ResourceGroupScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Resource Groups
func (a *ResourceGroupScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	return []models.AzqrServiceResult{}, nil
}

func (a *ResourceGroupScanner) ResourceTypes() []string {
	return []string{"Microsoft.Resources/resourceGroups"}
}

func (a *ResourceGroupScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
