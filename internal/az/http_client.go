// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package az

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/rs/zerolog/log"
)

// HttpClient wraps Azure SDK pipeline for authenticated HTTP requests with built-in retry logic
type HttpClient struct {
	pipeline runtime.Pipeline
	cred     azcore.TokenCredential
}

// HttpClientOptions configures the HTTP client behavior
type HttpClientOptions struct {
	Timeout       time.Duration
	MaxRetries    int32
	RetryDelay    time.Duration
	MaxRetryDelay time.Duration
}

// DefaultHttpClientOptions returns the default options for production use
func DefaultHttpClientOptions(timeout time.Duration) *HttpClientOptions {
	return &HttpClientOptions{
		Timeout:       timeout,
		MaxRetries:    3,
		RetryDelay:    2 * time.Second,
		MaxRetryDelay: 60 * time.Second,
	}
}

// NewHttpClient creates a new Azure HTTP client using Azure SDK's pipeline with built-in retry logic.
// Pass nil for opts to use default options with the specified timeout.
func NewHttpClient(cred azcore.TokenCredential, opts *HttpClientOptions) *HttpClient {
	if opts == nil {
		opts = DefaultHttpClientOptions(30 * time.Second)
	}

	// Configure retry options - Azure SDK handles exponential backoff automatically
	retryOptions := policy.RetryOptions{
		MaxRetries:    opts.MaxRetries,
		TryTimeout:    opts.Timeout,
		RetryDelay:    opts.RetryDelay,
		MaxRetryDelay: opts.MaxRetryDelay,
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
		Transport: &http.Client{Timeout: opts.Timeout},
	}

	// Create pipeline with authentication and retry policies
	pipeline := runtime.NewPipeline(
		"azqr-http-client",
		"v1.0.0",
		runtime.PipelineOptions{
			PerRetry: []policy.Policy{
				throttling.NewThrottlingPolicy(), // Apply throttling per retry attempt
			},
		},
		clientOpts,
	)

	return &HttpClient{
		pipeline: pipeline,
		cred:     cred,
	}
}

// Do performs an HTTP GET request with automatic authentication, throttling, and retries
// Pass scope as nil to skip authentication, or empty string to use default scope
func (c *HttpClient) Do(ctx context.Context, url string, scope *string) ([]byte, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, url, nil, scope)
	return body, err
}

// DoPost performs an HTTP POST request with automatic authentication, throttling, and retries
// Pass scope as nil to skip authentication, or empty string/specific scope to authenticate
func (c *HttpClient) DoPost(ctx context.Context, url string, body io.ReadSeekCloser, scope *string) ([]byte, *http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, url, body, scope)
}

// doRequest is the common implementation for HTTP requests
func (c *HttpClient) doRequest(ctx context.Context, method, url string, body io.ReadSeekCloser, scope *string) ([]byte, *http.Response, error) {
	// Create HTTP request
	req, err := runtime.NewRequest(ctx, method, url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set request body for POST/PUT/PATCH
	if body != nil {
		if err := req.SetBody(body, "application/json"); err != nil {
			return nil, nil, fmt.Errorf("failed to set request body: %w", err)
		}
	}

	// Add authentication if scope is provided (nil means no auth)
	if scope != nil {
		scopeVal := *scope
		if scopeVal == "" {
			scopeVal = "https://management.azure.com/.default"
		}
		token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{scopeVal},
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get access token: %w", err)
		}
		req.Raw().Header.Set("Authorization", "Bearer "+token.Token)
	}

	// Send request through pipeline (handles retries automatically)
	resp, err := c.pipeline.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("Failed to close response body")
		}
	}()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, resp, &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
			URL:        url,
		}
	}

	log.Debug().Msgf("Successfully executed %s request to %s (status: %d)", method, url, resp.StatusCode)
	return respBody, resp, nil
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
