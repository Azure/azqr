// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package gal

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["gal"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Compute/galleries"),
	}
}
