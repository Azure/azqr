package analyzers

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

type KeyVaultAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId string
	ctx context.Context
	cred azcore.TokenCredential
}

func NewKeyVaultAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *KeyVaultAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	analyzer := KeyVaultAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
	}
	return &analyzer
}

func (c KeyVaultAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
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
			SubscriptionId:     c.subscriptionId,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *vault.Name,
			Sku:                string(*vault.Properties.SKU.Name),
			Sla:                "99.99%",
			Type:               *vault.Type,
			AvailabilityZones:  true,
			PrivateEndpoints:   len(vault.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
			CAFNaming:          strings.HasPrefix(*vault.Name, "kv"),
		})
	}
	return results, nil
}

func (c KeyVaultAnalyzer) listVaults(resourceGroupName string) ([]*armkeyvault.Vault, error) {
	vaultsClient, err := armkeyvault.NewVaultsClient(c.subscriptionId, c.cred, nil)
	if err != nil {
		return nil, err
	}

	pager := vaultsClient.NewListByResourceGroupPager(resourceGroupName, nil)

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