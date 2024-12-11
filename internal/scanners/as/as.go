// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package as

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/analysisservices/armanalysisservices"
)

func init() {
	models.ScannerList["as"] = []models.IAzureScanner{&AnalysisServicesScanner{}}
}

// AnalysisServicesScanner - Scanner for Analysis Services
type AnalysisServicesScanner struct {
	config *models.ScannerConfig
	client *armanalysisservices.ServersClient
}

// Init - Initializes the AnalysisServicesScanner
func (c *AnalysisServicesScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armanalysisservices.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Analysis Services in a Resource Group
func (c *AnalysisServicesScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	workspaces, err := c.listWorkspaces()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, ws := range workspaces {
		rr := engine.EvaluateRecommendations(rules, ws, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*ws.ID),
			ServiceName:      *ws.Name,
			Type:             *ws.Type,
			Location:         *ws.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *AnalysisServicesScanner) listWorkspaces() ([]*armanalysisservices.Server, error) {
	pager := c.client.NewListPager(nil)

	registries := make([]*armanalysisservices.Server, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		registries = append(registries, resp.Value...)
	}
	return registries, nil
}

func (a *AnalysisServicesScanner) ResourceTypes() []string {
	return []string{"Microsoft.AnalysisServices/servers"}
}
