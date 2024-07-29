// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package it

import (
	"github.com/Azure/azqr/internal/azqr"
)

// ImageTemplateScanner - Scanner for Image Template
type ImageTemplateScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the Image Template Scanner
func (a *ImageTemplateScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all Image Template in a Resource Group
func (a *ImageTemplateScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *ImageTemplateScanner) ResourceTypes() []string {
	return []string{"Microsoft.VirtualMachineImages/imageTemplates"}
}

func (a *ImageTemplateScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
