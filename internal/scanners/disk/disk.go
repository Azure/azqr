// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package disk

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["disk"] = []models.IAzureScanner{NewDiskScanner()}
}

// NewDiskScanner creates a new DiskScanner
func NewDiskScanner() *DiskScanner {
	return &DiskScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Compute/disks"),
	}
}

// DiskScanner - Scanner for Disk
type DiskScanner struct {
	models.BaseScanner
}

// Init - Initializes the Disk Scanner
func (a *DiskScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
