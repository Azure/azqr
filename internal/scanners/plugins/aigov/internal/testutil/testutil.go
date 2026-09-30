// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package testutil holds test doubles shared by the aigov packages' tests.
package testutil

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/query/azmetrics"
)

type testCredential struct{}

func (testCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token:     "test-token",
		ExpiresOn: time.Now().Add(time.Hour),
	}, nil
}

type rewriteTransport struct {
	client *http.Client
	target *url.URL
}

func (r *rewriteTransport) Do(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = r.target.Scheme
	clone.URL.Host = r.target.Host
	clone.Host = r.target.Host
	return r.client.Do(clone)
}

// NewPricingClient returns an HTTP client whose requests are served by handler.
func NewPricingClient(t *testing.T, handler http.HandlerFunc) *az.HttpClient {
	t.Helper()

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse(server.URL) error = %v", err)
	}

	return az.NewHttpClient(testCredential{}, &az.HttpClientOptions{
		Timeout:          5 * time.Second,
		MaxRetries:       1,
		OperationTimeout: 10 * time.Second,
		Scope:            "https://management.azure.com/.default",
		Transport: &rewriteTransport{
			client: server.Client(),
			target: target,
		},
	})
}

// FakeQuerier is a metrics querier stub returning one 429 and one OK series per metric.
type FakeQuerier struct {
	// FailGrains makes required queries fail at these intervals.
	FailGrains map[string]bool
	Calls      []string
}

// Ptr returns a pointer to v.
func Ptr[T any](v T) *T { return &v }

func (f *FakeQuerier) QueryResources(_ context.Context, _ string, _ string, names []string, ids azmetrics.ResourceIDList, o *azmetrics.QueryResourcesOptions) (azmetrics.QueryResourcesResponse, error) {
	interval := *o.Interval
	f.Calls = append(f.Calls, interval+":"+strings.Join(names, ","))
	if f.FailGrains[interval] {
		return azmetrics.QueryResourcesResponse{}, errors.New("grain rejected")
	}
	start, _ := time.Parse(time.RFC3339, *o.StartTime)
	end, _ := time.Parse(time.RFC3339, *o.EndTime)
	var out azmetrics.QueryResourcesResponse
	for _, id := range ids.ResourceIDs {
		md := azmetrics.MetricData{ResourceID: Ptr(id)}
		for _, n := range names {
			ts := azmetrics.TimeSeriesElement{
				MetadataValues: []azmetrics.MetadataValue{
					{Name: &azmetrics.LocalizableString{Value: Ptr("ModelDeploymentName")}, Value: Ptr("dep")},
					{Name: &azmetrics.LocalizableString{Value: Ptr("ModelName")}, Value: Ptr("gpt-4o")},
				},
			}
			if strings.Contains(*o.Filter, "StatusCode") {
				ts.MetadataValues = append(ts.MetadataValues, azmetrics.MetadataValue{Name: &azmetrics.LocalizableString{Value: Ptr("StatusCode")}, Value: Ptr("429")})
			}
			val := float64(100)
			// One point at start, one exactly at end (must be ignored as it
			// belongs to the next half-open chunk).
			ts.Data = []azmetrics.MetricValue{
				{TimeStamp: Ptr(start), Total: Ptr(val), Average: Ptr(val)},
				{TimeStamp: Ptr(end), Total: Ptr(val), Average: Ptr(val)},
			}
			md.Values = append(md.Values, azmetrics.Metric{Name: &azmetrics.LocalizableString{Value: Ptr(n)}, TimeSeries: []azmetrics.TimeSeriesElement{ts}})
		}
		out.Values = append(out.Values, md)
	}
	return out, nil
}
