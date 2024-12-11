// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cog

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices"
)

func init() {
	models.ScannerList["cog"] = []models.IAzureScanner{&CognitiveScanner{}}
}

// CognitiveScanner - Scanner for Cognitive Services Accounts
type CognitiveScanner struct {
	config *models.ScannerConfig
	client *armcognitiveservices.AccountsClient
}

// Init - Initializes the CognitiveScanner
func (a *CognitiveScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armcognitiveservices.NewAccountsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Cognitive Services Accounts in a Resource Group
func (c *CognitiveScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	eventHubs, err := c.listEventHubs()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, eventHub := range eventHubs {
		rr := engine.EvaluateRecommendations(rules, eventHub, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*eventHub.ID),
			ServiceName:      *eventHub.Name,
			Type:             *eventHub.Type,
			Location:         *eventHub.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *CognitiveScanner) listEventHubs() ([]*armcognitiveservices.Account, error) {
	pager := c.client.NewListPager(nil)

	namespaces := make([]*armcognitiveservices.Account, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, resp.Value...)
	}
	return namespaces, nil
}

func (a *CognitiveScanner) ResourceTypes() []string {
	return []string{"Microsoft.CognitiveServices/accounts"}
}
