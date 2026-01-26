// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rsv

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["rsv"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.RecoveryServices/vaults"),
	}
}
