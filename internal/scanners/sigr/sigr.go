// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sigr

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/signalr/armsignalr"
)

func init() {
	models.ScannerList["sigr"] = []models.IAzureScanner{&SignalRScanner{}}
}

// SignalRScanner - Scanner for SignalR
type SignalRScanner struct {
	config        *models.ScannerConfig
	signalrClient *armsignalr.Client
}

// Init - Initializes the SignalRScanner
func (c *SignalRScanner) Init(config *models.ScannerConfig) error {
	c.config = config
	var err error
	c.signalrClient, err = armsignalr.NewClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan - Scans all SignalR in a Resource Group
func (c *SignalRScanner) Scan(scanContext *models.ScanContext) ([]models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(c.config.SubscriptionID, c.ResourceTypes()[0])

	signalr, err := c.listSignalR()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := c.GetRecommendations()
	results := []models.AzqrServiceResult{}

	for _, signalr := range signalr {
		rr := engine.EvaluateRecommendations(rules, signalr, scanContext)

		results = append(results, models.AzqrServiceResult{
			SubscriptionID:   c.config.SubscriptionID,
			SubscriptionName: c.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*signalr.ID),
			ServiceName:      *signalr.Name,
			Type:             *signalr.Type,
			Location:         *signalr.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (c *SignalRScanner) listSignalR() ([]*armsignalr.ResourceInfo, error) {
	pager := c.signalrClient.NewListBySubscriptionPager(nil)

	signalrs := make([]*armsignalr.ResourceInfo, 0)
	for pager.More() {
		resp, err := pager.NextPage(c.config.Ctx)
		if err != nil {
			return nil, err
		}
		signalrs = append(signalrs, resp.Value...)
	}
	return signalrs, nil
}

func (a *SignalRScanner) ResourceTypes() []string {
	return []string{"Microsoft.SignalRService/SignalR"}
}
