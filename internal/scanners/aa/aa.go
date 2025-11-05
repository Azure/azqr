// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aa

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["aa"] = []models.IAzureScanner{&AutomationAccountScanner{}}
}

// AutomationAccountScanner - Scanner for Automation Account
type AutomationAccountScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Automation Account Scanner
func (a *AutomationAccountScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Automation Accounts in a Resource Group
func (a *AutomationAccountScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []*models.AzqrServiceResult{}, nil
}

func (a *AutomationAccountScanner) ResourceTypes() []string {
	return []string{"Microsoft.Automation/automationAccounts"}
}

func (a *AutomationAccountScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
