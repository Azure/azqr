// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sap

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["sap"] = []models.IAzureScanner{
		models.NewBaseScanner("Specialized.Workload/SAP"),
	}
}
