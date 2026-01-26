// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avd

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["avd"] = []models.IAzureScanner{
		models.NewBaseScanner("Specialized.Workload/AVD"),
	}
}
