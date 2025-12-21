// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package avs

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["avs"] = []models.IAzureScanner{&AVSScanner{
		BaseScanner: models.NewBaseScanner(
			"Microsoft.AVS/privateClouds",
			"Specialized.Workload/AVS",
		),
	}}
}

// AVSScanner - Scanner for AVS
type AVSScanner struct {
	models.BaseScanner
}

// Init - Initializes the AVS Scanner
func (a *AVSScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
