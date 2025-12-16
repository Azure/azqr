// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sap

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["sap"] = []models.IAzureScanner{NewSAPScanner()}
}

// NewSAPScanner creates a new SAPScanner
func NewSAPScanner() *SAPScanner {
	return &SAPScanner{
		BaseScanner: models.NewBaseScanner("Specialized.Workload/SAP"),
	}
}

// SAPScanner - Scanner for SAP
type SAPScanner struct {
	models.BaseScanner
}

// Init - Initializes the SAP Scanner
func (a *SAPScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
