// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afw

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["afw"] = []models.IAzureScanner{NewFirewallScanner()}
}

// NewFirewallScanner - Creates a new Azure Firewall scanner
func NewFirewallScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.AzureFirewall, *armnetwork.AzureFirewallsClient]{
			ResourceTypes: []string{"Microsoft.Network/azureFirewalls", "Microsoft.Network/ipGroups"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.AzureFirewallsClient, error) {
				return armnetwork.NewAzureFirewallsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.AzureFirewallsClient, ctx context.Context) ([]*armnetwork.AzureFirewall, error) {
				pager := client.NewListAllPager(nil)
				services := make([]*armnetwork.AzureFirewall, 0)
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

			ExtractResourceInfo: func(firewall *armnetwork.AzureFirewall) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					firewall.ID,
					firewall.Name,
					firewall.Location,
					firewall.Type,
				)
			},
		},
	)
}
