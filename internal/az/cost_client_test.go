// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package az

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"golang.org/x/time/rate"
)

type costTestTransport func(*http.Request) (*http.Response, error)

func (f costTestTransport) Do(req *http.Request) (*http.Response, error) { return f(req) }

// TestCostManagementPipelines exercises HttpClient's cost pipeline directly. There is no
// longer a typed armcostmanagement.QueryClient to compare against: it has no pager for
// Usage, so both the region-selection plugin and the ordinary CostScanner now issue every
// page (including the first) through this raw HTTP pipeline via QueryCostManagementPages.
func TestCostManagementPipelines(t *testing.T) {
	// Virtual time exercises production SDK sleeps and the shared rate limiter,
	// including waits longer than one minute, without wall-clock delays.
	synctest.Test(t, func(t *testing.T) {
		cred := &mockCredential{token: "test-token"}
		limiter := &throttling.ThrottlingPolicy{CostLimiter: rate.NewLimiter(rate.Every(5*time.Second), 1)}
		for _, scenario := range []struct {
			name      string
			header    string
			delay     string
			failures  int
			status    int
			wantCalls int
			wantDelay time.Duration
			deadline  time.Duration
			retries   int32
			override  bool
			wantErr   bool
		}{
			{name: "long cooldown", header: "x-ms-ratelimit-microsoft.costmanagement-entity-retry-after", delay: "120", failures: 1, wantCalls: 2, wantDelay: 120 * time.Second},
			{name: "consumption cooldown", header: "x-ms-ratelimit-microsoft.consumption-retry-after", delay: "120", failures: 1, wantCalls: 2, wantDelay: 120 * time.Second},
			{name: "qpu cooldown", header: "x-ms-ratelimit-microsoft.costmanagement-qpu-retry-after", delay: "120", failures: 1, wantCalls: 2, wantDelay: 120 * time.Second},
			{name: "standard cooldown", header: "Retry-After", delay: "120", failures: 1, wantCalls: 2, wantDelay: 120 * time.Second},
			{name: "service unavailable", header: "Retry-After", delay: "120", failures: 1, status: http.StatusServiceUnavailable, wantCalls: 2, wantDelay: 120 * time.Second},
			{name: "missing hints", failures: 1, wantCalls: 2, wantDelay: 20 * time.Second},
			{name: "invalid hints", header: "x-ms-ratelimit-microsoft.costmanagement-qpu-retry-after", delay: "invalid", failures: 1, wantCalls: 2, wantDelay: 20 * time.Second},
			{name: "exhausted", header: "Retry-After", delay: "120", failures: 6, wantCalls: 6, wantDelay: 120 * time.Second, wantErr: true},
			{name: "request budget", header: "Retry-After", delay: "300", failures: 6, wantCalls: 3, wantDelay: 300 * time.Second, wantErr: true},
			{name: "caller deadline", header: "Retry-After", delay: "120", failures: 1, wantCalls: 1, deadline: time.Minute, wantErr: true},
			{name: "disabled", failures: 1, wantCalls: 1, retries: -1, wantErr: true},
			{name: "context override", failures: 1, wantCalls: 1, override: true, wantErr: true},
			{name: "beyond budget", header: "Retry-After", delay: "901", failures: 1, wantCalls: 1, wantErr: true},
		} {
			func() {
				t.Log(scenario.name)
				// Refill the limiter before the next scenario.
				time.Sleep(throttling.CostRequestTimeout)
				var attempts []time.Time
				var bodies [][]byte
				var firstDeadline time.Time
				transport := costTestTransport(func(req *http.Request) (*http.Response, error) {
					if req.Method != http.MethodPost || req.Header.Get("ClientType") != "azqr" || !strings.HasPrefix(req.Header.Get("Authorization"), "Bearer ") {
						t.Errorf("missing method, client type or authentication: method=%s clientType=%q auth=%q", req.Method, req.Header.Get("ClientType"), req.Header.Get("Authorization"))
					}
					if req.Context().Err() != nil {
						t.Errorf("transport called with cancelled context: %v", req.Context().Err())
					}
					if firstDeadline.IsZero() {
						firstDeadline, _ = req.Context().Deadline()
					}
					attempts = append(attempts, time.Now())
					body, err := io.ReadAll(req.Body)
					if err != nil {
						t.Fatal(err)
					}
					bodies = append(bodies, body)
					status := http.StatusOK
					payload := `{"properties":{"columns":[],"rows":[]}}`
					if len(attempts) <= scenario.failures {
						status = http.StatusTooManyRequests
						payload = `{"error":{"code":"429","message":"Too many requests"}}`
						if scenario.status != 0 {
							status = scenario.status
							payload = `{"error":{"code":"ServiceUnavailable","message":"Service temporarily unavailable"}}`
						}
					}
					headers := http.Header{"Content-Type": []string{"application/json"}}
					if scenario.header != "" {
						headers.Set(scenario.header, scenario.delay)
					}
					return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(payload)), Request: req}, nil
				})
				ctx := context.Background()
				if scenario.deadline > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, scenario.deadline)
					defer cancel()
				}
				if scenario.override {
					ctx = policy.WithRetryOptions(ctx, policy.RetryOptions{MaxRetries: -1})
				}
				retries := int32(5)
				if scenario.retries != 0 {
					retries = scenario.retries
				}
				start := time.Now()
				opts := DefaultHttpClientOptions(90 * time.Second)
				opts.ThrottlingPolicy = limiter
				opts.Transport = transport
				opts.MaxRetries = retries
				_, _, err := NewHttpClient(cred, opts).DoPost(ctx,
					"https://management.azure.com/subscriptions/test/providers/Microsoft.CostManagement/query?$skiptoken=private",
					NopReadSeekCloser{bytes.NewReader([]byte(`{"type":"ActualCost"}`))})
				if (err != nil) != scenario.wantErr {
					t.Fatalf("error = %v, wantErr %v", err, scenario.wantErr)
				}
				if len(attempts) != scenario.wantCalls {
					t.Fatalf("got %d attempts, want %d, err: %v", len(attempts), scenario.wantCalls, err)
				}
				for i := 1; i < len(attempts); i++ {
					tolerance := time.Duration(0)
					if scenario.wantDelay == 20*time.Second {
						// rate.Limiter converts floating-point token balances to nanoseconds.
						tolerance = time.Nanosecond
					}
					if elapsed := attempts[i].Sub(attempts[i-1]); (elapsed - scenario.wantDelay).Abs() > tolerance {
						t.Errorf("retry interval %s, want %s", elapsed, scenario.wantDelay)
					}
					if !bytes.Equal(bodies[0], bodies[i]) {
						t.Fatal("POST body changed on retry")
					}
				}
				if scenario.name == "request budget" {
					if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) != 15*time.Minute {
						t.Fatalf("budget not enforced: elapsed %s, err %v", time.Since(start), err)
					}
				}
				if scenario.name == "caller deadline" && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("caller deadline lost: %v", err)
				}
				if scenario.name == "exhausted" {
					var responseErr *azcore.ResponseError
					if !errors.As(err, &responseErr) || responseErr.StatusCode != 429 {
						t.Fatalf("response error lost: %v", err)
					}
				}
				if firstDeadline.Sub(attempts[0]) != min(90*time.Second, scenarioDeadline(scenario.deadline)) {
					t.Fatalf("network attempt timeout lost: %s", firstDeadline.Sub(attempts[0]))
				}
			}()
		}
		costQueuedRequests(t, cred, limiter)
		costCancellation(t, cred, limiter)
	})
}

func costQueuedRequests(t *testing.T, cred azcore.TokenCredential, limiter policy.Policy) {
	t.Helper()
	time.Sleep(throttling.CostRequestTimeout)
	var mu sync.Mutex
	var attempts []time.Time
	transport := costTestTransport(func(req *http.Request) (*http.Response, error) {
		deadline, ok := req.Context().Deadline()
		if !ok || time.Until(deadline) != 90*time.Second {
			t.Error("local queue wait consumed the network-attempt timeout")
		}
		mu.Lock()
		attempts = append(attempts, time.Now())
		mu.Unlock()
		return &http.Response{
			StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("{}")), Request: req,
		}, nil
	})
	var wg sync.WaitGroup
	start := time.Now()
	for range 7 {
		wg.Go(func() {
			opts := DefaultHttpClientOptions(90 * time.Second)
			opts.ThrottlingPolicy = limiter
			opts.Transport = transport
			// Different client instances must still coordinate the same scope.
			_, _, err := NewHttpClient(cred, opts).DoPost(context.Background(),
				"https://management.azure.com/subscriptions/queued/providers/Microsoft.CostManagement/query",
				NopReadSeekCloser{bytes.NewReader([]byte("{}"))})
			if err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if len(attempts) != 7 || time.Since(start) != 120*time.Second {
		t.Fatalf("queued requests failed or bypassed pacing: count %d, time %s", len(attempts), time.Since(start))
	}
}

func costCancellation(t *testing.T, cred azcore.TokenCredential, limiter policy.Policy) {
	t.Helper()
	time.Sleep(throttling.CostRequestTimeout)
	var calls int
	transport := costTestTransport(func(req *http.Request) (*http.Response, error) {
		calls++
		if req.Header.Get("ClientType") != "custom-client" {
			t.Error("explicit ClientType overwritten")
		}
		return &http.Response{
			StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("{}")), Request: req,
		}, nil
	})
	opts := DefaultHttpClientOptions(90 * time.Second)
	opts.ThrottlingPolicy = limiter
	opts.Transport = transport
	client := NewHttpClient(cred, opts)
	url := "https://management.azure.com/subscriptions/cancel/providers/Microsoft.CostManagement/query"
	ctx := policy.WithHTTPHeader(context.Background(), http.Header{"Clienttype": []string{"custom-client"}})
	if _, _, err := client.DoPost(ctx, url, NopReadSeekCloser{bytes.NewReader([]byte("{}"))}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	start := time.Now()
	_, _, err := client.DoPost(ctx, url, NopReadSeekCloser{bytes.NewReader([]byte("{}"))})
	if !errors.Is(err, context.Canceled) || calls != 1 || time.Since(start) != time.Second {
		t.Fatalf("cancelled queued request reached transport or kept retrying: calls %d, err %v", calls, err)
	}
}

func TestNonCostPipelineUnchanged(t *testing.T) {
	for _, url := range []string{
		"https://management.azure.com/subscriptions/test/providers/Microsoft.Compute/virtualMachines",
		"https://management.azure.com/providers/Microsoft.ResourceGraph/resources",
		"https://prices.azure.com/api/retail/prices",
	} {
		var calls int
		opts := DefaultHttpClientOptions(time.Second)
		opts.Transport = costTestTransport(func(req *http.Request) (*http.Response, error) {
			calls++
			if req.Header.Get("ClientType") != "" {
				t.Error("Cost Management ClientType added to unrelated endpoint")
			}
			return &http.Response{
				StatusCode: 429,
				Header: http.Header{
					"Retry-After": []string{"120"},
					"X-Ms-Ratelimit-Microsoft.Costmanagement-Entity-Retry-After": []string{"180"},
				},
				Body: io.NopCloser(strings.NewReader(`{"error":{"code":"429"}}`)), Request: req,
			}, nil
		})
		_, resp, err := NewHttpClient(&mockCredential{token: "test-token"}, opts).DoPost(context.Background(), url, nil)
		if err == nil || calls != 1 {
			t.Fatalf("non-Cost retry cap changed: calls %d, err %v", calls, err)
		}
		if resp.Header.Get("retry-after-ms") != "" {
			t.Fatal("Cost Management cooldown normalization leaked to another endpoint")
		}
	}
}

func scenarioDeadline(deadline time.Duration) time.Duration {
	if deadline == 0 {
		return throttling.CostRequestTimeout
	}
	return deadline
}
