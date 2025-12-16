// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nic

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["nic"] = []models.IAzureScanner{NewNICScanner()}
}

// NewNICScanner creates a new NICScanner
func NewNICScanner() *NICScanner {
	return &NICScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Network/networkInterfaces"),
	}
}

// NICScanner - Scanner for NIC
type NICScanner struct {
	models.BaseScanner
}

// Init - Initializes the NIC Scanner
func (a *NICScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
