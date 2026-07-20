// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sku

import (
	"context"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

// storageSKUItem is one item from the Storage SKUs global-list endpoint.
// The endpoint is not per-region; each item carries a locations[] list.
type storageSKUItem struct {
	Name         string           `json:"name"`
	Tier         string           `json:"tier"`
	Locations    []string         `json:"locations"`
	Restrictions []SKURestriction `json:"restrictions"`
	Capabilities []SKUCapability  `json:"capabilities"`
}

// StorageProvider checks Storage Account SKU availability.
// The Storage SKUs API returns a global list; this provider filters it to the
// requested region by checking item.Locations.
type StorageProvider struct{}

func (StorageProvider) ResourceType() string { return "microsoft.storage/storageaccounts" }

func (StorageProvider) FetchSKUs(ctx context.Context, subID, region string, client *az.HttpClient) (map[string]types.SKUAvailability, error) {
	endpoint := "https://management.azure.com/subscriptions/" + subID + "/providers/Microsoft.Storage/skus?api-version=2023-01-01"
	target := types.NormalizeRegionName(region)

	return FetchPagedSKUs(ctx, endpoint, client,
		func(item storageSKUItem) (string, types.SKUAvailability, bool) {
			// Filter: only accept items that list the target region.
			appliesTo := len(item.Locations) == 0 // no locations = globally available
			for _, loc := range item.Locations {
				if strings.EqualFold(types.NormalizeRegionName(loc), target) {
					appliesTo = true
					break
				}
			}
			if !appliesTo {
				return "", types.SKUAvailability{}, false
			}

			name := item.Name
			if name == "" {
				name = item.Tier
			}
			if name == "" {
				return "", types.SKUAvailability{}, false
			}

			return name, CheckStandardRestrictions(item.Restrictions, item.Capabilities), true
		},
	)
}

func init() { Register(StorageProvider{}) }
