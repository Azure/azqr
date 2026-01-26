// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pdnsz

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["pdnsz"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Network/privateDnsZones"),
	}
}

// TODO: version 6.1.0 of armentowrk does not allow listing per subscription yet.
