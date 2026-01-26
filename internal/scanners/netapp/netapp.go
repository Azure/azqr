// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package netapp

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["netapp"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.NetApp/netAppAccounts"),
	}
}
