// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package throttling

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

// terminatorPolicy is a policy.Policy that short-circuits the pipeline by
// returning a canned 200 response without calling req.Next() and without
// inspecting the request context. Placing it immediately after the
// ThrottlingPolicy lets the tests isolate the ThrottlingPolicy's own
// limiter/context handling from the SDK's downstream retry/transport policies.
type terminatorPolicy struct {
	called bool
}

func (p *terminatorPolicy) Do(req *policy.Request) (*http.Response, error) {
	p.called = true
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    req.Raw(),
	}, nil
}

// noopTransport satisfies the pipeline's transport requirement; it is never
// reached because terminatorPolicy short-circuits first.
type noopTransport struct{}

func (noopTransport) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    req,
	}, nil
}

// doWithThrottling sends a single request through a pipeline of
// [ThrottlingPolicy -> terminatorPolicy], returning the terminator (to assert
// whether the request got past the limiter) and any error.
func doWithThrottling(t *testing.T, ctx context.Context, url string) (*terminatorPolicy, error) {
	t.Helper()
	term := &terminatorPolicy{}
	pl := runtime.NewPipeline(
		"throttling-test",
		"v1.0.0",
		runtime.PipelineOptions{PerCall: []policy.Policy{NewThrottlingPolicy(), term}},
		&policy.ClientOptions{Transport: noopTransport{}},
	)
	req, err := runtime.NewRequest(ctx, http.MethodGet, url)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	_, err = pl.Do(req)
	return term, err
}

// TestThrottlingPolicy_PricesBypassesLimiter verifies that Retail Prices API
// requests skip the proactive rate limiter entirely. A cancelled context makes
// any limiter.Wait fail immediately, so a request that still reaches the next
// policy proves the limiter was bypassed.
func TestThrottlingPolicy_PricesBypassesLimiter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel: limiter.Wait would fail instantly if it were called

	term, err := doWithThrottling(t, ctx, "https://prices.azure.com/api/retail/prices?$filter=foo")
	if err != nil {
		t.Fatalf("expected prices request to bypass the limiter and succeed, got error: %v", err)
	}
	if !term.called {
		t.Fatal("expected the next policy to be reached for a prices.azure.com request (limiter bypassed)")
	}
}

// TestThrottlingPolicy_GraphIsLimited verifies that the proactive limiter is
// still applied to non-bypassed endpoints. With a cancelled context the limiter
// returns immediately with the context error, the policy wraps it, and the next
// policy is never reached.
func TestThrottlingPolicy_GraphIsLimited(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	url := "https://management.azure.com/providers/Microsoft.ResourceGraph/resources?api-version=2021-03-01"
	term, err := doWithThrottling(t, ctx, url)
	if err == nil {
		t.Fatal("expected a throttling/context error for a rate-limited endpoint, got nil")
	}
	if !strings.Contains(err.Error(), "throttling wait failed") {
		t.Fatalf("expected a 'throttling wait failed' error, got: %v", err)
	}
	if term.called {
		t.Fatal("expected the next policy NOT to be reached when the limiter blocks the request")
	}
}

// TestNewThrottlingPolicy verifies the constructor returns a usable policy.
func TestNewThrottlingPolicy(t *testing.T) {
	if NewThrottlingPolicy() == nil {
		t.Fatal("expected NewThrottlingPolicy to return a non-nil policy")
	}
}
