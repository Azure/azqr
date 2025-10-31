// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadYamlPluginValid tests loading a valid YAML plugin
func TestLoadYamlPluginValid(t *testing.T) {
	// Create a temporary YAML plugin file
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test-plugin.yaml")

	yamlContent := `---
name: test-plugin
version: 1.0.0
description: Test YAML plugin
author: Test Author
queries:
  - aprlGuid: test-guid-001
    description: Test recommendation
    longDescription: Test long description
    recommendationControl: High Availability
    recommendationImpact: High
    learnMoreLink:
      - name: Test Doc
        url: https://example.com/test
    recommendationResourceType: Microsoft.Test/resources
    query: |
      resources
      | where type == "microsoft.test/resources"
      | project id, name, resourceGroup, location
`

	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	// Load the plugin
	plugin, recommendations, err := LoadYamlPlugin(yamlPath)
	if err != nil {
		t.Fatalf("LoadYamlPlugin failed: %v", err)
	}

	// Verify plugin metadata
	if plugin.Metadata.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", plugin.Metadata.Name)
	}
	if plugin.Metadata.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", plugin.Metadata.Version)
	}
	if plugin.Metadata.Description != "Test YAML plugin" {
		t.Errorf("Expected description 'Test YAML plugin', got '%s'", plugin.Metadata.Description)
	}
	if plugin.Metadata.Type != PluginTypeYaml {
		t.Errorf("Expected type PluginTypeYaml, got %v", plugin.Metadata.Type)
	}

	// Verify recommendations are correctly converted
	if len(recommendations) != 1 {
		t.Errorf("Expected 1 recommendation, got %d", len(recommendations))
	}
	if len(plugin.YamlRecommendations) != 1 {
		t.Errorf("Expected 1 YAML recommendation in plugin, got %d", len(plugin.YamlRecommendations))
	}

	// Verify recommendation details
	rec := recommendations[0]
	if rec.RecommendationID != "test-guid-001" {
		t.Errorf("Expected APRL GUID 'test-guid-001', got '%s'", rec.RecommendationID)
	}
	if rec.Recommendation != "Test recommendation" {
		t.Errorf("Expected description 'Test recommendation', got '%s'", rec.Recommendation)
	}
	if rec.ResourceType != "Microsoft.Test/resources" {
		t.Errorf("Expected resource type 'Microsoft.Test/resources', got '%s'", rec.ResourceType)
	}
}

// TestLoadYamlPluginInvalidFile tests loading from a non-existent file
func TestLoadYamlPluginInvalidFile(t *testing.T) {
	_, _, err := LoadYamlPlugin("/nonexistent/path/plugin.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestLoadYamlPluginInvalidYaml tests loading invalid YAML
func TestLoadYamlPluginInvalidYaml(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidYaml := `name: test
version: 1.0.0
queries:
  - this is not valid yaml structure
    missing proper formatting
`

	if err := os.WriteFile(yamlPath, []byte(invalidYaml), 0644); err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	_, _, err := LoadYamlPlugin(yamlPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

// TestLoadYamlPluginMissingRequired tests loading YAML missing required fields
func TestLoadYamlPluginMissingRequired(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "missing.yaml")

	missingFields := `description: Missing name and version
queries:
  - aprlGuid: test-guid
    query: resources
`

	if err := os.WriteFile(yamlPath, []byte(missingFields), 0644); err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}

	_, _, err := LoadYamlPlugin(yamlPath)
	if err == nil {
		t.Error("Expected error for missing required fields, got nil")
	}
}

// TestLoadYamlPluginExternalKqlFile tests loading a plugin with external KQL file
func TestLoadYamlPluginExternalKqlFile(t *testing.T) {
	tmpDir := t.TempDir()
	kqlDir := filepath.Join(tmpDir, "kql")
	if err := os.MkdirAll(kqlDir, 0755); err != nil {
		t.Fatalf("Failed to create kql directory: %v", err)
	}

	// Create external KQL file
	kqlPath := filepath.Join(kqlDir, "test-query.kql")
	kqlContent := `resources
| where type == "microsoft.test/resources"
| where tags["environment"] == "production"
| project id, name, resourceGroup, location`

	if err := os.WriteFile(kqlPath, []byte(kqlContent), 0644); err != nil {
		t.Fatalf("Failed to create KQL file: %v", err)
	}

	// Create YAML plugin referencing external KQL
	yamlPath := filepath.Join(tmpDir, "plugin.yaml")
	yamlContent := `name: test-external-kql
version: 1.0.0
description: Test plugin with external KQL
queries:
  - aprlGuid: test-guid-external
    description: Test with external KQL
    recommendationControl: Security
    recommendationImpact: Medium
    learnMoreLink:
      - name: Example Doc
        url: https://example.com
    recommendationResourceType: Microsoft.Test/resources
    queryFile: ./kql/test-query.kql
`

	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create YAML file: %v", err)
	}

	// Load the plugin
	_, recommendations, err := LoadYamlPlugin(yamlPath)
	if err != nil {
		t.Fatalf("LoadYamlPlugin failed: %v", err)
	}

	// Verify the query was loaded from external file
	if len(recommendations) != 1 {
		t.Fatalf("Expected 1 recommendation, got %d", len(recommendations))
	}

	rec := recommendations[0]
	if !containsString(rec.GraphQuery, "tags[\"environment\"]") {
		t.Error("Expected query to contain content from external KQL file")
	}
}

// TestLoadYamlPluginMultipleQueries tests loading a plugin with multiple queries
func TestLoadYamlPluginMultipleQueries(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "multi.yaml")

	yamlContent := `name: multi-query-plugin
version: 1.0.0
description: Plugin with multiple queries
queries:
  - aprlGuid: query-001
    description: First query
    recommendationControl: Security
    recommendationImpact: High
    recommendationResourceType: Microsoft.Compute/virtualMachines
    query: resources | where type == "microsoft.compute/virtualmachines"
  
  - aprlGuid: query-002
    description: Second query
    recommendationControl: Cost Optimization
    recommendationImpact: Medium
    recommendationResourceType: Microsoft.Storage/storageAccounts
    query: resources | where type == "microsoft.storage/storageaccounts"
  
  - aprlGuid: query-003
    description: Third query
    recommendationControl: Operational Excellence
    recommendationImpact: Low
    recommendationResourceType: Microsoft.Network/publicIPAddresses
    query: resources | where type == "microsoft.network/publicipaddresses"
`

	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create YAML file: %v", err)
	}

	_, recommendations, err := LoadYamlPlugin(yamlPath)
	if err != nil {
		t.Fatalf("LoadYamlPlugin failed: %v", err)
	}

	if len(recommendations) != 3 {
		t.Errorf("Expected 3 recommendations, got %d", len(recommendations))
	}

	// Verify each recommendation
	expectedGuids := []string{"query-001", "query-002", "query-003"}
	for i, expectedGuid := range expectedGuids {
		if recommendations[i].RecommendationID != expectedGuid {
			t.Errorf("Recommendation %d: expected GUID '%s', got '%s'", i, expectedGuid, recommendations[i].RecommendationID)
		}
	}
}

// TestDiscoverYamlPluginsEmpty tests discovery in empty directories
func TestDiscoverYamlPluginsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	plugins, err := discoverYamlPlugins([]string{tmpDir})

	if err != nil {
		t.Fatalf("discoverYamlPlugins failed: %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins, got %d", len(plugins))
	}
}

// TestDiscoverYamlPluginsMultiple tests discovering multiple YAML plugins
func TestDiscoverYamlPluginsMultiple(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple YAML plugins
	plugins := []struct {
		name     string
		filename string
	}{
		{"plugin-one", "plugin-one.yaml"},
		{"plugin-two", "plugin-two.yaml"},
		{"plugin-three", "custom-checks.yaml"},
	}

	for _, p := range plugins {
		yamlPath := filepath.Join(tmpDir, p.filename)
		yamlContent := `name: ` + p.name + `
version: 1.0.0
description: Test plugin ` + p.name + `
queries:
  - aprlGuid: test-guid
    description: Test
    recommendationControl: Security
    recommendationImpact: High
    recommendationResourceType: Microsoft.Test/resources
    query: resources
`
		if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to create YAML file: %v", err)
		}
	}

	// Discover plugins
	discovered, err := discoverYamlPlugins([]string{tmpDir})
	if err != nil {
		t.Fatalf("discoverYamlPlugins failed: %v", err)
	}

	if len(discovered) != 3 {
		t.Errorf("Expected 3 plugins, got %d", len(discovered))
	}

	// Verify all plugins were discovered
	foundNames := make(map[string]bool)
	for _, plugin := range discovered {
		foundNames[plugin.Metadata.Name] = true
	}

	for _, p := range plugins {
		if !foundNames[p.name] {
			t.Errorf("Plugin '%s' was not discovered", p.name)
		}
	}
}

// TestDiscoverYamlPluginsIgnoresNonYaml tests that non-YAML files are ignored
func TestDiscoverYamlPluginsIgnoresNonYaml(t *testing.T) {
	tmpDir := t.TempDir()

	// Create various files
	files := map[string]string{
		"plugin.yaml": validPluginYaml(),
		"readme.md":   "# Plugin Documentation",
		"config.json": `{"setting": "value"}`,
		"script.sh":   "#!/bin/bash\necho test",
		"data.txt":    "some text data",
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	discovered, err := discoverYamlPlugins([]string{tmpDir})
	if err != nil {
		t.Fatalf("discoverYamlPlugins failed: %v", err)
	}

	// Should only discover the one YAML plugin
	if len(discovered) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(discovered))
	}
}

// TestDiscoverYamlPluginsDuplicates tests that duplicate plugins are handled correctly
func TestDiscoverYamlPluginsDuplicates(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create same plugin in two locations
	yamlContent := validPluginYaml()
	path1 := filepath.Join(tmpDir, "plugin.yaml")
	path2 := filepath.Join(subDir, "plugin.yaml")

	if err := os.WriteFile(path1, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create first YAML file: %v", err)
	}
	if err := os.WriteFile(path2, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create second YAML file: %v", err)
	}

	discovered, err := discoverYamlPlugins([]string{tmpDir})
	if err != nil {
		t.Fatalf("discoverYamlPlugins failed: %v", err)
	}

	// Should only get one plugin (first found wins)
	if len(discovered) != 1 {
		t.Errorf("Expected 1 plugin (duplicate removed), got %d", len(discovered))
	}
}

// Helper functions

func validPluginYaml() string {
	return `name: test-plugin
version: 1.0.0
description: Test YAML plugin
author: Test Author
queries:
  - aprlGuid: test-guid
    description: Test recommendation
    longDescription: Test long description
    recommendationControl: Security
    recommendationImpact: High
    learnMoreLink:
      - name: Test Doc
        url: https://example.com/test
    recommendationResourceType: Microsoft.Test/resources
    query: resources | where type == "microsoft.test/resources"
`
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
