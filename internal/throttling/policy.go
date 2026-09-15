package throttling

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

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

// Cost Management queries share an aggregate limit in addition to per-scope limits.
var costLimiter = rate.NewLimiter(rate.Every(5*time.Second), 1)

// ThrottlingPolicy applies rate limiting across Azure API clients.
type ThrottlingPolicy struct {
	// CostLimiter optionally overrides the shared Cost Management limiter.
	CostLimiter *rate.Limiter

	costMu          sync.Mutex
	costScopes      map[string]*costScopeLimiter
	lastCostCleanup time.Time
}

var sharedThrottlingPolicy = &ThrottlingPolicy{}

// NewThrottlingPolicy returns the shared throttling policy.
func NewThrottlingPolicy() policy.Policy {
	return sharedThrottlingPolicy
}

// Do implements the policy.Policy interface
func (p *ThrottlingPolicy) Do(req *policy.Request) (*http.Response, error) {
	if IsCostManagementQuery(req.Raw()) {
		return p.throttleCostRequest(req)
	}
	// Apply rate limiting based on URL before sending request
	url := req.Raw().URL.String()
	var err error
	switch {
	case strings.Contains(url, "Microsoft.ResourceGraph/resources"):
		log.Debug().
			Msg("Applying Graph API throttling limiter")
		err = graphLimiter.Wait(req.Raw().Context())
	case strings.Contains(url, "prices.azure.com"):
		// migration-advisor parity: the Retail Prices API gets NO proactive
		// rate cap. Instead we rely solely on reactive exponential backoff on
		// HTTP 429 (handled by the SDK retry policy), mirroring
		// migration-advisor's _get_with_retry. The previous 3 rps graphLimiter
		// was the dominant bottleneck for the region plugin's pricing phase.
		log.Debug().
			Msg("Bypassing proactive throttling for Retail Prices API (reactive 429 backoff only)")
		return req.Next()
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
