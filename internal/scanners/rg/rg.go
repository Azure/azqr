// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rg

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["rg"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Resources/resourceGroups"),
	}
}
