// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aa

import (
	"github.com/Azure/azqr/internal/scanners"
)

func init() {
	scanners.ScannerList["aa"] = []scanners.IAzureScanner{&AutomationAccountScanner{}}
}

// AutomationAccountScanner - Scanner for Automation Account
type AutomationAccountScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Automation Account Scanner
func (a *AutomationAccountScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Automation Accounts in a Resource Group
func (a *AutomationAccountScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *AutomationAccountScanner) ResourceTypes() []string {
	return []string{"Microsoft.Automation/automationAccounts"}
}

func (a *AutomationAccountScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
