// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package viewer

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExcelToDataStore parses an Excel file and converts it to the same DataStore format as JSON files.
func ExcelToDataStore(path string) (*DataStore, error) {
	// Clean and normalize the path for cross-platform compatibility
	cleanPath := filepath.Clean(path)
	f, err := excelize.OpenFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("open excel file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	ds := &DataStore{Data: map[string][]map[string]string{}}

	// Get all sheet names
	sheets := f.GetSheetList()

	for _, sheetName := range sheets {
		// Map sheet names to dataset names (matching azqr output conventions)
		datasetName := mapSheetToDataset(sheetName)
		if datasetName == "" {
			continue // Skip unknown sheets
		}

		// Get all rows from the sheet
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue // Skip sheets with read errors
		}

		if len(rows) < 2 {
			continue // Skip sheets with no data (need at least header + 1 row)
		}

		// Find the first non-empty row as headers (Excel files may have empty rows at the top)
		var headers []string
		headerRowIndex := -1
		for i, row := range rows {
			if len(row) > 0 {
				// Check if this row has non-empty content
				hasContent := false
				for _, cell := range row {
					if strings.TrimSpace(cell) != "" {
						hasContent = true
						break
					}
				}
				if hasContent {
					headers = row
					headerRowIndex = i
					break
				}
			}
		}

		if headerRowIndex == -1 || len(headers) == 0 {
			continue // No valid headers found
		}

		// Process data rows (skip header row and any empty rows before it)
		records := make([]map[string]string, 0, len(rows)-headerRowIndex-1)
		for i := headerRowIndex + 1; i < len(rows); i++ {
			row := rows[i]
			record := map[string]string{}

			// Map each cell to its header
			for j, header := range headers {
				if header == "" {
					continue // Skip empty headers
				}

				var cellValue string
				if j < len(row) {
					cellValue = strings.TrimSpace(row[j])
				}

				// Clean up header name (remove spaces, make consistent with JSON output)
				cleanHeader := cleanHeaderName(header)
				record[cleanHeader] = cellValue
			}

			// Only add non-empty records
			if len(record) > 0 {
				records = append(records, record)
			}
		}

		if len(records) > 0 {
			ds.Data[datasetName] = records
		}
	}

	return ds, nil
}

// mapSheetToDataset maps Excel sheet names to dataset names used in JSON format.
func mapSheetToDataset(sheetName string) string {
	// Clean the sheet name for comparison
	clean := strings.ToLower(strings.TrimSpace(sheetName))

	// Map common sheet names to dataset constants
	mapping := map[string]string{
		"recommendations":          DataSetRecommendations,
		"recommendation":           DataSetRecommendations,
		"impacted resources":       DataSetImpacted,
		"impactedresources":        DataSetImpacted,
		"impacted":                 DataSetImpacted,
		"resource types":           DataSetResourceType,
		"resource type":            DataSetResourceType,
		"resourcetype":             DataSetResourceType,
		"resourcetypes":            DataSetResourceType,
		"inventory":                DataSetInventory,
		"resource inventory":       DataSetInventory,
		"advisor":                  DataSetAdvisor,
		"azure advisor":            DataSetAdvisor,
		"policy":                   DataSetAzurePolicy,
		"azure policy":             DataSetAzurePolicy,
		"azurepolicy":              DataSetAzurePolicy,
		"policies":                 DataSetAzurePolicy,
		"arc sql":                  DataSetArcSQL,
		"arcsql":                   DataSetArcSQL,
		"defender":                 DataSetDefender,
		"microsoft defender":       DataSetDefender,
		"defender for cloud":       DataSetDefender,
		"defender recommendations": DataSetDefenderRecommendations,
		"defenderrecommendations":  DataSetDefenderRecommendations,
		"costs":                    DataSetCosts,
		"cost":                     DataSetCosts,
		"cost analysis":            DataSetCosts,
		"out of scope":             DataSetOutOfScope,
		"outofscope":               DataSetOutOfScope,
		"out-of-scope":             DataSetOutOfScope,
	}

	if dataset, exists := mapping[clean]; exists {
		return dataset
	}

	// Try partial matches for flexibility
	for key, dataset := range mapping {
		if strings.Contains(clean, key) || strings.Contains(key, clean) {
			return dataset
		}
	}

	return "" // Unknown sheet
}

// cleanHeaderName cleans up Excel header names to match JSON field names.
func cleanHeaderName(header string) string {
	// Remove extra spaces and convert to lowercase for processing
	clean := strings.TrimSpace(header)

	// Common header mappings to match JSON field names
	mappings := map[string]string{
		"Subscription ID":                  "subscriptionId",
		"Subscription Id":                  "subscriptionId",
		"Subscription Name":                "subscriptionName",
		"Resource Group":                   "resourceGroup",
		"Resource Type":                    "resourceType",
		"Resource Name":                    "resourceName",
		"Resource ID":                      "resourceId",
		"Resource Id":                      "resourceId",
		"Recommendation ID":                "recommendationId",
		"Recommendation Id":                "recommendationId",
		"Recommendation":                   "recommendation",
		"Category":                         "category",
		"Impact":                           "impact",
		"Implemented":                      "implemented",
		"Learn More":                       "learn",
		"Read More":                        "readMore",
		"Policy Display Name":              "policyDisplayName",
		"Policy Description":               "policyDescription",
		"Compliance State":                 "complianceState",
		"Time Stamp":                       "timeStamp",
		"Policy Definition Name":           "policyDefinitionName",
		"Policy Definition ID":             "policyDefinitionId",
		"Policy Assignment Name":           "policyAssignmentName",
		"Policy Assignment ID":             "policyAssignmentId",
		"Recommendation Severity":          "recommendationSeverity",
		"Recommendation Name":              "recommendationName",
		"Action Description":               "actionDescription",
		"Remediation Description":          "remediationDescription",
		"Azure Portal Link":                "azPortalLink",
		"SKU Name":                         "skuName",
		"SKU Tier":                         "skuTier",
		"Kind":                             "kind",
		"SLA":                              "sla",
		"Location":                         "location",
		"Description":                      "description",
		"Number of Impacted Resources":     "numberOfImpactedResources",
		"Number of Resources":              "numberOfResources",
		"Azure Service Well Architected":   "azureServiceWellArchitected",
		"Azure Service / Well-Architected": "azureServiceWellArchitected",
		"Azure Service Category / Well-Architected Area": "azureServiceCategoryWellArchitectedArea",
		"Azure Service / Well-Architected Topic":         "azureServiceWellArchitectedTopic",
		"Recommendation Source":                          "recommendationSource",
		"Best Practices Guidance":                        "bestPracticesGuidance",
		"Available in APRL":                              "availableInAprl",
		"Validated Using":                                "validatedUsing",
		"Source":                                         "source",
		"Service Name":                                   "serviceName",
		"Value":                                          "value",
		"Currency":                                       "currency",
		"From":                                           "from",
		"To":                                             "to",
		"Name":                                           "name",
		"Tier":                                           "tier",
		"Machine Name":                                   "machineName",
		"Machine Id":                                     "machineId",
		"Machine ID":                                     "machineId",
		"Tags":                                           "tags",
		"Status":                                         "status",
		"Provisioning State":                             "provisioningState",
		"License Type":                                   "licenseType",
		"ESU":                                            "esu",
		"Extension Version":                              "extensionVersion",
		"Excluded Instances":                             "excludedInstances",
		"Purview":                                        "purview",
		"Entra ID":                                       "entraId",
		"BPA":                                            "bpa",
		"Azure Arc Server":                               "azureArcServer",
		"SQL Instance":                                   "sqlInstance",
		"Version":                                        "version",
		"Build":                                          "build",
		"Patch Level":                                    "patchLevel",
		"Edition":                                        "edition",
		"VCores":                                         "vCores",
		"DPS Status":                                     "dpsStatus",
		"License":                                        "license",
		"TEL Status":                                     "telStatus",
		"Defender Status":                                "defenderStatus",
	}

	// Check for exact matches first
	if mapped, exists := mappings[clean]; exists {
		return mapped
	}

	// If no exact match, try to convert common patterns
	// Remove special characters and convert to camelCase
	result := clean
	result = strings.ReplaceAll(result, " ", "")
	result = strings.ReplaceAll(result, "-", "")
	result = strings.ReplaceAll(result, "_", "")

	// Convert to camelCase
	if len(result) > 0 {
		result = strings.ToLower(string(result[0])) + result[1:]
	}

	// Handle some common cases
	result = strings.ReplaceAll(result, "Id", "ID")
	result = strings.ReplaceAll(result, "id", "ID")

	// Fix double ID
	result = strings.ReplaceAll(result, "IDID", "ID")

	return result
}
