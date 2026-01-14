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

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "valid string",
			input:    map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "missing key",
			input:    map[string]interface{}{},
			key:      "key",
			expected: "",
		},
		{
			name:     "non-string value",
			input:    map[string]interface{}{"key": 123},
			key:      "key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.input, tt.key)
			if result != tt.expected {
				t.Errorf("getString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		key      string
		expected float64
	}{
		{
			name:     "valid float64",
			input:    map[string]interface{}{"key": float64(123.45)},
			key:      "key",
			expected: 123.45,
		},
		{
			name:     "valid int",
			input:    map[string]interface{}{"key": 123},
			key:      "key",
			expected: 123.0,
		},
		{
			name:     "missing key",
			input:    map[string]interface{}{},
			key:      "key",
			expected: 0,
		},
		{
			name:     "non-numeric value",
			input:    map[string]interface{}{"key": "text"},
			key:      "key",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFloat(tt.input, tt.key)
			if result != tt.expected {
				t.Errorf("getFloat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		key      string
		expected int64
	}{
		{
			name:     "valid int64",
			input:    map[string]interface{}{"key": int64(123)},
			key:      "key",
			expected: 123,
		},
		{
			name:     "valid int",
			input:    map[string]interface{}{"key": 123},
			key:      "key",
			expected: 123,
		},
		{
			name:     "valid float64",
			input:    map[string]interface{}{"key": float64(123.45)},
			key:      "key",
			expected: 123,
		},
		{
			name:     "missing key",
			input:    map[string]interface{}{},
			key:      "key",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInt(tt.input, tt.key)
			if result != tt.expected {
				t.Errorf("getInt() = %v, want %v", result, tt.expected)
			}
		})
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
