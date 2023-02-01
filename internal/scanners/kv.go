package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

// KeyVaultScanner - Analyzer for Key Vaults
type KeyVaultScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	vaultsClient        *armkeyvault.VaultsClient
	listVaultsFunc      func(resourceGroupName string) ([]*armkeyvault.Vault, error)
}

// Init - Initializes the KeyVaultScanner
func (c *KeyVaultScanner) Init(config *ScannerConfig) error {
	c.config = config
	var err error
	c.vaultsClient, err = armkeyvault.NewVaultsClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	c.diagnosticsSettings = DiagnosticsSettings{}
	err = c.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan - Scans all Key Vaults in a Resource Group
func (c *KeyVaultScanner) Scan(resourceGroupName string) ([]IAzureServiceResult, error) {
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
			SubscriptionID:     c.config.SubscriptionID,
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
