// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avs

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["avs"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"Microsoft.AVS/privateClouds",
			"Specialized.Workload/AVS",
		),
	}
}
