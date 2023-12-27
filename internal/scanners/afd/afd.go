// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afd

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
)

// FrontDoorScanner - Scanner for Front Door
type FrontDoorScanner struct {
	config *scanners.ScannerConfig
	client *armcdn.ProfilesClient
}

// Init - Initializes the FrontDoor Scanner
func (a *FrontDoorScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armcdn.NewProfilesClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Front Doors in a Resource Group
func (a *FrontDoorScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, "FrontDoor")

	gateways, err := a.list(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, g := range gateways {
		rr := engine.EvaluateRules(rules, g, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: a.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			Location:       *g.Location,
			Type:           *g.Type,
			ServiceName:    *g.Name,
			Rules:          rr,
		})
	}
	return results, nil
}

func (a *FrontDoorScanner) list(resourceGroupName string) ([]*armcdn.Profile, error) {
	pager := a.client.NewListByResourceGroupPager(resourceGroupName, nil)

	services := make([]*armcdn.Profile, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		services = append(services, resp.Value...)
	}
	return services, nil
}
