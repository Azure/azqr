// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package quota queries Azure resource quota usage for a subscription+region.
package quota

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/rs/zerolog/log"
)

const (
	nearLimitThreshold = 0.15 // resources with < 15% headroom are flagged at-risk
)

// UsageEntry holds quota data for one resource type in one subscription+region.
// It is returned by both FetchVMQuota and FetchNetworkQuota.
type UsageEntry struct {
	// ResourceName is the API identifier (e.g. "standardDSv3Family", "VirtualNetworks").
	ResourceName    string
	LocalizedName   string
	CurrentValue    int
	Limit           int
	Available       int
	HeadroomPct     float64 // Available/Limit * 100
	IsNearLimit     bool    // HeadroomPct >= 0 && HeadroomPct < nearLimitThreshold*100
	IsAtOrOverLimit bool    // Available <= 0
}

// VMFamilyUsage is a type alias for UsageEntry kept for backwards compatibility.
// New code should use UsageEntry directly.
type VMFamilyUsage = UsageEntry

// usagesResponse is the ARM JSON envelope returned by the Compute and Network usages APIs.
// Azure paginates this response; follow NextLink until empty.
type usagesResponse struct {
	Value    []usageItem `json:"value"`
	NextLink string      `json:"nextLink"`
}

type usageItem struct {
	Unit         string    `json:"unit"`
	CurrentValue int       `json:"currentValue"`
	Limit        int       `json:"limit"`
	Name         usageName `json:"name"`
}

// usageName represents the name field in Azure usages APIs.
// Most providers return an object: {"value": "...", "localizedValue": "..."}.
// Some providers (e.g. Microsoft.Sql) return a plain string.
// UnmarshalJSON handles both forms transparently.
type usageName struct {
	Value          string
	LocalizedValue string
}

func (n *usageName) UnmarshalJSON(data []byte) error {
	// Try object form: {"value": "...", "localizedValue": "..."}
	var obj struct {
		Value          string `json:"value"`
		LocalizedValue string `json:"localizedValue"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		n.Value = obj.Value
		n.LocalizedValue = obj.LocalizedValue
		return nil
	}
	// Fall back to plain string (e.g. Microsoft.Sql usages API)
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = s
	n.LocalizedValue = s
	return nil
}

// fetchUsages is the shared pagination loop for quota usages APIs.
// keepFn returns true for items that should be included in the result.
// Returns nil, nil when the subscription or endpoint is not supported by the RP
// (HTTP 400, 404, 405) — callers should treat nil as "no data available".
func fetchUsages(ctx context.Context, httpClient *az.HttpClient, startURL string, keepFn func(usageItem) bool) ([]UsageEntry, error) {
	entries := make([]UsageEntry, 0)
	url := startURL
	pageNum := 0

	for url != "" {
		body, err := httpClient.Do(ctx, url)
		if err != nil {
			var respErr *azcore.ResponseError
			if errors.As(err, &respErr) {
				switch respErr.StatusCode {
				case http.StatusMethodNotAllowed, http.StatusNotFound, http.StatusBadRequest:
					log.Debug().Msgf("fetchUsages: RP does not support this subscription/endpoint (%d) for %s — skipping", respErr.StatusCode, startURL)
					return nil, nil
				}
			}
			return nil, fmt.Errorf("usages API error: %w", err)
		}

		var resp usagesResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse usages response (page %d): %w", pageNum, err)
		}

		for _, item := range resp.Value {
			if !keepFn(item) {
				continue
			}
			if item.Limit <= 0 {
				log.Debug().Msgf("fetchUsages: skipping %s (limit=%d, treated as unlimited/unavailable)", item.Name.Value, item.Limit)
				continue // unlimited or unavailable — skip to avoid misleading math
			}

			available := item.Limit - item.CurrentValue
			headroomPct := float64(available) / float64(item.Limit) * 100
			entries = append(entries, UsageEntry{
				ResourceName:    item.Name.Value,
				LocalizedName:   item.Name.LocalizedValue,
				CurrentValue:    item.CurrentValue,
				Limit:           item.Limit,
				Available:       available,
				HeadroomPct:     headroomPct,
				IsNearLimit:     headroomPct < nearLimitThreshold*100,
				IsAtOrOverLimit: available <= 0,
			})
		}

		url = resp.NextLink
		pageNum++
	}
	return entries, nil
}

// FetchVMQuota queries Microsoft.Compute/locations/{region}/usages for VM family quotas.
// It returns only VM family-level counters (names containing "Family") so that the
// aggregate "cores" (Total Regional vCPUs) entry is excluded — that counter reflects ALL
// workloads in the region (including unrelated ones and stopped-but-not-deallocated VMs),
// making it noisy and misleading for migration planning. Family-level entries are
// actionable: they tell you whether the specific VM families you need have available quota.
// The caller should treat a nil return as "no data available".
func FetchVMQuota(ctx context.Context, cred azcore.TokenCredential, clientOpts *arm.ClientOptions, subscriptionID, region string) ([]UsageEntry, error) {
	client, err := armcompute.NewUsageClient(subscriptionID, cred, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute usage client for %s: %w", subscriptionID, err)
	}

	log.Debug().Msgf("Querying VM quota for subscription %s in %s", subscriptionID, region)

	var entries []UsageEntry
	pager := client.NewListPager(region, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("VM quota API error for %s in %s: %w", subscriptionID, region, err)
		}
		for _, item := range page.Value {
			if item.Name == nil || item.Name.Value == nil || !strings.Contains(*item.Name.Value, "Family") {
				continue
			}
			if item.Limit == nil {
				continue
			}
			if *item.Limit <= 0 {
				log.Debug().Msgf("fetchVMQuota: skipping %s (limit=%d, treated as unlimited/unavailable)", *item.Name.Value, *item.Limit)
				continue
			}

			cv := int(*item.CurrentValue)
			lim := int(*item.Limit)
			available := lim - cv
			headroomPct := float64(available) / float64(lim) * 100

			localizedName := ""
			if item.Name.LocalizedValue != nil {
				localizedName = *item.Name.LocalizedValue
			}
			log.Debug().Msgf("VM quota %s in %s: %s current=%d limit=%d", subscriptionID, region, *item.Name.Value, cv, lim)
			entries = append(entries, UsageEntry{
				ResourceName:    *item.Name.Value,
				LocalizedName:   localizedName,
				CurrentValue:    cv,
				Limit:           lim,
				Available:       available,
				HeadroomPct:     headroomPct,
				IsNearLimit:     headroomPct < nearLimitThreshold*100,
				IsAtOrOverLimit: available <= 0,
			})
		}
	}

	log.Debug().Msgf("Fetched %d VM family quota entries for %s in %s", len(entries), subscriptionID, region)
	return entries, nil
}

// AtRiskSummaries returns a human-readable list of at-risk entries from a UsageEntry slice.
// The format is "<LocalizedName> (<currentValue>/<limit>, <pctUsed>% used)" for each entry
// with < 15% headroom.
func AtRiskSummaries(usages []UsageEntry) []string {
	var risk []string
	for _, u := range usages {
		if u.IsNearLimit || u.IsAtOrOverLimit {
			pctUsed := 100.0 - u.HeadroomPct
			label := u.LocalizedName
			if label == "" {
				label = u.ResourceName
			}
			risk = append(risk, fmt.Sprintf("%s (%d/%d, %.0f%% used)", label, u.CurrentValue, u.Limit, pctUsed))
		}
	}
	return risk
}
