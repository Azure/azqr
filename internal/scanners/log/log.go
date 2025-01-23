// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package log

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights/v2"
)

func init() {
	scanners.ScannerList["log"] = []scanners.IAzureScanner{&LogAnalyticsScanner{}}
}

// LogAnalyticsScanner - Scanner for Log Analytics workspace
type LogAnalyticsScanner struct {
	config *scanners.ScannerConfig
	client *armoperationalinsights.WorkspacesClient
}

// Init - Initializes the Log Analytics workspace Scanner
func (a *LogAnalyticsScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armoperationalinsights.NewWorkspacesClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Log Analytics workspace in a Resource Group
func (c *LogAnalyticsScanner) Scan(scanContext *scanners.ScanContext) ([]scanners.AzqrServiceResult, error) {
	scanners.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	svcs, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := scanners.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []scanners.AzqrServiceResult{}

	for _, w := range svcs {
		rr := engine.EvaluateRecommendations(rules, w, scanContext)

		results = append(results, scanners.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    scanners.GetResourceGroupFromResourceID(*w.ID),
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
