package analyzers

import "strings"

type AzureServiceAnalyzer interface {
	Review(resourceGroupName string) ([]AzureServiceResult, error)
}

type AzureBaseServiceResult struct {
	SubscriptionId string
	ResourceGroup  string
	ServiceName    string
	Sku            string
	Sla            string
	Type           string
	Location       string
	CAFNaming      bool
}

type AzureServiceResult struct {
	AzureBaseServiceResult
	AvailabilityZones  bool
	PrivateEndpoints   bool
	DiagnosticSettings bool
}

func parseLocation(location *string) string {
	return strings.ToLower(strings.ReplaceAll(*location, " ", ""))
}
