// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package servicehealth

import (
	"testing"

	"github.com/Azure/azqr/internal/plugins"
)

func TestNewAvailabilityScanner(t *testing.T) {
	scanner := NewAvailabilityScanner()
	if scanner == nil {
		t.Fatal("NewAvailabilityScanner() should not return nil")
	}
}

func TestGetMetadata(t *testing.T) {
	scanner := NewAvailabilityScanner()
	metadata := scanner.GetMetadata()

	if metadata.Name != "service-health-availability" {
		t.Errorf("Expected name 'service-health-availability', got '%s'", metadata.Name)
	}

	if metadata.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", metadata.Version)
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
	scanner, exists := plugins.GetInternalPlugin("service-health-availability")
	if !exists {
		t.Fatal("Plugin 'service-health-availability' should be registered")
	}

	if scanner == nil {
		t.Fatal("Registered scanner should not be nil")
	}

	metadata := scanner.GetMetadata()
	if metadata.Name != "service-health-availability" {
		t.Errorf("Expected registered plugin name 'service-health-availability', got '%s'", metadata.Name)
	}
}
