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
func (a *ManagedGrafanaScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])

	workspaces, err := a.listWorkspaces(resourceGroupName)
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
			ResourceGroup:    resourceGroupName,
			Location:         *g.Location,
			Type:             *g.Type,
			ServiceName:      *g.Name,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *ManagedGrafanaScanner) listWorkspaces(resourceGroupName string) ([]*armdashboard.ManagedGrafana, error) {
	pager := a.grafanaClient.NewListByResourceGroupPager(resourceGroupName, nil)

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
