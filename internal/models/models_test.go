// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

import (
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func TestShouldSkipError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "MissingRegistrationForResourceProvider",
			err: &azcore.ResponseError{
				ErrorCode: "MissingRegistrationForResourceProvider",
			},
			expected: true,
		},
		{
			name: "MissingSubscriptionRegistration",
			err: &azcore.ResponseError{
				ErrorCode: "MissingSubscriptionRegistration",
			},
			expected: true,
		},
		{
			name: "DisallowedOperation",
			err: &azcore.ResponseError{
				ErrorCode: "DisallowedOperation",
			},
			expected: true,
		},
		{
			name: "ResourceNotFound",
			err: &azcore.ResponseError{
				ErrorCode: "ResourceNotFound",
			},
			expected: false,
		},
		{
			name:     "non-ResponseError",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldSkipError(tt.err)
			if got != tt.expected {
				t.Errorf("ShouldSkipError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetSubscriptionFromResourceID(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		want       string
	}{
		{
			name:       "standard ARM resource ID",
			resourceID: "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			want:       "12345678-1234-1234-1234-123456789012",
		},
		{
			name:       "empty string",
			resourceID: "",
			want:       "",
		},
		{
			name:       "malformed ID",
			resourceID: "/subscriptions/",
			want:       "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSubscriptionFromResourceID(tt.resourceID)
			if got != tt.want {
				t.Errorf("GetSubscriptionFromResourceID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetResourceGroupFromResourceID(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		want       string
	}{
		{
			name:       "standard ARM resource ID",
			resourceID: "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			want:       "myRG",
		},
		{
			name:       "empty string",
			resourceID: "",
			want:       "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetResourceGroupFromResourceID(tt.resourceID)
			if got != tt.want {
				t.Errorf("GetResourceGroupFromResourceID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetResourceGroupIDFromResourceID(t *testing.T) {
	resourceID := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM"
	want := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myRG"
	got := GetResourceGroupIDFromResourceID(resourceID)
	if got != want {
		t.Errorf("GetResourceGroupIDFromResourceID() = %q, want %q", got, want)
	}
}

func TestGetResourceTypeFromResourceID(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		want       string
	}{
		{
			name:       "standard ARM resource ID",
			resourceID: "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			want:       "Microsoft.Compute/virtualMachines",
		},
		{
			name:       "empty string",
			resourceID: "",
			want:       "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetResourceTypeFromResourceID(tt.resourceID)
			if got != tt.want {
				t.Errorf("GetResourceTypeFromResourceID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetResourceNameFromResourceID(t *testing.T) {
	resourceID := "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM"
	want := "myVM"
	got := GetResourceNameFromResourceID(resourceID)
	if got != want {
		t.Errorf("GetResourceNameFromResourceID() = %q, want %q", got, want)
	}
}

func TestRecommendationConstants(t *testing.T) {
	// Test Impact constants
	if ImpactHigh != "High" {
		t.Errorf("ImpactHigh = %s, want 'High'", ImpactHigh)
	}
	if ImpactMedium != "Medium" {
		t.Errorf("ImpactMedium = %s, want 'Medium'", ImpactMedium)
	}
	if ImpactLow != "Low" {
		t.Errorf("ImpactLow = %s, want 'Low'", ImpactLow)
	}

	// Test Category constants
	categories := []RecommendationCategory{
		CategoryBusinessContinuity,
		CategoryDisasterRecovery,
		CategoryGovernance,
		CategoryHighAvailability,
		CategoryMonitoringAndAlerting,
		CategoryOtherBestPractices,
		CategoryScalability,
		CategorySecurity,
		CategoryServiceUpgradeAndRetirement,
		CategorySLA,
	}

	for _, cat := range categories {
		if string(cat) == "" {
			t.Errorf("Category constant should not be empty")
		}
	}
}
