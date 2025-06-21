package throttling

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

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
	case strings.Contains(url, "prices.azure.com"):
		err = WaitGraph(req.Raw().Context())
	case strings.Contains(url, "Microsoft.ResourceGraph/resources"):
		// Azure Resource Graph API has stricter rate limits
		err = WaitGraph(req.Raw().Context())
	default: // Default to ARM throttling
		err = WaitARM(req.Raw().Context())
	}
	if err != nil {
		return nil, fmt.Errorf("throttling wait failed: %w", err)
	}

	// Forward to next policy in pipeline
	return req.Next()
}
