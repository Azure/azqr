// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

import (
	"strings"
	"testing"
)

func TestValidateResourceGroupID(t *testing.T) {
	tests := []struct {
		name            string
		resourceGroupID string
		expectError     bool
		errorContains   string
	}{
		{
			name:            "valid resource group ID",
			resourceGroupID: "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg",
			expectError:     false,
		},
		{
			name:            "invalid format - just resource group name",
			resourceGroupID: "test-rg",
			expectError:     true,
			errorContains:   "has incorrect format",
		},
		{
			name:            "invalid format - missing leading slash",
			resourceGroupID: "subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg",
			expectError:     true,
			errorContains:   "has incorrect format",
		},
		{
			name:            "invalid format - missing subscriptions segment",
			resourceGroupID: "/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg",
			expectError:     true,
			errorContains:   "has incorrect format",
		},
		{
			name:            "invalid format - missing resourceGroups segment",
			resourceGroupID: "/subscriptions/12345678-1234-1234-1234-123456789012/test-rg",
			expectError:     true,
			errorContains:   "has incorrect format",
		},
		{
			name:            "invalid format - wrong resourceGroups segment",
			resourceGroupID: "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroup/test-rg",
			expectError:     true,
			errorContains:   "has incorrect format",
		},
		{
			name:            "invalid format - empty subscription ID",
			resourceGroupID: "/subscriptions//resourceGroups/test-rg",
			expectError:     true,
			errorContains:   "has empty subscription ID",
		},
		{
			name:            "invalid format - empty resource group name",
			resourceGroupID: "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/",
			expectError:     true,
			errorContains:   "has empty resource group name",
		},
		{
			name:            "invalid format - too many segments",
			resourceGroupID: "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg/providers",
			expectError:     true,
			errorContains:   "has incorrect format",
		},
		{
			name:            "invalid format - too few segments",
			resourceGroupID: "/subscriptions/12345678-1234-1234-1234-123456789012",
			expectError:     true,
			errorContains:   "has incorrect format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceGroupID(tt.resourceGroupID)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for resourceGroupID '%s' but got none", tt.resourceGroupID)
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s' but got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for resourceGroupID '%s': %v", tt.resourceGroupID, err)
				}
			}
		})
	}
}

func TestLoadFiltersValidation(t *testing.T) {
	// Test with valid resource group IDs - should not panic or error
	validFilters := &Filters{
		Azqr: &AzqrFilter{
			Include: &IncludeFilter{
				Subscriptions:  []string{},
				ResourceGroups: []string{"/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg"},
				ResourceTypes:  []string{},
			},
			Exclude: &ExcludeFilter{
				Subscriptions:   []string{},
				ResourceGroups:  []string{"/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/exclude-rg"},
				Services:        []string{},
				Recommendations: []string{},
			},
		},
	}

	// This should not cause any validation errors
	if err := validateResourceGroupID("/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg"); err != nil {
		t.Errorf("valid resource group ID failed validation: %v", err)
	}

	if err := validateResourceGroupID("/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/exclude-rg"); err != nil {
		t.Errorf("valid resource group ID failed validation: %v", err)
	}

	// Test validation of individual components
	for _, rgID := range validFilters.Azqr.Include.ResourceGroups {
		if err := validateResourceGroupID(rgID); err != nil {
			t.Errorf("include resource group validation failed: %v", err)
		}
	}

	for _, rgID := range validFilters.Azqr.Exclude.ResourceGroups {
		if err := validateResourceGroupID(rgID); err != nil {
			t.Errorf("exclude resource group validation failed: %v", err)
		}
	}
}
