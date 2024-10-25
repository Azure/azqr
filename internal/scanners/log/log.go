// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package log

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights/v2"
)

// LogAnalyticsScanner - Scanner for Log Analytics workspace
type LogAnalyticsScanner struct {
	config *azqr.ScannerConfig
	client *armoperationalinsights.WorkspacesClient
}

// Init - Initializes the Log Analytics workspace Scanner
func (a *LogAnalyticsScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armoperationalinsights.NewWorkspacesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Log Analytics workspace in a Resource Group
func (c *LogAnalyticsScanner) Scan(scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	svcs, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, w := range svcs {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    azqr.GetResourceGroupFromResourceID(*w.ID),
			ServiceName:      *w.Name,
			Type:             *w.Type,
			Location:         *w.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *LogAnalyticsScanner) list() ([]*armoperationalinsights.Workspace, error) {
	pager := c.client.NewListPager(nil)

	svcs := make([]*armoperationalinsights.Workspace, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		svcs = append(svcs, resp.Value...)
	}
	return svcs, nil
}

func (a *LogAnalyticsScanner) ResourceTypes() []string {
	return []string{"Microsoft.OperationalInsights/workspaces"}
}
