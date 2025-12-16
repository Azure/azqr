// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package netapp

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["netapp"] = []models.IAzureScanner{NewNetAppScanner()}
}

// NewNetAppScanner creates a new NetAppScanner
func NewNetAppScanner() *NetAppScanner {
	return &NetAppScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.NetApp/netAppAccounts"),
	}
}

// NetAppScanner - Scanner for NetApp
type NetAppScanner struct {
	models.BaseScanner
}

// Init - Initializes the NetApp Scanner
func (a *NetAppScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
