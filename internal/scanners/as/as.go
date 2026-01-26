// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package as

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/analysisservices/armanalysisservices"
)

func init() {
	models.ScannerList["as"] = []models.IAzureScanner{NewAnalysisServicesScanner()}
}

// NewAnalysisServicesScanner - Creates a new Analysis Services scanner
func NewAnalysisServicesScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armanalysisservices.Server, *armanalysisservices.ServersClient]{
			ResourceTypes: []string{"Microsoft.AnalysisServices/servers"},

			ClientFactory: func(config *models.ScannerConfig) (*armanalysisservices.ServersClient, error) {
				return armanalysisservices.NewServersClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armanalysisservices.ServersClient, ctx context.Context) ([]*armanalysisservices.Server, error) {
				pager := client.NewListPager(nil)
				resources := make([]*armanalysisservices.Server, 0)
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

			ExtractResourceInfo: func(server *armanalysisservices.Server) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					server.ID,
					server.Name,
					server.Location,
					server.Type,
				)
			},
		},
	)
}
