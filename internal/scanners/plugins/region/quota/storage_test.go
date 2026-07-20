// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package quota

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
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

func newTestStorageClient(t *testing.T, handler http.HandlerFunc) *az.HttpClient {
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

func usageNames(usages []UsageEntry) []string {
	names := make([]string, 0, len(usages))
	for _, usage := range usages {
		names = append(names, usage.ResourceName)
	}
	return names
}

func TestFetchStorageQuota(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     []string
	}{
		{
			name: "keeps storage accounts and other regional counters",
			response: `{"value":[
				{"unit":"Count","currentValue":55,"limit":250,"name":{"value":"StorageAccounts","localizedValue":"Storage Accounts"}},
				{"unit":"Count","currentValue":10,"limit":500,"name":{"value":"ManagedDisks","localizedValue":"Managed Disks"}}
			]}`,
			want: []string{"StorageAccounts", "ManagedDisks"},
		},
		{
			name: "skips noisy sub-resource counters from skip list",
			response: `{"value":[
				{"unit":"Count","currentValue":120,"limit":100000,"name":{"value":"TotalBlobContainers","localizedValue":"Blob Containers"}},
				{"unit":"Count","currentValue":80,"limit":100000,"name":{"value":"TotalFileShares","localizedValue":"File Shares"}},
				{"unit":"Count","currentValue":12,"limit":250,"name":{"value":"StorageAccounts","localizedValue":"Storage Accounts"}}
			]}`,
			want: []string{"StorageAccounts"},
		},
		{
			name: "skips entries with non-positive limits",
			response: `{"value":[
				{"unit":"Count","currentValue":12,"limit":0,"name":{"value":"StorageAccounts","localizedValue":"Storage Accounts"}},
				{"unit":"Count","currentValue":10,"limit":-1,"name":{"value":"ManagedDisks","localizedValue":"Managed Disks"}},
				{"unit":"Count","currentValue":3,"limit":100,"name":{"value":"PremiumManagedDisks","localizedValue":"Premium Managed Disks"}}
			]}`,
			want: []string{"PremiumManagedDisks"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestStorageClient(t, func(w http.ResponseWriter, r *http.Request) {
				if got, want := r.URL.Path, "/subscriptions/sub-123/providers/Microsoft.Storage/locations/eastus/usages"; got != want {
					t.Fatalf("request path = %q, want %q", got, want)
				}
				if got, want := r.URL.Query().Get("api-version"), "2023-01-01"; got != want {
					t.Fatalf("api-version = %q, want %q", got, want)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.response))
			})

			got, err := FetchStorageQuota(context.Background(), client, "sub-123", "eastus")
			if err != nil {
				t.Fatalf("FetchStorageQuota() error = %v", err)
			}

			if gotNames := usageNames(got); !reflect.DeepEqual(gotNames, tt.want) {
				t.Errorf("FetchStorageQuota() names = %v, want %v", gotNames, tt.want)
			}
		})
	}
}
