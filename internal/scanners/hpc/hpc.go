// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hpc

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["hpc"] = []models.IAzureScanner{
		models.NewBaseScanner("Specialized.Workload/HPC"),
	}
}
