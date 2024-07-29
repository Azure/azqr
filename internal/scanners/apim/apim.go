// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package apim

import (
	"github.com/Azure/azqr/internal/azqr"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

// APIManagementScanner - Scanner for API Management Services
type APIManagementScanner struct {
	config        *azqr.ScannerConfig
	serviceClient *armapimanagement.ServiceClient
}

// Init - Initializes the APIManagementScanner
func (a *APIManagementScanner) Init(config *azqr.ScannerConfig) error {
	a.config = config
	var err error
	a.serviceClient, err = armapimanagement.NewServiceClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan -Scans all API Management Services in a Resource Group
func (a *APIManagementScanner) Scan(resourceGroupName string, scanContext *azqr.ScanContext) ([]azqr.AzqrServiceResult, error) {
	azqr.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, a.ResourceTypes()[0])

	services, err := a.listServices(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := azqr.RecommendationEngine{}
	rules := a.GetRecommendations()
	results := []azqr.AzqrServiceResult{}

	for _, s := range services {
		rr := engine.EvaluateRecommendations(rules, s, scanContext)

		results = append(results, azqr.AzqrServiceResult{
			SubscriptionID:   a.config.SubscriptionID,
			SubscriptionName: a.config.SubscriptionName,
			ResourceGroup:    resourceGroupName,
			ServiceName:      *s.Name,
			Type:             *s.Type,
			Location:         *s.Location,
			Recommendations:  rr,
		})
	}
	return results, nil
}

func (a *APIManagementScanner) listServices(resourceGroupName string) ([]*armapimanagement.ServiceResource, error) {
	pager := a.serviceClient.NewListByResourceGroupPager(resourceGroupName, nil)

	services := make([]*armapimanagement.ServiceResource, 0)
	for pager.More() {
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
