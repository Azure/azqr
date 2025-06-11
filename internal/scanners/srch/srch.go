// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package srch

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/search/armsearch"
)

func init() {
	models.ScannerList["srch"] = []models.IAzureScanner{&AISearchScanner{}}
}

// AISearchScanner - Scanner for Azure AI Search
type AISearchScanner struct {
	config *models.ScannerConfig
	client *armsearch.ServicesClient
}

// Init - Initializes the AISearchScanner Scanner
func (a *AISearchScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.client, _ = armsearch.NewServicesClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Azure AI Search
func (a *AISearchScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	workspaces, err := a.listWorkspaces()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, g := range workspaces {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*g.ID),
			Location:         *g.Location,
			Type:             *g.Type,
			ServiceName:      *g.Name,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *AISearchScanner) listWorkspaces() ([]*armsearch.Service, error) {
	pager := a.client.NewListBySubscriptionPager(&armsearch.SearchManagementRequestOptions{}, nil)

	workspaces := make([]*armsearch.Service, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, resp.Value...)
	}

	return workspaces, nil
}

func (a *AISearchScanner) ResourceTypes() []string {
	return []string{"Microsoft.Search/searchServices"}
}
