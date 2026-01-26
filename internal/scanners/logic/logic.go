// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package logic

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/logic/armlogic"
)

func init() {
	models.ScannerList["logic"] = []models.IAzureScanner{NewLogicAppScanner()}
}

// NewLogicAppScanner creates a new Logic App Scanner
func NewLogicAppScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armlogic.Workflow, *armlogic.WorkflowsClient]{
			ResourceTypes: []string{"Microsoft.Logic/workflows"},

			ClientFactory: func(config *models.ScannerConfig) (*armlogic.WorkflowsClient, error) {
				return armlogic.NewWorkflowsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
			},

			ListResources: func(client *armlogic.WorkflowsClient, ctx context.Context) ([]*armlogic.Workflow, error) {
				pager := client.NewListBySubscriptionPager(nil)
				logicApps := make([]*armlogic.Workflow, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					logicApps = append(logicApps, resp.Value...)
				}
				return logicApps, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(resource *armlogic.Workflow) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					resource.ID,
					resource.Name,
					resource.Location,
					resource.Type,
				)
			},
		},
	)
}
