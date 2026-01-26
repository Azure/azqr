// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

func init() {
	models.ScannerList["kv"] = []models.IAzureScanner{NewKeyVaultScanner()}
}

// NewKeyVaultScanner creates a new Key Vault scanner using the generic framework
func NewKeyVaultScanner() models.IAzureScanner {
	return models.NewGenericScanner(
		models.GenericScannerConfig[armkeyvault.Vault, *armkeyvault.VaultsClient]{
			ResourceTypes: []string{"Microsoft.KeyVault/vaults"},

			ClientFactory: func(config *models.ScannerConfig) (*armkeyvault.VaultsClient, error) {
				return armkeyvault.NewVaultsClient(
					config.SubscriptionID,
					config.Cred,
					config.ClientOptions,
				)
			},

			ListResources: func(client *armkeyvault.VaultsClient, ctx context.Context) ([]*armkeyvault.Vault, error) {
				pager := client.NewListBySubscriptionPager(nil)
				vaults := make([]*armkeyvault.Vault, 0)

				for pager.More() {
					resp, err := pager.NextPage(ctx)
					if err != nil {
						return nil, err
					}
					vaults = append(vaults, resp.Value...)
				}

				return vaults, nil
			},

			GetRecommendations: getRecommendations,

			ExtractResourceInfo: func(vault *armkeyvault.Vault) models.ResourceInfo {
				return models.ExtractStandardARMResourceInfo(
					vault.ID,
					vault.Name,
					vault.Location,
					vault.Type,
				)
			},
		},
	)
}
