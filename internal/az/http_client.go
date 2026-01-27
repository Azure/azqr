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
}

// HttpClientOptions configures the HTTP client behavior
type HttpClientOptions struct {
	Timeout          time.Duration // Per-attempt timeout
	MaxRetries       int32
	OperationTimeout time.Duration      // Total operation timeout including all retries
	Scope            string             // OAuth scope for authentication
	Transport        policy.Transporter // Optional custom transport (for testing)
}

// DefaultHttpClientOptions returns the default options for production use
func DefaultHttpClientOptions(timeout time.Duration) *HttpClientOptions {
	resourceManagerEndpoint := GetResourceManagerEndpoint()
	scope := fmt.Sprintf("%s/.default", resourceManagerEndpoint)

	return &HttpClientOptions{
		Timeout:          timeout,
		MaxRetries:       5,
		OperationTimeout: timeout * 10, // 10x per-attempt for retries + backoff
		Scope:            scope,
	}
}

// NewHttpClient creates a new Azure HTTP client using Azure SDK's pipeline with built-in retry logic.
// Pass nil for opts to use default options with the specified timeout.
func NewHttpClient(cred azcore.TokenCredential, opts *HttpClientOptions) *HttpClient {
	if opts == nil {
		opts = DefaultHttpClientOptions(30 * time.Second)
	}

	// Configure retry options - Azure SDK handles exponential backoff automatically
	// Leave StatusCodes as nil to use SDK's internal defaults (408, 429, 500, 502, 503, 504)
	// This ensures the SDK's retry policy can properly handle response bodies
	retryOptions := policy.RetryOptions{
		MaxRetries:    opts.MaxRetries,
		TryTimeout:    opts.Timeout,
		RetryDelay:    4 * time.Second,  // SDK default, explicit for clarity
		MaxRetryDelay: 60 * time.Second, // SDK default, explicit for clarity
	}

	// Create client options
	// Transport timeout should be slightly longer than per-attempt timeout
	// to allow the SDK's retry policy to handle timeouts gracefully
	var transport policy.Transporter
	if opts.Transport != nil {
		transport = opts.Transport
	} else {
		transport = &http.Client{Timeout: opts.Timeout + (5 * time.Second)}
	}

	clientOpts := &policy.ClientOptions{
		Retry:     retryOptions,
		Transport: transport,
	}

	// Create bearer token authentication policy using Azure SDK's built-in implementation
	// This provides automatic token caching and refresh via TokenCredential
	authPolicy := runtime.NewBearerTokenPolicy(cred, []string{opts.Scope}, nil)

	// Create pipeline with proper policy ordering
	// PerCall policies execute once per operation (before retry logic)
	// PerRetry policies execute on each retry attempt (after retry logic)
	pipeline := runtime.NewPipeline(
		"azqr-http-client",
		"v1.0.0",
		runtime.PipelineOptions{
			PerRetry: []policy.Policy{
				authPolicy,
				throttling.NewThrottlingPolicy(),
			},
		},
		clientOpts,
	)

	return &HttpClient{
		pipeline: pipeline,
	}
}

// Do performs an HTTP GET request with automatic authentication, throttling, and retries
func (c *HttpClient) Do(ctx context.Context, url string) ([]byte, error) {
	body, _, err := c.doRequest(ctx, http.MethodGet, url, nil)
	return body, err
}

// DoPost performs an HTTP POST request with automatic authentication, throttling, and retries
func (c *HttpClient) DoPost(ctx context.Context, url string, body io.ReadSeekCloser) ([]byte, *http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, url, body)
}

// doRequest is the common implementation for HTTP requests
func (c *HttpClient) doRequest(ctx context.Context, method, url string, body io.ReadSeekCloser) ([]byte, *http.Response, error) {
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

	// Send request through pipeline (handles authentication, throttling, and retries automatically)
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
