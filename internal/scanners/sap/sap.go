// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sap

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["sap"] = []scanners.IAzureScanner{&SAPScanner{}}
}

// SAPScanner - Scanner for SAP
type SAPScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the SAP Scanner
func (a *SAPScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all SAP in a Resource Group
func (a *SAPScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *SAPScanner) ResourceTypes() []string {
	return []string{"Specialized.Workload/SAP"}
}

func (a *SAPScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
