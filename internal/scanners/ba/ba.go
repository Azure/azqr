// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ba

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["ba"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Batch/batchAccounts"),
	}
}
