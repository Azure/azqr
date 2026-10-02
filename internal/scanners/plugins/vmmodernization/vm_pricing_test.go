// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmmodernization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// vmPricingMockCredential is a mock azcore.TokenCredential for vm_pricing_test.go,
// following the same pattern as internal/az/http_client_test.go's mockCredential
// (not reusable directly since that type is unexported in another package).
type vmPricingMockCredential struct{}

func (vmPricingMockCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "test-token", ExpiresOn: time.Now().Add(1 * time.Hour)}, nil
}

// testVMHttpClient builds an az.HttpClient whose transport points at the
// given httptest server, following the fake-transport test seam already
// proven in internal/az/http_client_test.go (HttpClientOptions.Transport).
func testVMHttpClient(server *httptest.Server) *az.HttpClient {
	opts := &az.HttpClientOptions{
		Timeout:          5 * time.Second,
		MaxRetries:       1,
		OperationTimeout: 10 * time.Second,
		Scope:            "https://management.azure.com/.default",
		Transport:        server.Client(),
	}
	return az.NewHttpClient(vmPricingMockCredential{}, opts)
}

func retailResponseJSON(t *testing.T, items []vmRetailPriceItem) []byte {
	t.Helper()
	body, err := json.Marshal(vmRetailPriceResponse{Items: items})
	if err != nil {
		t.Fatalf("failed to marshal fixture response: %v", err)
	}
	return body
}

// withVMRetailPricesURL redirects the package-level vmRetailPricesURL to the
// given test server's URL for the duration of the calling test, restoring
// the original value via t.Cleanup.
func withVMRetailPricesURL(t *testing.T, server *httptest.Server) {
	t.Helper()
	original := vmRetailPricesURL
	vmRetailPricesURL = server.URL
	t.Cleanup(func() { vmRetailPricesURL = original })
}

// TestSelectCanonicalPrice covers P02-T02's row-selection rule using fixture
// JSON only (no live HTTP call): single Linux row, single Windows row, mixed
// Windows+Linux rows, a Spot row that must be excluded, a Reservation row
// that must be excluded, two Consumption rows with different effective
// dates (expect the later one), and a zero-match case.
func TestSelectCanonicalPrice(t *testing.T) {
	linux := vmRetailPriceItem{
		ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope", RetailPrice: 0.096,
		Type: "Consumption", SkuName: "D2s v3", ProductName: "Virtual Machines Dsv3 Series",
		EffectiveStartDate: "2024-01-01T00:00:00Z",
	}
	windows := vmRetailPriceItem{
		ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope", RetailPrice: 0.192,
		Type: "Consumption", SkuName: "D2s v3", ProductName: "Virtual Machines Dsv3 Series Windows",
		EffectiveStartDate: "2024-01-01T00:00:00Z",
	}
	spot := vmRetailPriceItem{
		ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope", RetailPrice: 0.01,
		Type: "Consumption", SkuName: "D2s v3 Spot", ProductName: "Virtual Machines Dsv3 Series",
		EffectiveStartDate: "2024-01-01T00:00:00Z",
	}
	reservation := vmRetailPriceItem{
		ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope", RetailPrice: 0.05,
		Type: "Reservation", SkuName: "D2s v3", ProductName: "Virtual Machines Dsv3 Series",
		EffectiveStartDate: "2024-01-01T00:00:00Z",
	}
	older := vmRetailPriceItem{
		ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope", RetailPrice: 0.090,
		Type: "Consumption", SkuName: "D2s v3", ProductName: "Virtual Machines Dsv3 Series",
		EffectiveStartDate: "2023-01-01T00:00:00Z",
	}
	newer := vmRetailPriceItem{
		ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope", RetailPrice: 0.096,
		Type: "Consumption", SkuName: "D2s v3", ProductName: "Virtual Machines Dsv3 Series",
		EffectiveStartDate: "2024-06-01T00:00:00Z",
	}

	tests := []struct {
		name      string
		items     []vmRetailPriceItem
		osType    string
		wantPrice float64
		wantOK    bool
	}{
		{"single Linux row", []vmRetailPriceItem{linux}, "Linux", 0.096, true},
		{"single Windows row", []vmRetailPriceItem{windows}, "Windows", 0.192, true},
		{"mixed Windows+Linux rows selects Linux for Linux OS", []vmRetailPriceItem{linux, windows}, "Linux", 0.096, true},
		{"mixed Windows+Linux rows selects Windows for Windows OS", []vmRetailPriceItem{linux, windows}, "Windows", 0.192, true},
		{"Spot row excluded, Linux OS falls back to zero match", []vmRetailPriceItem{spot}, "Linux", 0, false},
		{"Reservation row excluded, Linux OS falls back to zero match", []vmRetailPriceItem{reservation}, "Linux", 0, false},
		{"two Consumption rows with different effective dates picks the later one", []vmRetailPriceItem{older, newer}, "Linux", 0.096, true},
		{"zero-match case", nil, "Linux", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrice, gotOK := selectCanonicalPrice(tt.items, tt.osType)
			if gotOK != tt.wantOK || (gotOK && gotPrice != tt.wantPrice) {
				t.Errorf("selectCanonicalPrice(%v, %q) = (%v, %v), want (%v, %v)", tt.items, tt.osType, gotPrice, gotOK, tt.wantPrice, tt.wantOK)
			}
		})
	}
}

// TestSelectCanonicalPrice_FutureDatedRowExcluded confirms a future-dated row
// is not selected even when it is the only candidate surviving the
// Spot/Consumption/OS filters, mirroring the region plugin's own
// EffectiveStartDate > nowStr exclusion.
func TestSelectCanonicalPrice_FutureDatedRowExcluded(t *testing.T) {
	future := vmRetailPriceItem{
		ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope", RetailPrice: 0.5,
		Type: "Consumption", SkuName: "D2s v3", ProductName: "Virtual Machines Dsv3 Series",
		EffectiveStartDate: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	}
	if _, ok := selectCanonicalPrice([]vmRetailPriceItem{future}, "Linux"); ok {
		t.Error("selectCanonicalPrice with only a future-dated row should report not-ok")
	}
}

// TestCollectVMPrices_CachesPerPair confirms two VMs sharing the same
// (armSkuName, armRegionName) tuple result in exactly one underlying price
// lookup call (NFR-002), using a call-counting fake transport.
func TestCollectVMPrices_CachesPerPair(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "")

	var callCount int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		item := vmRetailPriceItem{
			ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope", RetailPrice: 0.096,
			Type: "Consumption", SkuName: "D2s v3", ProductName: "Virtual Machines Dsv3 Series",
			EffectiveStartDate: "2024-01-01T00:00:00Z",
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(retailResponseJSON(t, []vmRetailPriceItem{item}))
	}))
	defer server.Close()

	withVMRetailPricesURL(t, server)
	client := testVMHttpClient(server)
	pairs := []skuRegionPair{
		{ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope"},
		{ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope"},
	}
	cache := collectVMPrices(context.Background(), client, pairs)

	if got := atomic.LoadInt32(&callCount); got != 1 {
		t.Errorf("collectVMPrices with two identical pairs issued %d HTTP calls, want 1", got)
	}
	if price, ok := cache.hourlyPrice("Standard_D2s_v3", "westeurope", "Linux"); !ok || price != 0.096 {
		t.Errorf("cache.hourlyPrice(...) = (%v, %v), want (0.096, true)", price, ok)
	}
}

// TestCollectVMPrices_BatchesAcrossManyPairs confirms more than
// vmPricingBatchSize unique pairs are split across multiple batched requests.
func TestCollectVMPrices_BatchesAcrossManyPairs(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "")

	const uniquePairs = 25 // more than one batch of vmPricingBatchSize (8)

	var callCount int32
	var maxConcurrent int32
	var inFlight int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		cur := atomic.AddInt32(&inFlight, 1)
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if cur <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, cur) {
				break
			}
		}
		// Hold the request open briefly so concurrent in-flight requests can
		// be observed, rather than completing so fast that a sequential and
		// a concurrent implementation would be indistinguishable.
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&inFlight, -1)

		query, _ := url.ParseQuery(r.URL.RawQuery)
		filter := query.Get("$filter")
		_ = filter // request content is not fixture-decoded; response below covers all pairs

		var items []vmRetailPriceItem
		for i := 0; i < uniquePairs; i++ {
			items = append(items, vmRetailPriceItem{
				ArmSkuName: fmt.Sprintf("Standard_Sku%d", i), ArmRegionName: "westeurope", RetailPrice: 0.1,
				Type: "Consumption", SkuName: fmt.Sprintf("Sku%d", i), ProductName: "Virtual Machines Series",
				EffectiveStartDate: "2024-01-01T00:00:00Z",
			})
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(retailResponseJSON(t, items))
	}))
	defer server.Close()

	withVMRetailPricesURL(t, server)
	client := testVMHttpClient(server)
	var pairs []skuRegionPair
	for i := 0; i < uniquePairs; i++ {
		pairs = append(pairs, skuRegionPair{ArmSkuName: fmt.Sprintf("Standard_Sku%d", i), ArmRegionName: "westeurope"})
	}

	cache := collectVMPrices(context.Background(), client, pairs)

	wantBatches := (uniquePairs + vmPricingBatchSize - 1) / vmPricingBatchSize
	if got := atomic.LoadInt32(&callCount); int(got) != wantBatches {
		t.Errorf("collectVMPrices with %d unique pairs issued %d HTTP calls, want %d batched requests", uniquePairs, got, wantBatches)
	}
	if got := atomic.LoadInt32(&maxConcurrent); got < 2 {
		t.Errorf("collectVMPrices observed max concurrent in-flight requests = %d, want >= 2 (concurrent, not strictly sequential)", got)
	}
	if price, ok := cache.hourlyPrice("Standard_Sku0", "westeurope", "Linux"); !ok || price != 0.1 {
		t.Errorf("cache.hourlyPrice(Standard_Sku0, westeurope, Linux) = (%v, %v), want (0.1, true)", price, ok)
	}
}

// TestCollectVMPrices_SuppressedOnExemptCloud confirms no HTTP call is made
// at all when the active cloud is exempt (FR-009), mirroring
// matchLifecycleNotes' suppression.
func TestCollectVMPrices_SuppressedOnExemptCloud(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "AzureGovernment")

	var callCount int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := testVMHttpClient(server)
	pairs := []skuRegionPair{{ArmSkuName: "Standard_D2s_v3", ArmRegionName: "westeurope"}}
	cache := collectVMPrices(context.Background(), client, pairs)

	if got := atomic.LoadInt32(&callCount); got != 0 {
		t.Errorf("collectVMPrices on an exempt cloud issued %d HTTP calls, want 0", got)
	}
	if _, ok := cache.hourlyPrice("Standard_D2s_v3", "westeurope", "Linux"); ok {
		t.Error("cache.hourlyPrice on an exempt cloud's empty cache should report not-ok")
	}
}

// TestCollectVMPrices_NoPairsNoHTTPCall confirms an empty pair set makes no
// HTTP call at all.
func TestCollectVMPrices_NoPairsNoHTTPCall(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "")

	var callCount int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := testVMHttpClient(server)
	cache := collectVMPrices(context.Background(), client, nil)

	if got := atomic.LoadInt32(&callCount); got != 0 {
		t.Errorf("collectVMPrices with no pairs issued %d HTTP calls, want 0", got)
	}
	if cache == nil {
		t.Fatal("collectVMPrices with no pairs should still return a non-nil, empty cache")
	}
}
