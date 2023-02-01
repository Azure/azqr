package scanners

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

// APIManagementScanner - Analyzer for API Management Services
type APIManagementScanner struct {
	config              *ScannerConfig
	diagnosticsSettings DiagnosticsSettings
	serviceClient       *armapimanagement.ServiceClient
	listServicesFunc    func(resourceGroupName string) ([]*armapimanagement.ServiceResource, error)
}

// Init - Initializes the APIManagementScanner
func (a *APIManagementScanner) Init(config *ScannerConfig) error {
	a.config = config
	var err error
	a.serviceClient, err = armapimanagement.NewServiceClient(config.SubscriptionID, config.Cred, nil)
	if err != nil {
		return err
	}
	a.diagnosticsSettings = DiagnosticsSettings{}
	err = a.diagnosticsSettings.Init(config)
	if err != nil {
		return err
	}
	return nil
}

// Scan -Scans all API Management Services in a Resource Group
func (a *APIManagementScanner) Scan(resourceGroupName string) ([]IAzureServiceResult, error) {
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
			SubscriptionID:     a.config.SubscriptionID,
			ResourceGroup:      resourceGroupName,
			ServiceName:        *s.Name,
			SKU:                sku,
			SLA:                sla,
			Type:               *s.Type,
			Location:           *s.Location,
			CAFNaming:          strings.HasPrefix(*s.Name, "apim"),
			AvailabilityZones:  len(s.Zones) > 0,
			PrivateEndpoints:   len(s.Properties.PrivateEndpointConnections) > 0,
			DiagnosticSettings: hasDiagnostics,
		})
	}
	return results, nil
}

func (a *APIManagementScanner) listServices(resourceGroupName string) ([]*armapimanagement.ServiceResource, error) {
	if a.listServicesFunc == nil {
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

	return a.listServicesFunc(resourceGroupName)
}
