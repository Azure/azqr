// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kv

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/cmendible/azqr/internal/scanners"
)

// KeyVaultScanner - Scanner for Key Vaults
type KeyVaultScanner struct {
	config              *scanners.ScannerConfig
	diagnosticsSettings scanners.DiagnosticsSettings
	vaultsClient        *armkeyvault.VaultsClient
	listVaultsFunc      func(resourceGroupName string) ([]*armkeyvault.Vault, error)
}

// Init - Initializes the KeyVaultScanner
func (c *KeyVaultScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.vaultsClient, err = armkeyvault.NewVaultsClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = scanners.DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Key Vaults in a Resource Group
func (c *KeyVaultScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning Key Vaults in Resource Group %s", resourceGroupName)

	vaults, err := c.listVaults(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, vault := range vaults {
		rr := engine.EvaluateRules(rules, vault, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *vault.Name,
			Type:           *vault.Type,
			Location:       *vault.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (c *KeyVaultScanner) listVaults(resourceGroupName string) ([]*armkeyvault.Vault, error) {
	if c.listVaultsFunc == nil {
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

	return c.listVaultsFunc(resourceGroupName)
}
