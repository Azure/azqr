package throttling

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

// ARMLimiter rate limits Azure Resource Manager API calls
// Allows 20 operations per second with burst capacity of 100
// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/request-limits-and-throttling#regional-throttling-and-token-bucket-algorithm
var armLimiter = rate.NewLimiter(rate.Limit(20), 100)

// GraphLimiter rate limits Azure Resource Graph API calls
// Allows 3 operations per second with burst capacity of 10
// With higher burst capacity to better utilize the 5-second window
// https://learn.microsoft.com/en-us/azure/governance/resource-graph/concepts/guidance-for-throttled-requests#staggering-queries
var graphLimiter = rate.NewLimiter(rate.Limit(3), 10)

// CostLimiter rate limits Azure Cost Management API calls
// Cost Management uses QPU (Query Processing Units): 1 QPU = 1 month of data queried
// Limits: 12 QPU per 10 seconds, 60 QPU per 1 minute, 600 QPU per 1 hour
// https://learn.microsoft.com/en-us/azure/cost-management-billing/costs/manage-automation#data-latency-and-rate-limits
var costLimiter = rate.NewLimiter(rate.Limit(0.2), 1)

// ThrottlingPolicy implements policy.Policy to apply rate limiting
type ThrottlingPolicy struct{}

// NewThrottlingPolicy creates a new throttling policy
func NewThrottlingPolicy() policy.Policy {
	return &ThrottlingPolicy{}
}

// Do implements the policy.Policy interface
func (p *ThrottlingPolicy) Do(req *policy.Request) (*http.Response, error) {
	// Apply rate limiting based on URL before sending request
	url := req.Raw().URL.String()
	var err error
	switch {
	case strings.Contains(url, "Microsoft.ResourceGraph/resources"):
		log.Debug().
			Msg("Applying Graph API throttling limiter")
		err = graphLimiter.Wait(req.Raw().Context())
	case strings.Contains(url, "Microsoft.CostManagement/query"):
		log.Debug().
			Msg("Applying Cost Management API throttling limiter")
		err = costLimiter.Wait(req.Raw().Context())
	case strings.Contains(url, "prices.azure.com"):
		log.Debug().
			Msg("Applying Price API throttling limiter")
		err = graphLimiter.Wait(req.Raw().Context())
	default: // Default to ARM throttling
		log.Debug().
			Msg("Applying ARM API throttling limiter")
		err = armLimiter.Wait(req.Raw().Context())
	}
	if err != nil {
		return nil, fmt.Errorf("throttling wait failed: %w", err)
	}

	// Forward to next policy in pipeline
	return req.Next()
}
