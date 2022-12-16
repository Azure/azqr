package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

type KeyVaultAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	vaultsClient        *armkeyvault.VaultsClient
	listVaultsFunc      func(resourceGroupName string) ([]*armkeyvault.Vault, error)
}

func NewKeyVaultAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *KeyVaultAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	vaultsClient, err := armkeyvault.NewVaultsClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := KeyVaultAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		vaultsClient:        vaultsClient,
	}
	return &analyzer
}

func (c KeyVaultAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing Key Vaults in Resource Group %s", resourceGroupName)

	vaults, err := c.listVaults(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, vault := range vaults {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*vault.ID)
		if err != nil {
			return nil, err
		}

		results = append(results, AzureServiceResult{
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionId: c.subscriptionId,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *vault.Name,
				Sku:            string(*vault.Properties.SKU.Name),
				Sla:            "99.99%",
				Type:           *vault.Type,
				Location:       parseLocation(vault.Location),
				CAFNaming:      strings.HasPrefix(*vault.Name, "kv")},
			AvailabilityZones:  true,
			PrivateEndpoints:   len(vault.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c KeyVaultAnalyzer) listVaults(resourceGroupName string) ([]*armkeyvault.Vault, error) {
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
	} else {
		return c.listVaultsFunc(resourceGroupName)
	}
}
