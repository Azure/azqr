// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pricing

import (
	"context"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/internal/testutil"
	"net/http"
	"testing"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/metrics"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/modelinfo"
)

func TestResolveAzureOpenAIPrice_ExactMeterMatch(t *testing.T) {
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.03, "unitOfMeasure": "1M", "meterName": "gpt-4o Inp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 0.06, "unitOfMeasure": "1M", "meterName": "gpt-4o Outp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 999, "unitOfMeasure": "1M", "meterName": "gpt-35-turbo Inp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	price, ok := resolveAzureOpenAIPrice(context.Background(), client, "gpt-4o", "eastus")
	if !ok {
		t.Fatal("expected resolveAzureOpenAIPrice to find an exact meter match")
	}
	if price.Source != "exact_retail" {
		t.Fatalf("expected Source %q, got %q", "exact_retail", price.Source)
	}
	if price.InputPerMTokUSD != 0.03 {
		t.Fatalf("expected InputPerMTokUSD 0.03, got %v", price.InputPerMTokUSD)
	}
	if price.OutputPerMTokUSD != 0.06 {
		t.Fatalf("expected OutputPerMTokUSD 0.06, got %v", price.OutputPerMTokUSD)
	}
}

func TestResolveAzureOpenAIPrice_NoMatchIsUnresolved(t *testing.T) {
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Items": [], "NextPageLink": null}`))
	})

	_, ok := resolveAzureOpenAIPrice(context.Background(), client, "some-unlisted-model", "eastus")
	if ok {
		t.Fatal("expected resolveAzureOpenAIPrice to report no match when the API returns no items")
	}
}

func TestResolveAzureOpenAIPrice_PartialMeterPairIsUnresolved(t *testing.T) {
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Only the input meter is present; output is missing, so the pair is incomplete.
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.03, "unitOfMeasure": "1M", "meterName": "gpt-4o Inp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	_, ok := resolveAzureOpenAIPrice(context.Background(), client, "gpt-4o", "eastus")
	if ok {
		t.Fatal("expected resolveAzureOpenAIPrice to report no match with only one side of the meter pair")
	}
}

func TestResolveAzureOpenAIPrice_MessyRealWorldMeterNames(t *testing.T) {
	// Real Azure Retail Prices meter names for "Foundry Models" are
	// inconsistently formatted and encode several billing axes at once.
	// This fixture mirrors meters observed live for gpt-6-sol: dropped
	// "gpt" prefix, short/long context, cached vs. standard, priority
	// processing vs. standard, and global vs. data-zone tier.
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.4,  "unitOfMeasure": "1M", "meterName": "6-sol ShortCo Cd Inp PP Gl 1M Tokens",  "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 2.0,  "unitOfMeasure": "1M", "meterName": "6-sol ShortCo Inp Std Gl 1M Tokens",   "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 2.2,  "unitOfMeasure": "1M", "meterName": "6-sol ShortCo Inp Std DZ 1M Tokens",   "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 4.4,  "unitOfMeasure": "1M", "meterName": "6-sol ShortCo Inp PP Gl 1M Tokens",    "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 10.0, "unitOfMeasure": "1M", "meterName": "6-sol ShortCo Opt Std Gl 1M Tokens",   "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 4.0,  "unitOfMeasure": "1M", "meterName": "6-sol LongCo Inp Std Gl 1M Tokens",    "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 15.0, "unitOfMeasure": "1M", "meterName": "6-sol LongCo Opt Std Gl 1M Tokens",    "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	price, ok := resolveAzureOpenAIPrice(context.Background(), client, "gpt-6-sol", "eastus")
	if !ok {
		t.Fatal("expected resolveAzureOpenAIPrice to resolve gpt-6-sol despite the messy meter names")
	}
	if price.Source != "exact_retail" {
		t.Fatalf("expected Source %q, got %q", "exact_retail", price.Source)
	}
	// Should prefer the Global, Standard (non-cached, non-priority-processing),
	// short-context pair: 2.0 input / 10.0 output.
	if price.InputPerMTokUSD != 2.0 {
		t.Fatalf("expected InputPerMTokUSD 2.0 (Std/Gl/ShortCo, not cached/PP/DZ), got %v", price.InputPerMTokUSD)
	}
	if price.OutputPerMTokUSD != 10.0 {
		t.Fatalf("expected OutputPerMTokUSD 10.0, got %v", price.OutputPerMTokUSD)
	}
}

func TestResolveAzureOpenAIPrice_EmbeddingModelHasNoOutputMeter(t *testing.T) {
	// Embedding models only ever bill on input tokens; there is no
	// "Outp"/"Opt" meter to find, so the output rate should resolve to
	// zero rather than being reported unavailable.
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.00013,  "unitOfMeasure": "1K", "meterName": "text-embedding-3-large-glbl Tokens",   "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 0.000143, "unitOfMeasure": "1K", "meterName": "text embedding 3 large DZ Tokens",     "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 0.00013,  "unitOfMeasure": "1K", "meterName": "text-embedding-3-large-grader Tokens", "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	price, ok := resolveAzureOpenAIPrice(context.Background(), client, "text-embedding-3-large", "eastus")
	if !ok {
		t.Fatal("expected resolveAzureOpenAIPrice to resolve the embedding model from its single input meter")
	}
	if price.OutputPerMTokUSD != 0 {
		t.Fatalf("expected OutputPerMTokUSD 0 for an embedding model, got %v", price.OutputPerMTokUSD)
	}
	// unitOfMeasure "1K": retailPrice 0.00013/1K tokens -> 0.13/1M tokens; prefers Global tier.
	if got, want := price.InputPerMTokUSD, 0.13; got < want-1e-9 || got > want+1e-9 {
		t.Fatalf("expected InputPerMTokUSD %.5f (normalized from 1K to 1M and preferring Global tier), got %.5f", want, got)
	}
}

func TestResolveAzureOpenAIPrice_ExcludesBatchAndPTUMeters(t *testing.T) {
	// Only batch/provisioned meters exist for the model; since standard
	// pay-as-you-go pricing can't be found, the result must stay
	// unresolved rather than silently reporting a batch/PTU rate.
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.015, "unitOfMeasure": "1M", "meterName": "gpt-4-8K-Batch-Inp-glbl Tokens",  "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 0.03,  "unitOfMeasure": "1M", "meterName": "gpt-4-8K-Batch-Outp-glbl Tokens", "serviceName": "Foundry Models", "armRegionName": "eastus"},
				{"retailPrice": 5.0,   "unitOfMeasure": "1M", "meterName": "gpt-4-8K PTU Inp Gl Tokens",      "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	_, ok := resolveAzureOpenAIPrice(context.Background(), client, "gpt-4-8K", "eastus")
	if ok {
		t.Fatal("expected batch/PTU-only meters to leave the price unresolved")
	}
}

func TestResolveAzureOpenAIPrice_FollowsPagination(t *testing.T) {
	mux := http.NewServeMux()
	var page2URL string
	mux.HandleFunc("/api/retail/prices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.03, "unitOfMeasure": "1M", "meterName": "gpt-4o Inp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": "` + page2URL + `"
		}`))
	})
	mux.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Items": [
				{"retailPrice": 0.06, "unitOfMeasure": "1M", "meterName": "gpt-4o Outp Global Tokens", "productName": "Azure OpenAI", "serviceName": "Foundry Models", "armRegionName": "eastus"}
			],
			"NextPageLink": null
		}`))
	})

	// The test client rewrites the host, so only the path needs to route to the mux.
	page2URL = "https://prices.azure.com/page2"
	client := testutil.NewPricingClient(t, mux.ServeHTTP)

	price, ok := resolveAzureOpenAIPrice(context.Background(), client, "gpt-4o", "eastus")
	if !ok {
		t.Fatal("expected resolveAzureOpenAIPrice to combine meters found across pages")
	}
	if price.InputPerMTokUSD != 0.03 || price.OutputPerMTokUSD != 0.06 {
		t.Fatalf("expected combined price 0.03/0.06, got %v/%v", price.InputPerMTokUSD, price.OutputPerMTokUSD)
	}
}

const modeAwareMeters = `{
	"Items": [
		{"retailPrice": 1, "unitOfMeasure": "1M", "meterName": "gpt-4.1 Inp Global Tokens", "serviceName": "Foundry Models", "armRegionName": "eastus"},
		{"retailPrice": 2, "unitOfMeasure": "1M", "meterName": "gpt-4.1 Outp Global Tokens", "serviceName": "Foundry Models", "armRegionName": "eastus"},
		{"retailPrice": 1.1, "unitOfMeasure": "1M", "meterName": "gpt-4.1 Inp DZone Tokens", "serviceName": "Foundry Models", "armRegionName": "eastus"},
		{"retailPrice": 2.2, "unitOfMeasure": "1M", "meterName": "gpt-4.1 Outp DZone Tokens", "serviceName": "Foundry Models", "armRegionName": "eastus"},
		{"retailPrice": 4, "unitOfMeasure": "1M", "meterName": "gpt-4.1 Inp Global LongCo Tokens", "serviceName": "Foundry Models", "armRegionName": "eastus"},
		{"retailPrice": 8, "unitOfMeasure": "1M", "meterName": "gpt-4.1 Outp Global LongCo Tokens", "serviceName": "Foundry Models", "armRegionName": "eastus"}
	],
	"NextPageLink": null
}`

func modeAwareClient(t *testing.T, body string) *az.HttpClient {
	return testutil.NewPricingClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})
}

func TestResolveAzureOpenAIPriceFor_ModeAndContext(t *testing.T) {
	client := modeAwareClient(t, modeAwareMeters)
	cases := []struct {
		name, mode, band string
		wantOK           bool
		wantInput        float64
	}{
		{"global short", modelinfo.ModeGlobal, "short", true, 1},
		{"unknown mode defaults to global", modelinfo.ModeUnknown, metrics.StatusUnknown, true, 1},
		{"data zone uses data zone meter", modelinfo.ModeDataZone, "short", true, 1.1},
		{"global long uses long meter", modelinfo.ModeGlobal, "long", true, 4},
		{"data zone long has no meter so is withheld", modelinfo.ModeDataZone, "long", false, 0},
		{"regional has no meter so is withheld", modelinfo.ModeRegional, "short", false, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			price, ok := resolveAzureOpenAIPriceFor(context.Background(), client, "gpt-4.1", "eastus", c.mode, c.band)
			if ok != c.wantOK {
				t.Fatalf("ok=%v want %v", ok, c.wantOK)
			}
			if ok && price.InputPerMTokUSD != c.wantInput {
				t.Errorf("input=%v want %v", price.InputPerMTokUSD, c.wantInput)
			}
		})
	}
}

// realEastUS2Meters are Foundry Models meter names and prices copied from the
// live Azure Retail Prices API (eastus2, Consumption). They include the
// look-alike families that used to be matched by mistake.
func realEastUS2Meters() []RetailItem {
	m := func(name string, price float64, unit string) RetailItem {
		return RetailItem{MeterName: name, RetailPrice: price, UnitOfMeasure: unit}
	}
	return []RetailItem{
		m("gpt 4.1 Inp glbl Tokens", 0.002, "1K"),
		m("gpt 4.1 Outp glbl Tokens", 0.008, "1K"),
		m("gpt 4.1 mini Inp glbl Tokens", 0.0004, "1K"),
		m("gpt 4.1 mini Outp glbl Tokens", 0.0016, "1K"),
		m("gpt 4.1 nano Inp glbl Tokens", 0.0001, "1K"),
		m("gpt 4.1 nano Outp glbl Tokens", 0.0004, "1K"),
		m("Grok 4.1 Inp Glbl Tokens", 0.0002, "1K"),
		m("Grok 4.1 Outp Glbl Tokens", 0.0005, "1K"),
		m("GPT 5 Inpt Glbl 1M Tokens", 1.25, "1M"),
		m("GPT 5 outpt Glbl 1M Tokens", 10, "1M"),
		m("gpt 5 pro inp glbl Tokens", 0.015, "1K"),
		m("gpt 5 pro out glbl Tokens", 0.12, "1K"),
		m("gpt-5-codex-inp-dzone Tokens", 0.001375, "1K"),
		m("gpt-5-codex-out-dzone Tokens", 0.011, "1K"),
		m("gpt-4o-0806-Inp-regnl Tokens", 0.00275, "1K"),
		m("gpt-4o-0806-Outp-regnl Tokens", 0.011, "1K"),
		m("gpt 4o 1120 Inp regnl Tokens", 0.00275, "1K"),
		m("gpt 4o 1120 Outp regnl Tokens", 0.011, "1K"),
		m("gpt 4o 1120 Outp glbl Tokens", 0.01, "1K"),
		m("gpt 4o 0513 Input global Tokens", 0.005, "1K"),
		m("gpt 4o 0513 Input regional Tokens", 0.0055, "1K"),
		m("gpt 4o 0513 Output regional Tokens", 0.0165, "1K"),
		m("gpt-4o-aud-1217 Outp glbl Tokens", 0.08, "1K"),
		m("gpt-4o-mini-0718-Outp-glbl Tokens", 0.0006, "1K"),
		m("gpt-4o-mini-transcribe-txt-out-glbl Tokens", 0.005, "1K"),
		m("gpt-4o-mini-tts-aud-out-rgnl Tokens", 0.0132, "1K"),
		m("gpt 4o mini 0718 cached Inp regnl Tokens", 0.000083, "1K"),
	}
}

func TestResolveFromRegionItems_RealMeters_NoLookAlikeMatches(t *testing.T) {
	items := realEastUS2Meters()
	tests := []struct {
		model        string
		wantOK       bool
		wantIn, want float64
	}{
		{"gpt-4.1", true, 2, 8},
		{"gpt-4.1-mini", true, 0.4, 1.6},
		{"gpt-4.1-nano", true, 0.1, 0.4},
		{"gpt-5", true, 1.25, 10},
		{"gpt-5-pro", true, 15, 120},
		{"grok-4.1", true, 0.2, 0.5},
	}
	for _, tt := range tests {
		got, ok := ResolveFromRegionItems(items, tt.model, modelinfo.ModeGlobal, "short", "")
		if ok != tt.wantOK {
			t.Errorf("%s: ok=%v, want %v (%+v)", tt.model, ok, tt.wantOK, got)
			continue
		}
		if ok && (got.InputPerMTokUSD != tt.wantIn || got.OutputPerMTokUSD != tt.want) {
			t.Errorf("%s: got %v/%v, want %v/%v", tt.model, got.InputPerMTokUSD, got.OutputPerMTokUSD, tt.wantIn, tt.want)
		}
	}
}

func TestResolveFromRegionItems_RealMeters_VersionAmbiguousIsUnpriced(t *testing.T) {
	// gpt-4o meters carry the model version (0513, 0806, 1120) in the name and
	// 0513 is priced differently, so a version-less model name cannot choose.
	if got, ok := ResolveFromRegionItems(realEastUS2Meters(), "gpt-4o", modelinfo.ModeRegional, "short", ""); ok {
		t.Errorf("gpt-4o: expected unpriced for version-ambiguous meters, got %+v", got)
	}
}

func TestResolveFromRegionItems_RealMeters_VariantFamiliesDoNotLeak(t *testing.T) {
	items := []RetailItem{
		{MeterName: "gpt-4o-aud-1217 Outp glbl Tokens", RetailPrice: 0.08, UnitOfMeasure: "1K"},
		{MeterName: "gpt-4o-mini-transcribe-txt-out-glbl Tokens", RetailPrice: 0.005, UnitOfMeasure: "1K"},
		{MeterName: "gpt-4o-mini-tts-aud-out-rgnl Tokens", RetailPrice: 0.0132, UnitOfMeasure: "1K"},
	}
	for _, model := range []string{"gpt-4o", "gpt-4o-mini"} {
		if got, ok := ResolveFromRegionItems(items, model, modelinfo.ModeGlobal, "short", ""); ok {
			t.Errorf("%s: expected no match from audio/transcribe/tts meters, got %+v", model, got)
		}
	}
}

func TestResolvePTUHourlyUSD_OnlyAzureOpenAIProduct(t *testing.T) {
	items := []RetailItem{
		{MeterName: "Provisioned Managed Regional Unit", ProductName: "Llama", UnitOfMeasure: "1/Hour", RetailPrice: 9},
		{MeterName: "Provisioned Managed Regional Unit", ProductName: "Azure OpenAI", UnitOfMeasure: "1/Hour", RetailPrice: 2},
		{MeterName: "Provisioned Managed Global Unit", ProductName: "Azure OpenAI", UnitOfMeasure: "1/Hour", RetailPrice: 1},
		{MeterName: "Provisioned Throughput Units", ProductName: "Azure OpenAI", UnitOfMeasure: "1/Month", RetailPrice: 312},
	}
	if got, ok := ResolvePTUHourlyUSD(items, modelinfo.ModeRegional); !ok || got != 2 {
		t.Errorf("regional = %v %v, want 2", got, ok)
	}
	if _, ok := ResolvePTUHourlyUSD(items, modelinfo.ModeDataZone); ok {
		t.Error("data zone should be unresolved when no meter exists")
	}
	if _, ok := ResolvePTUHourlyUSD(items, "unknown"); ok {
		t.Error("unknown mode should be unresolved")
	}
}

func TestResolveFromRegionItems_VersionSelectsMeter(t *testing.T) {
	m := func(name string, price float64) RetailItem {
		return RetailItem{MeterName: name, RetailPrice: price, UnitOfMeasure: "1K"}
	}
	items := []RetailItem{
		m("gpt 4o 0513 Inp regnl Tokens", 0.005), m("gpt 4o 0513 Outp regnl Tokens", 0.015),
		m("gpt-4o-0806-Inp-regnl Tokens", 0.00275), m("gpt-4o-0806-Outp-regnl Tokens", 0.011),
	}
	for version, want := range map[string][2]float64{"2024-08-06": {2.75, 11}, "2024-05-13": {5, 15}} {
		got, ok := ResolveFromRegionItems(items, "gpt-4o", modelinfo.ModeRegional, "short", version)
		if !ok || got.InputPerMTokUSD != want[0] || got.OutputPerMTokUSD != want[1] {
			t.Errorf("version %s: got %+v ok=%v, want %v", version, got, ok, want)
		}
	}
	got, ok := ResolveFromRegionItems(items, "gpt-4o", modelinfo.ModeRegional, "short", "")
	if ok || got.Note == "" {
		t.Errorf("unknown version should be unpriced with a note, got %+v ok=%v", got, ok)
	}
	if got, ok := ResolveFromRegionItems(items, "gpt-4o", modelinfo.ModeRegional, "short", "2099-01-01"); ok {
		t.Errorf("unmatched version must stay unpriced, got %+v", got)
	}
}

func TestResolveFromRegionItems_RealMeters_EachGPT4oVersionPriced(t *testing.T) {
	want := map[string][2]float64{"2024-05-13": {5.5, 16.5}, "2024-08-06": {2.75, 11}, "2024-11-20": {2.75, 11}}
	for version, w := range want {
		got, ok := ResolveFromRegionItems(realEastUS2Meters(), "gpt-4o", modelinfo.ModeRegional, "short", version)
		if !ok || got.InputPerMTokUSD != w[0] || got.OutputPerMTokUSD != w[1] {
			t.Errorf("version %s: got %+v ok=%v, want %v", version, got, ok, w)
		}
	}
}

func TestResolveFromRegionItems_PlaceholderVersionTreatedAsUnknown(t *testing.T) {
	items := realEastUS2Meters()
	for _, v := range []string{"N/A", "n/a", "unknown"} {
		g1, ok1 := ResolveFromRegionItems(items, "gpt-4o", modelinfo.ModeRegional, "short", v)
		g2, ok2 := ResolveFromRegionItems(items, "gpt-4o", modelinfo.ModeRegional, "short", "")
		if ok1 != ok2 || g1 != g2 {
			t.Errorf("version %q differs from blank: %+v %v vs %+v %v", v, g1, ok1, g2, ok2)
		}
	}
}
