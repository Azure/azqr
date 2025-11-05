// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pdnsz

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["pdnsz"] = []models.IAzureScanner{&PrivateDNSZoneScanner{}}
}

// PrivateDNSZoneScanner - Scanner for Private DNS Zone
type PrivateDNSZoneScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Private DNS Zone Scanner
func (a *PrivateDNSZoneScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Private DNS Zone in a Resource Group
func (a *PrivateDNSZoneScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []*models.AzqrServiceResult{}, nil
}

func (a *PrivateDNSZoneScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/privateDnsZones"}
}

func (a *PrivateDNSZoneScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}

// TODO: version 6.1.0 of armentowrk does not allow listing per subscription yet.
