// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package apim

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

func init() {
	models.ScannerList["apim"] = []models.IAzureScanner{NewAPIManagementScanner()}
}

// NewAPIManagementScanner creates a new API Management scanner using the generic framework
func NewAPIManagementScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armapimanagement.ServiceResource, *armapimanagement.ServiceClient]{
			ResourceTypes: []string{"Microsoft.ApiManagement/service"},

			ClientFactory: func(config *models.ScannerConfig) (*armapimanagement.ServiceClient, error) {
				return armapimanagement.NewServiceClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armapimanagement.ServiceClient, ctx context.Context) ([]*armapimanagement.ServiceResource, error) {
				pager := client.NewListPager(nil)
				services := make([]*armapimanagement.ServiceResource, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					services = append(services, resp.Value...)
				}

				return services, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(svc *armapimanagement.ServiceResource) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					svc.ID,
					svc.Name,
					svc.Location,
					svc.Type,
				)
			},
		},
	)
}
