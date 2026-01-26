// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package iot

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["iot"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Devices/IotHubs"),
	}
}
