// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"golang.org/x/time/rate"
)

type paginationCredential struct{}

func (paginationCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "test-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

type paginationTransport func(*http.Request) (*http.Response, error)

func (f paginationTransport) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestQueryMeterCostsPages(t *testing.T) {
	const next = "https://management.azure.com/subscriptions/test/providers/Microsoft.CostManagement/query?api-version=2022-10-01&$skiptoken="
	page := func(rows, nextLink string) string {
		return fmt.Sprintf(`{"properties":{"columns":[{"name":"ResourceGuid","type":"String"},{"name":"PreTaxCost","type":"Number"}],"rows":%s,"nextLink":%q}}`, rows, nextLink)
	}
	for _, tt := range []struct {
		name      string
		responses []string
		want      map[string]float64
		wantErr   bool
	}{
		{"single page", []string{page(`[["a",10],["a",20],["b",5]]`, "")}, map[string]float64{"a": 30, "b": 5}, false},
		{"three pages", []string{
			page(`[["a",10]]`, next+"2"), page(`[["a",20]]`, next+"3"), page(`[["b",5]]`, ""),
		}, map[string]float64{"a": 30, "b": 5}, false},
		{"empty first page", []string{page(`[]`, next+"2"), page(`[["a",20]]`, "")}, map[string]float64{"a": 20}, false},
		{"empty result", []string{page(`[]`, "")}, map[string]float64{}, false},
		{"missing first properties", []string{`{}`}, map[string]float64{}, false},
		{"204 first page", []string{""}, map[string]float64{}, false},
		{"204 continuation", []string{page(`[["a",10]]`, next+"2"), ""}, map[string]float64{"a": 10}, false},
		{"missing continuation properties", []string{page(`[["a",10]]`, next+"2"), `{}`}, map[string]float64{"a": 10}, false},
		{"invalid continuation", []string{page(`[["a",10]]`, next+"2"), `{`}, nil, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				calls := 0
				transport := paginationTransport(func(req *http.Request) (*http.Response, error) {
					if calls >= len(tt.responses) {
						t.Fatal("unexpected additional page request")
					}
					body := tt.responses[calls]
					calls++
					status := http.StatusOK
					if body == "" {
						status = http.StatusNoContent
					}
					return &http.Response{
						StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}},
						Body: io.NopCloser(strings.NewReader(body)), Request: req,
					}, nil
				})
				limiter := &throttling.ThrottlingPolicy{CostLimiter: rate.NewLimiter(rate.Every(5*time.Second), 1)}
				httpOpts := az.DefaultHttpClientOptions(90 * time.Second)
				httpOpts.ThrottlingPolicy = limiter
				httpOpts.Transport = transport
				meters, err := queryMeterCosts(context.Background(), az.NewHttpClient(paginationCredential{}, httpOpts), "test", 1)
				if calls != len(tt.responses) {
					t.Fatalf("got %d calls, want %d", calls, len(tt.responses))
				}
				if tt.wantErr {
					if err == nil || meters != nil || !strings.Contains(err.Error(), "failed to unmarshal Cost Management results") {
						t.Fatalf("expected no partial results on invalid JSON: %v, %v", meters, err)
					}
					return
				}
				if err != nil || meters == nil {
					t.Fatalf("expected successful, non-nil results: %v, %v", meters, err)
				}
				got := make(map[string]float64)
				for _, meter := range meters {
					got[meter.MeterID] = meter.HistoricalCost
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			})
		})
	}
}

func TestQueryMeterCostsPagination(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		limiter := &throttling.ThrottlingPolicy{CostLimiter: rate.NewLimiter(rate.Every(5*time.Second), 1)}
		for _, exhausted := range []bool{false, true} {
			time.Sleep(15 * time.Minute)
			var initialBody any
			var continuationTimes []time.Time
			start := time.Now()
			transport := paginationTransport(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost || req.Header.Get("ClientType") != "azqr" {
					t.Fatal("missing POST method or Cost Management policy")
				}
				var body any
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				headers := http.Header{"Content-Type": []string{"application/json"}}
				status := http.StatusOK
				var payload string
				if req.URL.Query().Get("$skiptoken") == "" {
					initialBody = body
					// The first page succeeds late in its own budget. A continuation
					// must still receive a fresh budget, not inherit this deadline.
					time.Sleep(10 * time.Minute)
					payload = `{"properties":{"columns":[{"name":"ResourceGuid","type":"String"},{"name":"PreTaxCost","type":"Number"}],"rows":[["meter-a",10]],"nextLink":"https://management.azure.com/subscriptions/test/providers/Microsoft.CostManagement/query?api-version=2022-10-01&$skiptoken=page2"}}`
				} else {
					if !reflect.DeepEqual(initialBody, body) {
						t.Fatal("query definition changed during pagination/retry")
					}
					continuationTimes = append(continuationTimes, time.Now())
					if exhausted || len(continuationTimes) == 1 {
						status = http.StatusTooManyRequests
						headers.Set("x-ms-ratelimit-microsoft.costmanagement-entity-retry-after", "600")
						if exhausted {
							headers.Set("x-ms-ratelimit-microsoft.costmanagement-entity-retry-after", "120")
						}
						payload = `{"error":{"code":"429","message":"Too many requests"}}`
					} else {
						payload = `{"properties":{"columns":[{"name":"ResourceGuid","type":"String"},{"name":"PreTaxCost","type":"Number"}],"rows":[["meter-a",20],["meter-b",5]]}}`
					}
				}
				return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(payload)), Request: req}, nil
			})
			cred := paginationCredential{}
			httpOpts := az.DefaultHttpClientOptions(90 * time.Second)
			httpOpts.ThrottlingPolicy = limiter
			httpOpts.Transport = transport
			meters, err := queryMeterCosts(context.Background(), az.NewHttpClient(cred, httpOpts), "test", 1)
			if exhausted {
				var responseErr *azcore.ResponseError
				if !errors.As(err, &responseErr) || responseErr.StatusCode != 429 || meters != nil {
					t.Fatalf("expected error and discarded earlier pages, got %v, %v", meters, err)
				}
				if len(continuationTimes) != 6 {
					t.Fatalf("got %d attempts, want 6", len(continuationTimes))
				}
				results := []types.RegionComparison{{SourceRegion: "eastus", TargetRegion: "westus"}}
				ApplyCostDiffs(results, meters, &types.CostComparisonData{
					RegionPricing: map[string]map[string]float64{"meter-a": {"eastus": 1, "westus": 2}},
				})
				if results[0].HasCostData || results[0].AvgCostDifference != 0 {
					t.Fatal("failed pagination introduced synthetic cost data")
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			got := map[string]float64{}
			for _, meter := range meters {
				got[meter.MeterID] = meter.HistoricalCost
			}
			if !reflect.DeepEqual(got, map[string]float64{"meter-a": 30, "meter-b": 5}) {
				t.Fatalf("incomplete aggregated pages: %v", got)
			}
			if len(continuationTimes) != 2 || continuationTimes[1].Sub(continuationTimes[0]) != 10*time.Minute {
				t.Fatalf("continuation cooldown not honored: %v", continuationTimes)
			}
			if time.Since(start) != 20*time.Minute {
				t.Fatalf("unexpected total query time: %s", time.Since(start))
			}
		}
	})
}
