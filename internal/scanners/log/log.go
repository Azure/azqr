// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package log

import (
	"github.com/Azure/azqr/internal/azqr"
)

// LogAnalyticsScanner - Scanner for Log Analytics workspace
type LogAnalyticsScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Log Analytics workspace Scanner
func (a *LogAnalyticsScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Log Analytics workspace in a Resource Group
func (a *LogAnalyticsScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *LogAnalyticsScanner) ResourceTypes() []string {
	return []string{"Microsoft.OperationalInsights/workspaces"}
}

func (a *LogAnalyticsScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
