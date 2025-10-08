// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package arc

import (
	"github.com/Azure/azqr/internal/models"
)

func init() {
	models.ScannerList["arc"] = []models.IAzureScanner{&ArcScanner{}}
}

// ArcScanner - Scanner for Azure Arc-enabled machines
type ArcScanner struct {
	config *models.ScannerConfig
}

// Init - Initializes the ArcScanner
func (c *ArcScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	return nil
}

// Scan - Scans all Azure Arc-enabled machines in a Resource Group
func (c *ArcScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])
	// This scanner doesn't perform actual scans - it's here to register the resource type
	// Actual Arc SQL scanning is done by the graph-based ArcSQLScanner
	return []models.AzqrServiceResult{}, nil
}

// ResourceTypes - Returns the resource types that this scanner covers
func (a *ArcScanner) ResourceTypes() []string {
	return []string{"Microsoft.AzureArcData/sqlServerInstances"}
}

// GetRecommendations - Returns empty recommendations as this scanner is only for resource type registration
func (a *ArcScanner) GetRecommendations() map[string]models.AzqrRecommendation {
	return map[string]models.AzqrRecommendation{}
}
