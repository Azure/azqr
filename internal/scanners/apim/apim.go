// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package apim

import (
	"github.com/Azure/azqr/internal/scanners"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

// APIManagementScanner - Scanner for API Management Services
type APIManagementScanner struct {
	config        *scanners.ScannerConfig
	serviceClient *armapimanagement.ServiceClient
}

// Init - Initializes the APIManagementScanner
func (a *APIManagementScanner) Init(config *scanners.ScannerConfig) error {
	a.config = config
	var err error
	a.serviceClient, err = armapimanagement.NewServiceClient(config.SubscriptionID, config.Cred, config.ClientOptions)
	return err
}

// Scan -Scans all API Management Services in a Resource Group
func (a *APIManagementScanner) Scan(resourceGroupName string, scanContext *scanners.ScanContext) ([]scanners.AzureServiceResult, error) {
	scanners.LogResourceGroupScan(a.config.SubscriptionID, resourceGroupName, "APIM")

	services, err := a.listServices(resourceGroupName)
	if err != nil {
		return nil, err
	}
	engine := scanners.RuleEngine{}
	rules := a.GetRules()
	results := []scanners.AzureServiceResult{}

	for _, s := range services {
		rr := engine.EvaluateRules(rules, s, scanContext)

		results = append(results, scanners.AzureServiceResult{
			SubscriptionID: a.config.SubscriptionID,
			ResourceGroup:  resourceGroupName,
			ServiceName:    *s.Name,
			Type:           *s.Type,
			Location:       *s.Location,
			Rules:          rr,
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
