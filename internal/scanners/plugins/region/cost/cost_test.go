// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package cost

import (
	"testing"

	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

func TestApplyCostDiffs_OwnedMetersOnly(t *testing.T) {
	results := []types.RegionComparison{{SourceRegion: "eastus", TargetRegion: "westus"}}
	subMeterCosts := []types.MeterCostData{{MeterID: "m1", HistoricalCost: 10}}
	shared := &types.CostComparisonData{
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 10, "westus": 20},
			"m2": {"eastus": 10, "westus": 1000},
		},
	}

	ApplyCostDiffs(results, subMeterCosts, shared)

	if !results[0].HasCostData {
		t.Fatal("expected HasCostData to be true")
	}
	if got := results[0].AvgCostDifference; got != 100 {
		t.Fatalf("expected AvgCostDifference 100, got %.2f", got)
	}
}

func TestApplyCostDiffs_NonPhysicalSourceSkipped(t *testing.T) {
	results := []types.RegionComparison{{SourceRegion: "global", TargetRegion: "eastus"}}
	subMeterCosts := []types.MeterCostData{{MeterID: "m1", HistoricalCost: 10}}
	shared := &types.CostComparisonData{
		RegionPricing: map[string]map[string]float64{
			"m1": {"global": 10, "eastus": 20},
		},
	}

	ApplyCostDiffs(results, subMeterCosts, shared)

	if results[0].HasCostData {
		t.Fatal("expected HasCostData to be false for non-physical source region")
	}
	if got := results[0].AvgCostDifference; got != 0 {
		t.Fatalf("expected AvgCostDifference 0, got %.2f", got)
	}
}

func TestApplyCostDiffs_NonPhysicalTargetSkipped(t *testing.T) {
	results := []types.RegionComparison{{SourceRegion: "eastus", TargetRegion: "global"}}
	subMeterCosts := []types.MeterCostData{{MeterID: "m1", HistoricalCost: 10}}
	shared := &types.CostComparisonData{
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 10, "global": 20},
		},
	}

	ApplyCostDiffs(results, subMeterCosts, shared)

	if results[0].HasCostData {
		t.Fatal("expected HasCostData to be false for non-physical target region")
	}
	if got := results[0].AvgCostDifference; got != 0 {
		t.Fatalf("expected AvgCostDifference 0, got %.2f", got)
	}
}

func TestApplyCostDiffs_NearZeroBothSkip(t *testing.T) {
	results := []types.RegionComparison{{SourceRegion: "eastus", TargetRegion: "westus"}}
	subMeterCosts := []types.MeterCostData{
		{MeterID: "m1", HistoricalCost: 3},
		{MeterID: "m2", HistoricalCost: 1},
	}
	shared := &types.CostComparisonData{
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 0.00005, "westus": 0.00004},
			"m2": {"eastus": 10, "westus": 20},
		},
	}

	ApplyCostDiffs(results, subMeterCosts, shared)

	if !results[0].HasCostData {
		t.Fatal("expected HasCostData to be true")
	}
	if got := results[0].AvgCostDifference; got != 25 {
		t.Fatalf("expected AvgCostDifference 25, got %.2f", got)
	}
}

func TestApplyCostDiffs_NearZeroSourceSkip(t *testing.T) {
	results := []types.RegionComparison{{SourceRegion: "eastus", TargetRegion: "westus"}}
	subMeterCosts := []types.MeterCostData{
		{MeterID: "m1", HistoricalCost: 3},
		{MeterID: "m2", HistoricalCost: 1},
	}
	shared := &types.CostComparisonData{
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 0.00005, "westus": 1},
			"m2": {"eastus": 10, "westus": 20},
		},
	}

	ApplyCostDiffs(results, subMeterCosts, shared)

	if !results[0].HasCostData {
		t.Fatal("expected HasCostData to be true")
	}
	if got := results[0].AvgCostDifference; got != 100 {
		t.Fatalf("expected AvgCostDifference 100, got %.2f", got)
	}
}

func TestApplyCostDiffs_HasCostDataSet(t *testing.T) {
	results := []types.RegionComparison{{SourceRegion: "eastus", TargetRegion: "westus"}}
	subMeterCosts := []types.MeterCostData{{MeterID: "m1", HistoricalCost: 5}}
	shared := &types.CostComparisonData{
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 10, "westus": 10},
		},
	}

	ApplyCostDiffs(results, subMeterCosts, shared)

	if !results[0].HasCostData {
		t.Fatal("expected HasCostData to be true when at least one meter is compared")
	}
	if got := results[0].AvgCostDifference; got != 0 {
		t.Fatalf("expected AvgCostDifference 0, got %.2f", got)
	}
}

func TestApplyCostDiffs_NoMatchingRegion(t *testing.T) {
	results := []types.RegionComparison{{SourceRegion: "eastus", TargetRegion: "westus"}}
	subMeterCosts := []types.MeterCostData{{MeterID: "m1", HistoricalCost: 5}}
	shared := &types.CostComparisonData{
		RegionPricing: map[string]map[string]float64{
			"m1": {"centralus": 10, "northcentralus": 10},
		},
	}

	ApplyCostDiffs(results, subMeterCosts, shared)

	if results[0].HasCostData {
		t.Fatal("expected HasCostData to be false when no source/target pricing pair exists")
	}
	if got := results[0].AvgCostDifference; got != 0 {
		t.Fatalf("expected AvgCostDifference 0, got %.2f", got)
	}
}

func TestMergeCostData_BothNil(t *testing.T) {
	result := MergeCostData(nil, nil)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestMergeCostData_DstNil(t *testing.T) {
	src := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{{MeterID: "m1", HistoricalCost: 10}},
	}
	result := MergeCostData(nil, src)
	if result != src {
		t.Fatal("expected src to be returned unchanged when dst is nil")
	}
}

func TestMergeCostData_SrcNil(t *testing.T) {
	dst := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{{MeterID: "m1", HistoricalCost: 10}},
	}
	result := MergeCostData(dst, nil)
	if result != dst {
		t.Fatal("expected dst to be returned unchanged when src is nil")
	}
}

func TestMergeCostData_HistoricalCostAccumulated(t *testing.T) {
	dst := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{{MeterID: "m1", MeterName: "meter1", HistoricalCost: 10.0}},
	}
	src := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{{MeterID: "m1", MeterName: "meter1", HistoricalCost: 5.0}},
	}
	result := MergeCostData(dst, src)
	if len(result.MeterInputs) != 1 {
		t.Fatalf("expected 1 meter, got %d", len(result.MeterInputs))
	}
	if result.MeterInputs[0].HistoricalCost != 15.0 {
		t.Errorf("expected HistoricalCost 15.0, got %.2f", result.MeterInputs[0].HistoricalCost)
	}
}

func TestMergeCostData_NewMeterAppended(t *testing.T) {
	dst := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{{MeterID: "m1", HistoricalCost: 10.0}},
	}
	src := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{{MeterID: "m2", HistoricalCost: 7.0}},
	}
	result := MergeCostData(dst, src)
	if len(result.MeterInputs) != 2 {
		t.Fatalf("expected 2 meters, got %d", len(result.MeterInputs))
	}
}

func TestMergeCostData_RegionPricingFirstWins(t *testing.T) {
	dst := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{},
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 1.0},
		},
	}
	src := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{},
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 99.0}, // should NOT overwrite
		},
	}
	result := MergeCostData(dst, src)
	if result.RegionPricing["m1"]["eastus"] != 1.0 {
		t.Errorf("expected first-wins price 1.0, got %.2f", result.RegionPricing["m1"]["eastus"])
	}
}

func TestMergeCostData_RegionPricingNewRegionAdded(t *testing.T) {
	dst := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{},
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 1.0},
		},
	}
	src := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{},
		RegionPricing: map[string]map[string]float64{
			"m1": {"westeurope": 2.0}, // new region for m1
			"m2": {"eastus": 3.0},     // new meterID
		},
	}
	result := MergeCostData(dst, src)
	if result.RegionPricing["m1"]["westeurope"] != 2.0 {
		t.Errorf("expected westeurope price 2.0, got %.2f", result.RegionPricing["m1"]["westeurope"])
	}
	if result.RegionPricing["m2"]["eastus"] != 3.0 {
		t.Errorf("expected m2/eastus price 3.0, got %.2f", result.RegionPricing["m2"]["eastus"])
	}
}

func TestMergeCostData_RegionPricingDstNilMap(t *testing.T) {
	dst := &types.CostComparisonData{
		MeterInputs:   []types.MeterCostData{},
		RegionPricing: nil, // intentionally nil
	}
	src := &types.CostComparisonData{
		MeterInputs: []types.MeterCostData{},
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 5.0},
		},
	}
	result := MergeCostData(dst, src)
	if result.RegionPricing["m1"]["eastus"] != 5.0 {
		t.Errorf("expected 5.0, got %.2f", result.RegionPricing["m1"]["eastus"])
	}
}

func TestMergeCostData_PriceItemsDeduplicated(t *testing.T) {
	item := types.RetailPriceItem{MeterName: "CPU", ProductID: "p1", SkuName: "D2s", ArmRegionName: "eastus"}
	dst := &types.CostComparisonData{PriceItems: []types.RetailPriceItem{item}}
	src := &types.CostComparisonData{PriceItems: []types.RetailPriceItem{item}} // duplicate
	result := MergeCostData(dst, src)
	if len(result.PriceItems) != 1 {
		t.Errorf("expected 1 PriceItem after dedup, got %d", len(result.PriceItems))
	}
}

func TestMergeCostData_PriceItemsNewItemAppended(t *testing.T) {
	item1 := types.RetailPriceItem{MeterName: "CPU", ProductID: "p1", SkuName: "D2s", ArmRegionName: "eastus"}
	item2 := types.RetailPriceItem{MeterName: "CPU", ProductID: "p1", SkuName: "D2s", ArmRegionName: "westeurope"}
	dst := &types.CostComparisonData{PriceItems: []types.RetailPriceItem{item1}}
	src := &types.CostComparisonData{PriceItems: []types.RetailPriceItem{item2}}
	result := MergeCostData(dst, src)
	if len(result.PriceItems) != 2 {
		t.Errorf("expected 2 PriceItems, got %d", len(result.PriceItems))
	}
}

func TestMergeCostData_UomErrorsMerged(t *testing.T) {
	dstErr := types.UoMError{OrigMeterID: "m-cpu"}
	srcErr := types.UoMError{OrigMeterID: "m-mem"}
	dst := &types.CostComparisonData{UomErrors: []types.UoMError{dstErr}}
	src := &types.CostComparisonData{UomErrors: []types.UoMError{srcErr}}
	result := MergeCostData(dst, src)
	if len(result.UomErrors) != 2 {
		t.Errorf("expected UomErrors to be merged (len 2), got %d", len(result.UomErrors))
	}
}

func TestMergeCostData_UomErrorsDeduped(t *testing.T) {
	err := types.UoMError{OrigMeterID: "m-cpu"}
	dst := &types.CostComparisonData{UomErrors: []types.UoMError{err}}
	src := &types.CostComparisonData{UomErrors: []types.UoMError{err}}
	result := MergeCostData(dst, src)
	if len(result.UomErrors) != 1 {
		t.Errorf("expected deduped UomErrors (len 1), got %d", len(result.UomErrors))
	}
}
