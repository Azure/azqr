// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avs

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["avs"] = []models.IAzureScanner{&AVSScanner{}}
}

// AVSScanner - Scanner for AVS
type AVSScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the AVS Scanner
func (a *AVSScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all AVS in a Resource Group
func (a *AVSScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *AVSScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.AVS/privateClouds",
		"Specialized.Workload/AVS",
	}
}

func (a *AVSScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
