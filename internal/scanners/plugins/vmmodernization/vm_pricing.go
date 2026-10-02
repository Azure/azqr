// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmmodernization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/az"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

const (
	// vmPricingBatchSize is lower than the region plugin's 15-meterId batch
	// size (P02-T01). This is not a URL-length constraint: the Retail Prices
	// API's $filter parser rejects the whole query with 400 "Invalid OData
	// parameters supplied" once a compound `(armSkuName eq '...' and
	// armRegionName eq '...')` OR-group exceeds 8 clauses, confirmed by
	// live testing (8 clauses: 200 OK; 9+: 400, independent of value
	// length). A rejected batch previously surfaced as every VM in it
	// showing Manual Review Required, even for SKUs with valid pricing.
	vmPricingBatchSize = 8
	// vmPricingMaxWorkers is lower than the region plugin's 10-worker pool.
	// Unlike the region plugin's pricing phase (few meterId batches, rarely
	// bursty), this plugin's SKU/region diversity across a whole VM estate
	// can produce many more batches, bursting well past the Retail Prices
	// API's observed per-client request budget (confirmed via the
	// x-ms-ratelimit-remaining-retailprices-requests response header) and
	// tripping its 60-second throttling penalty window. The shared
	// throttling policy deliberately bypasses prices.azure.com (see
	// internal/throttling/policy.go), so this plugin paces its own calls via
	// vmPricingLimiter instead of relying solely on reactive 429 backoff.
	vmPricingMaxWorkers = 5
)

// vmPricingLimiter paces this plugin's own Retail Prices API calls to stay
// under the endpoint's observed request budget (~10 requests/second before
// a 60-second penalty window kicks in). It is deliberately conservative and
// scoped to this plugin only — it does not affect the region plugin's own
// Retail Prices API usage, which bypasses the shared throttling policy.
var vmPricingLimiter = rate.NewLimiter(rate.Limit(5), 8)

// vmRetailPricesURL is the Azure Retail Prices API endpoint this plugin's
// lookup queries. It is a package-level variable (not a constant) so tests
// can redirect it to a local httptest server without a live network call;
// production code never reassigns it.
var vmRetailPricesURL = "https://prices.azure.com/api/retail/prices"

// skuRegionPair is the cache/query key for a VM pricing lookup (plan decision
// D4): OSType is deliberately excluded, since one fetched response already
// contains both OS variants (P02-T02 selects the OS-matched row client-side).
type skuRegionPair struct {
	ArmSkuName    string
	ArmRegionName string
}

// vmRetailPriceItem is a minimal local decode of the Azure Retail Prices API
// response shape, carrying only the fields this plugin's lookup and
// canonical-row selection need — it does not need every field the region
// plugin's types.RetailPriceItem declares.
type vmRetailPriceItem struct {
	ArmSkuName         string  `json:"armSkuName"`
	ArmRegionName      string  `json:"armRegionName"`
	RetailPrice        float64 `json:"retailPrice"`
	UnitOfMeasure      string  `json:"unitOfMeasure"`
	Type               string  `json:"type"`
	MeterName          string  `json:"meterName"`
	ProductName        string  `json:"productName"`
	SkuName            string  `json:"skuName"`
	EffectiveStartDate string  `json:"effectiveStartDate"`
}

// vmRetailPriceResponse is the paginated Azure Retail Prices API response
// envelope this plugin's lookup decodes.
type vmRetailPriceResponse struct {
	Items        []vmRetailPriceItem `json:"Items"`
	NextPageLink string              `json:"NextPageLink"`
}

// vmPriceCache holds the candidate retail price rows fetched for every unique
// (armSkuName, armRegionName) pair collected across one scan (P02-T03). It is
// populated once, in-memory, for the duration of a single Scan invocation —
// it is never persisted across runs.
type vmPriceCache struct {
	rows map[skuRegionPair][]vmRetailPriceItem
}

// hourlyPrice returns the OS-matched, Consumption, Pay-As-You-Go hourly
// retail price for the given SKU/region/OS (P02-T02's canonical-row
// selection applied to this pair's cached candidate rows). ok is false when
// the pair was never fetched (cache miss) or no row survives selection.
func (c *vmPriceCache) hourlyPrice(armSkuName, armRegionName, osType string) (float64, bool) {
	if c == nil {
		return 0, false
	}
	rows, found := c.rows[skuRegionPair{ArmSkuName: armSkuName, ArmRegionName: armRegionName}]
	if !found {
		return 0, false
	}
	return selectCanonicalPrice(rows, osType)
}

// odataEscape escapes a string value for use in an OData $filter expression
// by replacing single quotes with two single quotes, mirroring the region
// plugin's own helper (internal/scanners/plugins/region/cost/cost.go).
func odataEscape(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// collectVMPrices pre-collects and fetches the Retail Prices API rows for
// every unique (armSkuName, armRegionName) pair in one batched, concurrent
// pass (P02-T03/NFR-002), mirroring the region plugin's proven
// getMeterMetadataFromRetailAPI batch-and-worker-pool pattern. On an exempt
// cloud, no HTTP call is made and an empty cache is returned (FR-009).
func collectVMPrices(ctx context.Context, httpClient *az.HttpClient, pairs []skuRegionPair) *vmPriceCache {
	cache := &vmPriceCache{rows: make(map[skuRegionPair][]vmRetailPriceItem, len(pairs))}
	if isExemptCloud() || len(pairs) == 0 || httpClient == nil {
		return cache
	}

	type batchWork struct {
		idx   int
		batch []skuRegionPair
	}
	var batches []batchWork
	for i := 0; i < len(pairs); i += vmPricingBatchSize {
		end := i + vmPricingBatchSize
		if end > len(pairs) {
			end = len(pairs)
		}
		cp := make([]skuRegionPair, end-i)
		copy(cp, pairs[i:end])
		batches = append(batches, batchWork{idx: len(batches), batch: cp})
	}

	results := make([]map[skuRegionPair][]vmRetailPriceItem, len(batches))
	jobs := make(chan batchWork, len(batches))
	for _, b := range batches {
		jobs <- b
	}
	close(jobs)

	workers := vmPricingMaxWorkers
	if len(batches) < workers {
		workers = len(batches)
	}
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range jobs {
				results[work.idx] = fetchVMPriceBatch(ctx, httpClient, work.batch)
			}
		}()
	}
	wg.Wait()

	for _, r := range results {
		for pair, items := range r {
			cache.rows[pair] = items
		}
	}
	return cache
}

// fetchVMPriceBatch issues one Retail Prices API request (paginated via
// NextPageLink) for up to vmPricingBatchSize unique (armSkuName,
// armRegionName) pairs using a compound OR-filtered query (P02-T01), then
// splits the response Items back out per pair by matching armSkuName and
// armRegionName on each row. A non-2xx response or decode failure is treated
// as "no price available" for every pair in this batch rather than aborting
// the scan (FR-007); the caller's cache simply has no entry for those pairs.
func fetchVMPriceBatch(ctx context.Context, httpClient *az.HttpClient, batch []skuRegionPair) map[skuRegionPair][]vmRetailPriceItem {
	out := make(map[skuRegionPair][]vmRetailPriceItem, len(batch))

	clauses := make([]string, len(batch))
	for i, pair := range batch {
		clauses[i] = fmt.Sprintf("(armSkuName eq '%s' and armRegionName eq '%s')",
			odataEscape(pair.ArmSkuName), odataEscape(pair.ArmRegionName))
	}
	filter := fmt.Sprintf("serviceName eq 'Virtual Machines' and type eq 'Consumption' and currencyCode eq 'USD' and (%s)",
		strings.Join(clauses, " or "))
	pageURL := vmRetailPricesURL + "?$filter=" + url.QueryEscape(filter)

	for pageURL != "" {
		if err := vmPricingLimiter.Wait(ctx); err != nil {
			log.Warn().Err(err).Msg("Rate limiter wait canceled before querying Retail Prices API for VM pricing")
			break
		}
		body, err := httpClient.Do(ctx, pageURL)
		if err != nil {
			log.Warn().Err(err).Int("pairs", len(batch)).
				Msg("Failed to query Retail Prices API for VM pricing; affected SKU/region pairs will show Manual Review Required")
			break
		}
		var resp vmRetailPriceResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			log.Warn().Err(err).Int("pairs", len(batch)).
				Msg("Failed to decode Retail Prices API response for VM pricing; affected SKU/region pairs will show Manual Review Required")
			break
		}
		for _, item := range resp.Items {
			for _, pair := range batch {
				if item.ArmSkuName == pair.ArmSkuName && item.ArmRegionName == pair.ArmRegionName {
					out[pair] = append(out[pair], item)
				}
			}
		}
		pageURL = resp.NextPageLink
	}

	return out
}

// selectCanonicalPrice applies P02-T02's confirmed (D2) row-selection rule to
// a shared set of candidate rows for one (armSkuName, armRegionName) pair,
// returning the hourly retail price for the VM's operating system. Selection
// order: (1) discard any row whose meterName/skuName/productName contains
// "Spot" (case-insensitive) — the Retail Prices API's type field alone does
// not distinguish Spot from Pay-As-You-Go pricing; (2) keep only
// type eq 'Consumption' rows (re-verified client-side even though P02-T01's
// query already filters server-side); (3) match the row's product/SKU naming
// convention to the VM's OSType (Linux is the unqualified base case; Windows
// rows carry a "Windows" marker); (4) when more than one row remains,
// tie-break on the latest effectiveStartDate that is not in the future,
// mirroring the region plugin's own tie-break rule
// (getMeterPricingAcrossRegions in region/cost/cost.go).
func selectCanonicalPrice(items []vmRetailPriceItem, osType string) (float64, bool) {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	wantWindows := strings.EqualFold(osType, "Windows")

	var candidates []vmRetailPriceItem
	for _, item := range items {
		if containsFold(item.MeterName, "spot") || containsFold(item.SkuName, "spot") || containsFold(item.ProductName, "spot") {
			continue
		}
		if !strings.EqualFold(item.Type, "Consumption") {
			continue
		}
		isWindowsRow := containsFold(item.ProductName, "windows") || containsFold(item.SkuName, "windows")
		if isWindowsRow != wantWindows {
			continue
		}
		candidates = append(candidates, item)
	}

	var current []vmRetailPriceItem
	for _, c := range candidates {
		if c.EffectiveStartDate <= nowStr {
			current = append(current, c)
		}
	}
	if len(current) == 0 {
		// Every candidate is future-dated or no candidate survived the
		// Spot/Consumption/OS filters; report "no price available" (FR-007).
		return 0, false
	}

	best := current[0]
	for _, c := range current[1:] {
		if c.EffectiveStartDate > best.EffectiveStartDate {
			best = c
		}
	}
	return best.RetailPrice, true
}

// containsFold reports whether substr occurs within s, case-insensitively.
func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
