package analyzers

import "strings"

// AzureServiceAnalyzer - Interface for all Azure Service Analyzers
type AzureServiceAnalyzer interface {
	Review(resourceGroupName string) ([]AzureServiceResult, error)
}

// AzureBaseServiceResult - Base struct for all Azure Service Results
type AzureBaseServiceResult struct {
	SubscriptionID string
	ResourceGroup  string
	ServiceName    string
	SKU            string
	SLA            string
	Type           string
	Location       string
	CAFNaming      bool
}

// AzureServiceResult - Struct for all Azure Service Results
type AzureServiceResult struct {
	AzureBaseServiceResult
	AvailabilityZones  bool
	PrivateEndpoints   bool
	DiagnosticSettings bool
}

func parseLocation(location *string) string {
	return strings.ToLower(strings.ReplaceAll(*location, " ", ""))
}
