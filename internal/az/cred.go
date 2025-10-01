package az

import (
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/rs/zerolog/log"
)

// NewAzureCredential creates a new Azure credential using DefaultAzureCredential.
// The credential chain behavior can be customized using the AZURE_TOKEN_CREDENTIALS environment variable.
func NewAzureCredential() azcore.TokenCredential {
	var cred azcore.TokenCredential
	var err error

	opts := azcore.ClientOptions{Cloud: GetCloudConfiguration()}

	cred, err = azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{ClientOptions: opts})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get Azure credentials")
	}
	return cred
}

// GetCloudConfiguration returns the appropriate Azure cloud configuration
// based on environment variables. It supports both predefined clouds
// (AzurePublic, AzureGovernment, AzureChina) and custom cloud configurations.
//
// Environment variables:
//   - AZURE_CLOUD: Name of the predefined cloud (AzurePublic, AzureGovernment, AzureChina, etc.)
//   - AZURE_AUTHORITY_HOST: Custom Active Directory authority host (e.g., https://login.microsoftonline.us/)
//   - AZURE_RESOURCE_MANAGER_ENDPOINT: Custom ARM endpoint (e.g., https://management.usgovcloudapi.net)
//   - AZURE_RESOURCE_MANAGER_AUDIENCE: Custom ARM token audience (e.g., https://management.core.usgovcloudapi.net/)
//
// Priority:
//  1. If custom endpoints are provided (AZURE_AUTHORITY_HOST and AZURE_RESOURCE_MANAGER_ENDPOINT),
//     returns a custom configuration
//  2. If AZURE_CLOUD is set to a known cloud name, returns that predefined configuration
//  3. Otherwise, returns AzurePublic (default)
func GetCloudConfiguration() cloud.Configuration {
	// Check for custom cloud configuration first
	authHost := os.Getenv("AZURE_AUTHORITY_HOST")
	armEndpoint := os.Getenv("AZURE_RESOURCE_MANAGER_ENDPOINT")

	// If both custom endpoints are provided, use custom configuration
	if authHost != "" && armEndpoint != "" {
		config := cloud.Configuration{
			ActiveDirectoryAuthorityHost: authHost,
			Services: map[cloud.ServiceName]cloud.ServiceConfiguration{
				cloud.ResourceManager: {
					Endpoint: armEndpoint,
				},
			},
		}

		// Optionally add audience if provided
		if audience := os.Getenv("AZURE_RESOURCE_MANAGER_AUDIENCE"); audience != "" {
			config.Services[cloud.ResourceManager] = cloud.ServiceConfiguration{
				Endpoint: armEndpoint,
				Audience: audience,
			}
		}

		return config
	}

	// Otherwise, check for predefined cloud name
	cloudName := os.Getenv("AZURE_CLOUD")
	cloudName = strings.ToLower(strings.TrimSpace(cloudName))

	switch cloudName {
	case "azuregovernment", "azureusgovernment", "usgovernment":
		return cloud.AzureGovernment
	case "azurechina", "china":
		return cloud.AzureChina
	case "azurepublic", "public", "":
		// Empty string defaults to public cloud
		return cloud.AzurePublic
	default:
		// Unknown cloud name, default to public
		return cloud.AzurePublic
	}
}

func GetResourceManagerEndpoint() string {
	cloudConfig := GetCloudConfiguration()
	if service, ok := cloudConfig.Services[cloud.ResourceManager]; ok {
		return strings.TrimSuffix(service.Endpoint, "/")
	}
	// Default to public cloud endpoint if not found
	endpoint := cloud.AzurePublic.Services[cloud.ResourceManager].Endpoint
	return strings.TrimSuffix(endpoint, "/")
}
