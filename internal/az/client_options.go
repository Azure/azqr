// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package az

import (
	"time"

	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// NewDefaultClientOptions creates ARM client options with standard retry configuration and throttling policy
// This provides consistent retry behavior and throttling across all Azure SDK client instances
func NewDefaultClientOptions() *arm.ClientOptions {
	return &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Retry: policy.RetryOptions{
				// Only if the HTTP response does not contain a Retry-After header
				RetryDelay:    1 * time.Second, // More aggressive than default (4s)
				MaxRetries:    3,
				MaxRetryDelay: 60 * time.Second,
			},
			Cloud:            GetCloudConfiguration(),
			PerRetryPolicies: []policy.Policy{throttling.NewThrottlingPolicy()},
		},
	}
}
