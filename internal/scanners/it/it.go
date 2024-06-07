// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package it

import (
	"github.com/Azure/azqr/internal/scanners"
)

// ImageTemplateScanner - Scanner for Image Template
type ImageTemplateScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the Image Template Scanner
func (a *ImageTemplateScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Image Template in a Resource Group
func (a *ImageTemplateScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *ImageTemplateScanner) ResourceTypes() []string {
	return []string{"Microsoft.VirtualMachineImages/imageTemplates"}
}

func (a *ImageTemplateScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
