// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

func init() {
	scanners.ScannerList["kv"] = []scanners.IAzureScanner{&KeyVaultScanner{}}
}

// KeyVaultScanner - Scanner for Key Vaults
type KeyVaultScanner struct {
	config       *scanners.ScannerConfig
	vaultsClient *armkeyvault.VaultsClient
}

// Init - Initializes the KeyVaultScanner
func (c *KeyVaultScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.vaultsClient, err = armkeyvault.NewVaultsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Key Vaults in a Resource Group
func (c *KeyVaultScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	vaults, err := c.listVaults()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, vault := range vaults {
		rr := engine.EvaluateRecommendations(rules, vault, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*vault.ID),
			ServiceName:      *vault.Name,
			Type:             *vault.Type,
			Location:         *vault.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *KeyVaultScanner) listVaults() ([]*armkeyvault.Vault, error) {
	pager := c.vaultsClient.NewListBySubscriptionPager(nil)

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
