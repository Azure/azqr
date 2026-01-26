// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package fdfp

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["fdfp"] = []models.IAzureScanner{
		models.NewBaseScanner("Microsoft.Network/frontdoorWebApplicationFirewallPolicies"),
	}
}
