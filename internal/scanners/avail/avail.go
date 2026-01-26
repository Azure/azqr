// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avail

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["avail"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Compute/availabilitySets"),
	}
}
