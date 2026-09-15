// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package az

import (
	"slices"
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
				RetryDelay:    4 * time.Second,  // SDK default, explicit for clarity
				MaxRetryDelay: 60 * time.Second, // SDK default, explicit for clarity
				MaxRetries:    5,
			},
			Cloud:            GetCloudConfiguration(),
			PerRetryPolicies: []policy.Policy{throttling.NewThrottlingPolicy()},
		},
	}
}

func costClientOptions(base policy.ClientOptions) policy.ClientOptions {
	options := base
	options.Retry.MaxRetryDelay = throttling.CostRequestTimeout
	// Pacing belongs to the operation budget, not the network-attempt timeout.
	options.Retry.TryTimeout = 0
	options.PerCallPolicies = append([]policy.Policy{
		throttling.NewCostTimeoutPolicy(throttling.CostRequestTimeout),
	}, base.PerCallPolicies...)
	options.PerRetryPolicies = slices.Clone(base.PerRetryPolicies)
	if base.Retry.TryTimeout > 0 {
		options.PerRetryPolicies = append(options.PerRetryPolicies,
			throttling.NewCostTimeoutPolicy(base.Retry.TryTimeout))
	}
	return options
}
