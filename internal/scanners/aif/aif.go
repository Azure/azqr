// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aif

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices/v2"
)

func init() {
	models.ScannerList["aif"] = []models.IAzureScanner{&AIFoundryScanner{}}
}

// AIFoundryScanner - Scanner for Cognitive Services Accounts
type AIFoundryScanner struct {
	config *models.ScannerConfig
	client *armcognitiveservices.AccountsClient
}

// Init - Initializes the AIFoundryScanner
func (a *AIFoundryScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.client, err = armcognitiveservices.NewAccountsClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all Cognitive Services Accounts in a Resource Group
func (c *AIFoundryScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	services, err := c.list()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, service := range services {
		rr := engine.EvaluateRecommendations(rules, service, scanContext)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*service.ID),
			ServiceName:      *service.Name,
			Type:             *service.Type,
			Location:         *service.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *AIFoundryScanner) list() ([]*armcognitiveservices.Account, error) {
	pager := c.client.NewListPager(nil)

	namespaces := make([]*armcognitiveservices.Account, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		_ = throttling.WaitARM(c.config.Ctx); // nolint:errcheck
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, resp.Value...)
	}
	return namespaces, nil
}

func (a *AIFoundryScanner) ResourceTypes() []string {
	return []string{"Microsoft.CognitiveServices/accounts"}
}
