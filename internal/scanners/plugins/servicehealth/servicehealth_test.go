// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package servicehealth

import (
	"testing"

	"github.com/Azure/azqr/internal/plugins"
)

func TestNewServiceHealthScanner(t *testing.T) {
	scanner := NewServiceHealthScanner()
	if scanner == nil {
		t.Fatal("NewServiceHealthScanner() should not return nil")
	}
}

func TestGetMetadata(t *testing.T) {
	scanner := NewServiceHealthScanner()
	metadata := scanner.GetMetadata()

	if metadata.Name != "service-health" {
		t.Errorf("Expected name 'service-health', got '%s'", metadata.Name)
	}

	if metadata.Version != "0.1.0-beta" {
		t.Errorf("Expected version '0.1.0-beta', got '%s'", metadata.Version)
	}

	if metadata.Type != plugins.PluginTypeInternal {
		t.Errorf("Expected type PluginTypeInternal, got %v", metadata.Type)
	}

	if len(metadata.ColumnMetadata) != 6 {
		t.Errorf("Expected 6 column metadata entries, got %d", len(metadata.ColumnMetadata))
	}
}

func TestPluginRegistration(t *testing.T) {
	// Check that the plugin is registered
	scanner, exists := plugins.GetInternalPlugin("service-health")
	if !exists {
		t.Fatal("Plugin 'service-health' should be registered")
	}

	if scanner == nil {
		t.Fatal("Registered scanner should not be nil")
	}

	metadata := scanner.GetMetadata()
	if metadata.Name != "service-health" {
		t.Errorf("Expected registered plugin name 'service-health', got '%s'", metadata.Name)
	}
}
