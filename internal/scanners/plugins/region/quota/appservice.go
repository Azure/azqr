// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package quota queries Azure resource quota usage for a subscription+region.
package quota

import (
	"context"
	"fmt"

	"github.com/Azure/azqr/internal/az"
	"github.com/rs/zerolog/log"
)

// webSkipList contains Microsoft.Web usage counters that are clearly per-app or
// subscription-wide sub-resource limits rather than region-level capacity that
// affects migration placement decisions.
var webSkipList = map[string]bool{
	// Per-app hostname/binding limits, not independent regional capacity.
	"CustomDomains":    true,
	"HostNameBindings": true,
	"SslBindings":      true,
	"SslConnections":   true,

	// Certificate object counts are subscription-scoped inventory, not regional quota.
	"Certificates": true,
}

func keepAppServiceUsage(item usageItem) bool {
	return !webSkipList[item.Name.Value]
}

// FetchAppServiceQuota queries Microsoft.Web/locations/{region}/usages for App
// Service regional quotas. Returns nil, nil when the endpoint is not supported (405).
func FetchAppServiceQuota(ctx context.Context, httpClient *az.HttpClient, subscriptionID, region string) ([]UsageEntry, error) {
	url := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Web/locations/%s/usages?api-version=2023-01-01",
		subscriptionID, region,
	)
	log.Debug().Msgf("Querying App Service quota for subscription %s in %s", subscriptionID, region)

	entries, err := fetchUsages(ctx, httpClient, url, func(item usageItem) bool {
		if !keepAppServiceUsage(item) {
			return false
		}
		log.Debug().Msgf("App Service quota %s in %s: %s current=%d limit=%d",
			subscriptionID, region, item.Name.Value, item.CurrentValue, item.Limit)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("app service quota API error for %s in %s: %w", subscriptionID, region, err)
	}
	log.Debug().Msgf("Fetched %d App Service quota entries for %s in %s", len(entries), subscriptionID, region)
	return entries, nil
}
