// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package log

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights/v2"
)

func init() {
	models.ScannerList["log"] = []models.IAzureScanner{NewLogAnalyticsScanner()}
}

// NewLogAnalyticsScanner - Creates a new Log Analytics scanner
func NewLogAnalyticsScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armoperationalinsights.Workspace, *armoperationalinsights.WorkspacesClient]{
			ResourceTypes: []string{"Microsoft.OperationalInsights/workspaces"},

			ClientFactory: func(config *models.ScannerConfig) (*armoperationalinsights.WorkspacesClient, error) {
				return armoperationalinsights.NewWorkspacesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armoperationalinsights.WorkspacesClient, ctx context.Context) ([]*armoperationalinsights.Workspace, error) {
				pager := client.NewListPager(nil)
				svcs := make([]*armoperationalinsights.Workspace, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					svcs = append(svcs, resp.Value...)
				}
				return svcs, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(workspace *armoperationalinsights.Workspace) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					workspace.ID,
					workspace.Name,
					workspace.Location,
					workspace.Type,
				)
			},
		},
	)
}
