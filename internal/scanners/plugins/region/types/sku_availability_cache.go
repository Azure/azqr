// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package types

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/scanners/plugins/region/config"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
)

// SKUAvailabilityCache caches SKU availability results per region to avoid redundant API calls.
// A singleflight.Group ensures that concurrent cache misses for the same key trigger only one
// HTTP request; all waiters share the result.
type SKUAvailabilityCache struct {
	cache map[string]map[string]SKUAvailability // subscriptionID:resourceType:region -> map[skuName]SKUAvailability
	mu    sync.RWMutex
	group singleflight.Group
}

// NewSKUAvailabilityCache creates a new SKU availability cache
func NewSKUAvailabilityCache() *SKUAvailabilityCache {
	return &SKUAvailabilityCache{
		cache: make(map[string]map[string]SKUAvailability),
	}
}

// ExtractSKUsFromResponse extracts SKU availability from API response items.
// For global-list endpoints (regionalApi: false), each item is filtered to ensure
// it applies to targetRegion by checking locationInfo[].location (Compute) or
// locations[] (Storage). Items with no location data are accepted as globally available.
func (c *SKUAvailabilityCache) ExtractSKUsFromResponse(
	items []SKUAPIItem,
	config *config.PropertyMapConfig,
	targetRegion string,
) map[string]SKUAvailability {
	available := make(map[string]SKUAvailability)

	for i := range items {
		item := &items[i]

		if !config.RegionalAPI && !c.SKUAppliesToRegion(item, targetRegion) {
			continue
		}

		skuName := c.ExtractSKUName(item, config)
		if skuName == "" {
			continue
		}

		avail := c.CheckSKURestrictions(item)
		available[strings.ToLower(skuName)] = avail
	}

	return available
}

// SKUAppliesToRegion returns true if a SKU item from a global-list API applies
// to the given region. It checks locationInfo[].location (Microsoft.Compute/skus)
// and locations[] (Microsoft.Storage/skus). Items that carry neither field are
// accepted as globally available.
func (c *SKUAvailabilityCache) SKUAppliesToRegion(item *SKUAPIItem, targetRegion string) bool {
	target := NormalizeRegionName(targetRegion)

	if len(item.LocationInfo) > 0 {
		for _, li := range item.LocationInfo {
			if NormalizeRegionName(li.Location) == target {
				return true
			}
		}
		return false
	}

	if len(item.Locations) > 0 {
		for _, loc := range item.Locations {
			if NormalizeRegionName(loc) == target {
				return true
			}
		}
		return false
	}

	return true
}

// ExtractSKUName extracts the SKU name from an API response item using the
// configured property map.
func (c *SKUAvailabilityCache) ExtractSKUName(item *SKUAPIItem, config *config.PropertyMapConfig) string {
	// Use configured top-level property name for the SKU name field if set.
	if config.Properties.TopLevelProperties != nil {
		if nameProp, exists := config.Properties.TopLevelProperties["name"]; exists {
			switch nameProp {
			case "name":
				if item.Name != "" {
					return item.Name
				}
			case "size":
				if item.Size != "" {
					return item.Size
				}
			case "tier":
				if item.Tier != "" {
					return item.Tier
				}
			}
		}
	}

	// Fallback: use name > size > tier in priority order.
	if item.Name != "" {
		return item.Name
	}
	if item.Size != "" {
		return item.Size
	}
	return item.Tier
}

// CheckSKURestrictions returns the availability detail for a SKU based on its
// restrictions and capability flags.
//
// Priority order (highest severity first):
//  1. Location restriction → SKUUnavailable (hard block) or SKURestricted (quota-liftable)
//  2. Zone restriction     → SKUZoneRestricted with blocked zones populated
//  3. Capability flag      → SKUUnavailable
//  4. Default              → SKUAvailable
func (c *SKUAvailabilityCache) CheckSKURestrictions(item *SKUAPIItem) SKUAvailability {
	// 1. Check for regional (Location) restrictions — most severe.
	for _, r := range item.Restrictions {
		if !strings.EqualFold(r.Type, "Location") {
			continue
		}
		if strings.EqualFold(r.ReasonCode, "NotAvailableForSubscription") {
			return SKUAvailability{State: SKURestricted}
		}
		return SKUAvailability{State: SKUUnavailable}
	}

	// 2. Check for zone-level restrictions — SKU is usable in region but some zones are blocked.
	var hasZoneRestriction bool
	var blockedZones []string
	for _, r := range item.Restrictions {
		if !strings.EqualFold(r.Type, "Zone") {
			continue
		}
		hasZoneRestriction = true
		blockedZones = append(blockedZones, r.RestrictionInfo.Zones...)
	}
	if hasZoneRestriction {
		return SKUAvailability{State: SKUZoneRestricted, BlockedZones: blockedZones}
	}

	// 3. Check capability flags.
	for _, cap := range item.Capabilities {
		if strings.EqualFold(cap.Name, "available") && !strings.EqualFold(cap.Value, "true") {
			return SKUAvailability{State: SKUUnavailable}
		}
	}

	return SKUAvailability{State: SKUAvailable}
}

// GetSKUAvailability checks if SKUs are available in a target region.
// Returns a map of SKU name to SKUAvailability detail, or an error if the API could not be reached.
// Concurrent callers with the same key share a single in-flight request; errors are not cached.
func (c *SKUAvailabilityCache) GetSKUAvailability(
	ctx context.Context,
	subscriptionID string,
	resourceType string,
	targetRegion string,
	httpClient *az.HttpClient,
) (map[string]SKUAvailability, error) {
	// Normalize resource type and region
	resourceType = strings.ToLower(resourceType)
	targetRegion = NormalizeRegionName(targetRegion)

	// Cache key includes subscriptionID because SKU restrictions are subscription-scoped.
	// Use plain concatenation to avoid the fmt.Sprintf vararg heap alloc on every lookup.
	cacheKey := subscriptionID + ":" + resourceType + ":" + targetRegion

	// Fast path: already cached
	c.mu.RLock()
	if cached, exists := c.cache[cacheKey]; exists {
		c.mu.RUnlock()
		log.Debug().Msgf("Using cached SKU availability for %s in %s (%d SKUs)", resourceType, targetRegion, len(cached))
		return cached, nil
	}
	c.mu.RUnlock()

	// Get property map configuration for this resource type
	propertyMap := config.GetPropertyMapConfig(resourceType)
	if propertyMap == nil {
		log.Debug().Msgf("No property map configuration for resource type: %s", resourceType)
		return nil, fmt.Errorf("no API configuration for resource type: %s", resourceType)
	}

	// Coalesce concurrent cache misses for the same key into a single HTTP request.
	v, err, _ := c.group.Do(cacheKey, func() (interface{}, error) {
		// Re-check cache inside the singleflight to avoid a redundant API call when a
		// previous call for the same key just finished and wrote to the cache.
		c.mu.RLock()
		if cached, exists := c.cache[cacheKey]; exists {
			c.mu.RUnlock()
			return cached, nil
		}
		c.mu.RUnlock()

		available, err := c.querySKUAvailabilityAPI(ctx, subscriptionID, targetRegion, propertyMap, httpClient)
		if err != nil {
			// Do not cache errors — a subsequent call should retry the API.
			return nil, err
		}

		// Cache only successful results
		c.mu.Lock()
		c.cache[cacheKey] = available
		c.mu.Unlock()

		log.Debug().Msgf("Cached SKU availability for %s in %s (%d SKUs)", resourceType, targetRegion, len(available))
		return available, nil
	})
	if err != nil {
		return nil, err
	}

	return v.(map[string]SKUAvailability), nil
}

// querySKUAvailabilityAPI queries Azure SKU availability APIs based on configuration.
func (c *SKUAvailabilityCache) querySKUAvailabilityAPI(
	ctx context.Context,
	subscriptionID string,
	region string,
	config *config.PropertyMapConfig,
	httpClient *az.HttpClient,
) (map[string]SKUAvailability, error) {
	uri := config.URI
	uri = strings.ReplaceAll(uri, "{0}", subscriptionID)
	uri = strings.ReplaceAll(uri, "{1}", region)

	if !strings.HasPrefix(uri, "https://") {
		uri = "https://management.azure.com" + uri
	}

	log.Debug().Msgf("Querying SKU availability API: %s", uri)

	body, err := httpClient.Do(ctx, uri)
	if err != nil {
		if httpErr, ok := err.(*az.HTTPError); ok {
			return nil, fmt.Errorf("SKU availability API HTTP %d for %s: %s", httpErr.StatusCode, uri, httpErr.Body)
		}
		return nil, err
	}

	var response SKUAPIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	available := c.ExtractSKUsFromResponse(response.Value, config, region)
	log.Debug().Msgf("Found %d SKUs available for region %s", len(available), region)

	return available, nil
}
