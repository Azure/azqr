// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package afw

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["afw"] = []models.IAzureScanner{&FirewallScanner{}}
}

// FirewallScanner - Scanner for Azure Firewall
type FirewallScanner struct {
	config *models.ScannerConfig
	client *armnetwork.AzureFirewallsClient
}

// Init - Initializes the Azure Firewall
func (a *FirewallScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armnetwork.NewAzureFirewallsClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Azure Firewall in a Resource Group
func (a *FirewallScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	gateways, err := a.list()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, g := range gateways {
		rr := engine.EvaluateRecommendations(rules, g, scanContext)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*g.ID),
			Location:         *g.Location,
			Type:             *g.Type,
			ServiceName:      *g.Name,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *FirewallScanner) list() ([]*armnetwork.AzureFirewall, error) {
	pager := a.client.NewListAllPager(nil)

	services := make([]*armnetwork.AzureFirewall, 0)
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		services = append(services, resp.Value...)
	}
	return services, nil
}

func (a *FirewallScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/azureFirewalls", "Microsoft.Network/ipGroups"}
}
