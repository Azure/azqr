// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ptu

import (
	"fmt"
	"math"
	"time"

	"github.com/Azure/azqr/internal/scanners/plugins/aigov/metrics"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/modelinfo"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/pricing"
)

// CapacitySource is where the PTU sizing table comes from. It is generated
// by hack/code/ptu_capacity_gen (make ptu-capacity) into
// zz_generated.ptu_capacity.go. Capacity is matched on the exact canonical
// model key only; an unlisted model is reported as "capacity data required"
// rather than borrowing a related model's numbers.
//
// The recommendation decision shape and thresholds (not the capacity numbers)
// are adapted from the MIT-licensed https://github.com/msftse-org/ptu-advisor
// project, as ported through TokenLens-for-Azure
// (https://github.com/zakarel/tokenlens-for-azure).
const CapacitySource = "Microsoft Learn PTU sizing page (https://learn.microsoft.com/azure/foundry/openai/how-to/provisioned-throughput-sizing)"

// capacity describes the PTU purchasing shape for one canonical model.
type capacity struct {
	// PAYGQuotaTPM is the default pay-as-you-go tokens-per-minute quota, used
	// to judge capacity pressure.
	PAYGQuotaTPM int
	// InputTPMPerPTU is the sustained input tokens-per-minute one PTU buys.
	InputTPMPerPTU int
	// OutputRatio is how many input-token-equivalents one output token costs
	// against the same PTU capacity.
	OutputRatio     int
	GlobalMinPTU    int
	GlobalIncrement int
	RegionalMinPTU  int
	RegionalIncrem  int
}

// sizing is the documented PTU sizing shape for one model, as generated
// by hack/code/ptu_capacity_gen (make ptu-capacity).
type sizing struct {
	InputTPMPerPTU  int
	OutputRatio     int
	GlobalMinPTU    int
	GlobalIncrement int
	RegionalMinPTU  int
	RegionalIncrem  int
}

// paygQuotaTPM holds default pay-as-you-go TPM quotas, which Microsoft's
// sizing page does not publish. They are maintained by hand (originally from
// msftse-org/ptu-advisor); models absent here get 0, which skips the
// capacity-pressure utilization calculation.
var paygQuotaTPM = map[string]int{
	"gpt-5.5": 450_000, "gpt-5.4-mini": 450_000, "gpt-5.4": 450_000,
	"gpt-5.3-codex": 450_000, "gpt-5.2-codex": 450_000, "gpt-5.2": 450_000,
	"gpt-5.1-codex": 450_000, "gpt-5.1": 450_000, "gpt-5-mini": 450_000,
	"gpt-5": 450_000, "gpt-4.1": 720_000, "gpt-4.1-mini": 2_000_000,
	"gpt-4.1-nano": 4_000_000, "gpt-4o": 450_000, "gpt-4o-mini": 2_000_000,
	"o1": 450_000, "o3-mini": 2_000_000, "o4-mini": 2_000_000,
}

// capacityTable merges the generated sizing data with paygQuotaTPM.
var capacityTable = buildPTUCapacityTable()

func buildPTUCapacityTable() map[string]capacity {
	table := make(map[string]capacity, len(generatedPTUCapacity))
	for model, z := range generatedPTUCapacity {
		table[model] = capacity{
			PAYGQuotaTPM:    paygQuotaTPM[model],
			InputTPMPerPTU:  z.InputTPMPerPTU,
			OutputRatio:     z.OutputRatio,
			GlobalMinPTU:    z.GlobalMinPTU,
			GlobalIncrement: z.GlobalIncrement,
			RegionalMinPTU:  z.RegionalMinPTU,
			RegionalIncrem:  z.RegionalIncrem,
		}
	}
	return table
}

func inPTUTable(model string) bool {
	_, ok := capacityTable[modelinfo.CanonicalModelName(model)]
	return ok
}

// PTU recommendations, matching TokenLens's vocabulary.
const (
	RecPTURecommended = "PTU recommended"
	RecBorderline     = "Borderline"
	recPAYG           = "PAYG recommended"
	RecInsufficient   = "Insufficient evidence"
	recNotApplicable  = "PTU not applicable"
	RecCapacityNeeded = "Capacity data required"
)

// Dimension verdicts.
const (
	verdictPositive     = "positive"
	verdictNegative     = "negative"
	verdictNeutral      = "neutral"
	verdictInsufficient = "insufficient"
)

// Per-PTU hourly rates below are fallbacks (TokenLens's directional values)
// used only when Input.PTUHourlyUSD is not resolved from the Retail Prices API;
// Assess notes when a fallback was used.
const (
	ptuHourlyRateRegionalUSD = 2.0
	ptuHourlyRateGlobalUSD   = 1.0
)

// minimumActiveMinutes is the active time (100 five-minute buckets in
// TokenLens) required before distribution-based dimensions are scored.
const minimumActiveMinutes = 500

// minExtrapolationDays is the observed span below which monthly cost
// estimates are flagged as extrapolations of a short window.
const minExtrapolationDays = 7

// Input is everything the verdict engine needs for one deployment/model.
type Input struct {
	Model        string
	Mode         string
	Buckets      map[time.Time]*metrics.Bucket
	GrainMinutes int
	// ThrottleRate is a percentage; ThrottleKnown is false when withheld.
	ThrottleRate  float64
	ThrottleKnown bool
	Price         pricing.Resolution
	// DeploymentLimitTPM is the deployment's own rate limit in raw tokens
	// per minute (SKU capacity x 1K TPM for standard deployments); 0 when
	// unknown, in which case the model's default quota is used if listed.
	DeploymentLimitTPM float64
	// PTUHourlyUSD is the Retail price of one PTU hour for the deployment's
	// tier; 0 when unresolved and the fallback constants apply.
	PTUHourlyUSD float64
}

// Dimension is one scored signal with its supporting metric.
type Dimension struct {
	Verdict string
	Metric  string
}

// Assessment is one row of the "AI Gov PTU" sheet.
type Assessment struct {
	Recommendation                           string
	Shape, Predictability, Pressure, Latency Dimension
	// AvgTPM, P95TPM and BreakEvenTPM are PTU-weighted (input + output x
	// ratio) so they are comparable with each other and with PTU capacity.
	AvgTPM, P95TPM                              float64
	ActiveBuckets, ElapsedBuckets, GrainMinutes int
	// WindowDays spans the first to the last active bucket, idle gaps included.
	WindowDays                                  float64
	RecommendedPTUs                             int
	PAYGMonthlyUSD, PTUMonthlyUSD, BreakEvenTPM float64
	ThrottleRate                                float64
	ThrottleKnown                               bool
	Note                                        string
}

func insufficientDims(a *Assessment, metric string) {
	d := Dimension{Verdict: verdictInsufficient, Metric: metric}
	a.Shape, a.Predictability, a.Pressure, a.Latency = d, d, d, d
}

func minimumBuckets(grainMinutes int) int {
	if grainMinutes <= 0 {
		grainMinutes = 5
	}
	return int(math.Ceil(float64(minimumActiveMinutes) / float64(grainMinutes)))
}

// weightedSeries returns a continuous weighted-TPM series from the first to
// the last observed bucket (idle gaps count as zero), the active-bucket count,
// bucket-average latencies of active buckets, and raw token totals. Tokens are
// weighted as input + output*ratio.
func weightedSeries(buckets map[time.Time]*metrics.Bucket, grainMinutes, outputRatio int) (series []float64, active int, latencies []float64, in, out float64) {
	if len(buckets) == 0 || grainMinutes <= 0 {
		return nil, 0, nil, 0, 0
	}
	if outputRatio <= 0 {
		outputRatio = 1
	}
	step := time.Duration(grainMinutes) * time.Minute
	var first, last time.Time
	for ts := range buckets {
		if first.IsZero() || ts.Before(first) {
			first = ts
		}
		if ts.After(last) {
			last = ts
		}
	}
	n := int(last.Sub(first)/step) + 1
	series = make([]float64, n)
	latSum := make([]float64, n)
	latN := make([]int, n)
	for ts, b := range buckets {
		i := int(ts.Sub(first) / step)
		series[i] += (b.Input + b.Output*float64(outputRatio)) / float64(grainMinutes)
		in += b.Input
		out += b.Output
		latSum[i] += b.LatencySum
		latN[i] += b.LatencyN
	}
	for i, v := range series {
		if v > 0 {
			active++
			if latN[i] > 0 {
				latencies = append(latencies, latSum[i]/float64(latN[i]))
			}
		}
	}
	return series, active, latencies, in, out
}

func meanOf(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	var s float64
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

func stddevOf(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := meanOf(v)
	var s float64
	for _, x := range v {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(v)))
}

// lagCorrelation is the Pearson correlation between the series and itself
// shifted by lag buckets. ok is false when the series is shorter than two lags.
func lagCorrelation(v []float64, lag int) (float64, bool) {
	if lag <= 0 || len(v) < lag*2 {
		return 0, false
	}
	left, right := v[:len(v)-lag], v[lag:]
	lm, rm := meanOf(left), meanOf(right)
	var num, ls, rs float64
	for i := range left {
		num += (left[i] - lm) * (right[i] - rm)
		ls += (left[i] - lm) * (left[i] - lm)
		rs += (right[i] - rm) * (right[i] - rm)
	}
	den := math.Sqrt(ls * rs)
	if den == 0 {
		return 0.5, true
	}
	return math.Max(0, num/den), true
}

func roundUp(needed float64, minPTU, inc int) int {
	n := int(math.Ceil(needed))
	if n <= minPTU {
		return minPTU
	}
	if inc <= 0 {
		return n
	}
	return minPTU + ((n-minPTU+inc-1)/inc)*inc
}

func scoreShape(series []float64, avgTPM float64) Dimension {
	cv := 0.0
	if avgTPM > 0 {
		cv = stddevOf(series) / avgTPM
	}
	metric := fmt.Sprintf("TPM CV %.2f", cv)
	switch {
	case cv < 0.3:
		return Dimension{verdictPositive, metric}
	case cv > 0.7:
		return Dimension{verdictNegative, metric}
	default:
		return Dimension{verdictNeutral, metric}
	}
}

func scorePredictability(series []float64, grainMinutes int) Dimension {
	peak := 0.0
	for _, v := range series {
		peak = math.Max(peak, v)
	}
	sustained := 0.0
	if peak > 0 {
		cnt := 0
		for _, v := range series {
			if v >= peak*0.3 {
				cnt++
			}
		}
		sustained = float64(cnt) / float64(len(series))
	}
	similarity := 0.5
	if d, ok := lagCorrelation(series, 1440/grainMinutes); ok {
		similarity = math.Max(similarity, d)
	}
	if w, ok := lagCorrelation(series, 10080/grainMinutes); ok {
		similarity = math.Max(similarity, w)
	}
	metric := fmt.Sprintf("%.0f%% sustained, %.0f%% similarity", sustained*100, similarity*100)
	switch {
	case sustained >= 0.6 && similarity >= 0.7:
		return Dimension{verdictPositive, metric}
	case sustained < 0.3:
		return Dimension{verdictNegative, metric}
	default:
		return Dimension{verdictNeutral, metric}
	}
}

// scorePressure judges capacity pressure from throttling and from raw P95
// tokens per minute against the deployment's own limit when known, else the
// model's default quota. Quota is shared per subscription, region and model,
// so a per-deployment comparison is an approximation.
func scorePressure(in Input, capacity capacity, rawP95TPM float64) Dimension {
	limit, basis := in.DeploymentLimitTPM, "deployment limit"
	if limit <= 0 && capacity.PAYGQuotaTPM > 0 {
		limit, basis = float64(capacity.PAYGQuotaTPM), "default quota"
	}
	utilization, quotaText := 0.0, "limit unknown"
	if limit > 0 {
		utilization = rawP95TPM / limit
		quotaText = fmt.Sprintf("%.1f%% of %s", utilization*100, basis)
	}
	if !in.ThrottleKnown {
		return Dimension{verdictInsufficient, fmt.Sprintf("outcomes unavailable, %s", quotaText)}
	}
	metric := fmt.Sprintf("%.1f%% throttled, %s", in.ThrottleRate, quotaText)
	switch {
	case in.ThrottleRate > 5 || utilization > 0.85:
		return Dimension{verdictPositive, metric}
	case in.ThrottleRate < 1 && limit > 0 && utilization < 0.5:
		return Dimension{verdictNegative, metric}
	default:
		return Dimension{verdictNeutral, metric}
	}
}

func scoreLatency(latencies []float64) Dimension {
	if len(latencies) == 0 {
		return Dimension{verdictInsufficient, "latency metric not available"}
	}
	p50, p95, p99 := metrics.Percentile(latencies, 50), metrics.Percentile(latencies, 95), metrics.Percentile(latencies, 99)
	ratio := 0.0
	if p50 > 0 {
		ratio = p99 / p50
	}
	metric := fmt.Sprintf("P95 %.0f ms of bucket averages, P99/P50 %.1fx", p95, ratio)
	switch {
	case p95 > 5000 || ratio > 4:
		return Dimension{verdictPositive, metric}
	case p95 < 2000 && ratio < 2:
		return Dimension{verdictNegative, metric}
	default:
		return Dimension{verdictNeutral, metric}
	}
}

// recommend combines the four dimension verdicts into a recommendation and an
// optional explanatory note.
func recommend(dims ...Dimension) (recommendation, note string) {
	insufficient, positive := 0, 0
	for _, d := range dims {
		switch d.Verdict {
		case verdictInsufficient:
			insufficient++
		case verdictPositive:
			positive++
		}
	}
	switch {
	case insufficient >= 2:
		return RecInsufficient, "two or more dimensions lack evidence"
	case positive >= 3:
		return RecPTURecommended, ""
	case positive >= 2:
		return RecBorderline, ""
	default:
		return recPAYG, ""
	}
}

// Assess scores four dimensions (workload shape, capacity pressure,
// latency sensitivity and load predictability) and derives a recommendation
// using TokenLens's thresholds. Time-based thresholds scale with the grain
// actually collected.
func Assess(in Input) Assessment {
	a := Assessment{GrainMinutes: in.GrainMinutes, ThrottleRate: in.ThrottleRate, ThrottleKnown: in.ThrottleKnown}

	switch {
	case modelinfo.IsProvisionedMode(in.Mode):
		a.Recommendation = recNotApplicable
		a.Note = "deployment already uses provisioned throughput"
		insufficientDims(&a, "not evaluated")
		return a
	case modelinfo.IsBatchMode(in.Mode):
		a.Recommendation = recNotApplicable
		a.Note = "batch deployments are not eligible for PTU"
		insufficientDims(&a, "not evaluated")
		return a
	case modelinfo.ProviderFamilyForModel(in.Model) != modelinfo.FamilyAzureOpenAI && !inPTUTable(in.Model):
		a.Recommendation = recNotApplicable
		a.Note = "PTU capacity is only modeled for Azure OpenAI and documented PTU-capable models"
		insufficientDims(&a, "not evaluated")
		return a
	}

	capacity, ok := capacityTable[modelinfo.CanonicalModelName(in.Model)]
	if !ok {
		a.Recommendation = RecCapacityNeeded
		a.Note = fmt.Sprintf("model %q is not in the verified PTU capacity table (%s)", in.Model, CapacitySource)
		insufficientDims(&a, "capacity data required")
		return a
	}

	series, active, latencies, totalIn, totalOut := weightedSeries(in.Buckets, in.GrainMinutes, capacity.OutputRatio)
	a.ActiveBuckets = active
	a.ElapsedBuckets = len(series)
	a.WindowDays = float64(len(series)*in.GrainMinutes) / 1440
	a.AvgTPM = meanOf(series)
	a.P95TPM = metrics.Percentile(series, 95)

	minBuckets := minimumBuckets(in.GrainMinutes)
	if active < minBuckets {
		insufficientDims(&a, fmt.Sprintf("%d active of %d needed buckets", active, minBuckets))
		a.Recommendation = RecInsufficient
		a.Note = fmt.Sprintf("needs activity in at least %d %d-minute buckets", minBuckets, in.GrainMinutes)
		return a
	}

	a.Shape = scoreShape(series, a.AvgTPM)
	a.Predictability = scorePredictability(series, in.GrainMinutes)
	rawSeries, _, _, _, _ := weightedSeries(in.Buckets, in.GrainMinutes, 1)
	a.Pressure = scorePressure(in, capacity, metrics.Percentile(rawSeries, 95))
	a.Latency = scoreLatency(latencies)
	a.Recommendation, a.Note = recommend(a.Shape, a.Pressure, a.Latency, a.Predictability)

	minPTU, inc, rate := capacity.GlobalMinPTU, capacity.GlobalIncrement, ptuHourlyRateGlobalUSD
	if in.Mode == modelinfo.ModeRegional {
		minPTU, inc, rate = capacity.RegionalMinPTU, capacity.RegionalIncrem, ptuHourlyRateRegionalUSD
	}
	if in.PTUHourlyUSD > 0 {
		rate = in.PTUHourlyUSD
	} else {
		a.addNote(fmt.Sprintf("PTU hourly rate not found in Retail Prices; fallback $%.2f/h assumed", rate))
	}
	a.addNote(fmt.Sprintf("PTUs sized on the P95 of %d-minute averages; minute-level peaks may need more", in.GrainMinutes))
	a.RecommendedPTUs = roundUp(a.P95TPM/float64(capacity.InputTPMPerPTU), minPTU, inc)
	a.PTUMonthlyUSD = float64(a.RecommendedPTUs) * rate * 24 * 30
	if in.Price.Source != "" && a.WindowDays > 0 {
		a.PAYGMonthlyUSD = pricing.EstimatedCostUSD(in.Price, totalIn, totalOut) * 30 / a.WindowDays
		ratio := float64(capacity.OutputRatio)
		if ratio <= 0 {
			ratio = 1
		}
		weightedTPM := (totalIn + totalOut*ratio) / (a.WindowDays * 1440)
		if weightedTPM > 0 && a.PAYGMonthlyUSD > 0 {
			ratePerTPMMonth := a.PAYGMonthlyUSD / weightedTPM
			a.BreakEvenTPM = a.PTUMonthlyUSD / ratePerTPMMonth
		}
		if a.WindowDays < minExtrapolationDays {
			a.addNote(fmt.Sprintf("monthly estimates extrapolate %.1f days of observed activity to 30 days", a.WindowDays))
		}
	} else {
		a.addNote("pay-as-you-go pricing is unavailable so no cost comparison could be computed")
	}
	return a
}

// addNote appends a sentence to the assessment note.
func (a *Assessment) addNote(n string) {
	if a.Note == "" {
		a.Note = n
		return
	}
	a.Note += "; " + n
}
