package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

type StorageAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionId      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	storageClient       *armstorage.AccountsClient
}

func NewStorageAnalyzer(subscriptionId string, ctx context.Context, cred azcore.TokenCredential) *StorageAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(cred, ctx)
	storageClient, err := armstorage.NewAccountsClient(subscriptionId, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := StorageAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionId:      subscriptionId,
		ctx:                 ctx,
		cred:                cred,
		storageClient:       storageClient,
	}
	return &analyzer
}

func (c StorageAnalyzer) Review(resourceGroupName string) ([]AzureServiceResult, error) {
	log.Printf("Analyzing Storage in Resource Group %s", resourceGroupName)

	storage, err := c.listStorage(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []AzureServiceResult{}
	for _, storage := range storage {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*storage.ID)
		if err != nil {
			return nil, err
		}

		sku := string(*storage.SKU.Name)
		tier := string(*storage.SKU.Tier)
		zones := false
		if strings.Contains(sku, "ZRS") {
			zones = true
		}
		sla := "99.9%"
		if strings.Contains(sku, "RAGRS") || strings.Contains(tier, "Hot") {
			sla = "99.99%"
		}

		results = append(results, AzureServiceResult{
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionId: c.subscriptionId,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *storage.Name,
				Sku:            sku,
				Sla:            sla,
				Type:           *storage.Type,
				Location:       *storage.Location,
				CAFNaming:      strings.HasPrefix(*storage.Name, "st")},
			AvailabilityZones:  zones,
			PrivateEndpoints:   len(storage.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c StorageAnalyzer) listStorage(resourceGroupName string) ([]*armstorage.Account, error) {
	pager := c.storageClient.NewListByResourceGroupPager(resourceGroupName, nil)

	staccounts := make([]*armstorage.Account, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}
		staccounts = append(staccounts, resp.Value...)
	}
	return staccounts, nil
}
