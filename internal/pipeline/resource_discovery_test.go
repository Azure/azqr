// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

// excelMaxRows mirrors the constant in stage_resource_discovery.go.
const excelMaxRows = 1048566

// excelLimitCheck replicates the guard from ResourceDiscoveryStage.Execute so
// the condition can be unit-tested without invoking the real Azure scanner.
func excelLimitCheck(xlsx bool, resourceCount int) error {
	if xlsx && resourceCount > excelMaxRows {
		return fmt.Errorf(
			"resource count (%d) exceeds Excel's maximum row limit (%d); use --csv or --json instead",
			resourceCount, excelMaxRows)
	}
	return nil
}

// buildDiscoveryCtx builds a ScanContext pre-seeded with n resources.
func buildDiscoveryCtx(xlsx bool, n int) *ScanContext {
	resources := make([]*models.Resource, n)
	return &ScanContext{
		Params:     &models.ScanParams{Xlsx: xlsx},
		ReportData: &renderers.ReportData{Resources: resources},
	}
}

// --- guard logic tests ------------------------------------------------------

func TestExcelLimit_NotEnforcedWhenXlsxFalse(t *testing.T) {
	if err := excelLimitCheck(false, excelMaxRows+1); err != nil {
		t.Errorf("guard should not fire when xlsx=false, got: %v", err)
	}
}

func TestExcelLimit_EnforcedWhenXlsxTrue(t *testing.T) {
	if err := excelLimitCheck(true, excelMaxRows+1); err == nil {
		t.Error("guard should fire when xlsx=true and count exceeds limit")
	}
}

func TestExcelLimit_ExactlyAtLimit_NotTriggered(t *testing.T) {
	if err := excelLimitCheck(true, excelMaxRows); err != nil {
		t.Errorf("guard should not fire at exactly the limit, got: %v", err)
	}
}

func TestExcelLimit_BelowLimit_NotTriggered(t *testing.T) {
	if err := excelLimitCheck(true, 1000); err != nil {
		t.Errorf("guard should not fire below the limit, got: %v", err)
	}
}

func TestExcelLimit_ZeroResources_NotTriggered(t *testing.T) {
	if err := excelLimitCheck(true, 0); err != nil {
		t.Errorf("guard should not fire for zero resources, got: %v", err)
	}
}

// --- error message tests ----------------------------------------------------

func TestExcelLimit_ErrorSuggestsAlternatives(t *testing.T) {
	err := excelLimitCheck(true, excelMaxRows+1)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	for _, want := range []string{"--csv", "--json"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error message should mention %q as alternative, got: %s", want, err.Error())
		}
	}
}

func TestExcelLimit_ErrorIncludesActualCount(t *testing.T) {
	const count = excelMaxRows + 12345
	err := excelLimitCheck(true, count)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("%d", count)) {
		t.Errorf("error message should include the actual count %d, got: %s", count, err.Error())
	}
}

// --- ScanContext field alignment tests --------------------------------------

// TestBuildDiscoveryCtx_XlsxFlag verifies the ctx helper correctly reflects
// the xlsx flag — guards future refactors of ScanParams field names.
func TestBuildDiscoveryCtx_XlsxFlag(t *testing.T) {
	ctxOn := buildDiscoveryCtx(true, 0)
	if !ctxOn.Params.Xlsx {
		t.Error("buildDiscoveryCtx(true, ...) should set Params.Xlsx = true")
	}
	ctxOff := buildDiscoveryCtx(false, 0)
	if ctxOff.Params.Xlsx {
		t.Error("buildDiscoveryCtx(false, ...) should set Params.Xlsx = false")
	}
}
