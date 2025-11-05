// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hub

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
)

func init() {
	models.ScannerList["hub"] = []models.IAzureScanner{&AIFoundryHubScanner{}}
}

// AIFoundryHubScanner - Scanner for Azure AI Foundry Hubs
type AIFoundryHubScanner struct {
	config *models.ScannerConfig
	client *armmachinelearning.WorkspacesClient
}

// Init - Initializes the AIFoundryScanner Scanner
func (a *AIFoundryHubScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.client, _ = armmachinelearning.NewWorkspacesClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all AI Foundry Hubs
func (a *AIFoundryHubScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	workspaces, err := a.listWorkspaces()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, g := range workspaces {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, &models.AzqrServiceResult{
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

func (a *AIFoundryHubScanner) listWorkspaces() ([]*armmachinelearning.Workspace, error) {
	pager := a.client.NewListBySubscriptionPager(nil)

	workspaces := make([]*armmachinelearning.Workspace, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		_ = throttling.WaitARM(a.config.Ctx); // nolint:errcheck
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, resp.Value...)
	}

	return workspaces, nil
}

func (a *AIFoundryHubScanner) ResourceTypes() []string {
	return []string{"Microsoft.MachineLearningServices/workspaces"}
}
