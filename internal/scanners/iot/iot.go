// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package iot

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["iot"] = []models.IAzureScanner{&IoTHubScanner{}}
}

// IoTHubScanner - Scanner for IoT Hub
type IoTHubScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the IoT Hub Scanner
func (a *IoTHubScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	return nil
}

// Scan - Scans all IoT Hub in a Resource Group
func (a *IoTHubScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])
	return []*models.AzqrServiceResult{}, nil
}

func (a *IoTHubScanner) ResourceTypes() []string {
	return []string{"Microsoft.Devices/IotHubs"}
}

func (a *IoTHubScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
