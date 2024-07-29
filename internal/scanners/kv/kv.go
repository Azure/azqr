// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

// KeyVaultScanner - Scanner for Key Vaults
type KeyVaultScanner struct {
	config       *azqr.ScannerConfig
	vaultsClient *armkeyvault.VaultsClient
}

// Init - Initializes the KeyVaultScanner
func (c *KeyVaultScanner) Init(config *azqr.ScannerConfig) error {
	c.config = config
	var err error
	c.vaultsClient, err = armkeyvault.NewVaultsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Key Vaults in a Resource Group
func (c *KeyVaultScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(c.config.SubscriptionID, resourceGroupName, c.ResourceTypes()[0])

	vaults, err := c.listVaults(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, vault := range vaults {
		rr := engine.EvaluateRecommendations(rules, vault, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *vault.Name,
			Type:             *vault.Type,
			Location:         *vault.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *KeyVaultScanner) listVaults(resourceGroupName string) ([]*armkeyvault.Vault, error) {
	pager := c.vaultsClient.NewListByResourceGroupPager(resourceGroupName, nil)

	vaults := make([]*armkeyvault.Vault, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		vaults = append(vaults, resp.Value...)
	}
	return vaults, nil
}

func (a *KeyVaultScanner) ResourceTypes() []string {
	return []string{"Microsoft.KeyVault/vaults"}
}
