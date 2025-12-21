// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package fdfp

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["fdfp"] = []models.IAzureScanner{&FrontDoorWAFPolicyScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Network/frontdoorWebApplicationFirewallPolicies"),
	}}
}

// FrontDoorWAFPolicyScanner - Scanner for Front Door Web Application Policy
type FrontDoorWAFPolicyScanner struct {
	models.BaseScanner
}

// Init - Initializes the Front Door Web Application Policy Scanner
func (a *FrontDoorWAFPolicyScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
