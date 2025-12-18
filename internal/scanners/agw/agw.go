// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package agw

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
)

func init() {
	models.ScannerList["agw"] = []models.IAzureScanner{&ApplicationGatewayScanner{}}
}

// ApplicationGatewayScanner - Scanner for Application Gateways
type ApplicationGatewayScanner struct {
	config         *models.ScannerConfig
	gatewaysClient *armnetwork.ApplicationGatewaysClient
}

// Init - Initializes the ApplicationGatewayAnalyzer
func (a *ApplicationGatewayScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.gatewaysClient, err = armnetwork.NewApplicationGatewaysClient(config.SubscriptionID, a.config.Cred, a.config.ClientOptions)
	return err
}

// Scan - Scans all Application Gateways in a Resource Group
func (a *ApplicationGatewayScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	gateways, err := a.listGateways()
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
			ServiceName:      *g.Name,
			Type:             *g.Type,
			Location:         *g.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *ApplicationGatewayScanner) listGateways() ([]*armnetwork.ApplicationGateway, error) {
	pager := a.gatewaysClient.NewListAllPager(nil)
	results := []*armnetwork.ApplicationGateway{}
	for pager.More() {
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}
	return results, nil
}

func (a *ApplicationGatewayScanner) ResourceTypes() []string {
	return []string{"Microsoft.Network/applicationGateways"}
}
