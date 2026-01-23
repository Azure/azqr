// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"context"
	"strings"
	"testing"
)

func TestCreateAzqrTools(t *testing.T) {
	tools := createTools()

	if len(tools) == 0 {
		t.Fatal("expected tools to be created, got none")
	}

	// Verify we have the expected tools
	expectedTools := []string{"scan", "get-recommendations-catalog", "get-supported-services"}
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("expected tool %s not found", expected)
		}
	}

	t.Logf("Successfully created %d tools", len(tools))
}

func TestExecuteScanTool(t *testing.T) {
	t.Skip("skipping scan execution - requires Azure credentials")

	ctx := context.Background()
	params := ScanParams{
		Services: []string{"vm", "sql"},
		Defender: true,
		Advisor:  true,
		Cost:     true,
		Policy:   false,
		Arc:      false,
		Mask:     true,
	}

	result, err := executeScanTool(ctx, params)
	if err != nil {
		t.Errorf("executeScanTool() error = %v", err)
	}

	if !strings.Contains(result, "Scan completed") {
		t.Errorf("expected result to contain 'Scan completed', got %s", result)
	}
}

func TestGetRecommendationsCatalog(t *testing.T) {
	result, err := getRecommendationsCatalog()
	if err != nil {
		t.Fatalf("getRecommendationsCatalog() error = %v", err)
	}

	if !strings.Contains(result, "Recommendations List") {
		t.Errorf("expected result to contain 'Recommendations List', got %s", result)
	}

	// MCP server returns markdown format with resource type count
	if !strings.Contains(result, "Total Supported Azure Resource Types:") {
		t.Errorf("expected result to contain resource type count, got %s", result)
	}

	t.Logf("Recommendations catalog output has %d characters", len(result))
}

func TestGetSupportedServices(t *testing.T) {
	result, err := getSupportedServices()
	if err != nil {
		t.Fatalf("getSupportedServices() error = %v", err)
	}

	// MCP server returns table format with abbreviations and resource types
	if !strings.Contains(result, "Abbreviation") || !strings.Contains(result, "Resource Type") {
		t.Errorf("expected result to contain table headers, got %s", result)
	}

	// Verify some common Azure service types are present
	if !strings.Contains(result, "Microsoft.") {
		t.Errorf("expected result to contain Azure resource types (Microsoft.*), got %s", result)
	}

	t.Logf("Services list output has %d characters", len(result))
}

func TestScanParams(t *testing.T) {
	params := ScanParams{
		Services: []string{"vm", "sql"},
		Defender: true,
		Advisor:  true,
		Cost:     false,
		Policy:   true,
		Arc:      false,
		Mask:     true,
	}

	if len(params.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(params.Services))
	}

	if !params.Defender {
		t.Error("expected Defender to be true")
	}

	if !params.Advisor {
		t.Error("expected Advisor to be true")
	}

	if params.Cost {
		t.Error("expected Cost to be false")
	}

	if !params.Policy {
		t.Error("expected Policy to be true")
	}

	if params.Arc {
		t.Error("expected Arc to be false")
	}

	if !params.Mask {
		t.Error("expected Mask to be true")
	}
}

func TestEmptyParams(t *testing.T) {
	params := EmptyParams{}
	// EmptyParams should have no fields
	// This test just verifies it compiles
	_ = params
}

func TestToolIntegration(t *testing.T) {
	// Test that tools can be created and have proper structure
	tools := createTools()

	for _, tool := range tools {
		if tool.Name == "" {
			t.Error("tool has empty name")
		}

		if tool.Description == "" {
			t.Errorf("tool %s has empty description", tool.Name)
		}

		if tool.Handler == nil {
			t.Errorf("tool %s has nil handler", tool.Name)
		}

		t.Logf("Tool: %s - %s", tool.Name, tool.Description)
	}
}
