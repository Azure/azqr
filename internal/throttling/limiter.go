package throttling

import (
	"context"

	"golang.org/x/time/rate"
)

// ARMLimiter rate limits Azure Resource Manager API calls
// Allows 20 operations per second with burst capacity of 100
// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/request-limits-and-throttling#regional-throttling-and-token-bucket-algorithm
var ARMLimiter = rate.NewLimiter(rate.Limit(20), 100)

// GraphLimiter rate limits Azure Resource Graph API calls
// Allows 3 operations per second with burst capacity of 10
// With higher burst capacity to better utilize the 5-second window
// https://learn.microsoft.com/en-us/azure/governance/resource-graph/concepts/guidance-for-throttled-requests#staggering-queries
var GraphLimiter = rate.NewLimiter(rate.Limit(3), 10)

// CostLimiter rate limits Azure Cost Management API calls
// Cost Management uses QPU (Query Processing Units): 1 QPU = 1 month of data queried
// Limits: 12 QPU per 10 seconds, 60 QPU per 1 minute, 600 QPU per 1 hour
// Assuming 3 months per query (default): rate = 1 QPU/sec / 3 QPU = 0.33 req/sec, burst = 12 QPU / 3 = 4
// Using slightly more conservative rate to stay well within limits
// https://learn.microsoft.com/en-us/azure/cost-management-billing/costs/manage-automation#data-latency-and-rate-limits
var CostLimiter = rate.NewLimiter(rate.Limit(1), 10)

// WaitARM waits for permission to make an ARM API call using the provided context
func WaitARM(ctx context.Context) error {
	return ARMLimiter.Wait(ctx)
}

// WaitGraph waits for permission to make a Graph API call using the provided context
func WaitGraph(ctx context.Context) error {
	return GraphLimiter.Wait(ctx)
}

// WaitCost waits for permission to make a Cost Management API call using the provided context
func WaitCost(ctx context.Context) error {
	return CostLimiter.Wait(ctx)
}
