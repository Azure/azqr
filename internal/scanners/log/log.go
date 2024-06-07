// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package log

import (
	"github.com/Azure/azqr/internal/scanners"
)

// LogAnalyticsScanner - Scanner for Log Analytics workspace
type LogAnalyticsScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Log Analytics workspace Scanner
func (a *LogAnalyticsScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Log Analytics workspace in a Resource Group
func (a *LogAnalyticsScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *LogAnalyticsScanner) ResourceTypes() []string {
	return []string{"Microsoft.OperationalInsights/workspaces"}
}

func (a *LogAnalyticsScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
