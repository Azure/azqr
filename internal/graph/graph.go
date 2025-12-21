// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

// GraphQueryClient provides methods to query Azure Resource Graph using HTTP client.
type GraphQueryClient struct {
	httpClient *az.HttpClient // Azure HTTP client with built-in retry logic
	endpoint   string         // Resource Graph endpoint URL
}

// GraphResult holds the data returned from a Resource Graph query.
type GraphResult struct {
	Data []interface{} // Query result data
}

// QueryRequestOptions represents options for the Resource Graph query.
type QueryRequestOptions struct {
	ResultFormat             string  `json:"resultFormat,omitempty"`             // Format of the result
	Top                      *int32  `json:"$top,omitempty"`                     // Max number of results
	SkipToken                *string `json:"$skipToken,omitempty"`               // Token for pagination
	AuthorizationScopeFilter *string `json:"authorizationScopeFilter,omitempty"` // Filter by authorization scope
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
	SkipToken  *string       `json:"$skipToken,omitempty"`
	Quota      int           // Value of x-ms-user-quota-remaining header as int
	RetryAfter time.Duration // Value of x-ms-user-quota-resets-after header as timespan
}

// NewGraphQuery creates a new GraphQuery using the provided TokenCredential.
func NewGraphQuery(cred azcore.TokenCredential) *GraphQueryClient {
	// Create Azure HTTP client with built-in retry and throttling
	httpClient := az.NewHttpClient(cred, 60*time.Second)

	resourceManagerEndpoint := az.GetResourceManagerEndpoint()

	return &GraphQueryClient{
		httpClient: httpClient,
		endpoint:   fmt.Sprintf("%s/providers/Microsoft.ResourceGraph/resources?api-version=2024-04-01", resourceManagerEndpoint),
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
			ResultFormat:             format,
			Top:                      to.Ptr(int32(5000)),
			AuthorizationScopeFilter: to.Ptr("AtScopeAndAbove"), // Include management groups for Azure Policy queries
		}

		var skipToken *string = nil
		for ok := true; ok; ok = skipToken != nil {
			options.SkipToken = skipToken
			request := QueryRequest{
				Subscriptions: subscriptionIDs[i:j],
				Query:         query,
				Options:       options,
			}

			resp, err := q.doRequest(ctx, request)
			if err == nil {
				result.Data = append(result.Data, resp.Data...)
				skipToken = resp.SkipToken
				log.Debug().Msgf("Graph query batch %d-%d returned %d records, next skipToken: %v", i, j, len(resp.Data), skipToken)
			} else {
				log.Fatal().Err(err).Msgf("Failed to run Resource Graph query: %s", query)
				return nil
			}
		}
	}
	log.Debug().Msgf("Graph query returned %d records", len(result.Data))
	return &result
}

// doRequest sends the HTTP request to the Resource Graph API and returns the response.
func (q *GraphQueryClient) doRequest(ctx context.Context, request QueryRequest) (*QueryResponse, error) {
	// Serialize request to JSON
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create seekable request body reader (needed for retries)
	readSeekCloser := &bytesReadSeekCloser{
		Reader: bytes.NewReader(body),
	}

	// Get resource manager endpoint scope
	resourceManagerEndpoint := az.GetResourceManagerEndpoint()
	scope := to.Ptr(fmt.Sprintf("%s/.default", resourceManagerEndpoint))

	// Send POST request using HttpClient (handles retries and throttling automatically)
	respBody, resp, err := q.httpClient.DoPost(ctx, q.endpoint, readSeekCloser, scope)
	if err != nil {
		return nil, err
	}

	// Parse response JSON
	queryResp := QueryResponse{}

	// Extract quota headers from response
	quotaStr := resp.Header.Get("x-ms-user-quota-remaining")
	if quotaStr != "" {
		quota, parseErr := strconv.Atoi(quotaStr)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse quota header: %w", parseErr)
		}
		queryResp.Quota = quota
	}

	// Parse x-ms-user-quota-resets-after as timespan in "hh:mm:ss" format
	retryAfterStr := resp.Header.Get("x-ms-user-quota-resets-after")
	if retryAfterStr != "" {
		var h, m, s int
		_, scanErr := fmt.Sscanf(retryAfterStr, "%d:%d:%d", &h, &m, &s)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to parse retry-after header: %w", scanErr)
		}
		queryResp.RetryAfter = time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
	}

	log.Debug().Msgf("Graph query quota remaining: %d, Retry after: %s", queryResp.Quota, queryResp.RetryAfter)

	// Unmarshal response body
	if err := json.Unmarshal(respBody, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &queryResp, nil
}

// bytesReadSeekCloser wraps a bytes.Reader to implement io.ReadSeekCloser
// This allows the Azure SDK pipeline to seek back to the beginning for retries
type bytesReadSeekCloser struct {
	*bytes.Reader
}

// Close implements io.Closer (no-op for bytes.Reader)
func (b *bytesReadSeekCloser) Close() error {
	return nil
}
