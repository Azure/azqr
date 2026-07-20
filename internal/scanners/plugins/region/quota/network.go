// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package quota

import (
	"context"
	"fmt"

	"github.com/Azure/azqr/internal/az"
	"github.com/rs/zerolog/log"
)

// networkSkipList contains Microsoft.Network usage counters that are either
// singletons auto-managed by Azure (not user-provisioned resources) or
// informational totals rather than actionable per-resource quotas.
var networkSkipList = map[string]bool{
	// Azure auto-creates exactly one Network Watcher per region per subscription.
	// It is always 1/1 and cannot be avoided — not a migration concern.
	"NetworkWatchers": true,
	// Per-resource sub-limits (not standalone countable resources).
	"RouteFilterRulesPerRouteFilter":        true,
	"RouteFiltersPerExpressRouteBgpPeering": true,
	"RoutesPerExpressRouteCircuit":          true,
	"BgpCommunityFilterRulesPerRouteFilter": true,
}

// FetchNetworkQuota queries Microsoft.Network/locations/{region}/usages for network
// resource quotas in the target region. It returns all non-trivial usage entries so
// callers can determine whether the target region has room to absorb migrated network
// resources. The caller should treat a nil return as "no data available".
func FetchNetworkQuota(ctx context.Context, httpClient *az.HttpClient, subscriptionID, region string) ([]UsageEntry, error) {
	url := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Network/locations/%s/usages?api-version=2022-07-01",
		subscriptionID, region,
	)
	log.Debug().Msgf("Querying network quota for subscription %s in %s", subscriptionID, region)

	entries, err := fetchUsages(ctx, httpClient, url, func(item usageItem) bool {
		if networkSkipList[item.Name.Value] {
			return false
		}
		log.Debug().Msgf("Network quota %s in %s: %s current=%d limit=%d",
			subscriptionID, region, item.Name.Value, item.CurrentValue, item.Limit)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("network quota API error for %s in %s: %w", subscriptionID, region, err)
	}
	log.Debug().Msgf("Fetched %d network quota entries for %s in %s", len(entries), subscriptionID, region)
	return entries, nil
}
