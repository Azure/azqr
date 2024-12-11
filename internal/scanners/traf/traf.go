// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package traf

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
)

func init() {
	models.ScannerList["traf"] = []models.IAzureScanner{&TrafficManagerScanner{}}
}

// TrafficManagerScanner - Scanner for TrafficManager
type TrafficManagerScanner struct {
	config *models.ScannerConfig
	client *armtrafficmanager.ClientFactory
}

// Init - Initializes the TrafficManager
func (c *TrafficManagerScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armtrafficmanager.NewClientFactory(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all TrafficManager in a Resource Group
func (c *TrafficManagerScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	vnets, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, w := range vnets {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*w.ID),
			ServiceName:      *w.Name,
			Type:             *w.Type,
			Location:         *w.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *TrafficManagerScanner) list() ([]*armtrafficmanager.Profile, error) {
	pager := c.client.NewProfilesClient().NewListBySubscriptionPager(nil)

	vnets := make([]*armtrafficmanager.Profile, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		vnets = append(vnets, resp.Value...)
	}
	return vnets, nil
}

func (a *TrafficManagerScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/trafficManagerProfiles"}
}
