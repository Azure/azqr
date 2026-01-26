// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nic

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["nic"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Network/networkInterfaces"),
	}
}
