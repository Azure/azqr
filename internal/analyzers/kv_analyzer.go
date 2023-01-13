package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

// KeyVaultAnalyzer - Analyzer for Key Vaults
type KeyVaultAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	vaultsClient        *armkeyvault.VaultsClient
	listVaultsFunc      func(resourceGroupName string) ([]*armkeyvault.Vault, error)
}

// Init - Initializes the KeyVaultAnalyzer
func (c *KeyVaultAnalyzer) Init(config ServiceAnalizerConfig) error {
	c.subscriptionID = config.SubscriptionID
	c.ctx = config.Ctx
	c.cred = config.Cred
	var err error
	c.vaultsClient, err = armkeyvault.NewVaultsClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config.Ctx, config.Cred)
	if err != nil {
		return err
	}
	return nil
}

// Review - Analyzes all Key Vaults in a Resource Group
func (c *KeyVaultAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Key Vaults in Resource Group %s", resourceGroupName)

	vaults, err := c.listVaults(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, vault := range vaults {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*vault.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     c.subscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *vault.Name,
			SKU:                string(*vault.Properties.SKU.Name),
			SLA:                "99.99%",
			Type:               *vault.Type,
			Location:           *vault.Location,
			CAFNaming:          strings.HasPrefix(*vault.Name, "kv"),
			AvailabilityZones:  true,
			PrivateEndpoints:   len(vault.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c *KeyVaultAnalyzer) listVaults(resourceGroupName string) ([]*armkeyvault.Vault, error) {
	if c.listVaultsFunc == nil {
		pager := c.vaultsClient.NewListByResourceGroupPager(resourceGroupName, nil)

		vaults := make([]*armkeyvault.Vault, 0)
		for pager.More() {
			resp, err := pager.NextPage(c.ctx)
			if err != nil {
				return nil, err
			}
			vaults = append(vaults, resp.Value...)
		}
		return vaults, nil
	}

	return c.listVaultsFunc(resourceGroupName)
}
