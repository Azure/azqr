// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package fabric

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/fabric/armfabric"
)

func init() {
	models.ScannerList["fabric"] = []models.IAzureScanner{&FabricScanner{}}
}

// FabricScanner - Scanner for Microsoft Fabric Capacities
type FabricScanner struct {
	config *models.ScannerConfig
	client *armfabric.CapacitiesClient
}

// Init - Initializes the FabricScanner
func (c *FabricScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.client, err = armfabric.NewCapacitiesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Microsoft Fabric Capacities in a Subscription
func (c *FabricScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	capacities, err := c.listCapacities()
	if err != nil {
		if models.ShouldSkipError(err) {
			return nil, nil
		}
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, capacity := range capacities {
		rr := engine.EvaluateRecommendations(rules, capacity, scanContext)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*capacity.ID),
			ServiceName:      *capacity.Name,
			Type:             *capacity.Type,
			Location:         *capacity.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *FabricScanner) listCapacities() ([]*armfabric.Capacity, error) {
	pager := c.client.NewListBySubscriptionPager(nil)

	capacities := make([]*armfabric.Capacity, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		capacities = append(capacities, resp.Value...)
	}
	return capacities, nil
}

func (a *FabricScanner) ResourceTypes() []string {
	return []string{"Microsoft.Fabric/capacities"}
}
