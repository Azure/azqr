// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package az

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Azure/azqr/internal/throttling"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/costmanagement/armcostmanagement"
	"golang.org/x/time/rate"
)

// newFastCostHttpClient builds an HttpClient with a permissive Cost Management limiter so
// mechanics tests aren't gated by the shared 20s per-scope pacing (that pacing itself is
// covered separately by TestCostManagementPipelines).
func newFastCostHttpClient(transport costTestTransport) *HttpClient {
	opts := DefaultHttpClientOptions(5 * time.Second)
	opts.Transport = transport
	opts.ThrottlingPolicy = &throttling.ThrottlingPolicy{CostLimiter: rate.NewLimiter(rate.Every(time.Millisecond), 1)}
	return NewHttpClient(&mockCredential{token: "test-token"}, opts)
}

func TestQueryCostManagementPagesMechanics(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		for i, tt := range []struct {
			name      string
			responses []string // "" means HTTP 204
			wantPages int
			wantErr   string
		}{
			{name: "single page", responses: []string{`{"properties":{"rows":[["a",1]]}}`}, wantPages: 1},
			{
				name: "follows nextLink until empty",
				responses: []string{
					`{"properties":{"rows":[["a",1]],"nextLink":"https://management.azure.com/subscriptions/next/providers/Microsoft.CostManagement/query?$skiptoken=2"}}`,
					`{"properties":{"rows":[["b",2]]}}`,
				},
				wantPages: 2,
			},
			{name: "204 first page stops without pageFn", responses: []string{""}, wantPages: 0},
			{
				name: "204 continuation stops after first page",
				responses: []string{
					`{"properties":{"rows":[["a",1]],"nextLink":"https://management.azure.com/subscriptions/next/providers/Microsoft.CostManagement/query?$skiptoken=2"}}`,
					"",
				},
				wantPages: 1,
			},
			{name: "missing properties stops without pageFn", responses: []string{`{}`}, wantPages: 0},
			{name: "invalid JSON surfaces error", responses: []string{`{`}, wantErr: "failed to unmarshal Cost Management results"},
		} {
			// Each scenario uses a distinct scope so the shared per-scope pacing
			// (20s, burst 1) never forces scenarios to wait on each other.
			scope := fmt.Sprintf("/subscriptions/mechanics%d", i)
			func() {
				calls := 0
				transport := costTestTransport(func(req *http.Request) (*http.Response, error) {
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
				httpClient := newFastCostHttpClient(transport)

				var pages int
				err := QueryCostManagementPages(context.Background(), httpClient, scope, armcostmanagement.QueryDefinition{}, func(*armcostmanagement.QueryProperties) error {
					pages++
					return nil
				})

				if tt.wantErr != "" {
					if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
						t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if calls != len(tt.responses) {
					t.Fatalf("got %d calls, want %d", calls, len(tt.responses))
				}
				if pages != tt.wantPages {
					t.Fatalf("got %d pageFn invocations, want %d", pages, tt.wantPages)
				}
			}()
		}
	})
}

func TestQueryCostManagementPagesStopsOnPageFnError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		transport := costTestTransport(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:    io.NopCloser(strings.NewReader(`{"properties":{"rows":[["a",1]],"nextLink":"https://management.azure.com/subscriptions/next/providers/Microsoft.CostManagement/query?$skiptoken=2"}}`)),
				Request: req,
			}, nil
		})
		httpClient := newFastCostHttpClient(transport)

		wantErr := errStopEarly
		err := QueryCostManagementPages(context.Background(), httpClient, "/subscriptions/stopearly", armcostmanagement.QueryDefinition{}, func(*armcostmanagement.QueryProperties) error {
			return wantErr
		})
		if err != wantErr {
			t.Fatalf("error = %v, want %v (nextLink must not be followed after pageFn error)", err, wantErr)
		}
	})
}

var errStopEarly = &stopEarlyError{}

type stopEarlyError struct{}

func (*stopEarlyError) Error() string { return "stop early" }
