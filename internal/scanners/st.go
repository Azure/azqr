package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

// StorageScanner - Scanner for Storage
type StorageScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	storageClient       *armstorage.AccountsClient
	listStorageFunc     func(resourceGroupName string) ([]*armstorage.Account, error)
}

// Init - Initializes the StorageScanner
func (c *StorageScanner) Init(config *ScannerConfig) error {
	c.config = config
	var err error
	c.storageClient, err = armstorage.NewAccountsClient(config.SubscriptionID, config.Cred, nil)
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

// Scan - Scans all Storage in a Resource Group
func (c *StorageScanner) Scan(resourceGroupName string, scanContext *ScanContext) ([]IAzureServiceResult, error) {
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
		tier := ""
		if storage.Properties != nil {
			if storage.Properties.AccessTier != nil {
				tier = string(*storage.Properties.AccessTier)
			}
		}
		zones := false
		if strings.Contains(sku, "ZRS") {
			zones = true
		}
		sla := "99.9%"
		if strings.Contains(sku, "RAGRS") && strings.Contains(tier, "Hot") {
			sla = "99.99%"
		} else if strings.Contains(sku, "RAGRS") && !strings.Contains(tier, "Hot") {
			sla = "99.9%"
		} else if !strings.Contains(tier, "Hot") {
			sla = "99%"
		}

		results = append(results, AzureServiceResult{
			SubscriptionID:     c.config.SubscriptionID,
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
			resp, err := pager.NextPage(c.config.Ctx)
			if err != nil {
				return nil, err
			}
			staccounts = append(staccounts, resp.Value...)
		}
		return staccounts, nil
	}

	return c.listStorageFunc(resourceGroupName)
}
