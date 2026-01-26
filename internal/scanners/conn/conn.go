// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package conn

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["conn"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Network/connections"),
	}
}
