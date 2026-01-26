// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nsg

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["nsg"] = []models.IAzureScanner{NewNSGScanner()}
}

// NewNSGScanner creates a new Network Security Group scanner using the generic framework
func NewNSGScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armnetwork.SecurityGroup, *armnetwork.SecurityGroupsClient]{
			ResourceTypes: []string{"Microsoft.Network/networkSecurityGroups"},

			ClientFactory: func(config *models.ScannerConfig) (*armnetwork.SecurityGroupsClient, error) {
				return armnetwork.NewSecurityGroupsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armnetwork.SecurityGroupsClient, ctx context.Context) ([]*armnetwork.SecurityGroup, error) {
				pager := client.NewListAllPager(nil)
				svcs := make([]*armnetwork.SecurityGroup, 0)

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

			ExtractResourceInfo: func(nsg *armnetwork.SecurityGroup) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					nsg.ID,
					nsg.Name,
					nsg.Location,
					nsg.Type,
				)
			},
		},
	)
}
