// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"testing"
)

func TestSkuAppliesToRegion(t *testing.T) {
	tests := []struct {
		name         string
		item         skuAPIItem
		targetRegion string
		want         bool
	}{
		{
			name:         "locationInfo match",
			item:         skuAPIItem{LocationInfo: []skuLocationInfo{{Location: "eastus"}}},
			targetRegion: "eastus",
			want:         true,
		},
		{
			name:         "locationInfo no match",
			item:         skuAPIItem{LocationInfo: []skuLocationInfo{{Location: "westus"}}},
			targetRegion: "eastus",
			want:         false,
		},
		{
			name:         "locationInfo match with normalization",
			item:         skuAPIItem{LocationInfo: []skuLocationInfo{{Location: "East US"}}},
			targetRegion: "eastus",
			want:         true,
		},
		{
			name:         "locations match",
			item:         skuAPIItem{Locations: []string{"westeurope"}},
			targetRegion: "westeurope",
			want:         true,
		},
		{
			name:         "locations no match",
			item:         skuAPIItem{Locations: []string{"westeurope"}},
			targetRegion: "eastus",
			want:         false,
		},
		{
			name:         "locations match with mixed case and spaces",
			item:         skuAPIItem{Locations: []string{"West Europe"}},
			targetRegion: "westeurope",
			want:         true,
		},
		{
			name:         "no location data is globally available",
			item:         skuAPIItem{Name: "Standard_D2s_v3"},
			targetRegion: "eastus",
			want:         true,
		},
		{
			name: "locationInfo takes precedence over locations",
			item: skuAPIItem{
				LocationInfo: []skuLocationInfo{{Location: "westus"}},
				Locations:    []string{"eastus"},
			},
			targetRegion: "eastus",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tt.item
			if got := skuAppliesToRegion(&item, tt.targetRegion); got != tt.want {
				t.Errorf("skuAppliesToRegion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractSKUName(t *testing.T) {
	configWith := func(nameProp string) *propertyMapConfig {
		c := &propertyMapConfig{}
		c.Properties.TopLevelProperties = map[string]string{"name": nameProp}
		return c
	}

	tests := []struct {
		name   string
		item   skuAPIItem
		config *propertyMapConfig
		want   string
	}{
		{
			name:   "configured name property",
			item:   skuAPIItem{Name: "Standard_D2s_v3", Size: "D2s_v3", Tier: "Standard"},
			config: configWith("name"),
			want:   "Standard_D2s_v3",
		},
		{
			name:   "configured size property",
			item:   skuAPIItem{Name: "Standard_D2s_v3", Size: "D2s_v3", Tier: "Standard"},
			config: configWith("size"),
			want:   "D2s_v3",
		},
		{
			name:   "configured tier property",
			item:   skuAPIItem{Tier: "Premium"},
			config: configWith("tier"),
			want:   "Premium",
		},
		{
			name:   "configured property empty falls back to name",
			item:   skuAPIItem{Name: "fallbackName"},
			config: configWith("size"),
			want:   "fallbackName",
		},
		{
			name:   "no top-level properties uses name first",
			item:   skuAPIItem{Name: "n", Size: "s", Tier: "t"},
			config: &propertyMapConfig{},
			want:   "n",
		},
		{
			name:   "no top-level properties falls back to size",
			item:   skuAPIItem{Size: "s", Tier: "t"},
			config: &propertyMapConfig{},
			want:   "s",
		},
		{
			name:   "no top-level properties falls back to tier",
			item:   skuAPIItem{Tier: "t"},
			config: &propertyMapConfig{},
			want:   "t",
		},
		{
			name:   "all empty yields empty string",
			item:   skuAPIItem{},
			config: &propertyMapConfig{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tt.item
			if got := extractSKUName(&item, tt.config); got != tt.want {
				t.Errorf("extractSKUName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckSKURestrictions(t *testing.T) {
	tests := []struct {
		name string
		item skuAPIItem
		want skuAvailabilityState
	}{
		{
			name: "no restrictions or capabilities is available",
			item: skuAPIItem{Name: "sku"},
			want: skuAvailable,
		},
		{
			name: "location restriction NotAvailableForSubscription is restricted",
			item: skuAPIItem{Restrictions: []skuRestriction{{Type: "Location", ReasonCode: "NotAvailableForSubscription"}}},
			want: skuRestricted,
		},
		{
			name: "location restriction other reason is unavailable",
			item: skuAPIItem{Restrictions: []skuRestriction{{Type: "Location", ReasonCode: "NotAvailableForRegion"}}},
			want: skuUnavailable,
		},
		{
			name: "restriction type is case-insensitive",
			item: skuAPIItem{Restrictions: []skuRestriction{{Type: "location", ReasonCode: "notavailableforsubscription"}}},
			want: skuRestricted,
		},
		{
			name: "non-location restriction is ignored",
			item: skuAPIItem{Restrictions: []skuRestriction{{Type: "Zone", ReasonCode: "NotAvailableForSubscription"}}},
			want: skuAvailable,
		},
		{
			name: "capability available false is unavailable",
			item: skuAPIItem{Capabilities: []skuCapability{{Name: "available", Value: "false"}}},
			want: skuUnavailable,
		},
		{
			name: "capability available true is available",
			item: skuAPIItem{Capabilities: []skuCapability{{Name: "available", Value: "true"}}},
			want: skuAvailable,
		},
		{
			name: "restriction takes precedence over capabilities",
			item: skuAPIItem{
				Restrictions: []skuRestriction{{Type: "Location", ReasonCode: "NotAvailableForSubscription"}},
				Capabilities: []skuCapability{{Name: "available", Value: "false"}},
			},
			want: skuRestricted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tt.item
			if got := checkSKURestrictions(&item); got != tt.want {
				t.Errorf("checkSKURestrictions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractSKUsFromResponse(t *testing.T) {
	cache := newSKUAvailabilityCache()

	t.Run("regional API skips region filtering", func(t *testing.T) {
		config := &propertyMapConfig{RegionalAPI: true}
		items := []skuAPIItem{
			{Name: "SkuA", Locations: []string{"westus"}},
			{Name: "SkuB", Locations: []string{"eastus"}},
		}
		got := cache.extractSKUsFromResponse(items, config, "eastus")
		if len(got) != 2 {
			t.Fatalf("expected 2 SKUs (no filtering), got %d", len(got))
		}
		if _, ok := got["skua"]; !ok {
			t.Errorf("expected lowercased key 'skua' present")
		}
	})

	t.Run("global API filters by region", func(t *testing.T) {
		config := &propertyMapConfig{RegionalAPI: false}
		items := []skuAPIItem{
			{Name: "SkuA", Locations: []string{"westus"}},
			{Name: "SkuB", Locations: []string{"eastus"}},
		}
		got := cache.extractSKUsFromResponse(items, config, "eastus")
		if len(got) != 1 {
			t.Fatalf("expected 1 SKU after region filter, got %d", len(got))
		}
		if _, ok := got["skub"]; !ok {
			t.Errorf("expected 'skub' to be retained for eastus")
		}
	})

	t.Run("empty SKU names are skipped", func(t *testing.T) {
		config := &propertyMapConfig{RegionalAPI: true}
		items := []skuAPIItem{
			{Name: ""},
			{Name: "Valid"},
		}
		got := cache.extractSKUsFromResponse(items, config, "eastus")
		if len(got) != 1 {
			t.Fatalf("expected 1 SKU (empty name skipped), got %d", len(got))
		}
	})

	t.Run("availability state is captured", func(t *testing.T) {
		config := &propertyMapConfig{RegionalAPI: true}
		items := []skuAPIItem{
			{Name: "Restricted", Restrictions: []skuRestriction{{Type: "Location", ReasonCode: "NotAvailableForSubscription"}}},
		}
		got := cache.extractSKUsFromResponse(items, config, "eastus")
		if got["restricted"] != skuRestricted {
			t.Errorf("expected skuRestricted, got %v", got["restricted"])
		}
	})
}
