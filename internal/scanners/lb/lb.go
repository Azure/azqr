// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package lb

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["lb"] = []models.IAzureScanner{NewLoadBalancerScanner()}
}

// NewLoadBalancerScanner creates a new Load Balancer scanner using the generic framework
func NewLoadBalancerScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.LoadBalancer, *armnetwork.LoadBalancersClient]{
			ResourceTypes: []string{"Microsoft.Network/loadBalancers"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.LoadBalancersClient, error) {
				return armnetwork.NewLoadBalancersClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.LoadBalancersClient, ctx context.Context) ([]*armnetwork.LoadBalancer, error) {
				pager := client.NewListAllPager(nil)
				lbs := make([]*armnetwork.LoadBalancer, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					lbs = append(lbs, resp.Value...)
				}

				return lbs, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(lb *armnetwork.LoadBalancer) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					lb.ID,
					lb.Name,
					lb.Location,
					lb.Type,
				)
			},
		},
	)
}
