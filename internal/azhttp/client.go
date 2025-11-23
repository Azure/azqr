// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azhttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/rs/zerolog/log"
)

// Client wraps Azure SDK pipeline for authenticated HTTP requests with built-in retry logic
type Client struct {
	pipeline runtime.Pipeline
	cred     azcore.TokenCredential
}

// throttlingPolicy implements policy.Policy to apply rate limiting
type throttlingPolicy struct{}

// newThrottlingPolicy creates a new throttling policy
func newThrottlingPolicy() policy.Policy {
	return &throttlingPolicy{}
}

// Do implements the policy.Policy interface
func (p *throttlingPolicy) Do(req *policy.Request) (*http.Response, error) {
	// Apply rate limiting based on URL before sending request
	url := req.Raw().URL.String()
	if strings.Contains(url, "prices.azure.com") {
		if err := throttling.WaitRetailPrices(req.Raw().Context()); err != nil {
			return nil, fmt.Errorf("throttling wait failed: %w", err)
		}
	} else if strings.Contains(url, "Microsoft.ResourceGraph/resources") {
		// Azure Resource Graph API has stricter rate limits
		if err := throttling.WaitGraph(req.Raw().Context()); err != nil {
			return nil, fmt.Errorf("throttling wait failed: %w", err)
		}
	} else { // Default to ARM throttling
		if err := throttling.WaitARM(req.Raw().Context()); err != nil {
			return nil, fmt.Errorf("throttling wait failed: %w", err)
		}
	}

	// Forward to next policy in pipeline
	return req.Next()
}

// NewClient creates a new Azure HTTP client using Azure SDK's pipeline with built-in retry logic
func NewClient(cred azcore.TokenCredential, timeout time.Duration) *Client {
	// Configure retry options - Azure SDK handles exponential backoff automatically
	retryOptions := policy.RetryOptions{
		MaxRetries:    3,
		TryTimeout:    timeout,
		RetryDelay:    2 * time.Second,
		MaxRetryDelay: 60 * time.Second,
		StatusCodes: []int{
			http.StatusRequestTimeout,      // 408
			http.StatusTooManyRequests,     // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout,      // 504
		},
	}

	// Create client options
	clientOpts := &policy.ClientOptions{
		Retry:     retryOptions,
		Transport: &http.Client{Timeout: timeout},
	}

	// Create pipeline with authentication and retry policies
	pipeline := runtime.NewPipeline(
		"azqr-http-client",
		"v1.0.0",
		runtime.PipelineOptions{
			PerRetry: []policy.Policy{
				newThrottlingPolicy(), // Apply throttling per retry attempt
			},
		},
		clientOpts,
	)

	return &Client{
		pipeline: pipeline,
		cred:     cred,
	}
}

// Do performs an HTTP GET request with automatic authentication, throttling, and retries
func (c *Client) Do(ctx context.Context, url string, needsAuth bool, maxRetries int) ([]byte, error) {
	// Create HTTP request
	req, err := runtime.NewRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if needed
	if needsAuth {
		token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{"https://management.azure.com/.default"},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
		req.Raw().Header.Set("Authorization", "Bearer "+token.Token)
	}

	// Send request through pipeline (handles retries automatically)
	resp, err := c.pipeline.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("Failed to close response body")
		}
	}()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
			URL:        url,
		}
	}

	log.Debug().Msgf("Successfully executed request to %s (status: %d)", url, resp.StatusCode)
	return body, nil
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Body       string
	URL        string
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d from %s: %s", e.StatusCode, e.URL, e.Body)
}
