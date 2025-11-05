// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

func init() {
	models.ScannerList["kv"] = []models.IAzureScanner{&KeyVaultScanner{}}
}

// KeyVaultScanner - Scanner for Key Vaults
type KeyVaultScanner struct {
	config       *models.ScannerConfig
	vaultsClient *armkeyvault.VaultsClient
}

// Init - Initializes the KeyVaultScanner
func (c *KeyVaultScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.vaultsClient, err = armkeyvault.NewVaultsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Key Vaults in a Resource Group
func (c *KeyVaultScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	vaults, err := c.listVaults()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, vault := range vaults {
		rr := engine.EvaluateRecommendations(rules, vault, scanContext)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*vault.ID),
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
		// Wait for a token from the burstLimiter channel before making the request
		_ = throttling.WaitARM(c.config.Ctx); // nolint:errcheck
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
