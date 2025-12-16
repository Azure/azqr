// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package odb

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["odb"] = []models.IAzureScanner{NewOracleDatabaseScanner()}
}

// NewOracleDatabaseScanner creates a new OracleDatabaseScanner
func NewOracleDatabaseScanner() *OracleDatabaseScanner {
	return &OracleDatabaseScanner{
		BaseScanner: models.NewBaseScanner(
			"Oracle.Database/cloudExadataInfrastructures",
			"Oracle.Database/cloudVmClusters",
		),
	}
}

// OracleDatabaseScanner - Scanner for Oracle Database@Azure
type OracleDatabaseScanner struct {
	models.BaseScanner
}

// Init - Initializes the Oracle Database Scanner
func (a *OracleDatabaseScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
