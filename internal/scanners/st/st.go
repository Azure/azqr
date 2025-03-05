// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package st

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

func init() {
	scanners.ScannerList["st"] = []scanners.IAzureScanner{&StorageScanner{}}
}

// StorageScanner - Scanner for Storage
type StorageScanner struct {
	config             *scanners.ScannerConfig
	storageClient      *armstorage.AccountsClient
	blobServicesClient *armstorage.BlobServicesClient
}

// Init - Initializes the StorageScanner
func (c *StorageScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.storageClient, err = armstorage.NewAccountsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	if err != nil {
		return err
	}
	c.blobServicesClient, err = armstorage.NewBlobServicesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Storage in a Resource Group
func (c *StorageScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	storage, err := c.listStorage()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, storage := range storage {
		resourceGroupName := scanners.GetResourceGroupFromResourceID(*storage.ID)

		scanContext.BlobServiceProperties = nil
		blobServicesProperties, err := c.blobServicesClient.GetServiceProperties(c.config.Ctx, resourceGroupName, *storage.Name, nil)
		if err == nil {
			scanContext.BlobServiceProperties = &blobServicesProperties
		}

		rr := engine.EvaluateRecommendations(rules, storage, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *storage.Name,
			Type:             *storage.Type,
			Location:         *storage.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *StorageScanner) listStorage() ([]*armstorage.Account, error) {
	pager := c.storageClient.NewListPager(nil)

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

func (a *StorageScanner) ResourceTypes() []string {
	return []string{"Microsoft.Storage/storageAccounts"}
}
