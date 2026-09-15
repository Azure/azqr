// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package quota

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const quotaTestSubscription = "00000000-0000-0000-0000-000000000001"

var quotaTestProviders = []string{
	"Microsoft.Network", "Microsoft.Sql", "Microsoft.Storage", "Microsoft.Web", "Microsoft.Compute",
}

func fetchTestQuota(t *testing.T, provider string, handler http.HandlerFunc) ([]UsageEntry, error) {
	t.Helper()
	if provider == "Microsoft.Compute" {
		opts := &arm.ClientOptions{
			ClientOptions: policy.ClientOptions{
				Transport: newTestQuotaTransport(t, handler),
				Retry:     policy.RetryOptions{MaxRetries: 1},
			},
		}
		return FetchVMQuota(context.Background(), testCredential{}, opts, quotaTestSubscription, "eastus")
	}
	fetchers := map[string]func(context.Context, *az.HttpClient, string, string) ([]UsageEntry, error){
		"Microsoft.Network": FetchNetworkQuota,
		"Microsoft.Sql":     FetchSQLQuota,
		"Microsoft.Storage": FetchStorageQuota,
		"Microsoft.Web":     FetchAppServiceQuota,
	}
	fetch, ok := fetchers[provider]
	if !ok {
		t.Fatalf("unknown test provider %s", provider)
	}
	return fetch(context.Background(), newTestStorageClient(t, handler), quotaTestSubscription, "eastus")
}

func TestQuotaNoUsages(t *testing.T) {
	for _, provider := range quotaTestProviders {
		for _, paginated := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/paginated=%v", provider, paginated), func(t *testing.T) {
				var output bytes.Buffer
				old := log.Logger
				log.Logger = zerolog.New(&output).Level(zerolog.WarnLevel)
				defer func() { log.Logger = old }()
				calls := 0
				entries, err := fetchTestQuota(t, provider, func(w http.ResponseWriter, r *http.Request) {
					calls++
					w.Header().Set("Content-Type", "application/json")
					if paginated && calls == 1 {
						if _, err := fmt.Fprintf(w, `{"value":[{"currentValue":2,"limit":1000,"name":{"value":"testFamily"}}],
							"nextLink":"https://management.azure.com/subscriptions/%s/providers/%s/locations/eastus/usages?page=2"}`, quotaTestSubscription, provider); err != nil {
							t.Error(err)
						}
						return
					}
					w.WriteHeader(http.StatusConflict)
					if _, err := fmt.Fprintf(w, `{"error":{"code":"SubscriptionHasNoUsages","message":"Subscription %s has no usages."}}`, quotaTestSubscription); err != nil {
						t.Error(err)
					}
				})
				if err != nil || entries != nil {
					t.Fatalf("expected unknown quota without partial entries, got %v, %v", entries, err)
				}
				wantCalls := 1
				if paginated {
					wantCalls++
				}
				if calls != wantCalls {
					t.Fatalf("got %d calls, want %d (409 must not be retried)", calls, wantCalls)
				}
				message := output.String()
				if strings.Count(message, `"level":"warn"`) != 1 ||
					!strings.Contains(message, "quota remains unknown") ||
					!strings.Contains(message, renderers.MaskSubscriptionID(quotaTestSubscription, true)) ||
					!strings.Contains(message, provider) ||
					!strings.Contains(message, "eastus") {
					t.Fatalf("missing concise unavailable warning: %s", message)
				}
				if strings.Contains(message, quotaTestSubscription) || strings.Contains(message, "RESPONSE 409") {
					t.Fatalf("warning includes raw error details: %s", message)
				}
			})
		}
	}
}

func TestQuotaUnexpectedErrors(t *testing.T) {
	for _, provider := range quotaTestProviders {
		for _, tt := range []struct {
			status int
			code   string
		}{
			{http.StatusConflict, "AnotherConflict"},
			{http.StatusForbidden, "AuthorizationFailed"},
			{http.StatusForbidden, "SubscriptionHasNoUsages"},
		} {
			for _, paginated := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%d/%s/paginated=%v", provider, tt.status, tt.code, paginated), func(t *testing.T) {
					calls := 0
					entries, err := fetchTestQuota(t, provider, func(w http.ResponseWriter, r *http.Request) {
						calls++
						w.Header().Set("Content-Type", "application/json")
						if paginated && calls == 1 {
							if _, err := fmt.Fprintf(w, `{"value":[{"currentValue":2,"limit":1000,"name":{"value":"testFamily"}}],
								"nextLink":"https://management.azure.com/subscriptions/%s/providers/%s/locations/eastus/usages?page=2"}`, quotaTestSubscription, provider); err != nil {
								t.Error(err)
							}
							return
						}
						w.WriteHeader(tt.status)
						if _, err := fmt.Fprintf(w, `{"error":{"code":%q,"message":"Unexpected error"}}`, tt.code); err != nil {
							t.Error(err)
						}
					})
					var responseErr *azcore.ResponseError
					if entries != nil || !errors.As(err, &responseErr) ||
						responseErr.StatusCode != tt.status || responseErr.ErrorCode != tt.code {
						t.Fatalf("unexpected error was suppressed or changed: %v, %v", entries, err)
					}
				})
			}
		}
	}
}

func TestIsSubscriptionWithoutUsages(t *testing.T) {
	noUsages := &azcore.ResponseError{StatusCode: http.StatusConflict, ErrorCode: "SubscriptionHasNoUsages"}
	for _, tt := range []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"other error", errors.New("SubscriptionHasNoUsages"), false},
		{"canceled", context.Canceled, false},
		{"exact match", noUsages, true},
		{"wrapped", fmt.Errorf("usages API error: %w", noUsages), true},
		{"other status", &azcore.ResponseError{StatusCode: http.StatusForbidden, ErrorCode: "SubscriptionHasNoUsages"}, false},
		{"other conflict", &azcore.ResponseError{StatusCode: http.StatusConflict, ErrorCode: "AnotherConflict"}, false},
		{"missing code", &azcore.ResponseError{StatusCode: http.StatusConflict}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSubscriptionWithoutUsages(tt.err); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuotaPaginationSuccess(t *testing.T) {
	apiVersions := map[string]string{
		"Microsoft.Network": "2022-07-01",
		"Microsoft.Sql":     "2021-11-01",
		"Microsoft.Storage": "2023-01-01",
		"Microsoft.Web":     "2023-01-01",
	}
	for _, provider := range quotaTestProviders {
		t.Run(provider, func(t *testing.T) {
			calls := 0
			entries, err := fetchTestQuota(t, provider, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if want := "/subscriptions/" + quotaTestSubscription + "/providers/" + provider + "/locations/eastus/usages"; r.URL.Path != want {
					t.Errorf("path = %q, want %q", r.URL.Path, want)
				}
				if want := apiVersions[provider]; calls == 1 && want != "" && r.URL.Query().Get("api-version") != want {
					t.Errorf("api-version = %q, want %q", r.URL.Query().Get("api-version"), want)
				}
				w.Header().Set("Content-Type", "application/json")
				if calls == 1 {
					if _, err := fmt.Fprintf(w, `{"value":[{"currentValue":2,"limit":1000,"name":{"value":"testFamily"}}],
						"nextLink":"https://management.azure.com/subscriptions/%s/providers/%s/locations/eastus/usages?page=2"}`, quotaTestSubscription, provider); err != nil {
						t.Error(err)
					}
					return
				}
				if _, err := fmt.Fprint(w, `{"value":[{"currentValue":5,"limit":10,"name":{"value":"otherFamily"}}]}`); err != nil {
					t.Error(err)
				}
			})
			if err != nil || calls != 2 || len(entries) != 2 {
				t.Fatalf("expected two quota pages, got %v, %v (%d calls)", entries, err, calls)
			}
			if entries[0].Available != 998 || entries[1].Available != 5 {
				t.Fatalf("quota headroom changed: %v", entries)
			}
		})
	}
}

func TestQuotaUnsupportedEndpoint(t *testing.T) {
	for _, provider := range quotaTestProviders {
		for _, status := range []int{http.StatusBadRequest, http.StatusNotFound, http.StatusMethodNotAllowed} {
			t.Run(fmt.Sprintf("%s/%d", provider, status), func(t *testing.T) {
				entries, err := fetchTestQuota(t, provider, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					if _, err := fmt.Fprint(w, `{"error":{"code":"UnsupportedEndpoint","message":"Not supported"}}`); err != nil {
						t.Error(err)
					}
				})
				if entries != nil {
					t.Fatalf("expected no quota data, got %v", entries)
				}
				if provider == "Microsoft.Compute" {
					var responseErr *azcore.ResponseError
					if !errors.As(err, &responseErr) || responseErr.StatusCode != status {
						t.Fatalf("Compute error handling changed: %v", err)
					}
				} else if err != nil {
					t.Fatalf("unsupported generic endpoint should still be skipped: %v", err)
				}
			})
		}
	}
}

func TestFetchNetworkQuotaSuccess(t *testing.T) {
	client := newTestStorageClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprint(w, `{"value":[
			{"currentValue":2,"limit":1000,"name":{"value":"VirtualNetworks"}},
			{"currentValue":1,"limit":1,"name":{"value":"NetworkWatchers"}}]}`); err != nil {
			t.Error(err)
		}
	})
	entries, err := FetchNetworkQuota(context.Background(), client, "sub-123", "eastus")
	if err != nil || len(entries) != 1 || entries[0].ResourceName != "VirtualNetworks" ||
		entries[0].CurrentValue != 2 || entries[0].Available != 998 {
		t.Fatalf("network quota results changed: %v, %v", entries, err)
	}
}
