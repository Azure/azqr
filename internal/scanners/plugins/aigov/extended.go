// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/metrics"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/modelinfo"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/pricing"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/ptu"
	"github.com/rs/zerolog/log"
)

type priceVal struct {
	price pricing.Resolution
	ok    bool
}

// priceBook resolves prices for one scan. Each region's Retail meters are
// fetched once and every resolved price is memoized.
type priceBook struct {
	httpClient *az.HttpClient
	// A failed fetch caches nil items, which resolve to no Retail match.
	regionItems map[string][]pricing.RetailItem
	resolved    map[priceKey]priceVal
}

type priceKey struct{ model, version, region, mode, band string }

func newPriceBook(httpClient *az.HttpClient) *priceBook {
	return &priceBook{
		httpClient:  httpClient,
		regionItems: map[string][]pricing.RetailItem{},
		resolved:    map[priceKey]priceVal{},
	}
}

func (p *priceBook) lookup(ctx context.Context, region, model, version, mode, band string) priceVal {
	key := priceKey{model, version, region, mode, band}
	if v, ok := p.resolved[key]; ok {
		return v
	}
	items, fetched := p.regionItems[region]
	if !fetched {
		var ok bool
		items, ok = pricing.FetchRetailItems(ctx, p.httpClient, region)
		if !ok {
			log.Debug().Str("region", region).Msg("Retail prices unavailable, costs for this region will not be estimated")
		}
		p.regionItems[region] = items
	}
	price, ok := pricing.ResolveFromRegionItems(items, model, mode, band, version)
	v := priceVal{price, ok}
	p.resolved[key] = v
	return v
}

// ptuHourly resolves the per-PTU hourly Retail rate for a region and mode,
// reusing the region's already-fetched meters.
func (p *priceBook) ptuHourly(ctx context.Context, region, mode string) float64 {
	p.lookup(ctx, region, "", "", mode, metrics.StatusUnknown)
	rate, _ := pricing.ResolvePTUHourlyUSD(p.regionItems[region], mode)
	return rate
}

// skuCapacityText labels the raw SKU capacity with its unit: standard
// deployments are in 1K TPM units, provisioned in PTUs.
func skuCapacityText(info deploymentInfo) string {
	if info.SKUCapacity == "" || info.SKUCapacity == "N/A" {
		return "N/A"
	}
	switch {
	case modelinfo.IsProvisionedMode(info.Mode):
		return info.SKUCapacity + " PTUs"
	case modelinfo.IsBatchMode(info.Mode):
		return info.SKUCapacity
	default:
		return info.SKUCapacity + "K TPM"
	}
}

// deploymentLimitTPM converts standard deployment capacity (1K TPM units)
// to raw tokens per minute; 0 when not applicable or unknown.
func deploymentLimitTPM(info deploymentInfo) float64 {
	if !isTokenBilledMode(info.Mode) {
		return 0
	}
	n, err := strconv.ParseFloat(info.SKUCapacity, 64)
	if err != nil || n <= 0 {
		return 0
	}
	return n * 1000
}

// workloadRollup accumulates cost and evidence across all deployments that
// map to the same workload (see defaultWorkloadID).
type workloadRollup struct {
	Label         string
	Qualified     string
	CostUSD       float64
	ActiveBuckets int
	// PricedTokens and TotalTokens measure how much of the workload's token
	// volume is covered by CostUSD; unpriced volume contributes no cost.
	PricedTokens float64
	TotalTokens  float64
}

// costStatus describes whether a workload's cost covers all of its tokens.
func (w *workloadRollup) costStatus() string {
	switch {
	case w.TotalTokens == 0 || w.PricedTokens >= w.TotalTokens:
		return "Complete"
	case w.PricedTokens == 0:
		return "Not available"
	default:
		return "Partial"
	}
}

func sortedDepModels[V any](m map[metrics.DepModel]V) []metrics.DepModel {
	keys := make([]metrics.DepModel, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Deployment != keys[j].Deployment {
			return keys[i].Deployment < keys[j].Deployment
		}
		return keys[i].Model < keys[j].Model
	})
	return keys
}

func sortedTierContexts(m map[metrics.TierContext]*metrics.TokenTally) []metrics.TierContext {
	keys := make([]metrics.TierContext, 0, len(m))
	for tc := range m {
		keys = append(keys, tc)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Tier != keys[j].Tier {
			return keys[i].Tier < keys[j].Tier
		}
		return keys[i].Context < keys[j].Context
	})
	return keys
}

func joinTierContexts(set map[metrics.TierContext]bool) (tiers, bands string) {
	tierSet, bandSet := map[string]bool{}, map[string]bool{}
	for tc := range set {
		tierSet[tc.Tier] = true
		band := tc.Context
		if band == metrics.StatusUnknown {
			band = "N/A"
		}
		bandSet[band] = true
	}
	return strings.Join(metrics.SortedKeys(tierSet), ", "), strings.Join(metrics.SortedKeys(bandSet), ", ")
}

func maxGrainMinutes(d *metrics.AccountData, k metrics.DepModel) int {
	g := 0
	for _, b := range d.Buckets[k] {
		if b.Minutes > g {
			g = b.Minutes
		}
	}
	return g
}

// collectExtendedSheets builds the "AI Gov Usage", "AI Gov Cost", "AI Gov
// PTU", "AI Gov Workloads" and "AI Gov Insights" sheets from the collected
// Azure Monitor window. It ports the analytical capability of
// TokenLens-for-Azure (https://github.com/zakarel/tokenlens-for-azure)
// (usage aggregation, pricing, PTU suitability and aggregate findings) and
// excludes its reporting, UI, wizard and instrumentation layers.
func (s *AIGovScanner) collectExtendedSheets(ctx context.Context, httpClient *az.HttpClient, collected []*collectedAccount) []plugins.ExternalPluginOutput {
	usageTable := [][]string{usageSheetMetadata().HeaderRow()}
	costTable := [][]string{costSheetMetadata().HeaderRow()}
	ptuTable := [][]string{ptuSheetMetadata().HeaderRow()}
	workloadTable := [][]string{workloadsSheetMetadata().HeaderRow()}
	insightTable := [][]string{insightsSheetMetadata().HeaderRow()}

	workloadRollups := map[string]*workloadRollup{}
	prices := newPriceBook(httpClient)

	for _, acct := range collected {
		if ctx.Err() != nil {
			break
		}
		r, d := acct.Resource, acct.Data
		prefix := []string{acct.SubscriptionName, r.ResourceGroup, r.Name}
		kind := r.Kind
		if kind == "" {
			kind = "Unknown"
		}
		sku := r.SkuName
		if sku == "" {
			sku = "Unknown"
		}

		for _, k := range sortedDepModels(d.Buckets) {
			info := acct.deploymentInfo(k.Deployment)
			rollups := d.HourlyRollups(k)

			for _, h := range metrics.SortedKeys(rollups) {
				u := rollups[h]
				tiers, bands := joinTierContexts(d.HourTiers[k][h])
				latency := "N/A"
				if v := u.AvgLatencyMs(); v >= 0 {
					latency = fmt.Sprintf("%.0f", v)
				}
				usageTable = append(usageTable, append(append([]string{}, prefix...),
					k.Deployment, k.Model, h,
					fmt.Sprintf("%.0f", u.Input), fmt.Sprintf("%.0f", u.Output),
					fmt.Sprintf("%.0f", u.Requests), fmt.Sprintf("%.0f", u.RateLimited), fmt.Sprintf("%.0f", u.Failed),
					latency, fmt.Sprintf("%d", u.ActiveBuckets), tiers, bands,
				))
			}

			var totalInput, totalOutput float64
			for _, b := range d.Buckets[k] {
				totalInput += b.Input
				totalOutput += b.Output
			}

			lookup := func(mode, band string) priceVal {
				return prices.lookup(ctx, r.Location, k.Model, info.ModelVersion, mode, band)
			}
			defaultPrice := lookup(info.Mode, metrics.StatusUnknown)

			var deploymentCost, pricedTokens float64
			tally := d.Tally[k]
			tcKeys := sortedTierContexts(tally)
			if len(tcKeys) == 0 {
				tally = map[metrics.TierContext]*metrics.TokenTally{{Tier: "unknown", Context: "unknown"}: {Input: totalInput, Output: totalOutput}}
				tcKeys = []metrics.TierContext{{Tier: "unknown", Context: "unknown"}}
			}
			for _, tc := range tcKeys {
				t := tally[tc]
				pv := lookup(info.Mode, tc.Context)
				if !isStandardTier(tc.Tier) || !isTokenBilledMode(info.Mode) {
					pv = priceVal{}
				}
				cost, costText := 0.0, "N/A"
				priceSource := "unavailable"
				if pv.ok {
					cost = pricing.EstimatedCostUSD(pv.price, t.Input, t.Output)
					costText = fmt.Sprintf("%.2f", cost)
					priceSource = pv.price.Source
					pricedTokens += t.Input + t.Output
				}
				deploymentCost += cost
				costTable = append(costTable, append(append([]string{}, prefix...),
					k.Deployment, k.Model, info.Mode, tc.Tier, tc.Context,
					fmt.Sprintf("%.0f", t.Input), fmt.Sprintf("%.0f", t.Output),
					priceSource, costText,
					costAssumptions(info.Mode, tc, pv), costRemediation(pv.ok, priceSource, info.Mode, pv.price.Note),
				))
			}

			throttle, throttleKnown := d.WindowThrottleRate(k)
			assessment := ptu.Assess(ptu.Input{
				Model: k.Model, Mode: info.Mode, Buckets: d.Buckets[k],
				GrainMinutes: maxGrainMinutes(d, k),
				ThrottleRate: throttle, ThrottleKnown: throttleKnown, Price: defaultPrice.price,
				DeploymentLimitTPM: deploymentLimitTPM(info),
				PTUHourlyUSD:       prices.ptuHourly(ctx, r.Location, info.Mode),
			})
			throttleText := "N/A"
			if throttleKnown {
				throttleText = fmt.Sprintf("%.2f", throttle)
			}
			evaluated := assessment.ElapsedBuckets > 0
			sized := assessment.RecommendedPTUs > 0
			priced := assessment.PAYGMonthlyUSD > 0
			ptuTable = append(ptuTable, append(append([]string{}, prefix...),
				kind, sku, k.Deployment, k.Model,
				info.ModelVersion, info.ModelFormat, skuCapacityText(info), info.VersionUpgradeOption,
				info.SpilloverEnabled, info.SpilloverDeployment,
				info.SKUName, info.Mode, modelinfo.ProviderFamilyForModel(k.Model), assessment.Recommendation,
				dimensionText(assessment.Shape), dimensionText(assessment.Pressure),
				dimensionText(assessment.Latency), dimensionText(assessment.Predictability),
				naIf(!evaluated, fmt.Sprintf("%.1f", assessment.AvgTPM)), naIf(!evaluated, fmt.Sprintf("%.1f", assessment.P95TPM)),
				naIf(!evaluated, fmt.Sprintf("%d", assessment.ActiveBuckets)), naIf(!evaluated, fmt.Sprintf("%d", assessment.ElapsedBuckets)),
				fmt.Sprintf("%dm", assessment.GrainMinutes), throttleText,
				naIf(!sized, fmt.Sprintf("%d", assessment.RecommendedPTUs)),
				naIf(!sized || !priced, fmt.Sprintf("%.2f", assessment.PAYGMonthlyUSD)), naIf(!sized, fmt.Sprintf("%.2f", assessment.PTUMonthlyUSD)),
				naIf(!sized || !priced, fmt.Sprintf("%.0f", assessment.BreakEvenTPM)),
				ptuRemediation(assessment),
				naIf(!evaluated, fmt.Sprintf("%.1f", assessment.WindowDays)), assessment.Note,
			))

			active := 0
			for _, u := range rollups {
				active += u.ActiveBuckets
			}
			id := strings.ToLower(r.ID) + "|" + strings.ToLower(k.Deployment)
			w := workloadRollups[id]
			if w == nil {
				w = &workloadRollup{
					Label:     defaultWorkloadID(r.Name, k.Deployment),
					Qualified: fmt.Sprintf("%s/%s/%s", acct.SubscriptionName, r.ResourceGroup, defaultWorkloadID(r.Name, k.Deployment)),
				}
				workloadRollups[id] = w
			}
			w.CostUSD += deploymentCost
			w.ActiveBuckets += active
			w.PricedTokens += pricedTokens
			w.TotalTokens += totalInput + totalOutput
		}

		for _, f := range evaluateInsights(totalsFor(d)) {
			insightTable = append(insightTable, append(append([]string{}, prefix...),
				fmt.Sprintf("%d", f.Priority), f.ID, f.Category, f.Title, f.Observation, f.Evidence,
				f.Impact, f.Recommendation, f.Verification, f.Confidence, f.Applicability, f.Missing, f.Caveats,
			))
		}
	}

	labelCount := map[string]int{}
	for _, w := range workloadRollups {
		labelCount[w.Label]++
	}
	workloadKeys := metrics.SortedKeys(workloadRollups)
	display := func(w *workloadRollup) string {
		if labelCount[w.Label] > 1 {
			return w.Qualified
		}
		return w.Label
	}
	sort.SliceStable(workloadKeys, func(i, j int) bool {
		return display(workloadRollups[workloadKeys[i]]) < display(workloadRollups[workloadKeys[j]])
	})
	for _, key := range workloadKeys {
		w := workloadRollups[key]
		id := display(w)
		costText := "N/A"
		if w.PricedTokens > 0 || w.TotalTokens == 0 {
			costText = fmt.Sprintf("%.2f", w.CostUSD)
		}
		pricedShare := 100.0
		if w.TotalTokens > 0 {
			pricedShare = w.PricedTokens / w.TotalTokens * 100
		}
		workloadTable = append(workloadTable, []string{id, costText, fmt.Sprintf("%d", w.ActiveBuckets), fmt.Sprintf("%.0f", pricedShare), w.costStatus()})
	}

	return []plugins.ExternalPluginOutput{
		{Metadata: usageSheetMetadata(), SheetName: "AI Gov Usage", Description: "Hourly (UTC) usage, request outcomes and latency by deployment/model. Failed Requests exclude HTTP 429 (see Rate-Limited); statuses that are neither success, 429 nor failure are not counted in either. Avg Latency is the unweighted mean of per-bucket averages (not request-weighted)", Table: usageTable},
		{Metadata: costSheetMetadata(), SheetName: "AI Gov Cost", Description: "Estimated cost by deployment/model, tier and context band using verified pricing only; list USD retail prices, N/A when no verified price applies", Table: costTable},
		{Metadata: ptuSheetMetadata(), SheetName: "AI Gov PTU", Description: "PTU recommendation with four scored dimensions per deployment. Verdicts: positive = favours PTU, negative = favours pay-as-you-go, neutral = inconclusive, insufficient = metric not available. TPM columns are PTU-weighted (input + output x model ratio); the average spans the first to last active bucket. Capacity Pressure compares raw P95 tokens/min with the deployment limit (SKU Capacity x 1K) when known. N/A means the value was not computed. PTU dollars use Retail hourly meters when found, otherwise labelled fallbacks, and extrapolate the observed span to 30 days. Provisioned, batch and non-Microsoft deployments are not assessed; PTUs are sized on the P95 of the collected-grain averages, so minute-level peaks may need more", Table: ptuTable},
		{Metadata: workloadsSheetMetadata(), SheetName: "AI Gov Workloads", Description: "Estimated cost rolled up per technical workload; Cost Status shows whether unpriced usage is excluded", Table: workloadTable},
		{Metadata: insightsSheetMetadata(), SheetName: "AI Gov Insights", Description: "Evidence-grounded findings from aggregate Azure Monitor data", Table: insightTable},
	}
}

// naIf returns "N/A" for a value that was not computed.
func naIf(notComputed bool, text string) string {
	if notComputed {
		return "N/A"
	}
	return text
}

func dimensionText(d ptu.Dimension) string {
	if d.Metric == "" {
		return d.Verdict
	}
	return fmt.Sprintf("%s (%s)", d.Verdict, d.Metric)
}

// isStandardTier reports whether a service tier is priced by the standard
// meters. Priority, flex and batch tiers have separate meters that are not
// resolved, so their cost is withheld instead of priced at the standard rate.
func isStandardTier(tier string) bool {
	return tier == metrics.StatusUnknown || tier == "standard" || tier == "default" || tier == "auto"
}

// isTokenBilledMode reports whether per-token Retail meters describe how the
// deployment is billed. Provisioned deployments are billed per reserved PTU
// hour and batch deployments use batch meters that are not resolved, so neither
// is priced at the standard token rate.
func isTokenBilledMode(mode string) bool {
	return !modelinfo.IsProvisionedMode(mode) && !modelinfo.IsBatchMode(mode)
}

// costAssumptions states what the estimate assumed about mode, tier and
// context, since Azure Monitor does not expose cached tokens or per-prompt size.
func costAssumptions(mode string, tc metrics.TierContext, pv priceVal) string {
	parts := []string{"list USD retail prices; discounts, reservations and agreements not applied", "cached input priced as normal input"}
	switch {
	case modelinfo.IsProvisionedMode(mode):
		parts = append(parts, "provisioned deployment is billed per reserved PTU hour, not per token; token cost not shown")
	case modelinfo.IsBatchMode(mode):
		parts = append(parts, "batch deployment: batch meters are not resolved; cost withheld")
	}
	if mode == modelinfo.ModeUnknown || mode == "" {
		parts = append(parts, "deployment mode unknown; Global rate assumed")
	}
	if tc.Tier == metrics.StatusUnknown {
		parts = append(parts, "service tier unknown; Standard assumed")
	} else if !isStandardTier(tc.Tier) {
		parts = append(parts, fmt.Sprintf("service tier %q is not priced; cost withheld", tc.Tier))
	}
	switch tc.Context {
	case metrics.StatusUnknown:
		parts = append(parts, "context band unknown; short context assumed")
	case "long":
		if !pv.ok {
			parts = append(parts, "long-context meter not found; cost withheld")
		}
	default:
		parts = append(parts, "monitor buckets are interval totals, not single prompts; short context assumed")
	}
	if pv.ok && pv.price.Tier != "" {
		parts = append(parts, fmt.Sprintf("matched %s/%s meter", pv.price.Tier, pv.price.Context))
	}
	return strings.Join(parts, "; ")
}
