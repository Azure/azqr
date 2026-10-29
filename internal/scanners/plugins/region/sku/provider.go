// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package sku provides typed Azure SKU availability providers for the region selection plugin.
// Each Azure resource provider is implemented as a [SKUProvider] that knows how to fetch and
// interpret SKU availability for its specific ARM endpoint and response shape.
//
// To add a new provider, implement [SKUProvider] and call [Register] in an init() function.
// All providers in this package are registered automatically when the package is imported.
package sku

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

// SKUProvider knows how to fetch and interpret SKU availability for one Azure resource type.
type SKUProvider interface {
	// ResourceType returns the lowercase ARM resource type,
	// e.g. "microsoft.compute/virtualmachines".
	ResourceType() string

	// FetchSKUs returns a map of lowercase SKU name → availability state for the
	// given subscription and region. An API error (404, 403, throttling) must be
	// returned as a non-nil error, not as an empty map — the cache distinguishes
	// between "no SKUs available" (empty map) and "API call failed" (error).
	FetchSKUs(ctx context.Context, subscriptionID, region string, client *az.HttpClient) (map[string]types.SKUAvailability, error)
}

var registry = map[string]SKUProvider{}

// Register adds a provider to the global registry. It is typically called from init().
// Registering the same resource type twice panics.
func Register(p SKUProvider) {
	key := strings.ToLower(p.ResourceType())
	if _, exists := registry[key]; exists {
		panic(fmt.Sprintf("sku: duplicate provider registration for %q", key))
	}
	registry[key] = p
}

// Get returns the provider for the given resource type, or nil if none is registered.
func Get(resourceType string) SKUProvider {
	return registry[strings.ToLower(resourceType)]
}

// skuRank orders states from best (lowest) to worst for best-wins merging.
func skuRank(s types.SKUAvailabilityState) int {
	switch s {
	case types.SKUAvailable:
		return 0
	case types.SKUZoneRestricted:
		return 1
	case types.SKURestricted:
		return 2
	default:
		return 3
	}
}

// mergeBestSKU records avail under key; when the API returns several entries for the same
// SKU name (for example one per size), the best state wins.
func mergeBestSKU(m map[string]types.SKUAvailability, key string, avail types.SKUAvailability) {
	if existing, ok := m[key]; ok && skuRank(existing.State) <= skuRank(avail.State) {
		return
	}
	m[key] = avail
}

// armPage is the standard ARM list-response envelope used by most SKU endpoints.
type armPage[T any] struct {
	Value    []T    `json:"value"`
	NextLink string `json:"nextLink"`
}

// FetchPagedSKUs fetches one or more pages from a standard ARM { "value": [T] } endpoint
// and calls extract for each item. Returning ok=false from extract skips the item.
// Pagination via nextLink is handled automatically.
//
// Use this helper for providers whose endpoint returns the standard ARM list envelope.
// For non-standard response shapes, implement FetchSKUs directly without this helper.
func FetchPagedSKUs[T any](
	ctx context.Context,
	startURL string,
	client *az.HttpClient,
	extract func(item T) (skuName string, avail types.SKUAvailability, ok bool),
) (map[string]types.SKUAvailability, error) {
	items, err := FetchAllItems[T](ctx, startURL, client)
	if err != nil {
		return nil, err
	}
	result := make(map[string]types.SKUAvailability)
	for _, item := range items {
		if name, avail, ok := extract(item); ok {
			mergeBestSKU(result, strings.ToLower(name), avail)
		}
	}
	return result, nil
}

// FetchAllItems follows nextLink pagination and returns every item from a standard ARM list endpoint.
func FetchAllItems[T any](ctx context.Context, startURL string, client *az.HttpClient) ([]T, error) {
	var items []T
	url := startURL
	for url != "" {
		body, err := client.Do(ctx, url)
		if err != nil {
			return nil, err
		}
		var page armPage[T]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("sku: parse response from %s: %w", url, err)
		}
		items = append(items, page.Value...)
		url = page.NextLink
	}
	return items, nil
}
