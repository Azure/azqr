// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package amg

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dashboard/armdashboard"
)

func init() {
	models.ScannerList["amg"] = []models.IAzureScanner{NewManagedGrafanaScanner()}
}

// NewManagedGrafanaScanner - Returns a new ManagedGrafanaScanner
func NewManagedGrafanaScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armdashboard.ManagedGrafana, *armdashboard.GrafanaClient]{
			ResourceTypes: []string{"Microsoft.Dashboard/grafana"},

			ClientFactory: func(config *models.ScannerConfig) (*armdashboard.GrafanaClient, error) {
				return armdashboard.NewGrafanaClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armdashboard.GrafanaClient, ctx context.Context) ([]*armdashboard.ManagedGrafana, error) {
				pager := client.NewListPager(nil)
				resources := make([]*armdashboard.ManagedGrafana, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					resources = append(resources, resp.Value...)
				}
				return resources, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(grafana *armdashboard.ManagedGrafana) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					grafana.ID,
					grafana.Name,
					grafana.Location,
					grafana.Type,
				)
			},
		},
	)
}
