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

// storageSkipList contains Microsoft.Storage usage counters that represent
// nested data-plane objects inside storage accounts rather than standalone
// regional ARM resources. These counters are noisy for region-selection
// because they do not indicate whether the region can accept more storage
// accounts or other top-level storage resources.
var storageSkipList = map[string]bool{
	"TotalBlobContainers": true,
	"TotalBlobs":          true,
	"TotalContainers":     true,
	"TotalFileShares":     true,
	"TotalQueues":         true,
	"TotalTables":         true,
}

// FetchStorageQuota queries Microsoft.Storage/locations/{region}/usages for
// storage resource quotas in the target region. It keeps actionable regional
// counters such as StorageAccounts and any other non-noisy entries returned by
// the provider, while skipping nested sub-resource totals. The caller should
// treat a nil return as "no data available".
func FetchStorageQuota(ctx context.Context, httpClient *az.HttpClient, subscriptionID, region string) ([]UsageEntry, error) {
	url := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Storage/locations/%s/usages?api-version=2023-01-01",
		subscriptionID, region,
	)
	log.Debug().Msgf("Querying storage quota for subscription %s in %s", subscriptionID, region)

	entries, err := fetchUsages(ctx, httpClient, url, func(item usageItem) bool {
		if storageSkipList[item.Name.Value] {
			return false
		}
		log.Debug().Msgf("Storage quota %s in %s: %s current=%d limit=%d",
			subscriptionID, region, item.Name.Value, item.CurrentValue, item.Limit)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("storage quota API error for %s in %s: %w", subscriptionID, region, err)
	}
	log.Debug().Msgf("Fetched %d storage quota entries for %s in %s", len(entries), subscriptionID, region)
	return entries, nil
}
