// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmmodernization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/scanners/plugins/region/quota"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
)

const (
	skuAvailabilitySheetName = "VM Modernization - SKU Avail"

	availabilityAvailable      = "Available"
	availabilityRestricted     = "Restricted"
	availabilityZoneRestricted = "Zone restricted"
	availabilityNotOffered     = "Not offered"
	availabilityUnknown        = "Unknown"

	quotaSufficient    = "Sufficient"
	quotaNearLimit     = "Near limit"
	quotaInsufficient  = "Insufficient"
	quotaUnknown       = "Unknown"
	quotaNotApplicable = "N/A"

	// quotaNearLimitFraction flags a family whose remaining quota, after the
	// planned VMs are placed, drops below 15% of its limit. It matches the
	// threshold the region plugin uses for its own quota sheet.
	quotaNearLimitFraction = 0.15

	skuAvailabilityMaxWorkers = 5
	skuAvailabilityAPIVersion = "2021-07-01"
)

// resourceManagerBaseURL returns the ARM endpoint for the active cloud. It is
// a variable so tests can redirect it to a local server.
var resourceManagerBaseURL = az.GetResourceManagerEndpoint

// fetchVMFamilyQuota returns the per-family VM quota for one subscription and
// region. It is a variable so tests can replace the ARM SDK call.
var fetchVMFamilyQuota = func(ctx context.Context, cred azcore.TokenCredential, subscriptionID, region string) ([]quota.UsageEntry, error) {
	return quota.FetchVMQuotaIncludingZeroLimits(ctx, cred, az.NewDefaultClientOptions(), subscriptionID, region)
}

// subscriptionRegion is the cache key for one SKU and quota fetch.
type subscriptionRegion struct {
	SubscriptionID string
	Region         string
}

// vmSKUDetail is what the Compute SKUs API reports for one VM size in a region.
type vmSKUDetail struct {
	Family       string
	VCPUs        int
	Availability string
	Reasons      []string
	Zones        []string
}

// regionSKUData holds everything fetched for one subscription and region.
// SKUs is nil when the SKU fetch failed; Quota is nil when the quota fetch
// failed or returned no data.
type regionSKUData struct {
	SKUs  map[string]vmSKUDetail
	Quota map[string]quota.UsageEntry
}

type skuAPIRestriction struct {
	Type            string `json:"type"`
	ReasonCode      string `json:"reasonCode"`
	RestrictionInfo struct {
		Zones []string `json:"zones"`
	} `json:"restrictionInfo"`
}

type skuAPICapability struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type skuAPIItem struct {
	Name         string              `json:"name"`
	ResourceType string              `json:"resourceType"`
	Family       string              `json:"family"`
	Capabilities []skuAPICapability  `json:"capabilities"`
	Restrictions []skuAPIRestriction `json:"restrictions"`
}

type skuAPIPage struct {
	Value    []skuAPIItem `json:"value"`
	NextLink string       `json:"nextLink"`
}

// concreteTargetSKU returns the Suggested Target SKU only when it is an actual
// size to evaluate, not a pool-managed, already-on-target or manual-review row.
func (r vmModernizationRow) concreteTargetSKU() (string, bool) {
	patternCode, _ := classifyWorkloadPattern(bool(r.HasScaleSet), bool(r.IsAKSManaged), bool(r.IsDatabricksManaged), r.AVDHostPoolType, r.ImagePublisher, r.ImageOffer, r.IsAROManaged)
	if !isActionableWorkloadPattern(patternCode) {
		return "", false
	}
	if _, isTargetGeneration := classifySeriesGeneration(r.VMSize); isTargetGeneration {
		return "", false
	}
	name, _, ok := suggestTargetSKU(r.VMSize)
	return name, ok
}

// evaluateVMSKU interprets the restrictions and capabilities of one SKU.
// Location restrictions take priority over zone restrictions.
func evaluateVMSKU(item skuAPIItem) vmSKUDetail {
	detail := vmSKUDetail{Family: item.Family, Availability: availabilityAvailable}
	for _, c := range item.Capabilities {
		if strings.EqualFold(c.Name, "vCPUs") {
			if v, err := strconv.Atoi(strings.TrimSpace(c.Value)); err == nil {
				detail.VCPUs = v
			}
		}
	}

	reasons := map[string]bool{}
	zones := map[string]bool{}
	locationRestricted := false
	for _, r := range item.Restrictions {
		switch {
		case strings.EqualFold(r.Type, "Location"):
			locationRestricted = true
			if r.ReasonCode != "" {
				reasons[r.ReasonCode] = true
			}
		case strings.EqualFold(r.Type, "Zone"):
			for _, z := range r.RestrictionInfo.Zones {
				zones[z] = true
			}
			if r.ReasonCode != "" {
				reasons[r.ReasonCode] = true
			}
		}
	}
	detail.Reasons = sortedKeys(reasons)
	detail.Zones = sortedKeys(zones)

	switch {
	case locationRestricted:
		detail.Availability = availabilityRestricted
	case len(detail.Zones) > 0:
		detail.Availability = availabilityZoneRestricted
	default:
		for _, c := range item.Capabilities {
			if strings.EqualFold(c.Name, "available") && !strings.EqualFold(c.Value, "true") {
				detail.Availability = availabilityNotOffered
			}
		}
	}
	return detail
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// fetchRegionVMSKUs lists the virtual machine SKUs of one region, following
// nextLink, and returns them keyed by lowercase size name.
func fetchRegionVMSKUs(ctx context.Context, httpClient *az.HttpClient, subscriptionID, region string) (map[string]vmSKUDetail, error) {
	next := fmt.Sprintf("%s/subscriptions/%s/providers/Microsoft.Compute/skus?api-version=%s&$filter=%s",
		resourceManagerBaseURL(), subscriptionID, skuAvailabilityAPIVersion, url.QueryEscape("location eq '"+region+"'"))

	skus := map[string]vmSKUDetail{}
	for next != "" {
		body, err := httpClient.Do(ctx, next)
		if err != nil {
			return nil, err
		}
		var page skuAPIPage
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("failed to parse Compute SKUs response: %w", err)
		}
		for _, item := range page.Value {
			if !strings.EqualFold(item.ResourceType, "virtualMachines") || item.Name == "" {
				continue
			}
			skus[strings.ToLower(item.Name)] = evaluateVMSKU(item)
		}
		next = page.NextLink
	}
	return skus, nil
}

// collectRegionSKUData fetches SKU availability and family quota once for each
// unique (subscription, region) pair, concurrently. A failed fetch is logged
// and leaves the matching field nil, so its rows read Unknown instead of
// failing the scan.
func collectRegionSKUData(ctx context.Context, cred azcore.TokenCredential, httpClient *az.HttpClient, pairs []subscriptionRegion) map[subscriptionRegion]*regionSKUData {
	data := make(map[subscriptionRegion]*regionSKUData, len(pairs))
	if len(pairs) == 0 || httpClient == nil {
		return data
	}

	jobs := make(chan subscriptionRegion, len(pairs))
	for _, p := range pairs {
		jobs <- p
	}
	close(jobs)

	workers := skuAvailabilityMaxWorkers
	if len(pairs) < workers {
		workers = len(pairs)
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pair := range jobs {
				result := &regionSKUData{}
				maskedID := renderers.MaskSubscriptionID(pair.SubscriptionID, true)

				skus, err := fetchRegionVMSKUs(ctx, httpClient, pair.SubscriptionID, pair.Region)
				if err != nil {
					log.Warn().Err(err).Msgf("VM SKU availability unavailable for %s in %s; availability will read Unknown", maskedID, pair.Region)
				} else {
					result.SKUs = skus
				}

				usages, err := fetchVMFamilyQuota(ctx, cred, pair.SubscriptionID, pair.Region)
				if err != nil {
					log.Warn().Err(err).Msgf("VM family quota unavailable for %s in %s; quota status will read Unknown", maskedID, pair.Region)
				} else if len(usages) > 0 {
					result.Quota = make(map[string]quota.UsageEntry, len(usages))
					for _, u := range usages {
						result.Quota[strings.ToLower(u.ResourceName)] = u
					}
				}

				mu.Lock()
				data[pair] = result
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return data
}

// availabilityKey identifies one row of the availability sheet.
type availabilityKey struct {
	SubscriptionID string
	Region         string
	TargetSKU      string
}

// availabilityGroup accumulates the VMs that map to one availabilityKey.
type availabilityGroup struct {
	VMCount     int
	SourceSizes map[string]int
}

// groupAvailabilityTargets groups VM rows by (subscription, region, suggested
// target SKU). Rows without a concrete target SKU are skipped. When
// targetRegions is empty each VM is evaluated in its own region; otherwise
// every VM is evaluated once in each target region.
func groupAvailabilityTargets(rows []vmModernizationRow, targetRegions []string) map[availabilityKey]*availabilityGroup {
	groups := map[availabilityKey]*availabilityGroup{}
	for _, r := range rows {
		target, ok := r.concreteTargetSKU()
		if !ok {
			continue
		}
		regions := targetRegions
		if len(regions) == 0 {
			regions = []string{normalizeRegion(r.Location)}
		}
		for _, region := range regions {
			key := availabilityKey{SubscriptionID: r.SubscriptionID, Region: region, TargetSKU: target}
			g := groups[key]
			if g == nil {
				g = &availabilityGroup{SourceSizes: map[string]int{}}
				groups[key] = g
			}
			g.VMCount++
			g.SourceSizes[r.VMSize]++
		}
	}
	return groups
}

// parseTargetRegions reads the shared target-regions stage option (a
// comma-separated string) and returns de-duplicated normalized regions.
func parseTargetRegions(params *models.ScanParams) []string {
	if params == nil || params.Stages == nil {
		return nil
	}
	raw, ok := params.Stages.GetStageOptions(models.StageNamePlugin)["target-regions"].(string)
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	var regions []string
	for _, part := range strings.Split(raw, ",") {
		region := normalizeRegion(strings.TrimSpace(part))
		if region != "" && !seen[region] {
			seen[region] = true
			regions = append(regions, region)
		}
	}
	return regions
}

func regionPairs(groups map[availabilityKey]*availabilityGroup) []subscriptionRegion {
	seen := map[subscriptionRegion]bool{}
	var pairs []subscriptionRegion
	for key := range groups {
		pair := subscriptionRegion{SubscriptionID: key.SubscriptionID, Region: key.Region}
		if !seen[pair] {
			seen[pair] = true
			pairs = append(pairs, pair)
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].SubscriptionID != pairs[j].SubscriptionID {
			return pairs[i].SubscriptionID < pairs[j].SubscriptionID
		}
		return pairs[i].Region < pairs[j].Region
	})
	return pairs
}

// classifyQuotaStatus compares the vCPUs a group needs with the remaining
// family quota. Without a vCPU count or a quota entry the answer is Unknown.
func classifyQuotaStatus(neededVCPUs int, entry quota.UsageEntry, hasEntry bool) string {
	if !hasEntry || neededVCPUs <= 0 {
		return quotaUnknown
	}
	remaining := entry.Available - neededVCPUs
	switch {
	case remaining < 0:
		return quotaInsufficient
	case entry.Limit > 0 && float64(remaining)/float64(entry.Limit) < quotaNearLimitFraction:
		return quotaNearLimit
	default:
		return quotaSufficient
	}
}

func formatSourceSizes(sizes map[string]int) string {
	names := make([]string, 0, len(sizes))
	for name := range sizes {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%s (%d)", name, sizes[name]))
	}
	return strings.Join(parts, ", ")
}

func joinOrNA(values []string) string {
	if len(values) == 0 {
		return "N/A"
	}
	return strings.Join(values, ", ")
}

// buildSKUAvailabilitySheet renders the availability and quota sheet. It
// returns nil when there is nothing to evaluate. Like the notices sheet it has
// no ColumnMetadata, so the main sheet's schema is unaffected.
func buildSKUAvailabilitySheet(groups map[availabilityKey]*availabilityGroup, data map[subscriptionRegion]*regionSKUData, subscriptions map[string]string) *plugins.ExternalPluginOutput {
	if len(groups) == 0 {
		return nil
	}

	keys := make([]availabilityKey, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	subscriptionName := func(id string) string {
		if name := subscriptions[id]; name != "" {
			return name
		}
		return id
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if an, bn := subscriptionName(a.SubscriptionID), subscriptionName(b.SubscriptionID); an != bn {
			return an < bn
		}
		if a.SubscriptionID != b.SubscriptionID {
			return a.SubscriptionID < b.SubscriptionID
		}
		if a.Region != b.Region {
			return a.Region < b.Region
		}
		return a.TargetSKU < b.TargetSKU
	})

	table := [][]string{{
		"Subscription", "Region", "Suggested Target SKU", "Source VM Sizes", "VM Count",
		"Availability", "Quota", "Restriction Reasons", "Restricted Zones",
		"Quota Family", "vCPUs Needed", "Family Quota Available",
	}}
	for _, key := range keys {
		group := groups[key]
		regionData := data[subscriptionRegion{SubscriptionID: key.SubscriptionID, Region: key.Region}]

		availability := availabilityUnknown
		reasons, zones := "N/A", "N/A"
		family, vcpusNeeded, quotaAvailable, quotaStatus := "N/A", "N/A", "N/A", quotaUnknown

		if regionData != nil && regionData.SKUs != nil {
			detail, found := regionData.SKUs[strings.ToLower(key.TargetSKU)]
			if !found {
				availability = availabilityNotOffered
				quotaStatus = quotaNotApplicable
			} else {
				availability = detail.Availability
				reasons = joinOrNA(detail.Reasons)
				zones = joinOrNA(detail.Zones)
				if detail.Family != "" {
					family = detail.Family
				}
				needed := 0
				if detail.VCPUs > 0 {
					needed = group.VMCount * detail.VCPUs
					vcpusNeeded = strconv.Itoa(needed)
				}
				if regionData.Quota != nil && detail.Family != "" {
					entry, hasEntry := regionData.Quota[strings.ToLower(detail.Family)]
					if hasEntry {
						quotaAvailable = strconv.Itoa(entry.Available)
					}
					quotaStatus = classifyQuotaStatus(needed, entry, hasEntry)
				}
			}
		}

		table = append(table, []string{
			subscriptionName(key.SubscriptionID), key.Region, key.TargetSKU,
			formatSourceSizes(group.SourceSizes), strconv.Itoa(group.VMCount),
			availability, quotaStatus, reasons, zones,
			family, vcpusNeeded, quotaAvailable,
		})
	}

	return &plugins.ExternalPluginOutput{
		SheetName: skuAvailabilitySheetName,
		Description: "Regional availability and VM family quota for each suggested target SKU, per subscription and region. " +
			"Available means the SKU is offered with no reported restriction; it does not guarantee capacity. " +
			"Quota is a point-in-time view of the target VM family: a redeploy needs the full vCPU count, " +
			"while an in-place resize needs only the change in the new family",
		Table: table,
	}
}
