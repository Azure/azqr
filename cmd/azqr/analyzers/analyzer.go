package analyzers

import (
	"strconv"
	"strings"
)

type IAzureServiceResult interface {
	GetResourceType() string
	ToCommonResult() [][]string
}

// AzureServiceAnalyzer - Interface for all Azure Service Analyzers
type AzureServiceAnalyzer interface {
	Review(resourceGroupName string) ([]IAzureServiceResult, error)
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

// ToCommonResult - Returns a string representation of the Azure Service Result
func (r AzureServiceResult) ToCommonResult() [][]string {
	return [][]string{
		{r.SubscriptionID, r.ResourceGroup, r.Location, r.Type, r.ServiceName, r.SKU, r.SLA, strconv.FormatBool(r.AvailabilityZones), strconv.FormatBool(r.PrivateEndpoints), strconv.FormatBool(r.DiagnosticSettings), strconv.FormatBool(r.CAFNaming)},
	}
}

// GetResourceType - Returns the resource type of the Azure Service Result
func (r AzureServiceResult) GetResourceType() string {
	return r.Type
}

type IAzureFunctionAppResult interface {
	IAzureServiceResult
	ToFunctionResult() [][]string
}

// AzureFunctionAppResult - Struct for Azure Fucntion App Results
type AzureFunctionAppResult struct {
	AzureServiceResult
	AzureWebJobsDashboard         bool
	ScaleControllerLoggingEnabled bool // SCALE_CONTROLLER_LOGGING_ENABLED
	ContentOverVNET               bool // WEBSITE_CONTENTOVERVNET
	RunFromPackage                bool // WEBSITE_RUN_FROM_PACKAGE
	VNETRouteAll                  bool // WEBSITE_VNET_ROUTE_ALL
	AppInsightsEnabled            bool // APPINSIGHTS_INSTRUMENTATIONKEY or APPLICATIONINSIGHTS_CONNECTION_STRING
}

// ToFunctionResult - Returns a string representation of the Azure Function App Result
func (r AzureFunctionAppResult) ToFunctionResult() [][]string {
	return [][]string{
		{r.SubscriptionID, r.ResourceGroup, r.Location, r.Type, r.ServiceName, strconv.FormatBool(r.RunFromPackage), strconv.FormatBool(r.ContentOverVNET), strconv.FormatBool(r.VNETRouteAll), strconv.FormatBool(r.AzureWebJobsDashboard), strconv.FormatBool(r.AppInsightsEnabled), strconv.FormatBool(r.ScaleControllerLoggingEnabled)},
	}
}

func parseLocation(location *string) string {
	return strings.ToLower(strings.ReplaceAll(*location, " ", ""))
}
