// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package conn

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["con"] = []models.IAzureScanner{&ConnectionScanner{}}
}

// ConnectionScanner - Scanner for Automation Account
type ConnectionScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Automation Account Scanner
func (a *ConnectionScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Automation Accounts in a Resource Group
func (a *ConnectionScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *ConnectionScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/connections"}
}

func (a *ConnectionScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
