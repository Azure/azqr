// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package amg

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dashboard/armdashboard"
)

// ManagedGrafanaScanner - Scanner for Managed Grafana
type ManagedGrafanaScanner struct {
	config        *azqr.ScannerConfig
	grafanaClient *armdashboard.GrafanaClient
}

// Init - Initializes the ManagedGrafanaScanner Scanner
func (a *ManagedGrafanaScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.grafanaClient, _ = armdashboard.NewGrafanaClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Managed Grafana in a Resource Group
func (a *ManagedGrafanaScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	workspaces, err := a.listWorkspaces()
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, g := range workspaces {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    azqr.GetResourceGroupFromResourceID(*g.ID),
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
