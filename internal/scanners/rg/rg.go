// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rg

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["rg"] = []models.IAzureScanner{&ResourceGroupScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Resources/resourceGroups"),
	}}
}

// ResourceGroupScanner - Scanner for Resource Groups
type ResourceGroupScanner struct {
	models.BaseScanner
}

// Init - Initializes the Resource Groups Scanner
func (a *ResourceGroupScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
