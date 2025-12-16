// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package erc

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["erc"] = []models.IAzureScanner{NewExpressRouteScanner()}
}

// NewExpressRouteScanner creates a new ExpressRouteScanner
func NewExpressRouteScanner() *ExpressRouteScanner {
	return &ExpressRouteScanner{
		BaseScanner: models.NewBaseScanner(
			"Microsoft.Network/expressRouteCircuits",
			"Microsoft.Network/ExpressRoutePorts",
		),
	}
}

// ExpressRouteScanner - Scanner for Express Route
type ExpressRouteScanner struct {
	models.BaseScanner
}

// Init - Initializes the Express Route Scanner
func (a *ExpressRouteScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
