// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package srch

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/search/armsearch"
)

func init() {
	models.ScannerList["srch"] = []models.IAzureScanner{NewAISearchScanner()}
}

// NewAISearchScanner - Creates a new Azure AI Search scanner
func NewAISearchScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armsearch.Service, *armsearch.ServicesClient]{
			ResourceTypes: []string{"Microsoft.Search/searchServices"},

			ClientFactory: func(config *models.ScannerConfig) (*armsearch.ServicesClient, error) {
				return armsearch.NewServicesClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armsearch.ServicesClient, ctx context.Context) ([]*armsearch.Service, error) {
				pager := client.NewListBySubscriptionPager(&armsearch.SearchManagementRequestOptions{}, nil)
				workspaces := make([]*armsearch.Service, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					workspaces = append(workspaces, resp.Value...)
				}
				return workspaces, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(service *armsearch.Service) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					service.ID,
					service.Name,
					service.Location,
					service.Type,
				)
			},
		},
	)
}
