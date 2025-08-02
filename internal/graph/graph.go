// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/rs/zerolog/log"
)

// GraphQueryClient provides methods to query Azure Resource Graph using HTTP client.
type GraphQueryClient struct {
	httpClient  *http.Client // HTTP client for making requests
	endpoint    string       // Resource Graph endpoint URL
	accessToken string       // Bearer token for authentication
}

// GraphResult holds the data returned from a Resource Graph query.
type GraphResult struct {
	Data []interface{} // Query result data
}

// QueryRequestOptions represents options for the Resource Graph query.
type QueryRequestOptions struct {
	ResultFormat string  `json:"resultFormat,omitempty"` // Format of the result
	Top          *int32  `json:"top,omitempty"`          // Max number of results
	SkipToken    *string `json:"skipToken,omitempty"`    // Token for pagination
}

// QueryRequest represents the payload for a Resource Graph query.
type QueryRequest struct {
	Subscriptions []string             `json:"subscriptions"` // List of subscription IDs
	Query         string               `json:"query"`         // Kusto query string
	Options       *QueryRequestOptions `json:"options"`       // Query options
}

// QueryResponse represents the response from the Resource Graph API.
type QueryResponse struct {
	Data       []interface{} `json:"data"` // Query result data
	SkipToken  *string       `json:"skipToken,omitempty"`
	Quota      int           // Value of x-ms-user-quota-remaining header as int
	RetryAfter time.Duration // Value of x-ms-user-quota-resets-after header as timespan
}

// NewGraphQuery creates a new GraphQuery using the provided TokenCredential.
func NewGraphQuery(cred azcore.TokenCredential) *GraphQueryClient {
	// Create a new HTTP client with a timeout
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Acquire an access token using the provided credential
	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to acquire Azure access token")
	}

	// Set the access token string
	accessToken := token.Token

	return &GraphQueryClient{
		httpClient:  httpClient,
		endpoint:    "https://management.azure.com/providers/Microsoft.ResourceGraph/resources?api-version=2021-03-01",
		accessToken: accessToken,
	}
}

// Query executes a Resource Graph query for the given subscriptions and query string.
// It handles batching and pagination.
func (q *GraphQueryClient) Query(ctx context.Context, query string, subscriptions []*string) *GraphResult {
	result := GraphResult{
		Data: make([]interface{}, 0),
	}

	// Convert []*string to []string for serialization
	subscriptionIDs := make([]string, len(subscriptions))
	for i, s := range subscriptions {
		if s != nil {
			subscriptionIDs[i] = *s
		}
	}

	// Run the query in batches of 300 subscriptions
	batchSize := 300
	for i := 0; i < len(subscriptionIDs); i += batchSize {
		j := i + batchSize
		if j > len(subscriptionIDs) {
			j = len(subscriptionIDs)
		}

		format := "objectArray"
		options := &QueryRequestOptions{
			ResultFormat: format,
			Top:          to.Ptr(int32(1000)),
		}

		var skipToken *string = nil
		for ok := true; ok; ok = skipToken != nil {
			options.SkipToken = skipToken
			request := QueryRequest{
				Subscriptions: subscriptionIDs[i:j],
				Query:         query,
				Options:       options,
			}

			resp, err := q.retry(ctx, 3, request)
			if err == nil {
				result.Data = append(result.Data, resp.Data...)
				skipToken = resp.SkipToken
			} else {
				log.Fatal().Err(err).Msgf("Failed to run Resource Graph query: %s", query)
				return nil
			}
		}
	}
	return &result
}

// retry executes the Resource Graph query with retries when throttling occurs.
// Returns the QueryResponse or error.
func (q *GraphQueryClient) retry(ctx context.Context, attempts int, request QueryRequest) (*QueryResponse, error) {
	var err error
	for i := 0; ; i++ {
		resp, err := q.doRequest(ctx, request)
		if err == nil {
			return resp, nil
		}

		if i >= (attempts - 1) {
			log.Error().Msgf("Retry limit reached for Graph query: %s", request.Query)
			break
		}

		// Quota limit reached, sleep for the duration specified in the response header
		if resp.Quota == 0 {
			duration := resp.RetryAfter
			log.Debug().Msgf("Graph query quota limit reached. Sleeping for %s", duration)
			time.Sleep(duration)
		}

	}
	return nil, err
}

// doRequest sends the HTTP request to the Resource Graph API and returns the response.
func (q *GraphQueryClient) doRequest(ctx context.Context, request QueryRequest) (*QueryResponse, error) {
	// Serialize request to JSON
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, q.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+q.accessToken)

	// Wait for a token from the burstLimiter channel before making the request
	<-throttling.GraphLimiter

	// Send request
	resp, err := q.httpClient.Do(req)

	// Parse response JSON
	queryResp := QueryResponse{}

	// Extract quota headers and set them in the QueryResponse struct
	// Parse x-ms-user-quota-remaining as int
	quotaStr := resp.Header.Get("x-ms-user-quota-remaining")
	if quotaStr != "" {
		quota, err := strconv.Atoi(quotaStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse quota header: %w", err)
		}
		queryResp.Quota = quota
	}

	// Parse x-ms-user-quota-resets-after as timespan in "hh:mm:ss" format
	retryAfterStr := resp.Header.Get("x-ms-user-quota-resets-after")
	if retryAfterStr != "" {
		// If time.ParseDuration fails, fallback to manual parsing
		var h, m, s int
		_, scanErr := fmt.Sscanf(retryAfterStr, "%d:%d:%d", &h, &m, &s)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to parse retry-after header: %w", scanErr)
		}
		queryResp.RetryAfter = time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
	}

	log.Debug().Msgf("Graph query quota remaining: %d, Retry after: %s", queryResp.Quota, queryResp.RetryAfter)

	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal().Err(err).Msg("Failed to close response body")
		}
	}()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &queryResp, fmt.Errorf("received non-2xx status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &queryResp, nil
}
