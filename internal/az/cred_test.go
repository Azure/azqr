// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package az

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
)

func TestGetCloudConfiguration_AzurePublic(t *testing.T) {
	// Clear environment variables
	t.Setenv("AZURE_CLOUD", "")
	t.Setenv("AZURE_AUTHORITY_HOST", "")
	t.Setenv("AZURE_RESOURCE_MANAGER_ENDPOINT", "")
	t.Setenv("AZURE_RESOURCE_MANAGER_AUDIENCE", "")

	config := GetCloudConfiguration()

	if config.ActiveDirectoryAuthorityHost != cloud.AzurePublic.ActiveDirectoryAuthorityHost {
		t.Errorf("Expected AzurePublic authority host, got %s", config.ActiveDirectoryAuthorityHost)
	}
}

func TestGetCloudConfiguration_AzurePublicExplicit(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "AzurePublic")

	config := GetCloudConfiguration()

	if config.ActiveDirectoryAuthorityHost != cloud.AzurePublic.ActiveDirectoryAuthorityHost {
		t.Errorf("Expected AzurePublic authority host, got %s", config.ActiveDirectoryAuthorityHost)
	}
}

func TestGetCloudConfiguration_AzureGovernment(t *testing.T) {
	testCases := []string{"AzureGovernment", "azuregovernment", "AzureUSGovernment", "USGovernment"}

	for _, cloudName := range testCases {
		t.Run(cloudName, func(t *testing.T) {
			t.Setenv("AZURE_CLOUD", cloudName)

			config := GetCloudConfiguration()

			if config.ActiveDirectoryAuthorityHost != cloud.AzureGovernment.ActiveDirectoryAuthorityHost {
				t.Errorf("Expected AzureGovernment authority host for %s, got %s", cloudName, config.ActiveDirectoryAuthorityHost)
			}
		})
	}
}

func TestGetCloudConfiguration_AzureChina(t *testing.T) {
	testCases := []string{"AzureChina", "azurechina", "China", "china"}

	for _, cloudName := range testCases {
		t.Run(cloudName, func(t *testing.T) {
			t.Setenv("AZURE_CLOUD", cloudName)

			config := GetCloudConfiguration()

			if config.ActiveDirectoryAuthorityHost != cloud.AzureChina.ActiveDirectoryAuthorityHost {
				t.Errorf("Expected AzureChina authority host for %s, got %s", cloudName, config.ActiveDirectoryAuthorityHost)
			}
		})
	}
}

func TestGetCloudConfiguration_UnknownCloud(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "UnknownCloud")

	config := GetCloudConfiguration()

	// Should default to AzurePublic for unknown cloud names
	if config.ActiveDirectoryAuthorityHost != cloud.AzurePublic.ActiveDirectoryAuthorityHost {
		t.Errorf("Expected default AzurePublic authority host for unknown cloud, got %s", config.ActiveDirectoryAuthorityHost)
	}
}

func TestGetCloudConfiguration_CustomCloud(t *testing.T) {
	customAuthHost := "https://login.custom.cloud/"
	customARMEndpoint := "https://management.custom.cloud"

	t.Setenv("AZURE_AUTHORITY_HOST", customAuthHost)
	t.Setenv("AZURE_RESOURCE_MANAGER_ENDPOINT", customARMEndpoint)

	config := GetCloudConfiguration()

	if config.ActiveDirectoryAuthorityHost != customAuthHost {
		t.Errorf("Expected custom authority host %s, got %s", customAuthHost, config.ActiveDirectoryAuthorityHost)
	}

	if rmService, ok := config.Services[cloud.ResourceManager]; ok {
		if rmService.Endpoint != customARMEndpoint {
			t.Errorf("Expected custom ARM endpoint %s, got %s", customARMEndpoint, rmService.Endpoint)
		}
	} else {
		t.Error("Expected ResourceManager service in config")
	}
}

func TestGetCloudConfiguration_CustomCloudWithAudience(t *testing.T) {
	customAuthHost := "https://login.custom.cloud/"
	customARMEndpoint := "https://management.custom.cloud"
	customAudience := "https://management.custom.cloud/"

	t.Setenv("AZURE_AUTHORITY_HOST", customAuthHost)
	t.Setenv("AZURE_RESOURCE_MANAGER_ENDPOINT", customARMEndpoint)
	t.Setenv("AZURE_RESOURCE_MANAGER_AUDIENCE", customAudience)

	config := GetCloudConfiguration()

	if rmService, ok := config.Services[cloud.ResourceManager]; ok {
		if rmService.Audience != customAudience {
			t.Errorf("Expected custom audience %s, got %s", customAudience, rmService.Audience)
		}
	} else {
		t.Error("Expected ResourceManager service in config")
	}
}

func TestGetCloudConfiguration_PartialCustomConfig(t *testing.T) {
	// Test with only auth host (should not use custom config)
	t.Setenv("AZURE_AUTHORITY_HOST", "https://login.custom.cloud/")

	config := GetCloudConfiguration()

	// Should default to public cloud when custom config is incomplete
	if config.ActiveDirectoryAuthorityHost != cloud.AzurePublic.ActiveDirectoryAuthorityHost {
		t.Error("Expected AzurePublic when custom config is incomplete")
	}
}

func TestGetResourceManagerEndpoint_Default(t *testing.T) {
	// Clear environment variables
	t.Setenv("AZURE_CLOUD", "")
	t.Setenv("AZURE_AUTHORITY_HOST", "")
	t.Setenv("AZURE_RESOURCE_MANAGER_ENDPOINT", "")

	endpoint := GetResourceManagerEndpoint()

	if endpoint == "" {
		t.Error("GetResourceManagerEndpoint() returned empty string")
	}

	// Should return public cloud endpoint without trailing slash
	expectedEndpoint := "https://management.azure.com"
	if endpoint != expectedEndpoint {
		t.Errorf("Expected endpoint %s, got %s", expectedEndpoint, endpoint)
	}
}

func TestGetResourceManagerEndpoint_CustomCloud(t *testing.T) {
	customARMEndpoint := "https://management.custom.cloud/"

	t.Setenv("AZURE_AUTHORITY_HOST", "https://login.custom.cloud/")
	t.Setenv("AZURE_RESOURCE_MANAGER_ENDPOINT", customARMEndpoint)

	endpoint := GetResourceManagerEndpoint()

	// Should trim trailing slash
	expectedEndpoint := "https://management.custom.cloud"
	if endpoint != expectedEndpoint {
		t.Errorf("Expected endpoint %s (trailing slash removed), got %s", expectedEndpoint, endpoint)
	}
}

func TestGetResourceManagerEndpoint_Government(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "AzureGovernment")

	endpoint := GetResourceManagerEndpoint()

	expectedEndpoint := "https://management.usgovcloudapi.net"
	if endpoint != expectedEndpoint {
		t.Errorf("Expected Government endpoint %s, got %s", expectedEndpoint, endpoint)
	}
}
