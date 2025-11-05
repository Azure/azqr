// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package amg

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dashboard/armdashboard"
)

func init() {
	models.ScannerList["amg"] = []models.IAzureScanner{&ManagedGrafanaScanner{}}
}

// ManagedGrafanaScanner - Scanner for Managed Grafana
type ManagedGrafanaScanner struct {
	config        *models.ScannerConfig
	grafanaClient *armdashboard.GrafanaClient
}

// Init - Initializes the ManagedGrafanaScanner Scanner
func (a *ManagedGrafanaScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.grafanaClient, _ = armdashboard.NewGrafanaClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Managed Grafana in a Resource Group
func (a *ManagedGrafanaScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
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

func (a *ManagedGrafanaScanner) listWorkspaces() ([]*armdashboard.ManagedGrafana, error) {
	pager := a.grafanaClient.NewListPager(nil)

	workspaces := make([]*armdashboard.ManagedGrafana, 0)
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

func (a *ManagedGrafanaScanner) ResourceTypes() []string {
	return []string{"Microsoft.Dashboard/grafana"}
}
