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
	result := make(map[string]types.SKUAvailability)
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
		for _, item := range page.Value {
			if name, avail, ok := extract(item); ok {
				result[strings.ToLower(name)] = avail
			}
		}
		url = page.NextLink
	}
	return result, nil
}
