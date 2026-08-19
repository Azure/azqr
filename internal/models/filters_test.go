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
		{name: "valid", resourceGroupID: "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg"},
		{name: "name only", resourceGroupID: "test-rg", expectError: true, errorContains: "incorrect format"},
		{name: "missing leading slash", resourceGroupID: "subscriptions/123/resourceGroups/test-rg", expectError: true, errorContains: "incorrect format"},
		{name: "missing subscriptions segment", resourceGroupID: "/123/resourceGroups/test-rg", expectError: true, errorContains: "incorrect format"},
		{name: "wrong resource groups segment", resourceGroupID: "/subscriptions/123/resourceGroup/test-rg", expectError: true, errorContains: "incorrect format"},
		{name: "empty subscription", resourceGroupID: "/subscriptions//resourceGroups/test-rg", expectError: true, errorContains: "empty subscription ID"},
		{name: "empty resource group", resourceGroupID: "/subscriptions/123/resourceGroups/", expectError: true, errorContains: "empty resource group name"},
		{name: "too many segments", resourceGroupID: "/subscriptions/123/resourceGroups/test-rg/providers", expectError: true, errorContains: "incorrect format"},
		{name: "too few segments", resourceGroupID: "/subscriptions/123", expectError: true, errorContains: "incorrect format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceGroupID(tt.resourceGroupID)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected validation error")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Fatalf("error %q does not contain %q", err, tt.errorContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoadFiltersValidation(t *testing.T) {
	for _, id := range []string{
		"/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg",
		"/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/exclude-rg",
	} {
		if err := validateResourceGroupID(id); err != nil {
			t.Errorf("valid resource group ID failed validation: %v", err)
		}
	}
}

func TestTagFilters(t *testing.T) {
	filters := NewFilters()
	filters.Azqr.iResourceTypes = map[string]bool{"microsoft.test/widgets": true}
	filters.Azqr.iTags = map[string]string{"env": "prod", "team": "platform"}
	filters.Azqr.xTags = map[string]string{"lifecycle": "retired"}
	resourceID := "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Test/widgets/one"

	tests := []struct {
		name string
		tags map[string]string
		want bool
	}{
		{name: "all includes match", tags: map[string]string{"ENV": "prod", "Team": "platform"}},
		{name: "include missing", tags: map[string]string{"env": "prod"}, want: true},
		{name: "value comparison is exact", tags: map[string]string{"env": "Prod", "team": "platform"}, want: true},
		{name: "exclude wins", tags: map[string]string{"env": "prod", "team": "platform", "lifecycle": "retired"}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filters.Azqr.IsResourceExcluded(resourceID, tt.tags); got != tt.want {
				t.Fatalf("IsResourceExcluded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagScopeAppliesToChildResource(t *testing.T) {
	filters := NewFilters()
	filters.Azqr.iResourceTypes = map[string]bool{"microsoft.test/widgets": true}
	filters.Azqr.iTags = map[string]string{"env": "prod"}
	parent := "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Test/widgets/one"
	filters.Azqr.SetResourceScope(parent, true)

	if filters.Azqr.IsServiceExcluded(parent + "/slots/blue") {
		t.Fatal("child resource should inherit included parent scope")
	}
}
