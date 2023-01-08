package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

// StorageAnalyzer - Analyzer for Storage
type StorageAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	storageClient       *armstorage.AccountsClient
	listStorageFunc     func(resourceGroupName string) ([]*armstorage.Account, error)
}

// NewStorageAnalyzer - Creates a new StorageAnalyzer
func NewStorageAnalyzer(ctx context.Context, subscriptionID string, cred azcore.TokenCredential) *StorageAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(ctx, cred)
	storageClient, err := armstorage.NewAccountsClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	analyzer := StorageAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionID:      subscriptionID,
		ctx:                 ctx,
		cred:                cred,
		storageClient:       storageClient,
	}
	return &analyzer
}

// Review - Analyzes all Storage in a Resource Group
func (c StorageAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing Storage in Resource Group %s", resourceGroupName)

	storage, err := c.listStorage(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, storage := range storage {
		hasDiagnostics, err := c.diagnosticsSettings.HasDiagnostics(*storage.ID)
		if err != nil {
			return nil, err
		}

		sku := string(*storage.SKU.Name)
		tier := string(*storage.Properties.AccessTier)
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
				SubscriptionID: c.subscriptionID,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *storage.Name,
				SKU:            sku,
				SLA:            sla,
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
	if c.listStorageFunc == nil {
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

	return c.listStorageFunc(resourceGroupName)
}
