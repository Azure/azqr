// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package iot

import (
	"github.com/Azure/azqr/internal/azqr"
)

// IoTHubScanner - Scanner for IoT Hub
type IoTHubScanner struct {
	config *azqr.ScannerConfig
}

// Init - Initializes the IoT Hub Scanner
func (a *IoTHubScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all IoT Hub in a Resource Group
func (a *IoTHubScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []azqr.AzqrServiceResult{}, nil
}

func (a *IoTHubScanner) ResourceTypes() []string {
	return []string{"Microsoft.Devices/IotHubs"}
}

func (a *IoTHubScanner) GetRecommendations() map[string]azqr.AzqrRecommendation {
	return map[string]azqr.AzqrRecommendation{}
}
