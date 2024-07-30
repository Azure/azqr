// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ng

import (
	"github.com/Azure/azqr/internal/azqr"
)

// NatGatewwayScanner - Scanner for NAT Gateway
type NatGatewwayScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the NAT Gateway Scanner
func (a *NatGatewwayScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all NAT Gateway in a Resource Group
func (a *NatGatewwayScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *NatGatewwayScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/natGateways"}
}

func (a *NatGatewwayScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
