// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package hub

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning"
)

func init() {
	models.ScannerList["hub"] = []models.IAzureScanner{NewAIFoundryHubScanner()}
}

// NewAIFoundryHubScanner - Creates a new AI Foundry Hub scanner
func NewAIFoundryHubScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armmachinelearning.Workspace, *armmachinelearning.WorkspacesClient]{
			ResourceTypes: []string{"Microsoft.MachineLearningServices/workspaces"},

			ClientFactory: func(config *models.ScannerConfig) (*armmachinelearning.WorkspacesClient, error) {
				return armmachinelearning.NewWorkspacesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armmachinelearning.WorkspacesClient, ctx context.Context) ([]*armmachinelearning.Workspace, error) {
				pager := client.NewListBySubscriptionPager(nil)
				resources := make([]*armmachinelearning.Workspace, 0)
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

			ExtractResourceInfo: func(workspace *armmachinelearning.Workspace) models.ResourceInfo {
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
