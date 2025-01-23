// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package iot

import "github.com/Azure/azqr/internal/scanners"

func init() {
	scanners.ScannerList["iot"] = []scanners.IAzureScanner{&IoTHubScanner{}}
}

// IoTHubScanner - Scanner for IoT Hub
type IoTHubScanner struct {
	config *scanners.ScannerConfig
}

// Init - Initializes the IoT Hub Scanner
func (a *IoTHubScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all IoT Hub in a Resource Group
func (a *IoTHubScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []scanners.AzqrServiceResult{}, nil
}

func (a *IoTHubScanner) ResourceTypes() []string {
	return []string{"Microsoft.Devices/IotHubs"}
}

func (a *IoTHubScanner) GetRecommendations() map[string]scanners.AzqrRecommendation {
	return map[string]scanners.AzqrRecommendation{}
}
