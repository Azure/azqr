// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/xuri/excelize/v2"
)

// --- test helpers -----------------------------------------------------------

// testSubID is a valid 36-char UUID used in test data.
// MaskSubscriptionID returns "" for strings shorter than 36 chars.
const testSubID = "00000000-0000-0000-0000-000000000001"

// stagesWithOnly returns a StageConfigs with exactly one stage enabled.
func stagesWithOnly(stageName string) *models.StageConfigs {
	s := models.NewStageConfigs()
	_ = s.EnableStage(stageName)
	return s
}

// hasSheet reports whether sheetName exists in the file.
func hasSheet(f *excelize.File, sheetName string) bool {
	return slices.Contains(f.GetSheetList(), sheetName)
}

// cellAt returns the string value of the cell at the given 1-based (col, row).
func cellAt(t *testing.T, f *excelize.File, sheet string, col, row int) string {
	t.Helper()
	addr, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		t.Fatalf("CoordinatesToCellName(%d,%d): %v", col, row, err)
	}
	v, err := f.GetCellValue(sheet, addr)
	if err != nil {
		t.Fatalf("GetCellValue(%s, %s): %v", sheet, addr, err)
	}
	return v
}

// cellHasHyperlink reports whether a =HYPERLINK() formula is set on the cell at (col, row).
func cellHasHyperlink(t *testing.T, f *excelize.File, sheet string, col, row int) bool {
	t.Helper()
	addr, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		t.Fatalf("CoordinatesToCellName(%d,%d): %v", col, row, err)
	}
	formula, err := f.GetCellFormula(sheet, addr)
	if err != nil {
		t.Fatalf("GetCellFormula(%s, %s): %v", sheet, addr, err)
	}
	return len(formula) > 10 && formula[:10] == "HYPERLINK("
}

// freshFile creates a new excelize file with shared styles, and registers
// file.Close() as a test cleanup action.
func freshFile(t *testing.T) (*excelize.File, *StyleCache) {
	t.Helper()
	f := excelize.NewFile()
	t.Cleanup(func() { _ = f.Close() })
	styles, err := createSharedStyles(f)
	if err != nil {
		t.Fatalf("createSharedStyles: %v", err)
	}
	return f, styles
}

// saveAndReopen persists f to a temporary XLSX file and reopens it.
// Required after StreamWriter Flush because the in-memory cell map is cleared.
func saveAndReopen(t *testing.T, f *excelize.File) *excelize.File {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.xlsx")
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("saveAndReopen SaveAs: %v", err)
	}
	f2, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("saveAndReopen OpenFile: %v", err)
	}
	t.Cleanup(func() { _ = f2.Close() })
	return f2
}

// staticTable builds a tableFunc that always returns the given header + rows.
func staticTable(header []string, rows ...[]string) func() [][]string {
	all := make([][]string, 0, len(rows)+1)
	all = append(all, header)
	all = append(all, rows...)
	return func() [][]string { return all }
}

// --- TestRenderSheet: core helper behaviour ---------------------------------

func TestRenderSheet(t *testing.T) {
	const stage = models.StageNameAdvisor
	const sheet = "RenderSheetTest"
	header := []string{"Col1", "Col2", "Col3"}

	t.Run("stage_disabled_sheet_not_created", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{Stages: models.NewStageConfigs()} // no stage enabled

		renderSheet(f, data, sheetConfig{stageName: stage, sheetName: sheet,
			tableFunc: staticTable(header)}, styles)

		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when its stage is disabled", sheet)
		}
	})

	t.Run("no_data_sheet_created_headers_at_row4_row5_empty", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{Stages: stagesWithOnly(stage)}

		renderSheet(f, data, sheetConfig{stageName: stage, sheetName: sheet,
			tableFunc: staticTable(header) /* no data rows */}, styles)

		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q must be created even with no data rows", sheet)
		}
		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Col1" {
			t.Errorf("header A4 = %q, want Col1", got)
		}
		if got := cellAt(t, f, sheet, 1, 5); got != "" {
			t.Errorf("A5 = %q, want empty (no data written)", got)
		}
	})

	t.Run("data_rows_written_starting_at_row5", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{Stages: stagesWithOnly(stage)}

		renderSheet(f, data, sheetConfig{stageName: stage, sheetName: sheet,
			tableFunc: staticTable(header,
				[]string{"r1c1", "r1c2", "r1c3"},
				[]string{"r2c1", "r2c2", "r2c3"},
			)}, styles)

		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Col1" {
			t.Errorf("header A4 = %q, want Col1", got)
		}
		if got := cellAt(t, f, sheet, 1, 5); got != "r1c1" {
			t.Errorf("first data row A5 = %q, want r1c1", got)
		}
		if got := cellAt(t, f, sheet, 1, 6); got != "r2c1" {
			t.Errorf("second data row A6 = %q, want r2c1", got)
		}
	})

	t.Run("hyperlink_set_on_specified_column", func(t *testing.T) {
		const hyperlinkCol = 2
		f, styles := freshFile(t)
		data := &renderers.ReportData{Stages: stagesWithOnly(stage)}

		renderSheet(f, data, sheetConfig{stageName: stage, sheetName: sheet,
			hyperlinkCol: hyperlinkCol,
			tableFunc: staticTable(header,
				[]string{"value", "https://example.com", "value"},
			)}, styles)

		f = saveAndReopen(t, f)
		if !cellHasHyperlink(t, f, sheet, hyperlinkCol, 5) {
			t.Errorf("expected hyperlink at col %d row 5", hyperlinkCol)
		}
	})

	t.Run("no_hyperlink_when_hyperlinkCol_is_zero", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{Stages: stagesWithOnly(stage)}

		renderSheet(f, data, sheetConfig{stageName: stage, sheetName: sheet,
			hyperlinkCol: 0,
			tableFunc: staticTable(header,
				[]string{"https://example.com", "https://example.com", "val"},
			)}, styles)

		f = saveAndReopen(t, f)
		// Neither col 1 nor col 2 should get a hyperlink
		for col := 1; col <= 3; col++ {
			if cellHasHyperlink(t, f, sheet, col, 5) {
				t.Errorf("unexpected hyperlink at col %d row 5 when hyperlinkCol==0", col)
			}
		}
	})

	t.Run("isFirstSheet_renames_Sheet1", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{Stages: stagesWithOnly(stage)}

		renderSheet(f, data, sheetConfig{stageName: stage, sheetName: sheet,
			isFirstSheet: true,
			tableFunc:    staticTable(header)}, styles)

		if !hasSheet(f, sheet) {
			t.Errorf("sheet %q should exist after rename", sheet)
		}
		if hasSheet(f, "Sheet1") {
			t.Errorf("Sheet1 should have been renamed to %q", sheet)
		}
	})

	t.Run("isFirstSheet_false_does_not_rename_Sheet1", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{Stages: stagesWithOnly(stage)}

		renderSheet(f, data, sheetConfig{stageName: stage, sheetName: sheet,
			isFirstSheet: false,
			tableFunc:    staticTable(header)}, styles)

		// Both Sheet1 (the default) and the new sheet should exist
		if !hasSheet(f, "Sheet1") {
			t.Error("Sheet1 must survive when isFirstSheet==false")
		}
		if !hasSheet(f, sheet) {
			t.Errorf("new sheet %q must exist", sheet)
		}
	})
}

// --- Per-renderer tests -----------------------------------------------------
// Each renderer test covers:
//  (a) stage disabled → sheet not created
//  (b) stage enabled + data → correct sheet name, correct header at A4, correct data at A5
//  (c) hyperlink column where applicable

func TestRenderAdvisor(t *testing.T) {
	const sheet = "Advisor"

	t.Run("stage_disabled_no_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		renderAdvisor(f, &renderers.ReportData{Stages: models.NewStageConfigs()}, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when stage %s is disabled", sheet, models.StageNameAdvisor)
		}
	})

	t.Run("with_data_correct_sheet_and_header", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameAdvisor),
			Advisor: []*models.AdvisorResult{
				{SubscriptionID: testSubID, SubscriptionName: "Sub One", Category: "HighAvailability", Impact: "High"},
			},
		}
		renderAdvisor(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q not created", sheet)
		}
		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Subscription Id" {
			t.Errorf("A4 = %q, want %q", got, "Subscription Id")
		}
		if got := cellAt(t, f, sheet, 1, 5); got != testSubID {
			t.Errorf("A5 = %q, want %q", got, testSubID)
		}
	})
}

func TestRenderArcSQL(t *testing.T) {
	const sheet = "Arc SQL"

	t.Run("stage_disabled_no_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		renderArcSQL(f, &renderers.ReportData{Stages: models.NewStageConfigs()}, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when stage %s is disabled", sheet, models.StageNameArc)
		}
	})

	t.Run("with_data_correct_sheet_and_header", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameArc),
			ArcSQL: []*models.ArcSQLResult{
				{SubscriptionID: testSubID, SubscriptionName: "Sub One", Edition: "Enterprise"},
			},
		}
		renderArcSQL(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q not created", sheet)
		}
		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Subscription Id" {
			t.Errorf("A4 = %q, want %q", got, "Subscription Id")
		}
		if got := cellAt(t, f, sheet, 1, 5); got != testSubID {
			t.Errorf("A5 = %q, want %q", got, testSubID)
		}
	})
}

func TestRenderAzurePolicy(t *testing.T) {
	const sheet = "Azure Policy"

	t.Run("stage_disabled_no_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		renderAzurePolicy(f, &renderers.ReportData{Stages: models.NewStageConfigs()}, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when stage %s is disabled", sheet, models.StageNamePolicy)
		}
	})

	t.Run("with_data_correct_sheet_and_header", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNamePolicy),
			AzurePolicy: []*models.AzurePolicyResult{
				{SubscriptionID: testSubID, PolicyDisplayName: "Policy A"},
			},
		}
		renderAzurePolicy(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q not created", sheet)
		}
		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Subscription Id" {
			t.Errorf("A4 = %q, want %q", got, "Subscription Id")
		}
		if got := cellAt(t, f, sheet, 1, 5); got != testSubID {
			t.Errorf("A5 = %q, want %q", got, testSubID)
		}
	})
}

func TestRenderCosts(t *testing.T) {
	const sheet = "Costs"

	t.Run("stage_disabled_no_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		renderCosts(f, &renderers.ReportData{Stages: models.NewStageConfigs()}, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when stage %s is disabled", sheet, models.StageNameCost)
		}
	})

	t.Run("with_data_correct_sheet_and_header", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameCost),
			Cost: []*models.CostResult{
				{SubscriptionID: testSubID, ServiceName: "Compute", Value: "99.00", Currency: "USD",
					From: time.Now(), To: time.Now()},
			},
		}
		renderCosts(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q not created", sheet)
		}
		f = saveAndReopen(t, f)
		// CostTable header order: From, To, Subscription Id, Subscription Name, Service Name, Value, Currency
		if got := cellAt(t, f, sheet, 1, 4); got != "From" {
			t.Errorf("A4 = %q, want %q", got, "From")
		}
		// Col 3 is Subscription Id
		if got := cellAt(t, f, sheet, 3, 5); got != testSubID {
			t.Errorf("C5 (Subscription Id) = %q, want %q", got, testSubID)
		}
	})
}

func TestRenderDefender(t *testing.T) {
	const sheet = "Defender"

	t.Run("stage_disabled_no_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		renderDefender(f, &renderers.ReportData{Stages: models.NewStageConfigs()}, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when stage %s is disabled", sheet, models.StageNameDefender)
		}
	})

	t.Run("with_data_correct_sheet_and_header", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameDefender),
			Defender: []*models.DefenderResult{
				{SubscriptionID: testSubID, SubscriptionName: "Sub One", Name: "VirtualMachines", Tier: "Standard"},
			},
		}
		renderDefender(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q not created", sheet)
		}
		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Subscription Id" {
			t.Errorf("A4 = %q, want %q", got, "Subscription Id")
		}
		if got := cellAt(t, f, sheet, 1, 5); got != testSubID {
			t.Errorf("A5 = %q, want %q", got, testSubID)
		}
	})
}

func TestRenderDefenderRecommendations(t *testing.T) {
	const sheet = "DefenderRecommendations"
	// Column 11 in DefenderRecommendationsTable is "AzPortal Link"
	const azPortalLinkCol = 11

	t.Run("stage_disabled_no_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		renderDefenderRecommendations(f, &renderers.ReportData{Stages: models.NewStageConfigs()}, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when stage %s is disabled", sheet, models.StageNameDefenderRecommendations)
		}
	})

	t.Run("with_data_correct_sheet_and_header", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameDefenderRecommendations),
			DefenderRecommendations: []*models.DefenderRecommendation{
				{SubscriptionId: testSubID, SubscriptionName: "Sub One"},
			},
		}
		renderDefenderRecommendations(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q not created", sheet)
		}
		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Subscription Id" {
			t.Errorf("A4 = %q, want %q", got, "Subscription Id")
		}
	})

	// Regression: hyperlink must land on col 11 (AzPortal Link), not some other column.
	t.Run("hyperlink_on_col11_AzPortalLink", func(t *testing.T) {
		f, styles := freshFile(t)
		const portalURL = "https://portal.azure.com/recommendation1"
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameDefenderRecommendations),
			DefenderRecommendations: []*models.DefenderRecommendation{
				{SubscriptionId: "sub-1", AzPortalLink: portalURL},
			},
		}
		renderDefenderRecommendations(f, data, styles)
		f = saveAndReopen(t, f)
		if !cellHasHyperlink(t, f, sheet, azPortalLinkCol, 5) {
			t.Errorf("expected hyperlink at col %d (AzPortal Link) row 5", azPortalLinkCol)
		}
		// Adjacent columns must not accidentally receive a hyperlink
		if cellHasHyperlink(t, f, sheet, azPortalLinkCol-1, 5) {
			t.Errorf("unexpected hyperlink at col %d row 5", azPortalLinkCol-1)
		}
		if cellHasHyperlink(t, f, sheet, azPortalLinkCol+1, 5) {
			t.Errorf("unexpected hyperlink at col %d row 5", azPortalLinkCol+1)
		}
	})
}

func TestRenderImpactedResources(t *testing.T) {
	const sheet = "ImpactedResources"
	// Column 18 in ImpactedTable is "Learn"
	const learnCol = 18

	t.Run("stage_disabled_no_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		renderImpactedResources(f, &renderers.ReportData{Stages: models.NewStageConfigs()}, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when stage %s is disabled", sheet, models.StageNameGraph)
		}
	})

	t.Run("with_data_correct_sheet_and_header", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameGraph),
			Graph: []*models.GraphResult{
				{RecommendationID: "rec-001", SubscriptionID: "sub-1", Category: models.CategoryHighAvailability},
			},
		}
		renderImpactedResources(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q not created", sheet)
		}
		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Validated Using" {
			t.Errorf("A4 = %q, want %q", got, "Validated Using")
		}
	})

	// Regression: hyperlink must land on col 18 (Learn), not some other column.
	t.Run("hyperlink_on_col18_Learn", func(t *testing.T) {
		f, styles := freshFile(t)
		const learnURL = "https://learn.microsoft.com/recommendation"
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameGraph),
			Graph: []*models.GraphResult{
				{
					RecommendationID: "rec-001",
					SubscriptionID:   "sub-1",
					Category:         models.CategoryHighAvailability,
					Learn:            learnURL,
				},
			},
		}
		renderImpactedResources(f, data, styles)
		f = saveAndReopen(t, f)
		if !cellHasHyperlink(t, f, sheet, learnCol, 5) {
			t.Errorf("expected hyperlink at col %d (Learn) row 5", learnCol)
		}
	})
}

func TestRenderRecommendations(t *testing.T) {
	const sheet = "Recommendations"
	// Column 11 in RecommendationsTable is "Read More"
	const readMoreCol = 11

	t.Run("stage_disabled_Sheet1_not_renamed", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages:          models.NewStageConfigs(),
			Recommendations: map[string]map[string]*models.GraphRecommendation{},
		}
		renderRecommendations(f, data, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not exist when stage %s is disabled", sheet, models.StageNameGraph)
		}
		if !hasSheet(f, "Sheet1") {
			t.Error("Sheet1 must be preserved when stage is disabled")
		}
	})

	t.Run("stage_enabled_renames_Sheet1_to_Recommendations", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages:          stagesWithOnly(models.StageNameGraph),
			Recommendations: map[string]map[string]*models.GraphRecommendation{},
		}
		renderRecommendations(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Errorf("sheet %q must exist after stage-enabled render", sheet)
		}
		if hasSheet(f, "Sheet1") {
			t.Error("Sheet1 must be renamed to Recommendations, not kept")
		}
	})

	// Regression: hyperlink must land on col 11 (Read More), not some other column.
	t.Run("hyperlink_on_col11_ReadMore", func(t *testing.T) {
		f, styles := freshFile(t)
		const learnURL = "https://learn.microsoft.com/best-practices"
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameGraph),
			Recommendations: map[string]map[string]*models.GraphRecommendation{
				"microsoft.compute/virtualmachines": {
					"rec-001": {
						RecommendationID: "rec-001",
						ResourceType:     "microsoft.compute/virtualmachines",
						LearnMoreLink: []struct {
							Name string `yaml:"name"`
							Url  string `yaml:"url"`
						}{{Name: "Read More", Url: learnURL}},
					},
				},
			},
			Graph: []*models.GraphResult{
				{RecommendationID: "rec-001", ResourceType: "microsoft.compute/virtualmachines"},
			},
		}
		renderRecommendations(f, data, styles)
		f = saveAndReopen(t, f)
		if !cellHasHyperlink(t, f, sheet, readMoreCol, 5) {
			t.Errorf("expected hyperlink at col %d (Read More) row 5", readMoreCol)
		}
	})
}

func TestRenderResourceTypes(t *testing.T) {
	const sheet = "ResourceTypes"

	t.Run("stage_disabled_no_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		renderResourceTypes(f, &renderers.ReportData{Stages: models.NewStageConfigs()}, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when stage %s is disabled", sheet, models.StageNameGraph)
		}
	})

	t.Run("with_data_correct_sheet_header_and_values", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameGraph),
			ResourceTypeCount: []*models.ResourceTypeCount{
				{Subscription: "Sub One", ResourceType: "Microsoft.Compute/virtualMachines", Count: 42},
			},
		}
		renderResourceTypes(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q not created", sheet)
		}
		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Subscription Name" {
			t.Errorf("A4 = %q, want %q", got, "Subscription Name")
		}
		if got := cellAt(t, f, sheet, 1, 5); got != "Sub One" {
			t.Errorf("A5 = %q, want Sub One", got)
		}
	})
}

func TestRenderResources(t *testing.T) {
	const sheet = "Inventory"

	t.Run("stage_disabled_no_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		renderResources(f, &renderers.ReportData{Stages: models.NewStageConfigs()}, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when stage %s is disabled", sheet, models.StageNameGraph)
		}
	})

	t.Run("with_data_creates_inventory_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameGraph),
			Resources: []*models.Resource{
				{ID: "/subscriptions/" + testSubID + "/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1",
					SubscriptionID: testSubID, ResourceGroup: "rg",
					Type: "Microsoft.Compute/virtualMachines", Name: "vm1"},
			},
		}
		renderResources(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q not created", sheet)
		}
		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Subscription Id" {
			t.Errorf("A4 = %q, want %q", got, "Subscription Id")
		}
		if got := cellAt(t, f, sheet, 1, 5); got != testSubID {
			t.Errorf("A5 = %q, want %q", got, testSubID)
		}
	})
}

func TestRenderExcludedResources(t *testing.T) {
	const sheet = "OutOfScope"

	t.Run("stage_disabled_no_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		renderExcludedResources(f, &renderers.ReportData{Stages: models.NewStageConfigs()}, styles)
		if hasSheet(f, sheet) {
			t.Errorf("sheet %q must not be created when stage %s is disabled", sheet, models.StageNameGraph)
		}
	})

	t.Run("with_data_creates_outofscope_sheet", func(t *testing.T) {
		f, styles := freshFile(t)
		data := &renderers.ReportData{
			Stages: stagesWithOnly(models.StageNameGraph),
			ExludedResources: []*models.Resource{
				{ID: "/subscriptions/" + testSubID + "/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm2",
					SubscriptionID: testSubID, ResourceGroup: "rg",
					Type: "Microsoft.Compute/virtualMachines", Name: "vm2"},
			},
		}
		renderExcludedResources(f, data, styles)
		if !hasSheet(f, sheet) {
			t.Fatalf("sheet %q not created", sheet)
		}
		f = saveAndReopen(t, f)
		if got := cellAt(t, f, sheet, 1, 4); got != "Subscription Id" {
			t.Errorf("A4 = %q, want %q", got, "Subscription Id")
		}
		if got := cellAt(t, f, sheet, 1, 5); got != testSubID {
			t.Errorf("A5 = %q, want %q", got, testSubID)
		}
	})
}

// --- Integration test -------------------------------------------------------

// TestFullReport_AllSheetsPresent runs CreateExcelReport with all stages
// enabled and representative data, then opens the output file and verifies
// every expected sheet exists with the correct first header at row 4.
func TestFullReport_AllSheetsPresent(t *testing.T) {
	const outputName = "test_full_regression"
	filename := outputName + ".xlsx"
	t.Cleanup(func() { _ = os.Remove(filename) })

	stages := models.NewStageConfigsWithDefaults()
	_ = stages.EnableStage(models.StageNameDefenderRecommendations)
	_ = stages.EnableStage(models.StageNameArc)
	_ = stages.EnableStage(models.StageNamePolicy)
	_ = stages.EnableStage(models.StageNameCost)

	data := &renderers.ReportData{
		OutputFileName:  outputName,
		Stages:          stages,
		Recommendations: map[string]map[string]*models.GraphRecommendation{},
		Graph: []*models.GraphResult{
			{RecommendationID: "rec-001", SubscriptionID: testSubID,
				Category: models.CategoryHighAvailability, Learn: "https://learn.microsoft.com/x"},
		},
		Advisor: []*models.AdvisorResult{
			{SubscriptionID: testSubID, SubscriptionName: "Sub One", Category: "HighAvailability", Impact: "High"},
		},
		ArcSQL: []*models.ArcSQLResult{
			{SubscriptionID: testSubID, SubscriptionName: "Sub One", Edition: "Enterprise"},
		},
		AzurePolicy: []*models.AzurePolicyResult{
			{SubscriptionID: testSubID, PolicyDisplayName: "Policy A"},
		},
		Cost: []*models.CostResult{
			{SubscriptionID: testSubID, ServiceName: "Compute", Value: "10.00", Currency: "USD",
				From: time.Now(), To: time.Now()},
		},
		Defender: []*models.DefenderResult{
			{SubscriptionID: testSubID, SubscriptionName: "Sub One", Name: "VMs", Tier: "Standard"},
		},
		DefenderRecommendations: []*models.DefenderRecommendation{
			{SubscriptionId: testSubID, AzPortalLink: "https://portal.azure.com/rec"},
		},
		ResourceTypeCount: []*models.ResourceTypeCount{
			{Subscription: "Sub One", ResourceType: "Microsoft.Compute/virtualMachines", Count: 1},
		},
		Resources: []*models.Resource{
			{ID: "/subscriptions/" + testSubID + "/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1",
				SubscriptionID: testSubID, Name: "vm1"},
		},
		ExludedResources: []*models.Resource{
			{ID: "/subscriptions/" + testSubID + "/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm2",
				SubscriptionID: testSubID, Name: "vm2"},
		},
	}

	CreateExcelReport(data)

	f, err := excelize.OpenFile(filename)
	if err != nil {
		t.Fatalf("failed to open generated report %s: %v", filename, err)
	}
	t.Cleanup(func() { _ = f.Close() })

	// Every expected sheet must be present
	expectedSheets := []string{
		"Recommendations", "ImpactedResources", "ResourceTypes",
		"Inventory", "Advisor", "Azure Policy", "Arc SQL",
		"DefenderRecommendations", "Defender", "OutOfScope", "Costs",
	}
	for _, sheet := range expectedSheets {
		if !hasSheet(f, sheet) {
			t.Errorf("expected sheet %q is missing from the report", sheet)
		}
	}

	// Sheet1 must have been renamed (not left behind)
	if hasSheet(f, "Sheet1") {
		t.Error("Sheet1 must not appear in the final report (should be renamed to Recommendations)")
	}

	// First header cell (A4) must match the known first column of each table
	wantHeaderA4 := map[string]string{
		"Recommendations":         "Implemented",
		"ImpactedResources":       "Validated Using",
		"ResourceTypes":           "Subscription Name",
		"Inventory":               "Subscription Id",
		"Advisor":                 "Subscription Id",
		"Azure Policy":            "Subscription Id",
		"Arc SQL":                 "Subscription Id",
		"DefenderRecommendations": "Subscription Id",
		"Defender":                "Subscription Id",
		"OutOfScope":              "Subscription Id",
		"Costs":                   "From", // CostTable: From, To, Subscription Id, ...
	}
	for sheet, want := range wantHeaderA4 {
		if got := cellAt(t, f, sheet, 1, 4); got != want {
			t.Errorf("sheet %q A4 = %q, want %q", sheet, got, want)
		}
	}
}
