// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"context"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/internal/testutil"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/metrics"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/modelinfo"
)

func newTestAccount(t *testing.T, failGrains map[string]bool) *collectedAccount {
	t.Helper()
	q := &testutil.FakeQuerier{FailGrains: failGrains}
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	data := metrics.CollectBatch(context.Background(), q, "sub", []string{"r1"}, testPlan(), end, 30)["r1"]
	return &collectedAccount{
		Resource:         &models.Resource{ID: "r1", Name: "acct", ResourceGroup: "rg", Location: "eastus", Kind: "OpenAI"},
		SubscriptionName: "sub",
		Data:             data,
		Deployments:      map[string]deploymentInfo{"dep": {ModelVersion: "1", Mode: modelinfo.ModeGlobal}},
	}
}

func TestCollectExtendedSheets_RowWidthsAndRegionPriceCache(t *testing.T) {
	var calls int32
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Items": [
			{"retailPrice": 2.5, "unitOfMeasure": "1M", "meterName": "gpt-4o Inp Global Tokens"},
			{"retailPrice": 10, "unitOfMeasure": "1M", "meterName": "gpt-4o Outp Global Tokens"}
		], "NextPageLink": null}`))
	})

	// Two accounts in one region price the same model repeatedly; the
	// region's meters must be fetched once.
	accounts := []*collectedAccount{newTestAccount(t, nil), newTestAccount(t, nil)}
	accounts[1].Resource = &models.Resource{ID: "r2", Name: "acct2", ResourceGroup: "rg", Location: "eastus", Kind: "OpenAI"}

	outs := (&AIGovScanner{}).collectExtendedSheets(context.Background(), client, accounts)

	widths := map[string]int{
		"AI Gov Usage":     len(usageSheetMetadata().HeaderRow()),
		"AI Gov Cost":      len(costSheetMetadata().HeaderRow()),
		"AI Gov PTU":       len(ptuSheetMetadata().HeaderRow()),
		"AI Gov Workloads": len(workloadsSheetMetadata().HeaderRow()),
		"AI Gov Insights":  len(insightsSheetMetadata().HeaderRow()),
	}
	if len(outs) != len(widths) {
		t.Fatalf("sheets = %d, want %d", len(outs), len(widths))
	}
	for _, o := range outs {
		w, ok := widths[o.SheetName]
		if !ok {
			t.Fatalf("unexpected sheet %q", o.SheetName)
		}
		for i, row := range o.Table {
			if len(row) != w {
				t.Errorf("%s row %d width = %d, want %d", o.SheetName, i, len(row), w)
			}
		}
	}
	for _, name := range []string{"AI Gov Usage", "AI Gov PTU"} {
		for _, o := range outs {
			if o.SheetName == name && len(o.Table) < 2 {
				t.Errorf("%s has no data rows", name)
			}
		}
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("retail price fetches = %d, want 1 per region", got)
	}
}

func TestCollectExtendedSheets_ProvisionedAndBatchAreNotTokenPriced(t *testing.T) {
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Items": [
			{"retailPrice": 2.5, "unitOfMeasure": "1M", "meterName": "gpt-4o Inp Global Tokens"},
			{"retailPrice": 10, "unitOfMeasure": "1M", "meterName": "gpt-4o Outp Global Tokens"}
		], "NextPageLink": null}`))
	})

	for _, tc := range []struct {
		name       string
		mode       string
		wantCost   string
		wantStatus string
	}{
		{"standard", modelinfo.ModeGlobal, "", "Complete"},
		{"provisioned", modelinfo.ModeGlobalProvisioned, "N/A", "Not available"},
		{"batch", "batch", "N/A", "Not available"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			acct := newTestAccount(t, nil)
			acct.Deployments = map[string]deploymentInfo{"dep": {ModelVersion: "1", Mode: tc.mode}}
			outs := (&AIGovScanner{}).collectExtendedSheets(context.Background(), client, []*collectedAccount{acct})

			var cost, workloads [][]string
			for _, o := range outs {
				switch o.SheetName {
				case "AI Gov Cost":
					cost = o.Table
				case "AI Gov Workloads":
					workloads = o.Table
				}
			}
			if len(cost) < 2 || len(workloads) < 2 {
				t.Fatalf("expected data rows, cost=%d workloads=%d", len(cost), len(workloads))
			}
			costCol := len(costSheetMetadata().ColumnMetadata) - 3
			if tc.wantCost != "" && cost[1][costCol] != tc.wantCost {
				t.Errorf("Estimated Cost = %q, want %q", cost[1][costCol], tc.wantCost)
			}
			if tc.wantCost == "" && cost[1][costCol] == "N/A" {
				t.Errorf("standard deployment should be priced, got N/A")
			}
			if got := workloads[1][len(workloads[1])-1]; got != tc.wantStatus {
				t.Errorf("Cost Status = %q, want %q", got, tc.wantStatus)
			}
		})
	}
}

func TestCollectExtendedSheets_SameNamedAccountsStaySeparateWorkloads(t *testing.T) {
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"Items": [], "NextPageLink": null}`))
	})
	a, b := newTestAccount(t, nil), newTestAccount(t, nil)
	b.Resource = &models.Resource{ID: "r2", Name: "acct", ResourceGroup: "rg2", Location: "eastus", Kind: "OpenAI"}

	outs := (&AIGovScanner{}).collectExtendedSheets(context.Background(), client, []*collectedAccount{a, b})
	for _, o := range outs {
		if o.SheetName != "AI Gov Workloads" {
			continue
		}
		if len(o.Table) != 3 {
			t.Fatalf("workload rows = %d, want header + 2", len(o.Table))
		}
		if o.Table[1][0] == o.Table[2][0] || !strings.Contains(o.Table[1][0], "rg") {
			t.Errorf("expected qualified distinct labels, got %q and %q", o.Table[1][0], o.Table[2][0])
		}
	}
}

func TestCollectExtendedSheets_UncomputedPTUValuesAreNA(t *testing.T) {
	client := testutil.NewPricingClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"Items": [], "NextPageLink": null}`))
	})
	acct := newTestAccount(t, nil)
	acct.Deployments = map[string]deploymentInfo{"dep": {ModelVersion: "1", Mode: modelinfo.ModeGlobalProvisioned}}

	outs := (&AIGovScanner{}).collectExtendedSheets(context.Background(), client, []*collectedAccount{acct})
	for _, o := range outs {
		if o.SheetName != "AI Gov PTU" {
			continue
		}
		header := o.Table[0]
		for _, name := range []string{"Observed Avg TPM (PTU-weighted)", "Recommended PTUs", "Estimated PAYG Monthly (USD)", "Break-even TPM (PTU-weighted)"} {
			for i, h := range header {
				if h == name && o.Table[1][i] != "N/A" {
					t.Errorf("%s = %q, want N/A", name, o.Table[1][i])
				}
			}
		}
	}
}
