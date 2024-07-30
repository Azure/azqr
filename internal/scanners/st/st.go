// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package st

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

// StorageScanner - Scanner for Storage
type StorageScanner struct {
	config             *azqr.ScannerConfig
	storageClient      *armstorage.AccountsClient
	blobServicesClient *armstorage.BlobServicesClient
}

// Init - Initializes the StorageScanner
func (c *StorageScanner) Init(config *azqr.ScannerConfig) error {
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
func (c *StorageScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	storage, err := c.listStorage()
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, storage := range storage {
		resourceGroupName := azqr.GetResourceGroupFromResourceID(*storage.ID)

		scanContext.BlobServiceProperties = nil
		blobServicesProperties, err := c.blobServicesClient.GetServiceProperties(c.config.Ctx, resourceGroupName, *storage.Name, nil)
		if err == nil {
			scanContext.BlobServiceProperties = &blobServicesProperties
		}

		rr := engine.EvaluateRecommendations(rules, storage, scanContext)

		results = append(results, azqr.AzqrServiceResult{
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
