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
