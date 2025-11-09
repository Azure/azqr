// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package odb

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["odb"] = []models.IAzureScanner{&OracleDatabaseScanner{}}
}

// OracleDatabaseScanner - Scanner for Oracle Database@Azure
type OracleDatabaseScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the Oracle Database Scanner
func (a *OracleDatabaseScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Oracle Database@Azure resources in a Resource Group
func (a *OracleDatabaseScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []*models.AzqrServiceResult{}, nil
}

func (a *OracleDatabaseScanner) ResourceTypes() []string {
	return []string{
		"Oracle.Database/cloudExadataInfrastructures",
		"Oracle.Database/cloudVmClusters",
	}
}

func (a *OracleDatabaseScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
