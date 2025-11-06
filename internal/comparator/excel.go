// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package comparator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

// SheetComparison represents the comparison result for a single sheet
type SheetComparison struct {
	SheetName     string
	OldRowCount   int
	NewRowCount   int
	RowDifference int
	HasDuplicates bool
	DuplicateRows []int // Row numbers (1-indexed) that are duplicates
}

// RecommendationChange represents a change in a recommendation row
type RecommendationChange struct {
	RecommendationID      string
	Recommendation        string
	Category              string
	Impact                string
	ResourceType          string
	OldImpactedResources  string
	NewImpactedResources  string
	ImpactedResourcesDiff int
	ChangeType            string // "added", "removed", "changed", "unchanged"
}

// ExcelComparisonResult represents the complete comparison result
type ExcelComparisonResult struct {
	OldFilePath           string
	NewFilePath           string
	OldFileLabel          string // Label for older file
	NewFileLabel          string // Label for newer file
	SheetComparisons      []SheetComparison
	RecommendationChanges []RecommendationChange
}

// CompareExcelFiles compares two Excel files and returns row counts and duplicate detection
// Determines which file is older/newer based on modification time
func CompareExcelFiles(file1Path, file2Path string) (*ExcelComparisonResult, error) {
	// Clean and normalize paths
	cleanPath1 := filepath.Clean(file1Path)
	cleanPath2 := filepath.Clean(file2Path)

	// Get file modification times to determine old vs new
	stat1, err := os.Stat(cleanPath1)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file1 (%s): %w", cleanPath1, err)
	}
	stat2, err := os.Stat(cleanPath2)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file2 (%s): %w", cleanPath2, err)
	}

	// Determine which file is older
	var oldPath, newPath string
	var oldFile, newFile *excelize.File
	if stat1.ModTime().Before(stat2.ModTime()) {
		oldPath = cleanPath1
		newPath = cleanPath2
	} else {
		oldPath = cleanPath2
		newPath = cleanPath1
	}

	// Open old file
	oldFile, err = excelize.OpenFile(oldPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open old file (%s): %w", oldPath, err)
	}
	defer func() {
		_ = oldFile.Close()
	}()

	// Open new file
	newFile, err = excelize.OpenFile(newPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open new file (%s): %w", newPath, err)
	}
	defer func() {
		_ = newFile.Close()
	}()

	result := &ExcelComparisonResult{
		OldFilePath:      oldPath,
		NewFilePath:      newPath,
		OldFileLabel:     "Old",
		NewFileLabel:     "New",
		SheetComparisons: []SheetComparison{},
	}

	// Get all sheets from both files
	sheetsOld := oldFile.GetSheetList()
	sheetsNew := newFile.GetSheetList()

	// Create a map of all unique sheet names
	sheetMap := make(map[string]bool)
	for _, sheet := range sheetsOld {
		sheetMap[sheet] = true
	}
	for _, sheet := range sheetsNew {
		sheetMap[sheet] = true
	}

	// Compare each sheet
	for sheetName := range sheetMap {
		comparison := SheetComparison{
			SheetName:     sheetName,
			DuplicateRows: []int{},
		}

		// Get row count from old file
		rowsOld, err := oldFile.GetRows(sheetName)
		if err == nil {
			comparison.OldRowCount = len(rowsOld)
		} else {
			comparison.OldRowCount = 0
		}

		// Get row count from new file
		rowsNew, err := newFile.GetRows(sheetName)
		if err == nil {
			comparison.NewRowCount = len(rowsNew)
		} else {
			comparison.NewRowCount = 0
		}

		// Calculate difference
		comparison.RowDifference = comparison.NewRowCount - comparison.OldRowCount

		// Check for duplicates after row 4 in new file (if it exists)
		if comparison.NewRowCount > 0 {
			comparison.HasDuplicates, comparison.DuplicateRows = findDuplicatesAfterRow4(rowsNew)
		}

		result.SheetComparisons = append(result.SheetComparisons, comparison)
	}

	// Compare Recommendations sheet in detail
	result.RecommendationChanges = compareRecommendations(oldFile, newFile)

	return result, nil
}

// findDuplicatesAfterRow4 checks for duplicate rows starting from row 5 (index 4)
// Returns true if duplicates found and a list of duplicate row numbers
func findDuplicatesAfterRow4(rows [][]string) (bool, []int) {
	if len(rows) <= 4 {
		return false, []int{}
	}

	// Create a map to track seen rows (starting from row 5, which is index 4)
	seenRows := make(map[string]int)
	duplicateRows := []int{}

	for i := 4; i < len(rows); i++ {
		// Convert row to a string representation
		rowKey := rowToString(rows[i])

		// Check if we've seen this row before
		if firstOccurrence, exists := seenRows[rowKey]; exists {
			// This is a duplicate - record both the first occurrence and current row
			if !contains(duplicateRows, firstOccurrence+1) {
				duplicateRows = append(duplicateRows, firstOccurrence+1)
			}
			duplicateRows = append(duplicateRows, i+1)
		} else {
			// First time seeing this row
			seenRows[rowKey] = i
		}
	}

	return len(duplicateRows) > 0, duplicateRows
}

// rowToString converts a row slice to a string for comparison
func rowToString(row []string) string {
	result := ""
	for i, cell := range row {
		if i > 0 {
			result += "|"
		}
		result += cell
	}
	return result
}

// contains checks if a slice contains a value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// compareRecommendations compares the Recommendations sheet in detail
func compareRecommendations(f1, f2 *excelize.File) []RecommendationChange {
	changes := []RecommendationChange{}

	// Get rows from Recommendations sheet in both files
	rows1, err1 := f1.GetRows("Recommendations")
	rows2, err2 := f2.GetRows("Recommendations")

	// If either file doesn't have Recommendations sheet, return empty
	if err1 != nil || err2 != nil {
		return changes
	}

	// Skip to data rows (row 5+, indices 4+)
	if len(rows1) <= 4 || len(rows2) <= 4 {
		return changes
	}

	// Get headers from row 4 (index 3)
	headers := rows1[3]
	if len(headers) == 0 {
		return changes
	}

	// Find column indices for key fields
	recIDCol := findColumnIndex(headers, "Recommendation Id")
	recCol := findColumnIndex(headers, "Recommendation")
	categoryCol := findColumnIndex(headers, "Category")
	impactCol := findColumnIndex(headers, "Impact")
	resourceTypeCol := findColumnIndex(headers, "Azure Service Category")
	impactedCol := findColumnIndex(headers, "Number of Impacted Resources")

	if recIDCol == -1 || impactedCol == -1 {
		return changes
	}

	// Build maps of recommendations from both files
	recs1 := buildRecommendationMap(rows1[4:], recIDCol, recCol, categoryCol, impactCol, resourceTypeCol, impactedCol)
	recs2 := buildRecommendationMap(rows2[4:], recIDCol, recCol, categoryCol, impactCol, resourceTypeCol, impactedCol)

	// Track processed recommendations
	processed := make(map[string]bool)

	// Find changes and additions
	for id, rec2 := range recs2 {
		processed[id] = true
		if rec1, exists := recs1[id]; exists {
			// Check if impacted resources changed
			if rec1.impactedResources != rec2.impactedResources {
				diff := parseInt(rec2.impactedResources) - parseInt(rec1.impactedResources)
				changes = append(changes, RecommendationChange{
					RecommendationID:      id,
					Recommendation:        rec2.recommendation,
					Category:              rec2.category,
					Impact:                rec2.impact,
					ResourceType:          rec2.resourceType,
					OldImpactedResources:  rec1.impactedResources,
					NewImpactedResources:  rec2.impactedResources,
					ImpactedResourcesDiff: diff,
					ChangeType:            "changed",
				})
			}
		} else {
			// New recommendation in file2
			changes = append(changes, RecommendationChange{
				RecommendationID:      id,
				Recommendation:        rec2.recommendation,
				Category:              rec2.category,
				Impact:                rec2.impact,
				ResourceType:          rec2.resourceType,
				OldImpactedResources:  "0",
				NewImpactedResources:  rec2.impactedResources,
				ImpactedResourcesDiff: parseInt(rec2.impactedResources),
				ChangeType:            "added",
			})
		}
	}

	// Find removals
	for id, rec1 := range recs1 {
		if !processed[id] {
			changes = append(changes, RecommendationChange{
				RecommendationID:      id,
				Recommendation:        rec1.recommendation,
				Category:              rec1.category,
				Impact:                rec1.impact,
				ResourceType:          rec1.resourceType,
				OldImpactedResources:  rec1.impactedResources,
				NewImpactedResources:  "0",
				ImpactedResourcesDiff: -parseInt(rec1.impactedResources),
				ChangeType:            "removed",
			})
		}
	}

	return changes
}

// recommendationData holds recommendation information
type recommendationData struct {
	recommendation    string
	category          string
	impact            string
	resourceType      string
	impactedResources string
}

// buildRecommendationMap builds a map of recommendations from rows
func buildRecommendationMap(rows [][]string, recIDCol, recCol, categoryCol, impactCol, resourceTypeCol, impactedCol int) map[string]recommendationData {
	recMap := make(map[string]recommendationData)

	for _, row := range rows {
		if len(row) <= recIDCol {
			continue
		}

		id := getCell(row, recIDCol)
		if id == "" {
			continue
		}

		recMap[id] = recommendationData{
			recommendation:    getCell(row, recCol),
			category:          getCell(row, categoryCol),
			impact:            getCell(row, impactCol),
			resourceType:      getCell(row, resourceTypeCol),
			impactedResources: getCell(row, impactedCol),
		}
	}

	return recMap
}

// findColumnIndex finds the index of a column by header name (case-insensitive exact match first, then partial)
func findColumnIndex(headers []string, name string) int {
	lowerName := toLower(name)

	// First pass: exact match
	for i, header := range headers {
		if toLower(header) == lowerName {
			return i
		}
	}

	// Second pass: partial match (only if exact match not found)
	// But avoid matching substrings like "Recommendation" matching "Recommendation Source"
	for i, header := range headers {
		lowerHeader := toLower(header)
		// Only match if the search term is at the start or is the whole header
		if lowerHeader == lowerName || (len(lowerHeader) >= len(lowerName) && lowerHeader[:len(lowerName)] == lowerName) {
			return i
		}
	}

	return -1
}

// getCell safely gets a cell value from a row
func getCell(row []string, col int) string {
	if col >= 0 && col < len(row) {
		return row[col]
	}
	return ""
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}

// parseInt converts string to int, returns 0 if invalid
func parseInt(s string) int {
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		}
	}
	return result
}

// FormatComparisonResult formats the comparison result as a human-readable string
func FormatComparisonResult(result *ExcelComparisonResult) string {
	output := "Comparison Results:\n"
	output += fmt.Sprintf("Old File: %s\n", result.OldFilePath)
	output += fmt.Sprintf("New File: %s\n\n", result.NewFilePath)

	output += fmt.Sprintf("%-40s | %10s | %10s | %10s | %s\n",
		"Sheet Name", "Old Rows", "New Rows", "Difference", "Has Duplicates")
	output += fmt.Sprintf("%s\n", "-----------------------------------------------------------------------------------------------------------")

	for _, comparison := range result.SheetComparisons {
		duplicateStatus := "No"
		if comparison.HasDuplicates {
			duplicateStatus = fmt.Sprintf("Yes (%d)", len(comparison.DuplicateRows))
		}

		diffStr := fmt.Sprintf("%+d", comparison.RowDifference)
		output += fmt.Sprintf("%-40s | %10d | %10d | %10s | %s\n",
			comparison.SheetName,
			comparison.OldRowCount,
			comparison.NewRowCount,
			diffStr,
			duplicateStatus)

		// Show duplicate row numbers if present
		if comparison.HasDuplicates && len(comparison.DuplicateRows) > 0 {
			output += fmt.Sprintf("  Duplicate rows: %v\n", comparison.DuplicateRows)
		}
	}

	// Add Recommendations diff section if available
	if len(result.RecommendationChanges) > 0 {
		output += "\n" + formatRecommendationChanges(result.RecommendationChanges)
	}

	return output
}

// formatRecommendationChanges formats recommendation changes as a table
func formatRecommendationChanges(changes []RecommendationChange) string {
	output := "\nRecommendations Detailed Diff:\n"
	output += "==============================\n\n"

	// Group changes by type
	added := []RecommendationChange{}
	removed := []RecommendationChange{}
	changed := []RecommendationChange{}

	for _, change := range changes {
		switch change.ChangeType {
		case "added":
			added = append(added, change)
		case "removed":
			removed = append(removed, change)
		case "changed":
			changed = append(changed, change)
		}
	}

	// Format changed recommendations as table
	if len(changed) > 0 {
		output += fmt.Sprintf("Changed Recommendations (%d):\n", len(changed))
		output += fmt.Sprintf("%-20s | %-20s | %-10s | %-8s | %-50s | %10s | %10s | %10s\n",
			"Resource Type", "Category", "Impact", "Change", "Recommendation", "Old", "New", "Diff")
		output += "-------------------------------------------------------------------------------------------------------------------------------------------------------------\n"

		for _, change := range changed {
			diffStr := fmt.Sprintf("%+d", change.ImpactedResourcesDiff)
			resType := change.ResourceType
			if resType == "" {
				resType = "N/A"
			}
			output += fmt.Sprintf("%-20s | %-20s | %-10s | %-8s | %-50s | %10s | %10s | %10s\n",
				truncate(resType, 20),
				truncate(change.Category, 20),
				truncate(change.Impact, 10),
				"Changed",
				truncate(change.Recommendation, 50),
				change.OldImpactedResources,
				change.NewImpactedResources,
				diffStr)
		}
		output += "\n"
	}

	// Format added recommendations as table
	if len(added) > 0 {
		output += fmt.Sprintf("Added Recommendations (%d):\n", len(added))
		output += fmt.Sprintf("%-20s | %-20s | %-10s | %-8s | %-50s | %10s\n",
			"Resource Type", "Category", "Impact", "Change", "Recommendation", "Impacted")
		output += "---------------------------------------------------------------------------------------------------------------------------------------\n"

		for _, change := range added {
			resType := change.ResourceType
			if resType == "" {
				resType = "N/A"
			}
			output += fmt.Sprintf("%-20s | %-20s | %-10s | %-8s | %-50s | %10s\n",
				truncate(resType, 20),
				truncate(change.Category, 20),
				truncate(change.Impact, 10),
				"Added",
				truncate(change.Recommendation, 50),
				change.NewImpactedResources)
		}
		output += "\n"
	}

	// Format removed recommendations as table
	if len(removed) > 0 {
		output += fmt.Sprintf("Removed Recommendations (%d):\n", len(removed))
		output += fmt.Sprintf("%-20s | %-20s | %-10s | %-8s | %-50s | %10s\n",
			"Resource Type", "Category", "Impact", "Change", "Recommendation", "Impacted")
		output += "---------------------------------------------------------------------------------------------------------------------------------------\n"

		for _, change := range removed {
			resType := change.ResourceType
			if resType == "" {
				resType = "N/A"
			}
			output += fmt.Sprintf("%-20s | %-20s | %-10s | %-8s | %-50s | %10s\n",
				truncate(resType, 20),
				truncate(change.Category, 20),
				truncate(change.Impact, 10),
				"Removed",
				truncate(change.Recommendation, 50),
				change.OldImpactedResources)
		}
	}

	// Summary
	output += fmt.Sprintf("Summary: %d changed, %d added, %d removed\n",
		len(changed), len(added), len(removed))

	return output
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
