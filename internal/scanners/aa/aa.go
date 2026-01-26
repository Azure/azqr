// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aa

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["aa"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Automation/automationAccounts"),
	}
}
