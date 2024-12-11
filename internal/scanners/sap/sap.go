// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sap

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["sap"] = []models.IAzureScanner{&SAPScanner{}}
}

// SAPScanner - Scanner for SAP
type SAPScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the SAP Scanner
func (a *SAPScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all SAP in a Resource Group
func (a *SAPScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []models.AzqrServiceResult{}, nil
}

func (a *SAPScanner) ResourceTypes() []string {
	return []string{"Specialized.Workload/SAP"}
}

func (a *SAPScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
