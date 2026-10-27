// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmmodernization

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners/plugins/region/quota"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/spf13/cobra"
)

func TestEvaluateVMSKU(t *testing.T) {
	vcpus := []skuAPICapability{{Name: "vCPUs", Value: "8"}}
	tests := []struct {
		name        string
		item        skuAPIItem
		wantState   string
		wantReasons []string
		wantZones   []string
	}{
		{
			name:      "unrestricted",
			item:      skuAPIItem{Family: "standardDSv5Family", Capabilities: vcpus},
			wantState: availabilityAvailable,
		},
		{
			name: "not available for subscription",
			item: skuAPIItem{Capabilities: vcpus, Restrictions: []skuAPIRestriction{
				{Type: "Location", ReasonCode: "NotAvailableForSubscription"},
			}},
			wantState:   availabilityRestricted,
			wantReasons: []string{"NotAvailableForSubscription"},
		},
		{
			name: "location restricted for another reason",
			item: skuAPIItem{Capabilities: vcpus, Restrictions: []skuAPIRestriction{
				{Type: "Location", ReasonCode: "QuotaId"},
			}},
			wantState:   availabilityRestricted,
			wantReasons: []string{"QuotaId"},
		},
		{
			name: "zone restricted",
			item: func() skuAPIItem {
				r := skuAPIRestriction{Type: "Zone", ReasonCode: "NotAvailableForSubscription"}
				r.RestrictionInfo.Zones = []string{"3", "1"}
				return skuAPIItem{Capabilities: vcpus, Restrictions: []skuAPIRestriction{r}}
			}(),
			wantState:   availabilityZoneRestricted,
			wantReasons: []string{"NotAvailableForSubscription"},
			wantZones:   []string{"1", "3"},
		},
		{
			name:      "capability marks it unavailable",
			item:      skuAPIItem{Capabilities: []skuAPICapability{{Name: "available", Value: "false"}}},
			wantState: availabilityNotOffered,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateVMSKU(tt.item)
			if got.Availability != tt.wantState {
				t.Errorf("availability = %q, want %q", got.Availability, tt.wantState)
			}
			if !reflect.DeepEqual(got.Reasons, tt.wantReasons) && (len(got.Reasons) != 0 || len(tt.wantReasons) != 0) {
				t.Errorf("reasons = %v, want %v", got.Reasons, tt.wantReasons)
			}
			if !reflect.DeepEqual(got.Zones, tt.wantZones) && (len(got.Zones) != 0 || len(tt.wantZones) != 0) {
				t.Errorf("zones = %v, want %v", got.Zones, tt.wantZones)
			}
		})
	}
}

func TestEvaluateVMSKU_ParsesVCPUsAndFamily(t *testing.T) {
	got := evaluateVMSKU(skuAPIItem{Family: "standardDSv5Family", Capabilities: []skuAPICapability{{Name: "vCPUs", Value: "16"}}})
	if got.VCPUs != 16 || got.Family != "standardDSv5Family" {
		t.Errorf("got vCPUs=%d family=%q", got.VCPUs, got.Family)
	}
}

func withARMServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	original := resourceManagerBaseURL
	resourceManagerBaseURL = func() string { return server.URL }
	t.Cleanup(func() { resourceManagerBaseURL = original })
	return server
}

func withQuotaFetcher(t *testing.T, fn func(subscriptionID, region string) ([]quota.UsageEntry, error)) {
	t.Helper()
	original := fetchVMFamilyQuota
	fetchVMFamilyQuota = func(_ context.Context, _ azcore.TokenCredential, subscriptionID, region string) ([]quota.UsageEntry, error) {
		return fn(subscriptionID, region)
	}
	t.Cleanup(func() { fetchVMFamilyQuota = original })
}

func TestFetchRegionVMSKUs_FollowsPagingAndSkipsNonVMs(t *testing.T) {
	var server *httptest.Server
	var calls int32
	server = withARMServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`{"value":[{"name":"Standard_D4s_v5","resourceType":"virtualMachines","family":"standardDSv5Family"}]}`))
			return
		}
		if got := r.URL.Query().Get("$filter"); got != "location eq 'eastus'" {
			t.Errorf("filter = %q", got)
		}
		_, _ = w.Write([]byte(`{"value":[` +
			`{"name":"Premium_LRS","resourceType":"disks"},` +
			`{"name":"Standard_D2s_v5","resourceType":"virtualMachines","family":"standardDSv5Family"}],` +
			`"nextLink":"` + server.URL + `/next?page=2"}`))
	})

	skus, err := fetchRegionVMSKUs(context.Background(), testVMHttpClient(server), "sub1", "eastus")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
	if len(skus) != 2 {
		t.Fatalf("got %d SKUs, want 2: %v", len(skus), skus)
	}
	if _, ok := skus["standard_d4s_v5"]; !ok {
		t.Errorf("second page SKU missing: %v", skus)
	}
	if _, ok := skus["premium_lrs"]; ok {
		t.Errorf("non-VM SKU was kept")
	}
}

func TestGroupAvailabilityTargets(t *testing.T) {
	base := vmModernizationRow{SubscriptionID: "sub1", Location: "East US", VMSize: "Standard_D4s_v3"}
	target, ok := base.concreteTargetSKU()
	if !ok {
		t.Skip("fixture size has no suggested target")
	}

	second := base
	second.Name = "second"
	other := base
	other.VMSize = "Standard_D8s_v3"
	onTarget := base
	onTarget.VMSize = "Standard_D4s_v6"
	pooled := base
	pooled.HasScaleSet = true
	pooled.IsAKSManaged = true

	groups := groupAvailabilityTargets([]vmModernizationRow{base, second, onTarget, pooled, other}, nil)
	key := availabilityKey{SubscriptionID: "sub1", Region: "eastus", TargetSKU: target}
	g := groups[key]
	if g == nil || g.VMCount < 2 || g.SourceSizes["Standard_D4s_v3"] != 2 {
		t.Fatalf("expected two D4s_v3 VMs grouped, got %+v", g)
	}
	for k := range groups {
		if k.TargetSKU == "" || strings.HasPrefix(k.TargetSKU, "Already") {
			t.Errorf("non-concrete target in groups: %+v", k)
		}
	}
}

func TestClassifyQuotaStatus(t *testing.T) {
	entry := quota.UsageEntry{Limit: 100, Available: 50}
	tests := []struct {
		name     string
		needed   int
		entry    quota.UsageEntry
		hasEntry bool
		want     string
	}{
		{"sufficient", 10, entry, true, quotaSufficient},
		{"near limit after placement", 40, entry, true, quotaNearLimit},
		{"exactly uses the rest", 50, entry, true, quotaNearLimit},
		{"insufficient", 51, entry, true, quotaInsufficient},
		{"no quota entry", 10, entry, false, quotaUnknown},
		{"unknown vCPUs", 0, entry, true, quotaUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyQuotaStatus(tt.needed, tt.entry, tt.hasEntry); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSKUAvailabilitySheet_EndToEnd(t *testing.T) {
	server := withARMServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/subscriptions/sub-bad/") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_, _ = w.Write([]byte(`{"value":[` +
			`{"name":"Standard_D4s_v5","resourceType":"virtualMachines","family":"standardDSv5Family","capabilities":[{"name":"vCPUs","value":"4"}]},` +
			`{"name":"Standard_E4s_v5","resourceType":"virtualMachines","family":"standardESv5Family","capabilities":[{"name":"vCPUs","value":"4"}],` +
			`"restrictions":[{"type":"Location","reasonCode":"NotAvailableForSubscription"}]}]}`))
	})
	withQuotaFetcher(t, func(subscriptionID, region string) ([]quota.UsageEntry, error) {
		if subscriptionID == "sub-bad" {
			return nil, errors.New("quota failure")
		}
		return []quota.UsageEntry{{ResourceName: "standardDSv5Family", Limit: 100, Available: 6}}, nil
	})

	groups := map[availabilityKey]*availabilityGroup{
		{SubscriptionID: "sub-ok", Region: "eastus", TargetSKU: "Standard_D4s_v5"}:  {VMCount: 2, SourceSizes: map[string]int{"Standard_D4s_v3": 2}},
		{SubscriptionID: "sub-ok", Region: "eastus", TargetSKU: "Standard_E4s_v5"}:  {VMCount: 1, SourceSizes: map[string]int{"Standard_E4s_v3": 1}},
		{SubscriptionID: "sub-ok", Region: "eastus", TargetSKU: "Standard_M8_v5"}:   {VMCount: 1, SourceSizes: map[string]int{"Standard_M8_v3": 1}},
		{SubscriptionID: "sub-bad", Region: "eastus", TargetSKU: "Standard_D4s_v5"}: {VMCount: 1, SourceSizes: map[string]int{"Standard_D4s_v3": 1}},
	}
	data := collectRegionSKUData(context.Background(), nil, testVMHttpClient(server), regionPairs(groups))
	sheet := buildSKUAvailabilitySheet(groups, data, map[string]string{"sub-ok": "Prod", "sub-bad": "Broken"})
	if sheet == nil {
		t.Fatal("expected a sheet")
	}
	if len(sheet.SheetName) > 31 {
		t.Errorf("sheet name %q exceeds Excel's 31-character limit", sheet.SheetName)
	}

	rows := map[string][]string{}
	for _, row := range sheet.Table[1:] {
		rows[row[0]+"/"+row[2]] = row
	}
	// Columns: 5 Availability, 6 Quota, 7 Reasons, 9 Family, 10 vCPUs Needed, 11 Quota Available.
	check := func(key string, want map[int]string) {
		t.Helper()
		row, ok := rows[key]
		if !ok {
			t.Fatalf("missing row %s in %v", key, rows)
		}
		for col, v := range want {
			if row[col] != v {
				t.Errorf("%s col %d = %q, want %q", key, col, row[col], v)
			}
		}
	}
	check("Prod/Standard_D4s_v5", map[int]string{5: availabilityAvailable, 9: "standardDSv5Family", 10: "8", 11: "6", 6: quotaInsufficient})
	check("Prod/Standard_E4s_v5", map[int]string{5: availabilityRestricted, 7: "NotAvailableForSubscription", 6: quotaUnknown})
	check("Prod/Standard_M8_v5", map[int]string{5: availabilityNotOffered, 6: quotaNotApplicable})
	check("Broken/Standard_D4s_v5", map[int]string{5: availabilityUnknown, 6: quotaUnknown})

	if sheet.Table[1][0] != "Broken" {
		t.Errorf("rows are not sorted by subscription name: first row is %q", sheet.Table[1][0])
	}
}

func TestBuildSKUAvailabilitySheet_NoGroupsNoSheet(t *testing.T) {
	if sheet := buildSKUAvailabilitySheet(nil, nil, nil); sheet != nil {
		t.Errorf("expected nil sheet, got %+v", sheet)
	}
}

func TestGroupAvailabilityTargets_TargetRegions(t *testing.T) {
	base := vmModernizationRow{SubscriptionID: "sub1", Location: "East US", VMSize: "Standard_D4s_v3"}
	target, ok := base.concreteTargetSKU()
	if !ok {
		t.Skip("fixture size has no suggested target")
	}
	groups := groupAvailabilityTargets([]vmModernizationRow{base}, []string{"westeurope", "swedencentral"})
	if len(groups) != 2 {
		t.Fatalf("expected one group per target region, got %d", len(groups))
	}
	for _, region := range []string{"westeurope", "swedencentral"} {
		if g := groups[availabilityKey{SubscriptionID: "sub1", Region: region, TargetSKU: target}]; g == nil || g.VMCount != 1 {
			t.Errorf("missing group for %s: %+v", region, g)
		}
	}
	if _, ok := groups[availabilityKey{SubscriptionID: "sub1", Region: "eastus", TargetSKU: target}]; ok {
		t.Error("own region must not be evaluated when target regions are set")
	}
}

func TestParseTargetRegions(t *testing.T) {
	if got := parseTargetRegions(nil); got != nil {
		t.Errorf("nil params = %v, want nil", got)
	}
	stages := models.NewStageConfigs()
	if got := parseTargetRegions(&models.ScanParams{Stages: stages}); got != nil {
		t.Errorf("unset option = %v, want nil", got)
	}
	if err := stages.SetStageOptions(models.StageNamePlugin, map[string]any{"target-regions": "East US, westeurope,eastus,"}); err != nil {
		t.Fatal(err)
	}
	got := parseTargetRegions(&models.ScanParams{Stages: stages})
	if want := []string{"eastus", "westeurope"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRegisterFlags_TargetRegions(t *testing.T) {
	cmd := &cobra.Command{Use: "vm-modernization"}
	NewScanner().RegisterFlags(cmd)
	f := cmd.Flags().Lookup("target-regions")
	if f == nil || f.Value.Type() != "stringSlice" {
		t.Fatalf("target-regions string-slice flag not registered: %+v", f)
	}
}
