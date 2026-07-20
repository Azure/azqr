// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package quota queries Azure resource quota usage for a subscription+region.
package quota

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/rs/zerolog/log"
)

// sqlSkipList contains suffixes for Microsoft.Sql usage counters that represent
// per-resource sub-limits rather than regional quotas.
var sqlSkipList = map[string]bool{
	"PerServer":   true,
	"PerDatabase": true,
}

// FetchSQLQuota queries Microsoft.Sql/locations/{region}/usages for regional SQL
// quotas. It skips per-server and per-database sub-limits so callers only see
// actionable regional counters such as Servers, ElasticPools, DTUs, and vCores.
// The caller should treat a nil return as "no data available".
func FetchSQLQuota(ctx context.Context, httpClient *az.HttpClient, subscriptionID, region string) ([]UsageEntry, error) {
	url := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Sql/locations/%s/usages?api-version=2021-11-01",
		subscriptionID, region,
	)
	log.Debug().Msgf("Querying SQL quota for subscription %s in %s", subscriptionID, region)

	entries, err := fetchUsages(ctx, httpClient, url, func(item usageItem) bool {
		for suffix := range sqlSkipList {
			if strings.HasSuffix(item.Name.Value, suffix) {
				return false
			}
		}
		log.Debug().Msgf("SQL quota %s in %s: %s current=%d limit=%d",
			subscriptionID, region, item.Name.Value, item.CurrentValue, item.Limit)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("SQL quota API error for %s in %s: %w", subscriptionID, region, err)
	}
	if len(entries) == 0 {
		log.Debug().Msgf("SQL quota returned 0 usable entries for subscription %s in %s (all items may have limit=0 or limit=-1)", subscriptionID, region)
	} else {
		log.Debug().Msgf("Fetched %d SQL quota entries for %s in %s", len(entries), subscriptionID, region)
	}
	return entries, nil
}
