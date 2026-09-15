// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package throttling

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

func TestCostEndpoint(t *testing.T) {
	for _, tt := range []struct {
		url  string
		want bool
	}{
		{"https://management.azure.com/subscriptions/a/providers/Microsoft.CostManagement/query?api-version=1&$skiptoken=secret", true},
		{"https://management.usgovcloudapi.net/subscriptions/A/providers/MICROSOFT.COSTMANAGEMENT/QUERY/", true},
		{"https://management.azure.com/providers/Microsoft.Billing/billingAccounts/a/providers/Microsoft.CostManagement/query", true},
		{"https://management.azure.com/providers/Microsoft.ResourceGraph/resources", false},
		{"https://prices.azure.com/api/retail/prices", false},
		{"https://management.azure.com/subscriptions/a?filter=Microsoft.CostManagement/query", false},
		{"https://management.azure.com/subscriptions/a/providers/Microsoft.CostManagement/query/other", false},
	} {
		req, err := http.NewRequest(http.MethodPost, tt.url, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := IsCostManagementQuery(req); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.url, got, tt.want)
		}
	}
}

func TestCostLimiterScopes(t *testing.T) {
	if NewThrottlingPolicy() != NewThrottlingPolicy() {
		t.Fatal("production clients do not share a throttling policy")
	}
	synctest.Test(t, func(t *testing.T) {
		limiter := &ThrottlingPolicy{CostLimiter: rate.NewLimiter(rate.Every(5*time.Second), 1)}
		start := time.Now()
		for _, tt := range []struct {
			url string
			at  time.Duration
		}{
			{"https://management.azure.com/subscriptions/A/providers/Microsoft.CostManagement/query?api-version=1", 0},
			{"https://MANAGEMENT.AZURE.COM/SUBSCRIPTIONS/a/PROVIDERS/MICROSOFT.COSTMANAGEMENT/QUERY/?$skiptoken=next", 20 * time.Second},
			{"https://management.azure.com/subscriptions/b/providers/Microsoft.CostManagement/query", 25 * time.Second},
			{"https://management.usgovcloudapi.net/subscriptions/a/providers/Microsoft.CostManagement/query", 30 * time.Second},
			{"https://management.azure.com/subscriptions/a/providers/Microsoft.CostManagement/query?$skiptoken=last", 40 * time.Second},
		} {
			pl := runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{}, &policy.ClientOptions{
				PerRetryPolicies: []policy.Policy{limiter},
				Transport:        noopTransport{},
			})
			req, err := runtime.NewRequest(context.Background(), http.MethodPost, tt.url)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := pl.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if err := resp.Body.Close(); err != nil {
				t.Fatal(err)
			}
			if got := time.Since(start); got != tt.at {
				t.Fatalf("scope dispatched at %s, want %s", got, tt.at)
			}
		}
	})
}

func TestCostLimiterConcurrent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		limiter := &ThrottlingPolicy{CostLimiter: rate.NewLimiter(rate.Every(5*time.Second), 1)}
		sent := make(chan time.Time, 12)
		pl := runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{}, &policy.ClientOptions{
			PerRetryPolicies: []policy.Policy{limiter},
			Transport: costTransportFunc(func(req *http.Request) (*http.Response, error) {
				sent <- time.Now()
				return noopTransport{}.Do(req)
			}),
		})
		start := time.Now()
		var wg sync.WaitGroup
		for i := range 12 {
			wg.Go(func() {
				url := fmt.Sprintf("https://management.azure.com/subscriptions/%d%s", i, costQuerySuffix)
				req, err := runtime.NewRequest(context.Background(), http.MethodPost, url)
				if err != nil {
					t.Error(err)
					return
				}
				resp, err := pl.Do(req)
				if err != nil {
					t.Error(err)
					return
				}
				if err := resp.Body.Close(); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		close(sent)
		var times []time.Time
		for at := range sent {
			times = append(times, at)
		}
		if len(times) != 12 {
			t.Fatalf("got %d requests, want 12", len(times))
		}
		sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
		for i, at := range times {
			if got := at.Sub(start); got != time.Duration(i)*5*time.Second {
				t.Fatalf("request %d dispatched at %s", i, got)
			}
		}
	})
}

func TestCostScopeLimiterCleanup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := &ThrottlingPolicy{}
		request := func(scope string) *http.Request {
			t.Helper()
			req, err := http.NewRequest(http.MethodPost, "https://management.azure.com/subscriptions/"+scope+costQuerySuffix, nil)
			if err != nil {
				t.Fatal(err)
			}
			return req
		}
		a := p.acquireCostScope(request("a"))
		time.Sleep(2 * CostRequestTimeout)
		b := p.acquireCostScope(request("b"))
		aAgain := p.acquireCostScope(request("a"))
		if a != aAgain || len(p.costScopes) != 2 {
			t.Fatal("cleanup evicted an active scope")
		}
		p.releaseCostScope(a)
		p.releaseCostScope(aAgain)
		p.releaseCostScope(b)
		time.Sleep(CostRequestTimeout)
		c := p.acquireCostScope(request("c"))
		defer p.releaseCostScope(c)
		if len(p.costScopes) != 1 {
			t.Fatalf("idle scope limiters were not cleaned up: %d", len(p.costScopes))
		}
	})
}

func TestCostScopeBacklogDoesNotBlockOtherScopes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := &ThrottlingPolicy{CostLimiter: rate.NewLimiter(rate.Every(5*time.Second), 1)}
		pl := runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{}, &policy.ClientOptions{
			PerRetryPolicies: []policy.Policy{p},
			Transport:        noopTransport{},
		})
		send := func(scope string) {
			t.Helper()
			req, err := runtime.NewRequest(context.Background(), http.MethodPost,
				"https://management.azure.com/subscriptions/"+scope+costQuerySuffix)
			if err != nil {
				t.Error(err)
				return
			}
			resp, err := pl.Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			if err := resp.Body.Close(); err != nil {
				t.Error(err)
			}
		}
		start := time.Now()
		send("a")
		done := make(chan struct{})
		go func() {
			send("a")
			close(done)
		}()
		synctest.Wait()
		send("b")
		if got := time.Since(start); got != 5*time.Second {
			t.Fatalf("scope a's backlog delayed scope b: %s", got)
		}
		<-done
		if got := time.Since(start); got != 20*time.Second {
			t.Fatalf("scope a bypassed its own limit: %s", got)
		}
	})
}

func TestCostLimiterThroughput(t *testing.T) {
	for _, tt := range []struct {
		subscriptions int
		pages         int
		balanced      time.Duration
	}{
		{1, 30, 580 * time.Second},
		{2, 15, 285 * time.Second},
		{3, 10, 190 * time.Second},
		{5, 6, 145 * time.Second},
		{30, 1, 145 * time.Second},
	} {
		t.Run(fmt.Sprintf("%d_subscriptions_%d_pages", tt.subscriptions, tt.pages), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				measure := func(globalInterval time.Duration) time.Duration {
					p := &ThrottlingPolicy{CostLimiter: rate.NewLimiter(rate.Every(globalInterval), 1)}
					pl := runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{}, &policy.ClientOptions{
						PerCallPolicies:  []policy.Policy{NewCostTimeoutPolicy(CostRequestTimeout)},
						PerRetryPolicies: []policy.Policy{p},
						Transport:        noopTransport{},
					})
					start := time.Now()
					var wg sync.WaitGroup
					for sub := range tt.subscriptions {
						wg.Go(func() {
							for page := range tt.pages {
								url := fmt.Sprintf("https://management.azure.com/subscriptions/%d%s?$skiptoken=%d", sub, costQuerySuffix, page)
								req, err := runtime.NewRequest(context.Background(), http.MethodPost, url)
								if err != nil {
									t.Error(err)
									return
								}
								resp, err := pl.Do(req)
								if err != nil {
									t.Error(err)
									return
								}
								if err := resp.Body.Close(); err != nil {
									t.Error(err)
								}
							}
						})
					}
					wg.Wait()
					return time.Since(start)
				}
				before := measure(20 * time.Second)
				after := measure(5 * time.Second)
				if (before-580*time.Second).Abs() > time.Microsecond || (after-tt.balanced).Abs() > time.Microsecond {
					t.Fatalf("pacing-only duration: before %s, after %s (want %s)", before, after, tt.balanced)
				}
				t.Logf("30 cost requests: %s -> %s (%.2fx), excluding API latency and retries",
					before.Round(time.Millisecond), after.Round(time.Millisecond), float64(before)/float64(after))
			})
		})
	}
}

func TestNormalizeCostRetryAfter(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{"missing", nil, ""},
		{"seconds", map[string]string{"Retry-After": "120"}, "120000"},
		{"consumption", map[string]string{"x-ms-ratelimit-microsoft.consumption-retry-after": "120"}, "120000"},
		{"qpu", map[string]string{"x-ms-ratelimit-microsoft.costmanagement-qpu-retry-after": "120"}, "120000"},
		{"date", map[string]string{"Retry-After": now.Add(120 * time.Second).Format(http.TimeFormat)}, "120000"},
		{"past date", map[string]string{"Retry-After": now.Add(-time.Hour).Format(http.TimeFormat)}, ""},
		{"precedence", map[string]string{"retry-after-ms": "1", "Retry-After": "120"}, "120000"},
		{"milliseconds", map[string]string{"x-ms-retry-after-ms": "120001", "Retry-After": "120"}, "120001"},
		{"negative", map[string]string{"x-ms-ratelimit-microsoft.costmanagement-qpu-retry-after": "-1"}, ""},
		{"malformed", map[string]string{"x-ms-ratelimit-microsoft.costmanagement-qpu-retry-after": "secret-invalid"}, ""},
		{"overflow", map[string]string{"x-ms-ratelimit-microsoft.costmanagement-qpu-retry-after": "9223372036854775807"}, ""},
		{"zero", map[string]string{"x-ms-ratelimit-microsoft.costmanagement-qpu-retry-after": "0"}, ""},
		{"longest", map[string]string{
			"x-ms-ratelimit-microsoft.costmanagement-entity-retry-after": "90",
			"x-ms-ratelimit-microsoft.costmanagement-qpu-retry-after":    "120",
			"Retry-After": "5",
		}, "120000"},
		{"consumption takes precedence", map[string]string{
			"x-ms-ratelimit-microsoft.consumption-retry-after":           "240",
			"x-ms-ratelimit-microsoft.costmanagement-entity-retry-after": "60",
			"retry-after-ms": "120000",
		}, "240000"},
	}
	for _, dimension := range costQuotaDimensions {
		tests = append(tests, struct {
			name    string
			headers map[string]string
			want    string
		}{dimension, map[string]string{"x-ms-ratelimit-microsoft.costmanagement-" + dimension + "-retry-after": "120"}, "120000"})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			for k, v := range tt.headers {
				headers.Set(k, v)
			}
			normalizeCostRetryAfter(headers, now)
			if got := headers.Get("retry-after-ms"); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCostRetryDiagnostics(t *testing.T) {
	var output bytes.Buffer
	old := log.Logger
	log.Logger = zerolog.New(&output).Level(zerolog.DebugLevel)
	defer func() { log.Logger = old }()
	headers := http.Header{}
	headers.Set("x-ms-ratelimit-microsoft.costmanagement-entity-retry-after", "120")
	headers.Set("x-ms-ratelimit-microsoft.costmanagement-qpu-retry-after", "sensitive-invalid-value")
	headers.Set("x-ms-ratelimit-remaining-microsoft.costmanagement-entity-requests", "DefaultQuota:0")
	headers.Set("x-ms-ratelimit-microsoft.costmanagement-qpu-consumed", "12")
	headers.Set("x-ms-ratelimit-microsoft.costmanagement-qpu-remaining", "QueriesPerHour:588")
	headers.Set("Location", "https://example.com/subscriptions/private?token=private")
	normalizeCostRetryAfter(headers, time.Now())
	for _, want := range []string{"entity", "120000", "DefaultQuota:0", "Ignoring invalid", `"qpuConsumed":"12"`, `"qpuRemaining":"QueriesPerHour:588"`} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("missing diagnostic %q in %s", want, output.String())
		}
	}
	for _, secret := range []string{"sensitive-invalid-value", "private"} {
		if strings.Contains(output.String(), secret) {
			t.Errorf("diagnostics exposed %q", secret)
		}
	}
}

type costTransportFunc func(*http.Request) (*http.Response, error)

func (f costTransportFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func BenchmarkCostScopeLimiter(b *testing.B) {
	req, err := http.NewRequest(http.MethodPost, "https://management.azure.com/subscriptions/test/providers/Microsoft.CostManagement/query", nil)
	if err != nil {
		b.Fatal(err)
	}
	p := &ThrottlingPolicy{}
	b.ReportAllocs()
	for b.Loop() {
		scope := p.acquireCostScope(req)
		p.releaseCostScope(scope)
	}
}

func BenchmarkIsCostManagementQuery(b *testing.B) {
	req, err := http.NewRequest(http.MethodPost, "https://management.azure.com/subscriptions/test/providers/Microsoft.CostManagement/query", nil)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if !IsCostManagementQuery(req) {
			b.Fatal("cost query not recognized")
		}
	}
}

func TestCostTimeoutBodyLifetime(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		for _, closeBody := range []bool{false, true} {
			var requestContext context.Context
			pl := runtime.NewPipeline("test", "v1.0.0", runtime.PipelineOptions{}, &policy.ClientOptions{
				PerCallPolicies: []policy.Policy{NewCostTimeoutPolicy(CostRequestTimeout)},
				Transport: costTransportFunc(func(req *http.Request) (*http.Response, error) {
					requestContext = req.Context()
					return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("ok")), Request: req}, nil
				}),
			})
			req, err := runtime.NewRequest(context.Background(), http.MethodPost, "https://management.azure.com/subscriptions/a"+costQuerySuffix)
			if err != nil {
				t.Fatal(err)
			}
			runtime.SkipBodyDownload(req)
			resp, err := pl.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if requestContext.Err() != nil {
				t.Fatal("operation context cancelled before body consumption")
			}
			if closeBody {
				if err := resp.Body.Close(); err != nil {
					t.Fatal(err)
				}
			} else if _, err := io.ReadAll(resp.Body); err != nil {
				t.Fatal(err)
			}
			if !errors.Is(requestContext.Err(), context.Canceled) {
				t.Fatal("operation context not cancelled after body consumption")
			}
		}
	})
}
