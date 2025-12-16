// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avail

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["avail"] = []models.IAzureScanner{NewAvailabilitySetScanner()}
}

// NewAvailabilitySetScanner creates a new AvailabilitySetScanner
func NewAvailabilitySetScanner() *AvailabilitySetScanner {
	return &AvailabilitySetScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Compute/availabilitySets"),
	}
}

// AvailabilitySetScanner - Scanner for Availability Set
type AvailabilitySetScanner struct {
	models.BaseScanner
}

// Init - Initializes the Availability Set Scanner
func (a *AvailabilitySetScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
