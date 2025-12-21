// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pdnsz

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["pdnsz"] = []models.IAzureScanner{&PrivateDNSZoneScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Network/privateDnsZones"),
	}}
}

// PrivateDNSZoneScanner - Scanner for Private DNS Zone
type PrivateDNSZoneScanner struct {
	models.BaseScanner
}

// Init - Initializes the Private DNS Zone Scanner
func (a *PrivateDNSZoneScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}

// TODO: version 6.1.0 of armentowrk does not allow listing per subscription yet.
