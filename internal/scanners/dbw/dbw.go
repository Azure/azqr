// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package dbw

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/databricks/armdatabricks"
)

func init() {
	models.ScannerList["dbw"] = []models.IAzureScanner{NewDatabricksScanner()}
}

// NewDatabricksScanner - Creates a new Databricks scanner
func NewDatabricksScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armdatabricks.Workspace, *armdatabricks.WorkspacesClient]{
			ResourceTypes: []string{"Microsoft.Databricks/workspaces"},

			ClientFactory: func(config *models.ScannerConfig) (*armdatabricks.WorkspacesClient, error) {
				return armdatabricks.NewWorkspacesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armdatabricks.WorkspacesClient, ctx context.Context) ([]*armdatabricks.Workspace, error) {
				pager := client.NewListBySubscriptionPager(nil)
				resources := make([]*armdatabricks.Workspace, 0)
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

			ExtractResourceInfo: func(workspace *armdatabricks.Workspace) models.ResourceInfo {
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
