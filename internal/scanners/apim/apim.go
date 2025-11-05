// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package apim

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

func init() {
	models.ScannerList["apim"] = []models.IAzureScanner{&APIManagementScanner{}}
}

// APIManagementScanner - Scanner for API Management Services
type APIManagementScanner struct {
	config        *models.ScannerConfig
	serviceClient *armapimanagement.ServiceClient
}

// Init - Initializes the APIManagementScanner
func (a *APIManagementScanner) Init(config *models.ScannerConfig) error {
	a.config = config
	var err error
	a.serviceClient, err = armapimanagement.NewServiceClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan -Scans all API Management Services in a Resource Group
func (a *APIManagementScanner) Scan(scanContext *models.ScanContext) ([]*models.AzqrServiceResult, error) {
	models.LogSubscriptionScan(a.config.SubscriptionID, a.ResourceTypes()[0])

	services, err := a.listServices()
	if err != nil {
		return nil, err
	}
	engine := models.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []*models.AzqrServiceResult{}

	for _, s := range services {
		rr := engine.EvaluateRecommendations(rules, s, scanContext)

		results = append(results, &models.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    models.GetResourceGroupFromResourceID(*s.ID),
			ServiceName:      *s.Name,
			Type:             *s.Type,
			Location:         *s.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *APIManagementScanner) listServices() ([]*armapimanagement.ServiceResource, error) {
	pager := a.serviceClient.NewListPager(nil)

	services := make([]*armapimanagement.ServiceResource, 0)
	for pager.More() {
		// Wait for a token from the burstLimiter channel before making the request
		<-throttling.ARMLimiter
		resp, err := pager.NextPage(a.config.Ctx)
		if err != nil {
			return nil, err
		}
		services = append(services, resp.Value...)
	}
	return services, nil
}

func (a *APIManagementScanner) ResourceTypes() []string {
	return []string{"Microsoft.ApiManagement/service"}
}
