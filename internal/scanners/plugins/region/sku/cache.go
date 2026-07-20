// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sku

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
)

// Cache caches SKU availability results per region to avoid redundant API calls.
// A singleflight.Group ensures that concurrent cache misses for the same key
// trigger only one HTTP request; all waiters share the result.
//
// This replaces types.SKUAvailabilityCache, which has been removed.
type Cache struct {
	cache map[string]map[string]types.SKUAvailability // key → map[skuName]SKUAvailability
	mu    sync.RWMutex
	group singleflight.Group
}

// NewCache creates a new SKU availability cache.
func NewCache() *Cache {
	return &Cache{
		cache: make(map[string]map[string]types.SKUAvailability),
	}
}

// GetSKUAvailability returns the SKU availability map for the given resource type
// and region. Concurrent callers with the same key share a single in-flight
// request; errors are not cached so a subsequent call will retry the API.
//
// Returns (nil, error) when the provider API call fails.
// Returns (emptyMap, nil) when the API succeeds but reports no SKUs for the region.
// Returns (nil, error) with a descriptive message when no provider is registered.
func (c *Cache) GetSKUAvailability(
	ctx context.Context,
	subscriptionID string,
	resourceType string,
	targetRegion string,
	httpClient *az.HttpClient,
) (map[string]types.SKUAvailability, error) {
	resourceType = strings.ToLower(resourceType)
	targetRegion = types.NormalizeRegionName(targetRegion)

	provider := Get(resourceType)
	if provider == nil {
		return nil, fmt.Errorf("no SKU provider registered for resource type: %s", resourceType)
	}

	// Cache key: subscriptionID:resourceType:region
	// SKU restrictions are subscription-scoped, so subscriptionID must be part of the key.
	cacheKey := subscriptionID + ":" + resourceType + ":" + targetRegion

	// Fast path: already cached.
	c.mu.RLock()
	if cached, exists := c.cache[cacheKey]; exists {
		c.mu.RUnlock()
		log.Debug().Msgf("Using cached SKU availability for %s in %s (%d SKUs)", resourceType, targetRegion, len(cached))
		return cached, nil
	}
	c.mu.RUnlock()

	// Coalesce concurrent cache misses for the same key into a single HTTP request.
	v, err, _ := c.group.Do(cacheKey, func() (any, error) {
		// Re-check cache inside singleflight to avoid a redundant API call if a
		// concurrent request for the same key finished while we were waiting.
		c.mu.RLock()
		if cached, exists := c.cache[cacheKey]; exists {
			c.mu.RUnlock()
			return cached, nil
		}
		c.mu.RUnlock()

		available, err := provider.FetchSKUs(ctx, subscriptionID, targetRegion, httpClient)
		if err != nil {
			// Do not cache errors — a subsequent call should retry the API.
			return nil, err
		}

		// Normalize keys to lowercase so callers can look up with strings.ToLower
		// without knowing the case convention of the underlying Azure API.
		normalised := make(map[string]types.SKUAvailability, len(available))
		for k, v := range available {
			normalised[strings.ToLower(strings.TrimSpace(k))] = v
		}

		c.mu.Lock()
		c.cache[cacheKey] = normalised
		c.mu.Unlock()

		log.Debug().Msgf("Cached SKU availability for %s in %s (%d SKUs)", resourceType, targetRegion, len(normalised))
		return normalised, nil
	})
	if err != nil {
		return nil, err
	}

	return v.(map[string]types.SKUAvailability), nil
}
