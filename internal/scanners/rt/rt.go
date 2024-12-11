// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package rt

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["rt"] = []models.IAzureScanner{&RouteTableScanner{}}
}

// RouteTableScanner - Scanner for Route Table
type RouteTableScanner struct {
	config *models.ScannerConfig
	client *armnetwork.RouteTablesClient
}

// Init - Initializes the Route Table Scanner
func (a *RouteTableScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armnetwork.NewRouteTablesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Route Table in a Resource Group
func (c *RouteTableScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	svcs, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, w := range svcs {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*w.ID),
			ServiceName:      *w.Name,
			Type:             *w.Type,
			Location:         parseLocation(w.Location),
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *RouteTableScanner) list() ([]*armnetwork.RouteTable, error) {
	pager := c.client.NewListAllPager(nil)

	svcs := make([]*armnetwork.RouteTable, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		svcs = append(svcs, resp.Value...)
	}
	return svcs, nil
}

func (a *RouteTableScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/routeTables"}
}

func parseLocation(l *string) string {
	if l == nil {
		return ""
	}
	return *l
}
