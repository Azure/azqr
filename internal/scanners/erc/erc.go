// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package erc

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["erc"] = []models.IAzureScanner{
		models.NewBaseScanner(
			"Microsoft.Network/expressRouteCircuits",
			"Microsoft.Network/ExpressRoutePorts",
		),
	}
}
