// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"golang.org/x/time/rate"
)

type costScannerCredential struct{}

func (costScannerCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "test-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

type costScannerTransport func(*http.Request) (*http.Response, error)

func (f costScannerTransport) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestCostScannerThrottling(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var attempts []time.Time
		opts := az.NewDefaultClientOptions()
		opts.PerRetryPolicies = []policy.Policy{&throttling.ThrottlingPolicy{CostLimiter: rate.NewLimiter(rate.Every(5*time.Second), 1)}}
		opts.Transport = costScannerTransport(func(req *http.Request) (*http.Response, error) {
			attempts = append(attempts, time.Now())
			if req.Header.Get("ClientType") != "azqr" {
				t.Error("ordinary cost scanner did not apply Cost Management policy")
			}
			headers := http.Header{"Content-Type": []string{"application/json"}}
			status := http.StatusOK
			body := `{"properties":{"rows":[[10,"Storage","USD"]]}}`
			if len(attempts) == 1 {
				status = http.StatusTooManyRequests
				headers.Set("x-ms-ratelimit-microsoft.costmanagement-tenant-retry-after", "120")
				body = `{"error":{"code":"429"}}`
			}
			return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
		})
		scanner := &CostScanner{}
		if err := scanner.init(&models.ScannerConfig{
			Ctx: context.Background(), Cred: costScannerCredential{}, ClientOptions: opts,
			SubscriptionID: "00000000-0000-0000-0000-000000000001", SubscriptionName: "test",
		}); err != nil {
			t.Fatal(err)
		}
		results, err := scanner.QueryCosts()
		if err != nil {
			t.Fatal(err)
		}
		if len(attempts) != 2 || attempts[1].Sub(attempts[0]) != 120*time.Second {
			t.Fatalf("CostScanner did not honor cooldown: %v", attempts)
		}
		if len(results) != 1 || results[0].ServiceName != "Storage" || results[0].Value != "10" {
			t.Fatalf("cost result changed: %v", results)
		}
		if opts.Retry.MaxRetryDelay != time.Minute {
			t.Fatal("ordinary ARM configuration mutated")
		}
	})
}

func TestCostScannerPagination(t *testing.T) {
	// CostScanner previously issued only the initial Cost Management request and discarded
	// any nextLink, silently truncating large result sets. This proves it now follows
	// pagination through the same az.QueryCostManagementPages helper used by region-selection.
	synctest.Test(t, func(t *testing.T) {
		var calls int
		opts := az.NewDefaultClientOptions()
		opts.PerRetryPolicies = []policy.Policy{&throttling.ThrottlingPolicy{CostLimiter: rate.NewLimiter(rate.Every(5*time.Second), 1)}}
		opts.Transport = costScannerTransport(func(req *http.Request) (*http.Response, error) {
			calls++
			headers := http.Header{"Content-Type": []string{"application/json"}}
			var body string
			if req.URL.Query().Get("$skiptoken") == "" {
				body = `{"properties":{"rows":[[10,"Storage","USD"]],"nextLink":"https://management.azure.com/subscriptions/test/providers/Microsoft.CostManagement/query?$skiptoken=page2"}}`
			} else {
				body = `{"properties":{"rows":[[20,"Compute","USD"]]}}`
			}
			return &http.Response{StatusCode: http.StatusOK, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
		})
		scanner := &CostScanner{}
		if err := scanner.init(&models.ScannerConfig{
			Ctx: context.Background(), Cred: costScannerCredential{}, ClientOptions: opts,
			SubscriptionID: "00000000-0000-0000-0000-000000000001", SubscriptionName: "test",
		}); err != nil {
			t.Fatal(err)
		}
		results, err := scanner.QueryCosts()
		if err != nil {
			t.Fatal(err)
		}
		if calls != 2 {
			t.Fatalf("expected 2 pages to be fetched, got %d", calls)
		}
		if len(results) != 2 || results[0].ServiceName != "Storage" || results[1].ServiceName != "Compute" {
			t.Fatalf("results from both pages not merged: %v", results)
		}
	})
}

func TestCostTimeRangePreviousMonth(t *testing.T) {
	now := time.Date(2026, time.February, 3, 12, 0, 0, 0, time.UTC)
	start, end := costTimeRange(now)

	wantStart := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)

	if !start.Equal(wantStart) {
		t.Fatalf("start = %v, want %v", start, wantStart)
	}

	if !end.Equal(wantEnd) {
		t.Fatalf("end = %v, want %v", end, wantEnd)
	}
}
