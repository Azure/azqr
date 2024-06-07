// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ng

import (
	"github.com/Azure/azqr/internal/scanners"
)

// NatGatewwayScanner - Scanner for NAT Gateway
type NatGatewwayScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the NAT Gateway Scanner
func (a *NatGatewwayScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all NAT Gateway in a Resource Group
func (a *NatGatewwayScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *NatGatewwayScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/natGateways"}
}

func (a *NatGatewwayScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
