// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package iot

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["iot"] = []models.IAzureScanner{NewIoTHubScanner()}
}

// NewIoTHubScanner creates a new IoTHubScanner
func NewIoTHubScanner() *IoTHubScanner {
	return &IoTHubScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Devices/IotHubs"),
	}
}

// IoTHubScanner - Scanner for IoT Hub
type IoTHubScanner struct {
	models.BaseScanner
}

// Init - Initializes the IoT Hub Scanner
func (a *IoTHubScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
