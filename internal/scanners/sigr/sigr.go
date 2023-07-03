// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sigr

import (
	"log"

	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/signalr/armsignalr"
)

// SignalRScanner - Scanner for SignalR
type SignalRScanner struct {
	config        *scanners.ScannerConfig
	signalrClient *armsignalr.Client
}

// Init - Initializes the SignalRScanner
func (c *SignalRScanner) Init(config *scanners.ScannerConfig) error {
	c.config = config
	var err error
	c.signalrClient, err = armsignalr.NewClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all SignalR in a Resource Group
func (c *SignalRScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	log.Printf("Scanning SignalR in Resource Group %s", resourceGroupName)

	signalr, err := c.listSignalR(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := c.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, signalr := range signalr {
		rr := engine.EvaluateRules(rules, signalr, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: c.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *signalr.Name,
			Type:           *signalr.Type,
			Location:       *signalr.Location,
			Rules:          rr,
		})
	}
	return results, nil
}

func (c *SignalRScanner) listSignalR(resourceGroupName string) ([]*armsignalr.ResourceInfo, error) {
	pager := c.signalrClient.NewListByResourceGroupPager(resourceGroupName, nil)

	signalrs := make([]*armsignalr.ResourceInfo, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		signalrs = append(signalrs, resp.Value...)
	}
	return signalrs, nil
}
