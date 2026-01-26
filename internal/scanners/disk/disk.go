// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package disk

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["disk"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Compute/disks"),
	}
}
