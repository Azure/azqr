// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aa

import (
	"github.com/Azure/azqr/internal/azqr"
)

// AutomationAccountScanner - Scanner for Automation Account
type AutomationAccountScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Automation Account Scanner
func (a *AutomationAccountScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Automation Accounts in a Resource Group
func (a *AutomationAccountScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *AutomationAccountScanner) ResourceTypes() []string {
	return []string{"Microsoft.Automation/automationAccounts"}
}

func (a *AutomationAccountScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
