// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices/v2"
)

func TestNewAIGovScanner(t *testing.T) {
	scanner := NewScanner()
	if scanner == nil {
		t.Fatal("NewAIGovScanner() returned nil")
	}
}

func TestAIGovScanner_GetMetadata(t *testing.T) {
	scanner := NewScanner()
	metadata := scanner.GetMetadata()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Name", metadata.Name, "ai-gov"},
		{"Version", metadata.Version, "1.0.0"},
		{"Description", metadata.Description, "Checks AI Governance"},
		{"Author", metadata.Author, "Azure Quick Review Team"},
		{"License", metadata.License, "MIT"},
		{"Type", metadata.Type, plugins.PluginTypeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("GetMetadata().%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestAIGovScanner_GetMetadata_ColumnMetadata(t *testing.T) {
	scanner := NewScanner()
	metadata := scanner.GetMetadata()

	expectedNames := []string{
		"Subscription",
		"Resource Group",
		"Account Name",
		"Kind",
		"SKU",
		"Deployment Name",
		"Model Name",
		"Model Version",
		"Model Format",
		"SKU Capacity",
		"Version Upgrade Option",
		"Spillover Enabled",
		"Spillover Deployment",
		"Hour",
		"Status Code",
		"Request Count",
	}

	if len(metadata.ColumnMetadata) != len(expectedNames) {
		t.Errorf("Expected %d columns, got %d", len(expectedNames), len(metadata.ColumnMetadata))
	}

	for i, expected := range expectedNames {
		if i >= len(metadata.ColumnMetadata) {
			break
		}
		col := metadata.ColumnMetadata[i]
		if col.Name != expected {
			t.Errorf("Column[%d].Name = %s, want %s", i, col.Name, expected)
		}
	}
}

func TestAIGovScanner_groupResourcesForBatch(t *testing.T) {
	scanner := NewScanner()

	tests := []struct {
		name          string
		resources     []*models.Resource
		expectedCount int
	}{
		{
			name:          "Empty resources",
			resources:     []*models.Resource{},
			expectedCount: 0,
		},
		{
			name: "Single resource",
			resources: []*models.Resource{
				{
					ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account1",
					SubscriptionID: "sub1",
					Location:       "eastus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
			},
			expectedCount: 1,
		},
		{
			name: "Same subscription and region",
			resources: []*models.Resource{
				{
					ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account1",
					SubscriptionID: "sub1",
					Location:       "eastus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
				{
					ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account2",
					SubscriptionID: "sub1",
					Location:       "eastus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
			},
			expectedCount: 1,
		},
		{
			name: "Different regions",
			resources: []*models.Resource{
				{
					ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account1",
					SubscriptionID: "sub1",
					Location:       "eastus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
				{
					ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account2",
					SubscriptionID: "sub1",
					Location:       "westus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
			},
			expectedCount: 2,
		},
		{
			name: "Different subscriptions",
			resources: []*models.Resource{
				{
					ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account1",
					SubscriptionID: "sub1",
					Location:       "eastus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
				{
					ID:             "/subscriptions/sub2/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account2",
					SubscriptionID: "sub2",
					Location:       "eastus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
			},
			expectedCount: 2,
		},
		{
			name: "Mixed subscriptions and regions",
			resources: []*models.Resource{
				{
					ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account1",
					SubscriptionID: "sub1",
					Location:       "eastus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
				{
					ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account2",
					SubscriptionID: "sub1",
					Location:       "eastus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
				{
					ID:             "/subscriptions/sub1/resourceGroups/rg2/providers/Microsoft.CognitiveServices/accounts/account3",
					SubscriptionID: "sub1",
					Location:       "westus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
				{
					ID:             "/subscriptions/sub2/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account4",
					SubscriptionID: "sub2",
					Location:       "eastus",
					Type:           "Microsoft.CognitiveServices/accounts",
				},
			},
			expectedCount: 3, // sub1+eastus, sub1+westus, sub2+eastus
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := scanner.groupResourcesForBatch(tt.resources)
			if len(groups) != tt.expectedCount {
				t.Errorf("groupResourcesForBatch() returned %d groups, want %d", len(groups), tt.expectedCount)
			}

			// Verify all resources are included
			totalResources := 0
			for _, group := range groups {
				totalResources += len(group.Resources)
				// Verify all resources in group have same subscription and region
				if len(group.Resources) > 0 {
					expectedSub := group.SubscriptionID
					expectedRegion := group.Region
					for _, resource := range group.Resources {
						if resource.SubscriptionID != expectedSub {
							t.Errorf("Resource %s has subscription %s, expected %s", resource.ID, resource.SubscriptionID, expectedSub)
						}
						if resource.Location != expectedRegion {
							t.Errorf("Resource %s has location %s, expected %s", resource.ID, resource.Location, expectedRegion)
						}
					}
				}
			}
			if totalResources != len(tt.resources) {
				t.Errorf("Total resources in groups = %d, want %d", totalResources, len(tt.resources))
			}
		})
	}
}

func TestAIGovScanner_groupResourcesForBatch_Grouping(t *testing.T) {
	scanner := NewScanner()

	// Test that resources are correctly grouped
	resources := []*models.Resource{
		{
			ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account1",
			SubscriptionID: "sub1",
			Location:       "eastus",
			Name:           "account1",
		},
		{
			ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account2",
			SubscriptionID: "sub1",
			Location:       "eastus",
			Name:           "account2",
		},
		{
			ID:             "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.CognitiveServices/accounts/account3",
			SubscriptionID: "sub1",
			Location:       "eastus",
			Name:           "account3",
		},
	}

	groups := scanner.groupResourcesForBatch(resources)

	if len(groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(groups))
	}

	group := groups[0]
	if group.SubscriptionID != "sub1" {
		t.Errorf("Group SubscriptionID = %s, want sub1", group.SubscriptionID)
	}
	if group.Region != "eastus" {
		t.Errorf("Group Region = %s, want eastus", group.Region)
	}
	if len(group.Resources) != 3 {
		t.Errorf("Group has %d resources, want 3", len(group.Resources))
	}

	// Verify all resources are in the group
	resourceNames := make(map[string]bool)
	for _, r := range group.Resources {
		resourceNames[r.Name] = true
	}
	for _, expectedName := range []string{"account1", "account2", "account3"} {
		if !resourceNames[expectedName] {
			t.Errorf("Expected resource %s not found in group", expectedName)
		}
	}
}

// Test that init function registers the plugin
func TestPluginRegistration(t *testing.T) {
	// The init() function should have registered the plugin
	// This is a basic sanity check that the plugin can be created
	scanner := NewScanner()
	metadata := scanner.GetMetadata()
	if metadata.Name != "ai-gov" {
		t.Errorf("Plugin registration failed or wrong plugin registered")
	}
}

func TestExtractDeploymentInfo(t *testing.T) {
	upgradeOption := armcognitiveservices.DeploymentModelVersionUpgradeOptionOnceNewDefaultVersionAvailable

	tests := []struct {
		name     string
		deploy   *armcognitiveservices.Deployment
		expected deploymentInfo
	}{
		{
			name: "All fields populated",
			deploy: &armcognitiveservices.Deployment{
				Properties: &armcognitiveservices.DeploymentProperties{
					Model: &armcognitiveservices.DeploymentModel{
						Version: to.Ptr("2024-05-13"),
						Format:  to.Ptr("OpenAI"),
					},
					VersionUpgradeOption:    &upgradeOption,
					SpilloverDeploymentName: to.Ptr("spillover-deploy"),
				},
				SKU: &armcognitiveservices.SKU{
					Capacity: to.Ptr(int32(120)),
				},
			},
			expected: deploymentInfo{
				ModelVersion:         "2024-05-13",
				ModelFormat:          "OpenAI",
				SKUCapacity:          "120",
				VersionUpgradeOption: "OnceNewDefaultVersionAvailable",
				SpilloverEnabled:     "Yes",
				SpilloverDeployment:  "spillover-deploy",
			},
		},
		{
			name: "Nil properties",
			deploy: &armcognitiveservices.Deployment{
				Properties: nil,
				SKU:        nil,
			},
			expected: deploymentInfo{
				ModelVersion:         "N/A",
				ModelFormat:          "N/A",
				SKUCapacity:          "N/A",
				VersionUpgradeOption: "N/A",
				SpilloverEnabled:     "No",
				SpilloverDeployment:  "N/A",
			},
		},
		{
			name: "Model without version",
			deploy: &armcognitiveservices.Deployment{
				Properties: &armcognitiveservices.DeploymentProperties{
					Model: &armcognitiveservices.DeploymentModel{
						Format: to.Ptr("OpenAI"),
					},
				},
				SKU: &armcognitiveservices.SKU{
					Capacity: to.Ptr(int32(50)),
				},
			},
			expected: deploymentInfo{
				ModelVersion:         "N/A",
				ModelFormat:          "OpenAI",
				SKUCapacity:          "50",
				VersionUpgradeOption: "N/A",
				SpilloverEnabled:     "No",
				SpilloverDeployment:  "N/A",
			},
		},
		{
			name: "SKU without capacity",
			deploy: &armcognitiveservices.Deployment{
				Properties: &armcognitiveservices.DeploymentProperties{},
				SKU: &armcognitiveservices.SKU{
					Name: to.Ptr("Standard"),
				},
			},
			expected: deploymentInfo{
				ModelVersion:         "N/A",
				ModelFormat:          "N/A",
				SKUCapacity:          "N/A",
				VersionUpgradeOption: "N/A",
				SpilloverEnabled:     "No",
				SpilloverDeployment:  "N/A",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDeploymentInfo(tt.deploy)
			if got.ModelVersion != tt.expected.ModelVersion {
				t.Errorf("ModelVersion = %s, want %s", got.ModelVersion, tt.expected.ModelVersion)
			}
			if got.ModelFormat != tt.expected.ModelFormat {
				t.Errorf("ModelFormat = %s, want %s", got.ModelFormat, tt.expected.ModelFormat)
			}
			if got.SKUCapacity != tt.expected.SKUCapacity {
				t.Errorf("SKUCapacity = %s, want %s", got.SKUCapacity, tt.expected.SKUCapacity)
			}
			if got.VersionUpgradeOption != tt.expected.VersionUpgradeOption {
				t.Errorf("VersionUpgradeOption = %s, want %s", got.VersionUpgradeOption, tt.expected.VersionUpgradeOption)
			}
			if got.SpilloverEnabled != tt.expected.SpilloverEnabled {
				t.Errorf("SpilloverEnabled = %s, want %s", got.SpilloverEnabled, tt.expected.SpilloverEnabled)
			}
			if got.SpilloverDeployment != tt.expected.SpilloverDeployment {
				t.Errorf("SpilloverDeployment = %s, want %s", got.SpilloverDeployment, tt.expected.SpilloverDeployment)
			}
		})
	}
}
