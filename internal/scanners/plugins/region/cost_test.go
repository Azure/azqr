// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package region

import (
	"testing"
)

func TestMergeCostData_BothNil(t *testing.T) {
	result := mergeCostData(nil, nil)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestMergeCostData_DstNil(t *testing.T) {
	src := &CostComparisonData{
		MeterInputs: []meterCostData{{meterID: "m1", historicalCost: 10}},
	}
	result := mergeCostData(nil, src)
	if result != src {
		t.Fatal("expected src to be returned unchanged when dst is nil")
	}
}

func TestMergeCostData_SrcNil(t *testing.T) {
	dst := &CostComparisonData{
		MeterInputs: []meterCostData{{meterID: "m1", historicalCost: 10}},
	}
	result := mergeCostData(dst, nil)
	if result != dst {
		t.Fatal("expected dst to be returned unchanged when src is nil")
	}
}

func TestMergeCostData_HistoricalCostAccumulated(t *testing.T) {
	dst := &CostComparisonData{
		MeterInputs: []meterCostData{{meterID: "m1", meterName: "meter1", historicalCost: 10.0}},
	}
	src := &CostComparisonData{
		MeterInputs: []meterCostData{{meterID: "m1", meterName: "meter1", historicalCost: 5.0}},
	}
	result := mergeCostData(dst, src)
	if len(result.MeterInputs) != 1 {
		t.Fatalf("expected 1 meter, got %d", len(result.MeterInputs))
	}
	if result.MeterInputs[0].historicalCost != 15.0 {
		t.Errorf("expected historicalCost 15.0, got %.2f", result.MeterInputs[0].historicalCost)
	}
}

func TestMergeCostData_NewMeterAppended(t *testing.T) {
	dst := &CostComparisonData{
		MeterInputs: []meterCostData{{meterID: "m1", historicalCost: 10.0}},
	}
	src := &CostComparisonData{
		MeterInputs: []meterCostData{{meterID: "m2", historicalCost: 7.0}},
	}
	result := mergeCostData(dst, src)
	if len(result.MeterInputs) != 2 {
		t.Fatalf("expected 2 meters, got %d", len(result.MeterInputs))
	}
}

func TestMergeCostData_RegionPricingFirstWins(t *testing.T) {
	dst := &CostComparisonData{
		MeterInputs: []meterCostData{},
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 1.0},
		},
	}
	src := &CostComparisonData{
		MeterInputs: []meterCostData{},
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 99.0}, // should NOT overwrite
		},
	}
	result := mergeCostData(dst, src)
	if result.RegionPricing["m1"]["eastus"] != 1.0 {
		t.Errorf("expected first-wins price 1.0, got %.2f", result.RegionPricing["m1"]["eastus"])
	}
}

func TestMergeCostData_RegionPricingNewRegionAdded(t *testing.T) {
	dst := &CostComparisonData{
		MeterInputs: []meterCostData{},
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 1.0},
		},
	}
	src := &CostComparisonData{
		MeterInputs: []meterCostData{},
		RegionPricing: map[string]map[string]float64{
			"m1": {"westeurope": 2.0}, // new region for m1
			"m2": {"eastus": 3.0},     // new meterID
		},
	}
	result := mergeCostData(dst, src)
	if result.RegionPricing["m1"]["westeurope"] != 2.0 {
		t.Errorf("expected westeurope price 2.0, got %.2f", result.RegionPricing["m1"]["westeurope"])
	}
	if result.RegionPricing["m2"]["eastus"] != 3.0 {
		t.Errorf("expected m2/eastus price 3.0, got %.2f", result.RegionPricing["m2"]["eastus"])
	}
}

func TestMergeCostData_RegionPricingDstNilMap(t *testing.T) {
	dst := &CostComparisonData{
		MeterInputs:   []meterCostData{},
		RegionPricing: nil, // intentionally nil
	}
	src := &CostComparisonData{
		MeterInputs: []meterCostData{},
		RegionPricing: map[string]map[string]float64{
			"m1": {"eastus": 5.0},
		},
	}
	result := mergeCostData(dst, src)
	if result.RegionPricing["m1"]["eastus"] != 5.0 {
		t.Errorf("expected 5.0, got %.2f", result.RegionPricing["m1"]["eastus"])
	}
}

func TestMergeCostData_PriceItemsDeduplicated(t *testing.T) {
	item := retailPriceItem{MeterName: "CPU", ProductID: "p1", SkuName: "D2s", ArmRegionName: "eastus"}
	dst := &CostComparisonData{PriceItems: []retailPriceItem{item}}
	src := &CostComparisonData{PriceItems: []retailPriceItem{item}} // duplicate
	result := mergeCostData(dst, src)
	if len(result.PriceItems) != 1 {
		t.Errorf("expected 1 PriceItem after dedup, got %d", len(result.PriceItems))
	}
}

func TestMergeCostData_PriceItemsNewItemAppended(t *testing.T) {
	item1 := retailPriceItem{MeterName: "CPU", ProductID: "p1", SkuName: "D2s", ArmRegionName: "eastus"}
	item2 := retailPriceItem{MeterName: "CPU", ProductID: "p1", SkuName: "D2s", ArmRegionName: "westeurope"}
	dst := &CostComparisonData{PriceItems: []retailPriceItem{item1}}
	src := &CostComparisonData{PriceItems: []retailPriceItem{item2}}
	result := mergeCostData(dst, src)
	if len(result.PriceItems) != 2 {
		t.Errorf("expected 2 PriceItems, got %d", len(result.PriceItems))
	}
}

func TestMergeCostData_UomErrorsNotMerged(t *testing.T) {
	dstErr := uomError{OrigMeterID: "m-cpu"}
	srcErr := uomError{OrigMeterID: "m-mem"}
	dst := &CostComparisonData{UomErrors: []uomError{dstErr}}
	src := &CostComparisonData{UomErrors: []uomError{srcErr}}
	result := mergeCostData(dst, src)
	if len(result.UomErrors) != 1 {
		t.Errorf("expected UomErrors not to be merged (len 1), got %d", len(result.UomErrors))
	}
	if result.UomErrors[0].OrigMeterID != "m-cpu" {
		t.Errorf("expected dst UomError retained, got %s", result.UomErrors[0].OrigMeterID)
	}
}
