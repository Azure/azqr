// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avail

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["avail"] = []models.IAzureScanner{&AvailabilitySetScanner{}}
}

// AvailabilitySetScanner - Scanner for Availability Sets
type AvailabilitySetScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Availability Sets Scanner
func (a *AvailabilitySetScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Availability Sets in a Resource Group
func (a *AvailabilitySetScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *AvailabilitySetScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/availabilitySets"}
}

func (a *AvailabilitySetScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
