package scanners

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

// StorageScanner - Analyzer for Storage
type StorageScanner struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	storageClient       *armstorage.AccountsClient
	listStorageFunc     func(resourceGroupName string) ([]*armstorage.Account, error)
}

// Init - Initializes the StorageScanner
func (c *StorageScanner) Init(config ScannerConfig) error {
	c.subscriptionID = config.SubscriptionID
	c.ctx = config.Ctx
	c.cred = config.Cred
	var err error
	c.storageClient, err = armstorage.NewAccountsClient(config.SubscriptionID, config.Cred, nil)
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

// Review - Analyzes all Storage in a Resource Group
func (c *StorageScanner) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
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
			SubscriptionID:     c.subscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *storage.Name,
			SKU:                sku,
			SLA:                sla,
			Type:               *storage.Type,
			Location:           *storage.Location,
			CAFNaming:          strings.HasPrefix(*storage.Name, "st"),
			AvailabilityZones:  zones,
			PrivateEndpoints:   len(storage.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (c *StorageScanner) listStorage(resourceGroupName string) ([]*armstorage.Account, error) {
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
