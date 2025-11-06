package throttling

import (
	"context"

	"golang.org/x/time/rate"
)

// ARMLimiter rate limits Azure Resource Manager API calls
// Allows 3 operations per second with burst capacity of 100
// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/request-limits-and-throttling#regional-throttling-and-token-bucket-algorithm
var ARMLimiter = rate.NewLimiter(rate.Limit(3), 100)

// GraphLimiter rate limits Azure Resource Graph API calls
// Allows 3 operations per second with burst capacity of 10
// With higher burst capacity to better utilize the 5-second window
// https://learn.microsoft.com/en-us/azure/governance/resource-graph/concepts/guidance-for-throttled-requests#staggering-queries
var GraphLimiter = rate.NewLimiter(rate.Limit(2), 10)

// WaitARM waits for permission to make an ARM API call using the provided context
func WaitARM(ctx context.Context) error {
	return ARMLimiter.Wait(ctx)
}

// WaitGraph waits for permission to make a Graph API call using the provided context
func WaitGraph(ctx context.Context) error {
	return GraphLimiter.Wait(ctx)
}
