// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nic

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["nic"] = []models.IAzureScanner{&NICScanner{}}
}

// NICScanner - Scanner for NICs
type NICScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the NICs Scanner
func (a *NICScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all NICs
func (a *NICScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []*models.AzqrServiceResult{}, nil
}

func (a *NICScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/networkInterfaces"}
}

func (a *NICScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
