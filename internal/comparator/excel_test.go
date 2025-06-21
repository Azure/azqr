// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package comparator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

// createTestExcelFile creates a test Excel file with specified sheets and data
func createTestExcelFile(t *testing.T, filename string, sheets map[string][][]string) string {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("Failed to close test Excel file: %v", err)
		}
	}()

	// Delete default Sheet1 if we have custom sheets
	if len(sheets) > 0 {
		_ = f.DeleteSheet("Sheet1")
	}

	// Create sheets with data
	for sheetName, rows := range sheets {
		_, err := f.NewSheet(sheetName)
		if err != nil {
			t.Fatalf("Failed to create sheet %s: %v", sheetName, err)
		}

		// Write rows
		for rowIdx, row := range rows {
			for colIdx, cell := range row {
				cellName, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
				if err != nil {
					t.Fatalf("Failed to get cell coordinates: %v", err)
				}
				err = f.SetCellValue(sheetName, cellName, cell)
				if err != nil {
					t.Fatalf("Failed to set cell value: %v", err)
				}
			}
		}
	}

	// Save file
	if err := f.SaveAs(filePath); err != nil {
		t.Fatalf("Failed to save test Excel file: %v", err)
	}

	return filePath
}

func TestCompareExcelFiles_BasicComparison(t *testing.T) {
	// Create first test file
	file1Data := map[string][][]string{
		"Recommendations": {
			{"Header1", "Header2", "Header3"},
			{"", "", ""},                 // Row 2 - empty
			{"", "", ""},                 // Row 3 - empty
			{"ID", "Name", "Status"},     // Row 4 - headers
			{"1", "Resource1", "Active"}, // Row 5 - data
			{"2", "Resource2", "Active"}, // Row 6 - data
		},
		"Inventory": {
			{"Header1", "Header2"},
			{"", ""},       // Row 2
			{"", ""},       // Row 3
			{"ID", "Name"}, // Row 4
			{"A", "Item1"}, // Row 5
		},
	}

	// Create second test file with different row counts
	file2Data := map[string][][]string{
		"Recommendations": {
			{"Header1", "Header2", "Header3"},
			{"", "", ""},                 // Row 2
			{"", "", ""},                 // Row 3
			{"ID", "Name", "Status"},     // Row 4
			{"1", "Resource1", "Active"}, // Row 5
			{"2", "Resource2", "Active"}, // Row 6
			{"3", "Resource3", "Active"}, // Row 7 - new row
		},
		"Inventory": {
			{"Header1", "Header2"},
			{"", ""},       // Row 2
			{"", ""},       // Row 3
			{"ID", "Name"}, // Row 4
			{"A", "Item1"}, // Row 5
			{"B", "Item2"}, // Row 6 - new row
		},
		"NewSheet": {
			{"Col1"},
			{"", ""},   // Row 2
			{"", ""},   // Row 3
			{"Header"}, // Row 4
			{"Data1"},  // Row 5
		},
	}

	file1 := createTestExcelFile(t, "test1.xlsx", file1Data)
	file2 := createTestExcelFile(t, "test2.xlsx", file2Data)

	// Compare files
	result, err := CompareExcelFiles(file1, file2)
	if err != nil {
		t.Fatalf("CompareExcelFiles failed: %v", err)
	}

	// Verify results
	if result.OldFilePath != file1 {
		t.Errorf("Expected OldFilePath=%s, got %s", file1, result.OldFilePath)
	}
	if result.NewFilePath != file2 {
		t.Errorf("Expected NewFilePath=%s, got %s", file2, result.NewFilePath)
	}

	// Check sheet comparisons (should have at least 3, may have default sheet)
	if len(result.SheetComparisons) < 3 {
		t.Errorf("Expected at least 3 sheet comparisons, got %d", len(result.SheetComparisons))
	}

	// Find specific sheet comparisons
	var recsComparison, invComparison, newSheetComparison *SheetComparison
	for i := range result.SheetComparisons {
		switch result.SheetComparisons[i].SheetName {
		case "Recommendations":
			recsComparison = &result.SheetComparisons[i]
		case "Inventory":
			invComparison = &result.SheetComparisons[i]
		case "NewSheet":
			newSheetComparison = &result.SheetComparisons[i]
		}
	}

	// Test Recommendations sheet
	if recsComparison == nil {
		t.Fatal("Recommendations sheet comparison not found")
		return
	}
	if recsComparison.OldRowCount != 6 {
		t.Errorf("Recommendations OldRowCount: expected 6, got %d", recsComparison.OldRowCount)
	}
	if recsComparison.NewRowCount != 7 {
		t.Errorf("Recommendations NewRowCount: expected 7, got %d", recsComparison.NewRowCount)
	}
	if recsComparison.RowDifference != 1 {
		t.Errorf("Recommendations RowDifference: expected 1, got %d", recsComparison.RowDifference)
	}

	// Test Inventory sheet
	if invComparison == nil {
		t.Fatal("Inventory sheet comparison not found")
		return
	}
	if invComparison.OldRowCount != 5 {
		t.Errorf("Inventory OldRowCount: expected 5, got %d", invComparison.OldRowCount)
	}
	if invComparison.NewRowCount != 6 {
		t.Errorf("Inventory NewRowCount: expected 6, got %d", invComparison.NewRowCount)
	}

	// Test new sheet (only in file2)
	if newSheetComparison == nil {
		t.Fatal("NewSheet comparison not found")
		return
	}
	if newSheetComparison.OldRowCount != 0 {
		t.Errorf("NewSheet OldRowCount: expected 0, got %d", newSheetComparison.OldRowCount)
	}
	if newSheetComparison.NewRowCount != 5 {
		t.Errorf("NewSheet NewRowCount: expected 5, got %d", newSheetComparison.NewRowCount)
	}
}

func TestCompareExcelFiles_DuplicateDetection(t *testing.T) {
	// Create test file with duplicates after row 4
	fileData := map[string][][]string{
		"TestSheet": {
			{"Logo", "Area"},
			{"", ""},                       // Row 2
			{"", ""},                       // Row 3
			{"ID", "Name", "Status"},       // Row 4 - headers
			{"1", "Resource1", "Active"},   // Row 5 - data
			{"2", "Resource2", "Active"},   // Row 6 - data
			{"1", "Resource1", "Active"},   // Row 7 - duplicate of row 5
			{"3", "Resource3", "Inactive"}, // Row 8 - unique
			{"2", "Resource2", "Active"},   // Row 9 - duplicate of row 6
		},
	}

	file := createTestExcelFile(t, "test_duplicates.xlsx", fileData)

	// Compare file with itself to trigger duplicate detection
	result, err := CompareExcelFiles(file, file)
	if err != nil {
		t.Fatalf("CompareExcelFiles failed: %v", err)
	}

	// Find TestSheet comparison
	var testSheetComparison *SheetComparison
	for i := range result.SheetComparisons {
		if result.SheetComparisons[i].SheetName == "TestSheet" {
			testSheetComparison = &result.SheetComparisons[i]
			break
		}
	}

	if testSheetComparison == nil {
		t.Fatal("TestSheet comparison not found")
		return
	}

	// Check duplicate detection
	if !testSheetComparison.HasDuplicates {
		t.Error("Expected duplicates to be detected")
	}

	if len(testSheetComparison.DuplicateRows) == 0 {
		t.Error("Expected duplicate rows to be listed")
	}

	// Expected duplicate rows: 5, 6, 7, 9 (1-indexed)
	// Row 5 and 7 are duplicates, Row 6 and 9 are duplicates
	expectedDuplicates := map[int]bool{5: true, 6: true, 7: true, 9: true}
	for _, rowNum := range testSheetComparison.DuplicateRows {
		if !expectedDuplicates[rowNum] {
			t.Errorf("Unexpected duplicate row number: %d", rowNum)
		}
	}
}

func TestCompareExcelFiles_NoDuplicates(t *testing.T) {
	// Create test file without duplicates
	fileData := map[string][][]string{
		"TestSheet": {
			{"Logo", "Area"},
			{"", ""},                       // Row 2
			{"", ""},                       // Row 3
			{"ID", "Name", "Status"},       // Row 4 - headers
			{"1", "Resource1", "Active"},   // Row 5
			{"2", "Resource2", "Active"},   // Row 6
			{"3", "Resource3", "Inactive"}, // Row 7
		},
	}

	file := createTestExcelFile(t, "test_no_duplicates.xlsx", fileData)

	result, err := CompareExcelFiles(file, file)
	if err != nil {
		t.Fatalf("CompareExcelFiles failed: %v", err)
	}

	// Find TestSheet comparison
	var testSheetComparison *SheetComparison
	for i := range result.SheetComparisons {
		if result.SheetComparisons[i].SheetName == "TestSheet" {
			testSheetComparison = &result.SheetComparisons[i]
			break
		}
	}

	if testSheetComparison == nil {
		t.Fatal("TestSheet comparison not found")
		return
	}

	// Check no duplicates detected
	if testSheetComparison.HasDuplicates {
		t.Error("Expected no duplicates to be detected")
	}

	if len(testSheetComparison.DuplicateRows) > 0 {
		t.Errorf("Expected no duplicate rows, got %v", testSheetComparison.DuplicateRows)
	}
}

func TestCompareExcelFiles_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "nonexistent.xlsx")

	// Create a valid test file
	validFileData := map[string][][]string{
		"Sheet1": {{"A", "B"}, {"1", "2"}},
	}
	validFile := createTestExcelFile(t, "valid.xlsx", validFileData)

	// Test with non-existent file1
	_, err := CompareExcelFiles(nonExistentFile, validFile)
	if err == nil {
		t.Error("Expected error for non-existent file1, got nil")
	}

	// Test with non-existent file2
	_, err = CompareExcelFiles(validFile, nonExistentFile)
	if err == nil {
		t.Error("Expected error for non-existent file2, got nil")
	}

	// Test with invalid Excel file
	invalidFile := filepath.Join(tmpDir, "invalid.xlsx")
	if err := os.WriteFile(invalidFile, []byte("not an excel file"), 0600); err != nil {
		t.Fatalf("Failed to create invalid test file: %v", err)
	}

	_, err = CompareExcelFiles(invalidFile, validFile)
	if err == nil {
		t.Error("Expected error for invalid Excel file, got nil")
	}
}

func TestFormatComparisonResult(t *testing.T) {
	result := &ExcelComparisonResult{
		OldFilePath: "/path/to/file1.xlsx",
		NewFilePath: "/path/to/file2.xlsx",
		SheetComparisons: []SheetComparison{
			{
				SheetName:     "Recommendations",
				OldRowCount:   10,
				NewRowCount:   12,
				RowDifference: 2,
				HasDuplicates: false,
				DuplicateRows: []int{},
			},
			{
				SheetName:     "Inventory",
				OldRowCount:   5,
				NewRowCount:   5,
				RowDifference: 0,
				HasDuplicates: true,
				DuplicateRows: []int{5, 7},
			},
		},
	}

	output := FormatComparisonResult(result)

	// Check that output contains expected information
	expectedStrings := []string{
		"Comparison Results:",
		"Old File: /path/to/file1.xlsx",
		"New File: /path/to/file2.xlsx",
		"Sheet Name",
		"Old Rows",
		"New Rows",
		"Difference",
		"Has Duplicates",
		"Recommendations",
		"Inventory",
		"10",
		"12",
		"+2",
		"Duplicate rows: [5 7]",
	}

	for _, expected := range expectedStrings {
		if !containsString(output, expected) {
			t.Errorf("Expected output to contain %q, but it doesn't. Output:\n%s", expected, output)
		}
	}
}

func TestRowToString(t *testing.T) {
	tests := []struct {
		name     string
		row      []string
		expected string
	}{
		{
			name:     "simple row",
			row:      []string{"A", "B", "C"},
			expected: "A|B|C",
		},
		{
			name:     "empty row",
			row:      []string{},
			expected: "",
		},
		{
			name:     "single cell",
			row:      []string{"single"},
			expected: "single",
		},
		{
			name:     "row with empty cells",
			row:      []string{"A", "", "C"},
			expected: "A||C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rowToString(tt.row)
			if result != tt.expected {
				t.Errorf("rowToString(%v) = %q, expected %q", tt.row, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		val      int
		expected bool
	}{
		{
			name:     "contains value",
			slice:    []int{1, 2, 3, 4, 5},
			val:      3,
			expected: true,
		},
		{
			name:     "does not contain value",
			slice:    []int{1, 2, 3, 4, 5},
			val:      6,
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []int{},
			val:      1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.val)
			if result != tt.expected {
				t.Errorf("contains(%v, %d) = %t, expected %t", tt.slice, tt.val, result, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
