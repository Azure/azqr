// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sku

import (
	"context"
	"fmt"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

// ── VM Sizes ────────────────────────────────────────────────────────────────

// vmSizeItem is one item from the Compute VM Sizes endpoint.
// This endpoint lists available sizes; it does NOT include restrictions or
// capabilities, so every returned item is considered Available.
type vmSizeItem struct {
	Name string `json:"name"`
}

// VMProvider checks VM SKU availability via the per-region vmSizes endpoint.
// The vmSizes endpoint lists only sizes deployable in the region for this
// subscription — absent entries are treated as unavailable.
type VMProvider struct{}

func (VMProvider) ResourceType() string { return "microsoft.compute/virtualmachines" }

func (VMProvider) FetchSKUs(ctx context.Context, subID, region string, client *az.HttpClient) (map[string]types.SKUAvailability, error) {
	return fetchVMSizes(ctx, subID, region, client)
}

// ── VMSS ────────────────────────────────────────────────────────────────────

// VMSSProvider reuses the same vmSizes endpoint as VMProvider — VMSS node SKUs
// are the same VM size names.
type VMSSProvider struct{}

func (VMSSProvider) ResourceType() string { return "microsoft.compute/virtualmachinescalesets" }

func (VMSSProvider) FetchSKUs(ctx context.Context, subID, region string, client *az.HttpClient) (map[string]types.SKUAvailability, error) {
	return fetchVMSizes(ctx, subID, region, client)
}

// fetchVMSizes fetches available VM sizes for a region via the per-region vmSizes endpoint.
// The endpoint lists only sizes deployable in the region for this subscription.
func fetchVMSizes(ctx context.Context, subID, region string, client *az.HttpClient) (map[string]types.SKUAvailability, error) {
	url := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Compute/locations/%s/vmSizes?api-version=2023-03-01",
		subID, region,
	)
	return FetchPagedSKUs(ctx, url, client, func(item vmSizeItem) (string, types.SKUAvailability, bool) {
		if item.Name == "" {
			return "", types.SKUAvailability{}, false
		}
		return item.Name, types.SKUAvailability{State: types.SKUAvailable}, true
	})
}

// ── Managed Disks ────────────────────────────────────────────────────────────

// diskSKUItem is one item from the Compute SKUs endpoint (filtered to a region).
// Unlike vmSizes, this endpoint includes restrictions and capabilities.
type diskSKUItem struct {
	Name         string           `json:"name"`
	Tier         string           `json:"tier"`
	Restrictions []SKURestriction `json:"restrictions"`
	Capabilities []SKUCapability  `json:"capabilities"`
}

// DiskProvider checks Managed Disk SKU availability via the global Compute SKUs
// endpoint filtered to a specific region.
type DiskProvider struct{}

func (DiskProvider) ResourceType() string { return "microsoft.compute/disks" }

func (DiskProvider) FetchSKUs(ctx context.Context, subID, region string, client *az.HttpClient) (map[string]types.SKUAvailability, error) {
	// The $filter parameter pre-filters the global list to the requested region,
	// so every returned item applies to this region.
	url := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Compute/skus?api-version=2023-03-01&$filter=location%%20eq%%20%%27%s%%27",
		subID, region,
	)
	return FetchPagedSKUs(ctx, url, client, func(item diskSKUItem) (string, types.SKUAvailability, bool) {
		name := item.Name
		if name == "" {
			name = item.Tier
		}
		if name == "" {
			return "", types.SKUAvailability{}, false
		}
		return name, CheckStandardRestrictions(item.Restrictions, item.Capabilities), true
	})
}

func init() {
	Register(VMProvider{})
	Register(VMSSProvider{})
	Register(DiskProvider{})
}
