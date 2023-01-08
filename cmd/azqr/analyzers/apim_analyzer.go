package analyzers

import (
	"context"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

// APIManagementAnalyzer - Analyzer for API Management Services
type APIManagementAnalyzer struct {
	diagnosticsSettings DiagnosticsSettings
	subscriptionID      string
	ctx                 context.Context
	cred                azcore.TokenCredential
	serviceClient       *armapimanagement.ServiceClient
	listServicesFunc    func(resourceGroupName string) ([]*armapimanagement.ServiceResource, error)
}

// NewAPIManagementAnalyzer - Creates a new APIManagementAnalyzer
func NewAPIManagementAnalyzer(ctx context.Context, subscriptionID string, cred azcore.TokenCredential) *APIManagementAnalyzer {
	diagnosticsSettings, _ := NewDiagnosticsSettings(ctx, cred)
	serviceClient, err := armapimanagement.NewServiceClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	analyzer := APIManagementAnalyzer{
		diagnosticsSettings: *diagnosticsSettings,
		subscriptionID:      subscriptionID,
		ctx:                 ctx,
		cred:                cred,
		serviceClient:       serviceClient,
	}
	return &analyzer
}

// Review -Analyzes all API Management Services in a Resource Group
func (a APIManagementAnalyzer) Review(resourceGroupName string) ([]IAzureServiceResult, error) {
	log.Printf("Analyzing API Management Services in Resource Group %s", resourceGroupName)

	services, err := a.listServices(resourceGroupName)
	if err != nil {
		return nil, err
	}
	results := []IAzureServiceResult{}
	for _, s := range services {
		hasDiagnostics, err := a.diagnosticsSettings.HasDiagnostics(*s.ID)
		if err != nil {
			return nil, err
		}

		sku := string(*s.SKU.Name)
		sla := "99.95%"
		if strings.Contains(sku, "Premium") && (len(s.Zones) > 0 || len(s.Properties.AdditionalLocations) > 0) {
			sla = "99.99%"
		} else if strings.Contains(sku, "Developer") {
			sla = "None"
		}

		results = append(results, AzureServiceResult{
			AzureBaseServiceResult: AzureBaseServiceResult{
				SubscriptionID: a.subscriptionID,
				ResourceGroup:  resourceGroupName,
				ServiceName:    *s.Name,
				SKU:            sku,
				SLA:            sla,
				Type:           *s.Type,
				Location:       parseLocation(s.Location),
				CAFNaming:      strings.HasPrefix(*s.Name, "apim")},
			AvailabilityZones:  len(s.Zones) > 0,
			PrivateEndpoints:   len(s.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (a APIManagementAnalyzer) listServices(resourceGroupName string) ([]*armapimanagement.ServiceResource, error) {
	if a.listServicesFunc == nil {
		pager := a.serviceClient.NewListByResourceGroupPager(resourceGroupName, nil)

		services := make([]*armapimanagement.ServiceResource, 0)
		for pager.More() {
			resp, err := pager.NextPage(a.ctx)
			if err != nil {
				return nil, err
			}
			services = append(services, resp.Value...)
		}
		return services, nil
	}

	return a.listServicesFunc(resourceGroupName)
}
