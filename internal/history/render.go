// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package history

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/Azure/azqr/internal/models"
)

// Render formats selected history records as table or JSON.
func Render(records []Record, scopeID, format string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		output := struct {
			ScopeID string   `json:"scopeId"`
			Records []Record `json:"records"`
		}{ScopeID: scopeID, Records: records}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal trend output: %w", err)
		}
		return string(data), nil
	case "table", "":
		return renderTable(records, scopeID), nil
	default:
		return "", fmt.Errorf("unsupported trend format %q: supported values are table and json", format)
	}
}

func renderTable(records []Record, scopeID string) string {
	var output strings.Builder
	fmt.Fprintf(&output, "Scope: %s\n", scopeID)
	if len(records) == 0 {
		output.WriteString("No history records found.\n")
		return output.String()
	}

	writer := tabwriter.NewWriter(&output, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(writer, "TIME\tRESOURCES\tFINDINGS\tHIGH\tMEDIUM\tLOW\tTREND")
	maxFindings := 1
	for _, record := range records {
		if record.ImpactedResources > maxFindings {
			maxFindings = record.ImpactedResources
		}
	}
	for _, record := range records {
		_, _ = fmt.Fprintf(
			writer,
			"%s\t%d\t%d\t%d\t%d\t%d\t%s\n",
			record.Timestamp.Format("2006-01-02 15:04"),
			record.Resources,
			record.ImpactedResources,
			record.ImpactedByImpact[string(models.ImpactHigh)],
			record.ImpactedByImpact[string(models.ImpactMedium)],
			record.ImpactedByImpact[string(models.ImpactLow)],
			trendBar(record.ImpactedResources, maxFindings),
		)
	}
	_ = writer.Flush()
	return output.String()
}

func trendBar(value, maximum int) string {
	if value == 0 {
		return "-"
	}
	width := max(1, value*12/maximum)
	return strings.Repeat("#", width)
}
