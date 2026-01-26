// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aif

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices/v2"
)

func init() {
	models.ScannerList["aif"] = []models.IAzureScanner{NewAIFoundryScanner()}
}

// NewAIFoundryScanner creates a new AIFoundry Scanner
func NewAIFoundryScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armcognitiveservices.Account, *armcognitiveservices.AccountsClient]{
			ResourceTypes: []string{"Microsoft.CognitiveServices/accounts"},

			ClientFactory: func(config *models.ScannerConfig) (*armcognitiveservices.AccountsClient, error) {
				return armcognitiveservices.NewAccountsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
			},

			ListResources: func(client *armcognitiveservices.AccountsClient, ctx context.Context) ([]*armcognitiveservices.Account, error) {
				pager := client.NewListPager(nil)
				namespaces := make([]*armcognitiveservices.Account, 0)
				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					namespaces = append(namespaces, resp.Value...)
				}
				return namespaces, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(resource *armcognitiveservices.Account) models.ResourceInfo {
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
