// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package conn

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["conn"] = []models.IAzureScanner{&ConnectionScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.Network/connections"),
	}}
}

// ConnectionScanner - Scanner for Connection
type ConnectionScanner struct {
	models.BaseScanner
}

// Init - Initializes the Connection Scanner
func (a *ConnectionScanner) Init(config *models.ScannerConfig) error {
	return a.BaseScanner.Init(config)
}
