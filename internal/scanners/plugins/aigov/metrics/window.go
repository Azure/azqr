// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/query/azmetrics"
	"github.com/rs/zerolog/log"
)

const (
	metricsNamespace = "Microsoft.CognitiveServices/accounts"

	// MaxWindowDays is the maximum observation window. Azure Monitor retains
	// platform metrics for about 93 days; the window stays below that so every
	// chunk is inside retention. Chunks that are rejected are reported as
	// coverage gaps rather than silently dropped.
	MaxWindowDays = 90
	chunkDays     = 30
)

// Grain is one Azure Monitor time Grain the collector can query at.
type Grain struct {
	ISO     string
	Minutes int
}

// grainLadder is tried in order for every chunk: TokenLens collects at 5
// minutes, and coarser grains are used only when a chunk cannot be served at
// the finer one.
var grainLadder = []Grain{{"PT5M", 5}, {"PT15M", 15}, {"PT1H", 60}}

// metricsQuerier is the subset of azmetrics.Client the collector uses.
type metricsQuerier interface {
	QueryResources(ctx context.Context, subscriptionID string, metricNamespace string, metricNames []string, resourceIDs azmetrics.ResourceIDList, options *azmetrics.QueryResourcesOptions) (azmetrics.QueryResourcesResponse, error)
}

type DepModel struct {
	Deployment string
	Model      string
}

type TierContext struct {
	Tier    string
	Context string
}

// Bucket is the evidence observed for one deployment/model in one time Bucket.
type Bucket struct {
	Input, Output                      float64
	Requests                           float64
	Success, RateLimited, Failed, Rest float64
	LatencySum                         float64
	LatencyN                           int
	Minutes                            int
}

type TokenTally struct {
	Input, Output float64
}

// AccountData is everything collected for one account over the window.
type AccountData struct {
	ResourceID string
	Plan       Plan
	Buckets    map[DepModel]map[time.Time]*Bucket
	// Tally: deployment/model -> tier/context -> tokens across the window.
	Tally map[DepModel]map[TierContext]*TokenTally
	// Tiers and Bands observed per deployment/model/hour for the usage sheet.
	HourTiers map[DepModel]map[string]map[TierContext]bool

	// latencyMetric is the first latency metric that returned data; points
	// from lower-priority metrics are ignored so sources are never mixed.
	latencyMetric string
	Gaps          []string
	Grain         string // coarsest grain that was needed
	grainMins     int
	Elapsed       int
	observed      map[int64]struct{}
	// RequestsComplete is false when the request denominator is missing or a
	// request chunk failed, so a window-level throttle rate must be withheld.
	RequestsComplete bool
	WindowStart      time.Time
	WindowEnd        time.Time
}

func newAccountData(resourceID string, plan Plan, start, end time.Time) *AccountData {
	return &AccountData{
		ResourceID:       resourceID,
		Plan:             plan,
		Buckets:          map[DepModel]map[time.Time]*Bucket{},
		Tally:            map[DepModel]map[TierContext]*TokenTally{},
		HourTiers:        map[DepModel]map[string]map[TierContext]bool{},
		observed:         map[int64]struct{}{},
		RequestsComplete: plan.Requests != "",
		WindowStart:      start,
		WindowEnd:        end,
	}
}

// Observed returns the number of distinct buckets that returned any data.
func (a *AccountData) Observed() int { return len(a.observed) }

// CoveragePercent is observed buckets over elapsed buckets, capped at 100.
func (a *AccountData) CoveragePercent() float64 {
	if a.Elapsed == 0 {
		return 0
	}
	pct := float64(a.Observed()) / float64(a.Elapsed) * 100
	if pct > 100 {
		return 100
	}
	return pct
}

func (a *AccountData) bucketAt(k DepModel, ts time.Time, minutes int) *Bucket {
	m := a.Buckets[k]
	if m == nil {
		m = map[time.Time]*Bucket{}
		a.Buckets[k] = m
	}
	b := m[ts]
	if b == nil {
		b = &Bucket{Minutes: minutes}
		m[ts] = b
	}
	a.observed[ts.Unix()] = struct{}{}
	return b
}

func hourKey(ts time.Time) string { return ts.UTC().Format("2006-01-02 15:00") }

// querySpec is one Azure Monitor query: metrics sharing an aggregation and filter.
type querySpec struct {
	Kind        string // tokens, requests, throttled, latency
	Names       []string
	Aggregation string
	Filter      string
	Required    bool
}

func (p Plan) specs() []querySpec {
	base := dimDeployment + " eq '*' and " + dimModel + " eq '*'"
	var specs []querySpec
	var tokenNames []string
	for _, n := range []string{p.Input, p.Output} {
		if n != "" {
			tokenNames = append(tokenNames, n)
		}
	}
	if len(tokenNames) > 0 {
		f := base
		if p.TierDim {
			f += " and " + dimTier + " eq '*'"
		}
		if p.ContextDim {
			f += " and " + dimContext + " eq '*'"
		}
		specs = append(specs, querySpec{Kind: "tokens", Names: tokenNames, Aggregation: "Total", Filter: f, Required: true})
	}
	if p.Requests != "" {
		f := base
		if p.StatusSplit {
			f = dimStatusCode + " eq '*' and " + base
		}
		specs = append(specs, querySpec{Kind: "requests", Names: []string{p.Requests}, Aggregation: "Total", Filter: f, Required: true})
	}
	if p.Throttled != "" {
		specs = append(specs, querySpec{Kind: "throttled", Names: []string{p.Throttled}, Aggregation: "Total", Filter: base})
	}
	if p.Latency != "" {
		for _, name := range append([]string{p.Latency}, p.LatencyFallbacks...) {
			specs = append(specs, querySpec{Kind: "latency", Names: []string{name}, Aggregation: "Average", Filter: base})
		}
	}
	return specs
}

// chunk is one non-overlapping, half-open [Start, End) slice of the window.
type chunk struct{ Start, End time.Time }

// windowChunks splits the window ending at end into chunkDays slices, newest
// last. Slices never overlap, so merged points cannot be double counted.
func windowChunks(end time.Time, days int) []chunk {
	start := end.AddDate(0, 0, -days)
	var out []chunk
	for s := start; s.Before(end); {
		e := s.AddDate(0, 0, chunkDays)
		if e.After(end) {
			e = end
		}
		out = append(out, chunk{s, e})
		s = e
	}
	return out
}

// CollectBatch collects the full window for a set of accounts that share the
// same metric plan. Chunks that fail at every grain are recorded as gaps.
func CollectBatch(ctx context.Context, q metricsQuerier, subscriptionID string, resourceIDs []string, plan Plan, end time.Time, days int) map[string]*AccountData {
	start := end.AddDate(0, 0, -days)
	result := make(map[string]*AccountData, len(resourceIDs))
	for _, id := range resourceIDs {
		d := newAccountData(id, plan, start, end)
		d.Gaps = append(d.Gaps, plan.Notes...)
		for _, n := range plan.Notes {
			log.Debug().Str("resourceID", id).Msg("metric plan note: " + n)
		}
		result[id] = d
	}
	specs := plan.specs()
	if len(specs) == 0 {
		for _, d := range result {
			d.Gaps = append(d.Gaps, "no supported metrics were available for this account")
		}
		return result
	}

	for _, c := range windowChunks(end, days) {
		if ctx.Err() != nil {
			return result
		}
		chunkMinutes := int(c.End.Sub(c.Start).Minutes())
		var lastErr error
		done := false
		for _, g := range grainLadder {
			attempt := make(map[string]*AccountData, len(resourceIDs))
			for _, id := range resourceIDs {
				attempt[id] = newAccountData(id, plan, start, end)
			}
			var optionalGaps []string
			failed := false
			for _, spec := range specs {
				err := runSpec(ctx, q, subscriptionID, resourceIDs, spec, g, c, plan, attempt)
				if err == nil {
					continue
				}
				if spec.Required {
					lastErr = err
					failed = true
					break
				}
				log.Debug().Err(err).Str("kind", spec.Kind).Strs("metrics", spec.Names).Str("grain", g.ISO).Str("filter", spec.Filter).Msg("optional metric query failed")
				optionalGaps = append(optionalGaps, fmt.Sprintf("%s metrics unavailable for %s to %s: %v", spec.Kind, c.Start.Format("2006-01-02"), c.End.Format("2006-01-02"), err))
			}
			if failed {
				log.Debug().Err(lastErr).Str("grain", g.ISO).Msg("chunk query failed; trying coarser grain")
				continue
			}
			for id, part := range attempt {
				d := result[id]
				d.merge(part, g, chunkMinutes)
				d.Gaps = append(d.Gaps, optionalGaps...)
			}
			done = true
			break
		}
		if !done {
			for _, d := range result {
				d.Elapsed += chunkMinutes / grainLadder[0].Minutes
				d.Gaps = append(d.Gaps, fmt.Sprintf("no data for %s to %s: %v", c.Start.Format("2006-01-02"), c.End.Format("2006-01-02"), lastErr))
				d.RequestsComplete = false
			}
		}
	}
	return result
}

// merge folds one successfully collected chunk into the account totals.
func (a *AccountData) merge(part *AccountData, g Grain, chunkMinutes int) {
	a.Elapsed += chunkMinutes / g.Minutes
	if g.Minutes > a.grainMins {
		a.grainMins = g.Minutes
		a.Grain = g.ISO
	}
	for k, buckets := range part.Buckets {
		if a.Buckets[k] == nil {
			a.Buckets[k] = map[time.Time]*Bucket{}
		}
		for ts, b := range buckets {
			a.Buckets[k][ts] = b
		}
	}
	for k, tcs := range part.Tally {
		for tc, t := range tcs {
			cur := a.tallyFor(k, tc)
			cur.Input += t.Input
			cur.Output += t.Output
		}
	}
	for k, hours := range part.HourTiers {
		for h, tcs := range hours {
			for tc := range tcs {
				a.markHourTier(k, h, tc)
			}
		}
	}
	for ts := range part.observed {
		a.observed[ts] = struct{}{}
	}
}

// runSpec executes one query for the chunk and folds the points into attempt.
func runSpec(ctx context.Context, q metricsQuerier, subscriptionID string, resourceIDs []string, spec querySpec, g Grain, c chunk, plan Plan, attempt map[string]*AccountData) error {
	opts := &azmetrics.QueryResourcesOptions{
		StartTime:   to.Ptr(c.Start.Format(time.RFC3339)),
		EndTime:     to.Ptr(c.End.Format(time.RFC3339)),
		Interval:    to.Ptr(g.ISO),
		Aggregation: to.Ptr(spec.Aggregation),
		Filter:      to.Ptr(spec.Filter),
	}
	resp, err := q.QueryResources(ctx, subscriptionID, metricsNamespace, spec.Names, azmetrics.ResourceIDList{ResourceIDs: resourceIDs}, opts)
	if err != nil {
		return err
	}
	for _, rm := range resp.Values {
		if rm.ResourceID == nil {
			continue
		}
		acct := attempt[*rm.ResourceID]
		if acct == nil {
			continue
		}
		for _, metric := range rm.Values {
			metricName := ""
			if metric.Name != nil && metric.Name.Value != nil {
				metricName = *metric.Name.Value
			}
			for _, ts := range metric.TimeSeries {
				dims := map[string]string{}
				for _, mv := range ts.MetadataValues {
					if mv.Name != nil && mv.Name.Value != nil && mv.Value != nil {
						dims[canonicalDimension(*mv.Name.Value)] = *mv.Value
					}
				}
				for _, pt := range ts.Data {
					if pt.TimeStamp == nil || pt.TimeStamp.Before(c.Start) || !pt.TimeStamp.Before(c.End) {
						continue
					}
					applyPoint(acct, spec.Kind, metricName, plan, dims, pt, g)
				}
			}
		}
	}
	return nil
}

func dimOr(dims map[string]string, key string) string {
	if v := strings.TrimSpace(dims[key]); v != "" {
		return v
	}
	return "Unknown"
}

// applyPoint adds one data point to the account's buckets and tallies.
func applyPoint(a *AccountData, kind, metricName string, plan Plan, dims map[string]string, pt azmetrics.MetricValue, g Grain) {
	key := DepModel{Deployment: dimOr(dims, dimDeployment), Model: dimOr(dims, dimModel)}
	ts := pt.TimeStamp.UTC()

	switch kind {
	case "tokens":
		if pt.Total == nil {
			return
		}
		b := a.bucketAt(key, ts, g.Minutes)
		tc := TierContext{Tier: normalizeTier(dims[dimTier]), Context: contextBand(dims[dimContext])}
		t := a.tallyFor(key, tc)
		if strings.EqualFold(metricName, plan.Input) {
			b.Input += *pt.Total
			t.Input += *pt.Total
		} else {
			b.Output += *pt.Total
			t.Output += *pt.Total
		}
		a.markHourTier(key, hourKey(ts), tc)
	case "requests":
		if pt.Total == nil {
			return
		}
		b := a.bucketAt(key, ts, g.Minutes)
		b.Requests += *pt.Total
		if plan.StatusSplit {
			switch ClassifyStatus(dims[dimStatusCode]) {
			case statusSuccess:
				b.Success += *pt.Total
			case statusRateLimited:
				b.RateLimited += *pt.Total
			case statusFailed:
				b.Failed += *pt.Total
			default:
				b.Rest += *pt.Total
			}
		}
	case "throttled":
		if pt.Total == nil || *pt.Total == 0 {
			return
		}
		b := a.bucketAt(key, ts, g.Minutes)
		b.RateLimited += *pt.Total
	case "latency":
		if pt.Average == nil {
			return
		}
		if a.latencyMetric == "" {
			a.latencyMetric = metricName
		} else if !strings.EqualFold(a.latencyMetric, metricName) {
			return
		}
		b := a.bucketAt(key, ts, g.Minutes)
		b.LatencySum += *pt.Average
		b.LatencyN++
	}
}

// tallyFor returns the token tally for a deployment/model and tier/context,
// creating it on first use.
func (a *AccountData) tallyFor(k DepModel, tc TierContext) *TokenTally {
	if a.Tally[k] == nil {
		a.Tally[k] = map[TierContext]*TokenTally{}
	}
	t := a.Tally[k][tc]
	if t == nil {
		t = &TokenTally{}
		a.Tally[k][tc] = t
	}
	return t
}

// markHourTier records that a tier/context was observed in the given hour.
func (a *AccountData) markHourTier(k DepModel, hour string, tc TierContext) {
	if a.HourTiers[k] == nil {
		a.HourTiers[k] = map[string]map[TierContext]bool{}
	}
	if a.HourTiers[k][hour] == nil {
		a.HourTiers[k][hour] = map[TierContext]bool{}
	}
	a.HourTiers[k][hour][tc] = true
}

// hourRollup is one deployment/model hour aggregated from its buckets.
type hourRollup struct {
	Input, Output, Requests, RateLimited, Failed float64
	LatencySum                                   float64
	LatencyN                                     int
	ActiveBuckets                                int
}

// AvgLatencyMs is the mean of bucket-average latencies, or -1 when absent.
func (h hourRollup) AvgLatencyMs() float64 {
	if h.LatencyN == 0 {
		return -1
	}
	return h.LatencySum / float64(h.LatencyN)
}

// HourlyRollups aggregates buckets to hours.
func (a *AccountData) HourlyRollups(k DepModel) map[string]*hourRollup {
	out := map[string]*hourRollup{}
	for ts, b := range a.Buckets[k] {
		h := hourKey(ts)
		r := out[h]
		if r == nil {
			r = &hourRollup{}
			out[h] = r
		}
		r.Input += b.Input
		r.Output += b.Output
		r.Requests += b.Requests
		r.RateLimited += b.RateLimited
		r.Failed += b.Failed
		r.LatencySum += b.LatencySum
		r.LatencyN += b.LatencyN
		if b.Input+b.Output > 0 {
			r.ActiveBuckets++
		}
	}
	return out
}

// WindowThrottleRate is rate-limited requests over total requests across the
// window for one deployment/model. ok is false (withheld) when the request
// denominator is incomplete or zero, matching TokenLens.
func (a *AccountData) WindowThrottleRate(k DepModel) (rate float64, ok bool) {
	if !a.RequestsComplete {
		return 0, false
	}
	var req, limited float64
	for _, b := range a.Buckets[k] {
		req += b.Requests
		limited += b.RateLimited
	}
	if req <= 0 {
		return 0, false
	}
	return limited / req * 100, true
}
