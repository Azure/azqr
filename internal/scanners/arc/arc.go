// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package arc

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["arc"] = []models.IAzureScanner{&ArcScanner{
		BaseScanner: models.NewBaseScanner("Microsoft.AzureArcData/sqlServerInstances"),
	}}
}

// ArcScanner - Scanner for Arc SQL
type ArcScanner struct {
	models.BaseScanner
}

// Init - Initializes the Arc SQL Scanner
func (c *ArcScanner) Init(config *models.ScannerConfig) error {
	return c.BaseScanner.Init(config)
}

// Scan - Scans all Azure Arc-enabled machines in a Resource Group
func (c *ArcScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.BaseScanner.GetConfig().SubscriptionID, c.ResourceTypes()[0])
	// This scanner doesn't perform actual scans - it's here to register the resource type
	// Actual Arc SQL scanning is done by the graph-based ArcSQLScanner
	return []*models.AzqrServiceResult{}, nil
}
