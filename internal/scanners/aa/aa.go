// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aa

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["aa"] = []models.IAzureScanner{&AutomationAccountScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Automation/automationAccounts"),
	}}
}

// AutomationAccountScanner - Scanner for Automation Account
type AutomationAccountScanner struct {
	models.BaseScanner
}

// Init - Initializes the Automation Account Scanner
func (a *AutomationAccountScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
