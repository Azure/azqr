// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package availability

import (
	"testing"

	"github.com/Azure/azqr/internal/scanners/plugins/region/config"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

func TestSkuAppliesToRegion(t *testing.T) {
	cache := types.NewSKUAvailabilityCache()
	tests := []struct {
		name         string
		item         types.SKUAPIItem
		targetRegion string
		want         bool
	}{
		{
			name:         "locationInfo match",
			item:         types.SKUAPIItem{LocationInfo: []types.SKULocationInfo{{Location: "eastus"}}},
			targetRegion: "eastus",
			want:         true,
		},
		{
			name:         "locationInfo no match",
			item:         types.SKUAPIItem{LocationInfo: []types.SKULocationInfo{{Location: "westus"}}},
			targetRegion: "eastus",
			want:         false,
		},
		{
			name:         "locationInfo match with normalization",
			item:         types.SKUAPIItem{LocationInfo: []types.SKULocationInfo{{Location: "East US"}}},
			targetRegion: "eastus",
			want:         true,
		},
		{
			name:         "locations match",
			item:         types.SKUAPIItem{Locations: []string{"westeurope"}},
			targetRegion: "westeurope",
			want:         true,
		},
		{
			name:         "locations no match",
			item:         types.SKUAPIItem{Locations: []string{"westeurope"}},
			targetRegion: "eastus",
			want:         false,
		},
		{
			name:         "locations match with mixed case and spaces",
			item:         types.SKUAPIItem{Locations: []string{"West Europe"}},
			targetRegion: "westeurope",
			want:         true,
		},
		{
			name:         "no location data is globally available",
			item:         types.SKUAPIItem{Name: "Standard_D2s_v3"},
			targetRegion: "eastus",
			want:         true,
		},
		{
			name: "locationInfo takes precedence over locations",
			item: types.SKUAPIItem{
				LocationInfo: []types.SKULocationInfo{{Location: "westus"}},
				Locations:    []string{"eastus"},
			},
			targetRegion: "eastus",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tt.item
			if got := cache.SKUAppliesToRegion(&item, tt.targetRegion); got != tt.want {
				t.Errorf("SKUAppliesToRegion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractSKUName(t *testing.T) {
	cache := types.NewSKUAvailabilityCache()
	configWith := func(nameProp string) *config.PropertyMapConfig {
		c := &config.PropertyMapConfig{}
		c.Properties.TopLevelProperties = map[string]string{"name": nameProp}
		return c
	}

	tests := []struct {
		name   string
		item   types.SKUAPIItem
		config *config.PropertyMapConfig
		want   string
	}{
		{
			name:   "configured name property",
			item:   types.SKUAPIItem{Name: "Standard_D2s_v3", Size: "D2s_v3", Tier: "Standard"},
			config: configWith("name"),
			want:   "Standard_D2s_v3",
		},
		{
			name:   "configured size property",
			item:   types.SKUAPIItem{Name: "Standard_D2s_v3", Size: "D2s_v3", Tier: "Standard"},
			config: configWith("size"),
			want:   "D2s_v3",
		},
		{
			name:   "configured tier property",
			item:   types.SKUAPIItem{Tier: "Premium"},
			config: configWith("tier"),
			want:   "Premium",
		},
		{
			name:   "configured property empty falls back to name",
			item:   types.SKUAPIItem{Name: "fallbackName"},
			config: configWith("size"),
			want:   "fallbackName",
		},
		{
			name:   "no top-level properties uses name first",
			item:   types.SKUAPIItem{Name: "n", Size: "s", Tier: "t"},
			config: &config.PropertyMapConfig{},
			want:   "n",
		},
		{
			name:   "no top-level properties falls back to size",
			item:   types.SKUAPIItem{Size: "s", Tier: "t"},
			config: &config.PropertyMapConfig{},
			want:   "s",
		},
		{
			name:   "no top-level properties falls back to tier",
			item:   types.SKUAPIItem{Tier: "t"},
			config: &config.PropertyMapConfig{},
			want:   "t",
		},
		{
			name:   "all empty yields empty string",
			item:   types.SKUAPIItem{},
			config: &config.PropertyMapConfig{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tt.item
			if got := cache.ExtractSKUName(&item, tt.config); got != tt.want {
				t.Errorf("ExtractSKUName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckSKURestrictions(t *testing.T) {
	cache := types.NewSKUAvailabilityCache()
	tests := []struct {
		name string
		item types.SKUAPIItem
		want types.SKUAvailabilityState
	}{
		{
			name: "no restrictions or capabilities is available",
			item: types.SKUAPIItem{Name: "sku"},
			want: types.SKUAvailable,
		},
		{
			name: "location restriction NotAvailableForSubscription is restricted",
			item: types.SKUAPIItem{Restrictions: []types.SKURestriction{{Type: "Location", ReasonCode: "NotAvailableForSubscription"}}},
			want: types.SKURestricted,
		},
		{
			name: "location restriction other reason is unavailable",
			item: types.SKUAPIItem{Restrictions: []types.SKURestriction{{Type: "Location", ReasonCode: "NotAvailableForRegion"}}},
			want: types.SKUUnavailable,
		},
		{
			name: "restriction type is case-insensitive",
			item: types.SKUAPIItem{Restrictions: []types.SKURestriction{{Type: "location", ReasonCode: "notavailableforsubscription"}}},
			want: types.SKURestricted,
		},
		{
			name: "non-location restriction is ignored",
			item: types.SKUAPIItem{Restrictions: []types.SKURestriction{{Type: "Zone", ReasonCode: "NotAvailableForSubscription"}}},
			want: types.SKUAvailable,
		},
		{
			name: "capability available false is unavailable",
			item: types.SKUAPIItem{Capabilities: []types.SKUCapability{{Name: "available", Value: "false"}}},
			want: types.SKUUnavailable,
		},
		{
			name: "capability available true is available",
			item: types.SKUAPIItem{Capabilities: []types.SKUCapability{{Name: "available", Value: "true"}}},
			want: types.SKUAvailable,
		},
		{
			name: "restriction takes precedence over capabilities",
			item: types.SKUAPIItem{
				Restrictions: []types.SKURestriction{{Type: "Location", ReasonCode: "NotAvailableForSubscription"}},
				Capabilities: []types.SKUCapability{{Name: "available", Value: "false"}},
			},
			want: types.SKURestricted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tt.item
			if got := cache.CheckSKURestrictions(&item); got != tt.want {
				t.Errorf("CheckSKURestrictions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractSKUsFromResponse(t *testing.T) {
	cache := types.NewSKUAvailabilityCache()

	t.Run("regional API skips region filtering", func(t *testing.T) {
		config := &config.PropertyMapConfig{RegionalAPI: true}
		items := []types.SKUAPIItem{
			{Name: "SkuA", Locations: []string{"westus"}},
			{Name: "SkuB", Locations: []string{"eastus"}},
		}
		got := cache.ExtractSKUsFromResponse(items, config, "eastus")
		if len(got) != 2 {
			t.Fatalf("expected 2 SKUs (no filtering), got %d", len(got))
		}
		if _, ok := got["skua"]; !ok {
			t.Errorf("expected lowercased key 'skua' present")
		}
	})

	t.Run("global API filters by region", func(t *testing.T) {
		config := &config.PropertyMapConfig{RegionalAPI: false}
		items := []types.SKUAPIItem{
			{Name: "SkuA", Locations: []string{"westus"}},
			{Name: "SkuB", Locations: []string{"eastus"}},
		}
		got := cache.ExtractSKUsFromResponse(items, config, "eastus")
		if len(got) != 1 {
			t.Fatalf("expected 1 SKU after region filter, got %d", len(got))
		}
		if _, ok := got["skub"]; !ok {
			t.Errorf("expected 'skub' to be retained for eastus")
		}
	})

	t.Run("empty SKU names are skipped", func(t *testing.T) {
		config := &config.PropertyMapConfig{RegionalAPI: true}
		items := []types.SKUAPIItem{
			{Name: ""},
			{Name: "Valid"},
		}
		got := cache.ExtractSKUsFromResponse(items, config, "eastus")
		if len(got) != 1 {
			t.Fatalf("expected 1 SKU (empty name skipped), got %d", len(got))
		}
	})

	t.Run("availability state is captured", func(t *testing.T) {
		config := &config.PropertyMapConfig{RegionalAPI: true}
		items := []types.SKUAPIItem{
			{Name: "Restricted", Restrictions: []types.SKURestriction{{Type: "Location", ReasonCode: "NotAvailableForSubscription"}}},
		}
		got := cache.ExtractSKUsFromResponse(items, config, "eastus")
		if got["restricted"] != types.SKURestricted {
			t.Errorf("expected types.SKURestricted, got %v", got["restricted"])
		}
	})
}
