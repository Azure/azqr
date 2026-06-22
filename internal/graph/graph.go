// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	Data []json.RawMessage // Query result rows as raw JSON; each consumer unmarshals into its own typed struct.
}

// UnmarshalRows decodes each raw JSON row into a value of type T. Rows that fail
// to unmarshal are logged and skipped rather than aborting the whole result. The
// label is used only for the warning message (e.g. "Advisor", "Defender status").
// If all rows fail (e.g. due to an API schema change) an error-level message is
// logged so that a silent empty report is clearly visible in logs.
func UnmarshalRows[T any](data []json.RawMessage, label string) []T {
	rows := make([]T, 0, len(data))
	skipped := 0
	for _, raw := range data {
		var r T
		if err := json.Unmarshal(raw, &r); err != nil {
			log.Warn().Err(err).Msgf("Skipping malformed %s row", label)
			skipped++
			continue
		}
		rows = append(rows, r)
	}
	if skipped > 0 {
		if skipped == len(data) {
			log.Error().Msgf("All %d %s rows failed to unmarshal — possible API schema change; results will be empty", skipped, label)
		} else {
			log.Warn().Msgf("Skipped %d/%d malformed %s rows", skipped, len(data), label)
		}
	}
	return rows
}

// QueryOptions controls optional behaviour of a Resource Graph query.
type QueryOptions struct {
	// ManagementGroupScope sets AuthorizationScopeFilter to "AtScopeAndAbove",
	// which traverses management groups. Required for PolicyResources queries;
	// avoid for all other queries as it significantly increases latency.
	ManagementGroupScope bool
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
	Data       []json.RawMessage `json:"data"` // Query result rows; each element is a raw JSON object.
	SkipToken  *string           `json:"$skipToken,omitempty"`
	Quota      int               // Value of x-ms-user-quota-remaining header as int
	RetryAfter time.Duration     // Value of x-ms-user-quota-resets-after header as timespan
}

// NewGraphQuery creates a new GraphQuery using the provided TokenCredential.
func NewGraphQuery(cred azcore.TokenCredential) *GraphQueryClient {
	// Create Azure HTTP client with built-in retry and throttling
	httpClient := az.NewHttpClient(cred, az.DefaultHttpClientOptions(120*time.Second))

	resourceManagerEndpoint := az.GetResourceManagerEndpoint()

	return &GraphQueryClient{
		httpClient: httpClient,
		endpoint:   fmt.Sprintf("%s/providers/Microsoft.ResourceGraph/resources?api-version=2024-04-01", resourceManagerEndpoint),
	}
}

// Query executes a Resource Graph query for the given subscriptions and query string.
// It handles batching and pagination. Pass a QueryOptions value to enable optional
// features such as management-group scope (needed only for PolicyResources queries).
func (q *GraphQueryClient) Query(ctx context.Context, query string, subscriptions map[string]string, opts ...QueryOptions) (*GraphResult, error) {
	result := GraphResult{
		Data: make([]json.RawMessage, 0, 5000),
	}

	subscriptionIDs := make([]string, 0, len(subscriptions))
	for s := range subscriptions {
		subscriptionIDs = append(subscriptionIDs, s)
	}

	// Run the query in batches of 300 subscriptions
	const batchSize = 300

	options := &QueryRequestOptions{
		ResultFormat: "objectArray",
		Top:          to.Ptr(int32(5000)),
	}
	if len(opts) > 0 && opts[0].ManagementGroupScope {
		options.AuthorizationScopeFilter = to.Ptr("AtScopeAndAbove")
	}

	for i := 0; i < len(subscriptionIDs); i += batchSize {
		j := min(i+batchSize, len(subscriptionIDs))

		// QueryRequest is hoisted outside the pagination loop: Subscriptions, Query,
		// and Options do not change per page — only options.SkipToken does (via pointer).
		request := QueryRequest{
			Subscriptions: subscriptionIDs[i:j],
			Query:         query,
			Options:       options,
		}

		var skipToken *string
		for ok := true; ok; ok = skipToken != nil {
			options.SkipToken = skipToken

			resp, err := q.doRequest(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("failed to run resource graph query: %w", err)
			}

			result.Data = append(result.Data, resp.Data...)
			skipToken = resp.SkipToken
			log.Debug().Msgf("Graph query batch %d-%d returned %d records, next skipToken: %v", i, j, len(resp.Data), skipToken)
		}
	}
	log.Debug().Msgf("Graph query returned %d records", len(result.Data))
	return &result, nil
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

	// Send POST request using HttpClient (handles retries and throttling automatically)
	resp, err := q.httpClient.DoPostStream(ctx, q.endpoint, readSeekCloser)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("Failed to close response body")
		}
	}()

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

	// Decode response body directly from the stream, then drain the remainder so
	// the underlying TCP connection is returned to the pool for reuse.
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)

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
