// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package odb

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["odb"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"Oracle.Database/cloudExadataInfrastructures",
			"Oracle.Database/cloudVmClusters",
		),
	}
}
