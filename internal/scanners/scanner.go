package scanners

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

type (
	// IAzureServiceResult - Interface for all Azure Service Results
	IAzureServiceResult interface {
		GetResourceType() string
		GetHeathers() []string
		GetDetailHeathers() []string
		ToMap(mask bool) map[string]string
		ToDetailMap(mask bool) map[string]string
		Value() AzureServiceResult
	}

	// ScannerConfig - Struct for Scanner Config
	ScannerConfig struct {
		Ctx                context.Context
		Cred               azcore.TokenCredential
		SubscriptionID     string
		EnableDetailedScan bool
	}

	// ScanContext - Struct for Scanner Context
	ScanContext struct {
		PrivateEndpoints map[string]bool
	}

	// IAzureScanner - Interface for all Azure Scanners
	IAzureScanner interface {
		Init(config *ScannerConfig) error
		Scan(resourceGroupName string, scanContext *ScanContext) ([]IAzureServiceResult, error)
	}

	// AzureServiceResult - Struct for all Azure Service Results
	AzureServiceResult struct {
		SubscriptionID     string
		ResourceGroup      string
		ServiceName        string
		SKU                string
		SLA                string
		Type               string
		Location           string
		CAFNaming          bool
		AvailabilityZones  bool
		PrivateEndpoints   bool
		DiagnosticSettings bool
	}
)

// ToMap - Returns a map representation of the Azure Service Result
func (r AzureServiceResult) ToMap(mask bool) map[string]string {
	return map[string]string{
		"SubscriptionID": maskSubscriptionID(r.SubscriptionID, mask),
		"ResourceGroup":  r.ResourceGroup,
		"Location":       parseLocation(r.Location),
		"Type":           r.Type,
		"Name":           r.ServiceName,
		"SKU":            r.SKU,
		"SLA":            r.SLA,
		"AZ":             strconv.FormatBool(r.AvailabilityZones),
		"PE":             strconv.FormatBool(r.PrivateEndpoints),
		"DS":             strconv.FormatBool(r.DiagnosticSettings),
		"CAF":            strconv.FormatBool(r.CAFNaming),
	}
}

// ToDetail - Returns a map representation of the Azure Service Result
func (r AzureServiceResult) ToDetailMap(mask bool) map[string]string {
	return map[string]string{}
}

// GetResourceType - Returns the resource type of the Azure Service Result
func (r AzureServiceResult) GetResourceType() string {
	return r.Type
}

// GetHeathers - Returns the headers of the Azure Service Result
func (r AzureServiceResult) GetHeathers() []string {
	return []string{
		"SubscriptionID",
		"ResourceGroup",
		"Location",
		"Type",
		"Name",
		"SKU",
		"SLA",
		"AZ",
		"PE",
		"DS",
		"CAF",
	}
}

// GetDeatilsHeathers - Returns the detail headers of the Azure Service Result
func (r AzureServiceResult) GetDetailHeathers() []string {
	return []string{}
}

// Get - Returns the Azure Service Result
func (r AzureServiceResult) Value() AzureServiceResult {
	return r
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

// ToDetail - Returns a map representation of the Azure Function App Result
func (r AzureFunctionAppResult) ToDetailMap(mask bool) map[string]string {
	return map[string]string{
		"SubscriptionID":                maskSubscriptionID(r.SubscriptionID, mask),
		"ResourceGroup":                 r.ResourceGroup,
		"Location":                      parseLocation(r.Location),
		"Type":                          r.Type,
		"Name":                          r.ServiceName,
		"RunFromPackage":                strconv.FormatBool(r.RunFromPackage),
		"ContentOverVNET":               strconv.FormatBool(r.ContentOverVNET),
		"VNETRouteAll":                  strconv.FormatBool(r.VNETRouteAll),
		"AzureWebJobsDashboard":         strconv.FormatBool(r.AzureWebJobsDashboard),
		"AppInsightsEnabled":            strconv.FormatBool(r.AppInsightsEnabled),
		"ScaleControllerLoggingEnabled": strconv.FormatBool(r.ScaleControllerLoggingEnabled),
	}
}

// GetDetailProperties - Returns the detail properties of the Azure Function App Result
func (r AzureFunctionAppResult) GetDetailProperties() []string {
	return []string{
		"SubscriptionID",
		"ResourceGroup",
		"Location",
		"Type",
		"Name",
		"RunFromPackage",
		"ContentOverVNET",
		"VNETRouteAll",
		"AzureWebJobsDashboard",
		"AppInsightsEnabled",
		"ScaleControllerLoggingEnabled",
	}
}

func parseLocation(location string) string {
	return strings.ToLower(strings.ReplaceAll(location, " ", ""))
}

func maskSubscriptionID(subscriptionID string, mask bool) string {
	if !mask {
		return subscriptionID
	}

	// Show only last 7 chars of the subscription ID
	return fmt.Sprintf("xxxxxxxx-xxxx-xxxx-xxxx-xxxxx%s", subscriptionID[29:])
}
