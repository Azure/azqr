// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package as

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/analysisservices/armanalysisservices"
)

func init() {
	scanners.ScannerList["as"] = []scanners.IAzureScanner{&AnalysisServicesScanner{}}
}

// AnalysisServicesScanner - Scanner for Analysis Services
type AnalysisServicesScanner struct {
	config *scanners.ScannerConfig
	client *armanalysisservices.ServersClient
}

// Init - Initializes the AnalysisServicesScanner
func (c *AnalysisServicesScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armanalysisservices.NewServersClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Analysis Services in a Resource Group
func (c *AnalysisServicesScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	workspaces, err := c.listWorkspaces()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, ws := range workspaces {
		rr := engine.EvaluateRecommendations(rules, ws, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*ws.ID),
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
