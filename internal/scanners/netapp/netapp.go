// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package netapp

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["netapp"] = []models.IAzureScanner{&NetAppScanner{}}
}

// NetAppScanner - Scanner for NetApp
type NetAppScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the NetApp Scanner
func (a *NetAppScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all NetApp in a Resource Group
func (a *NetAppScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []*models.AzqrServiceResult{}, nil
}

func (a *NetAppScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.NetApp/netAppAccounts",
	}
}

func (a *NetAppScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
