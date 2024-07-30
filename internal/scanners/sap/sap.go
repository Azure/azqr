// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sap

import (
	"github.com/Azure/azqr/internal/azqr"
)

// SAPScanner - Scanner for SAP
type SAPScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the SAP Scanner
func (a *SAPScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all SAP in a Resource Group
func (a *SAPScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *SAPScanner) ResourceTypes() []string {
	return []string{"Specialized.Workload/SAP"}
}

func (a *SAPScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
