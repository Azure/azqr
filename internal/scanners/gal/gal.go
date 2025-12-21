// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package gal

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["gal"] = []models.IAzureScanner{&GalleryScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Compute/galleries"),
	}}
}

// GalleryScanner - Scanner for Gallery
type GalleryScanner struct {
	models.BaseScanner
}

// Init - Initializes the Gallery Scanner
func (a *GalleryScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
