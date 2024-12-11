// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package erc

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["erc"] = []models.IAzureScanner{&ExpressRouteScanner{}}
}

// ExpressRouteScanner - Scanner for Express Route
type ExpressRouteScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Express Route Scanner
func (a *ExpressRouteScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Express Routes in a Resource Group
func (a *ExpressRouteScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *ExpressRouteScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.Network/expressRouteCircuits",
		"Microsoft.Network/ExpressRoutePorts",
	}
}

func (a *ExpressRouteScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
