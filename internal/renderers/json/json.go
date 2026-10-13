// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package json

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"os"

	"github.com/Azure/azqr/internal/models"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/iancoleman/strcase"

	"github.com/rs/zerolog/log"
)

// buildConsolidatedReport builds the consolidated JSON structure from report data
func buildConsolidatedReport(data *renderers.ReportData) map[string]interface{} {
	consolidatedReport := map[string]interface{}{}

	// Only include AZQR-related data if the feature is enabled
	if data.Stages.IsStageEnabled(models.StageNameGraph) {
		consolidatedReport["recommendations"] = convertToJSON(data.RecommendationsTable())
		consolidatedReport["impacted"] = convertToJSON(data.ImpactedTable())
		consolidatedReport["resourceType"] = convertToJSON(data.ResourceTypesTable())
		consolidatedReport["inventory"] = convertToJSON(data.ResourcesTable())
		consolidatedReport["outOfScope"] = convertToJSON(data.ExcludedResourcesTable())
	} else {
		log.Debug().Msg("Skipping AZQR data in JSON. Feature is disabled")
	}

	// Only include Advisor data if the feature is enabled
	if data.Stages.IsStageEnabled(models.StageNameAdvisor) {
		consolidatedReport["advisor"] = convertToJSON(data.AdvisorTable())
	} else {
		log.Debug().Msg("Skipping Advisor data in JSON. Feature is disabled")
	}

	// Only include Azure Policy data if the feature is enabled
	if data.Stages.IsStageEnabled(models.StageNamePolicy) {
		consolidatedReport["azurePolicy"] = convertToJSON(data.AzurePolicyTable())
	} else {
		log.Debug().Msg("Skipping Azure Policy data in JSON. Feature is disabled")
	}

	// Only include Arc SQL data if the feature is enabled
	if data.Stages.IsStageEnabled(models.StageNameArc) {
		consolidatedReport["arcSQL"] = convertToJSON(data.ArcSQLTable())
	} else {
		log.Debug().Msg("Skipping Arc SQL data in JSON. Feature is disabled")
	}

	// Only include Defender data if the feature is enabled
	if data.Stages.IsStageEnabled(models.StageNameDefender) {
		consolidatedReport["defender"] = convertToJSON(data.DefenderTable())
	} else {
		log.Debug().Msg("Skipping Defender data in JSON. Feature is disabled")
	}

	// Only include Defender Recommendations data if the feature is enabled
	if data.Stages.IsStageEnabled(models.StageNameDefenderRecommendations) {
		consolidatedReport["defenderRecommendations"] = convertToJSON(data.DefenderRecommendationsTable())
	} else {
		log.Debug().Msg("Skipping Defender Recommendations data in JSON. Feature is disabled")
	}

	// Only include Cost data if the feature is enabled
	if data.Stages.IsStageEnabled(models.StageNameCost) {
		consolidatedReport["costs"] = convertToJSON(data.CostTable())
	} else {
		log.Debug().Msg("Skipping Cost data in JSON. Feature is disabled")
	}

	// Add external plugin results
	if len(data.PluginResults) > 0 {
		// Use a slice so that plugins returning multiple sheets (e.g. region-selection)
		// are all preserved. A map keyed by PluginName would silently overwrite
		// earlier sheets when multiple results share the same plugin name.
		var pluginsList []interface{}
		for _, result := range data.PluginResults {
			dataKey := strcase.ToLowerCamel(result.SheetName)
			pluginData := map[string]interface{}{
				"pluginName":  result.PluginName,
				"sheetName":   result.SheetName,
				"description": result.Description,
				dataKey:       convertToJSON(result.Table),
			}
			pluginsList = append(pluginsList, pluginData)
		}
		consolidatedReport["externalPlugins"] = pluginsList
	}

	return consolidatedReport
}

// CreateJsonReport generates a single consolidated JSON report file, writing each top-level
// section as it is computed and encoded rather than first assembling the whole consolidated
// report as one in-memory map (see streamConsolidatedReport): peak extra memory for this path
// is bounded to one section's converted data at a time, not the full report.
func CreateJsonReport(data *renderers.ReportData) {
	filename := fmt.Sprintf("%s.json", data.OutputFileName)
	log.Info().Msgf("Generating Report: %s", filename)

	f, err := os.Create(filename) //nolint:gosec // filename is generated from user's output-name flag
	if err != nil {
		log.Fatal().Err(err).Msg("error creating json:")
		return
	}
	defer func() {
		// Handle error during file close
		if cerr := f.Close(); cerr != nil {
			log.Error().Err(cerr).Msg("error closing file:")
		}
	}()

	if err := streamConsolidatedReport(f, data); err != nil {
		log.Error().Err(err).Msg("error writing json:")
	}
}

// CreateJsonOutput generates the same consolidated JSON structure as CreateJsonReport
// but returns it as a string for console output instead of writing to a file
func CreateJsonOutput(data *renderers.ReportData) string {
	consolidatedReport := buildConsolidatedReport(data)

	js, err := json.MarshalIndent(consolidatedReport, "", "\t")
	if err != nil {
		log.Fatal().Err(err).Msg("error marshaling data:")
	}

	return string(js)
}

// jsonSection pairs a consolidated-report top-level key with a lazily-computed value. The
// value is produced only when the section is actually written, and goes out of scope (eligible
// for garbage collection) immediately afterward, rather than every section's converted data
// being held in one map for the lifetime of the whole write.
type jsonSection struct {
	key   string
	value func() interface{}
}

// streamConsolidatedReport writes the same logical JSON document buildConsolidatedReport would
// produce (same keys, same values, same nesting, same "\t"-indented formatting), but encodes
// and writes each top-level section to w as it is computed instead of first building one
// in-memory map of every section's converted data. Sections are written in the same
// lexicographic key order encoding/json uses when marshaling a map, so output is identical to
// the prior full-buffer implementation for any given ReportData.
func streamConsolidatedReport(w io.Writer, data *renderers.ReportData) error {
	bw := bufio.NewWriter(w)

	sections := consolidatedReportSections(data)

	if _, err := bw.WriteString("{\n"); err != nil {
		return err
	}
	for i, section := range sections {
		value := section.value()
		valueJSON, err := json.MarshalIndent(value, "\t", "\t")
		if err != nil {
			return fmt.Errorf("marshal section %q: %w", section.key, err)
		}
		keyJSON, err := json.Marshal(section.key)
		if err != nil {
			return fmt.Errorf("marshal section key %q: %w", section.key, err)
		}

		if _, err := bw.WriteString("\t"); err != nil {
			return err
		}
		if _, err := bw.Write(keyJSON); err != nil {
			return err
		}
		if _, err := bw.WriteString(": "); err != nil {
			return err
		}
		if _, err := bw.Write(valueJSON); err != nil {
			return err
		}
		if i != len(sections)-1 {
			if _, err := bw.WriteString(","); err != nil {
				return err
			}
		}
		if _, err := bw.WriteString("\n"); err != nil {
			return err
		}
		// value and valueJSON are not referenced again after this iteration and may be
		// collected before the next section is computed.
	}
	if _, err := bw.WriteString("}"); err != nil {
		return err
	}

	return bw.Flush()
}

// consolidatedReportSections returns the same top-level sections buildConsolidatedReport would
// include, in the lexicographic key order encoding/json uses when marshaling a map (so
// streamConsolidatedReport's output matches buildConsolidatedReport's previous output byte for
// byte). Each section's value is wrapped in a closure so it is computed only when written.
func consolidatedReportSections(data *renderers.ReportData) []jsonSection {
	var sections []jsonSection

	if data.Stages.IsStageEnabled(models.StageNameAdvisor) {
		sections = append(sections, jsonSection{"advisor", func() interface{} { return convertToJSON(data.AdvisorTable()) }})
	} else {
		log.Debug().Msg("Skipping Advisor data in JSON. Feature is disabled")
	}

	if data.Stages.IsStageEnabled(models.StageNameArc) {
		sections = append(sections, jsonSection{"arcSQL", func() interface{} { return convertToJSON(data.ArcSQLTable()) }})
	} else {
		log.Debug().Msg("Skipping Arc SQL data in JSON. Feature is disabled")
	}

	if data.Stages.IsStageEnabled(models.StageNamePolicy) {
		sections = append(sections, jsonSection{"azurePolicy", func() interface{} { return convertToJSON(data.AzurePolicyTable()) }})
	} else {
		log.Debug().Msg("Skipping Azure Policy data in JSON. Feature is disabled")
	}

	if data.Stages.IsStageEnabled(models.StageNameCost) {
		sections = append(sections, jsonSection{"costs", func() interface{} { return convertToJSON(data.CostTable()) }})
	} else {
		log.Debug().Msg("Skipping Cost data in JSON. Feature is disabled")
	}

	if data.Stages.IsStageEnabled(models.StageNameDefender) {
		sections = append(sections, jsonSection{"defender", func() interface{} { return convertToJSON(data.DefenderTable()) }})
	} else {
		log.Debug().Msg("Skipping Defender data in JSON. Feature is disabled")
	}

	if data.Stages.IsStageEnabled(models.StageNameDefenderRecommendations) {
		sections = append(sections, jsonSection{"defenderRecommendations", func() interface{} { return convertToJSON(data.DefenderRecommendationsTable()) }})
	} else {
		log.Debug().Msg("Skipping Defender Recommendations data in JSON. Feature is disabled")
	}

	if len(data.PluginResults) > 0 {
		sections = append(sections, jsonSection{"externalPlugins", func() interface{} { return buildExternalPluginsSection(data) }})
	}

	if data.Stages.IsStageEnabled(models.StageNameGraph) {
		sections = append(sections,
			jsonSection{"impacted", func() interface{} { return convertToJSON(data.ImpactedTable()) }},
			jsonSection{"inventory", func() interface{} { return convertToJSON(data.ResourcesTable()) }},
			jsonSection{"outOfScope", func() interface{} { return convertToJSON(data.ExcludedResourcesTable()) }},
			jsonSection{"recommendations", func() interface{} { return convertToJSON(data.RecommendationsTable()) }},
			jsonSection{"resourceType", func() interface{} { return convertToJSON(data.ResourceTypesTable()) }},
		)
	} else {
		log.Debug().Msg("Skipping AZQR data in JSON. Feature is disabled")
	}

	sort.Slice(sections, func(i, j int) bool { return sections[i].key < sections[j].key })
	return sections
}

// buildExternalPluginsSection builds the "externalPlugins" section's value exactly as
// buildConsolidatedReport does. See the comment on that field there for why a slice (rather
// than a map keyed by PluginName) is required.
func buildExternalPluginsSection(data *renderers.ReportData) []interface{} {
	var pluginsList []interface{}
	for _, result := range data.PluginResults {
		dataKey := strcase.ToLowerCamel(result.SheetName)
		pluginData := map[string]interface{}{
			"pluginName":  result.PluginName,
			"sheetName":   result.SheetName,
			"description": result.Description,
			dataKey:       convertToJSON(result.Table),
		}
		pluginsList = append(pluginsList, pluginData)
	}
	return pluginsList
}

func convertToJSON(data [][]string) []map[string]string {
	var result []map[string]string
	headers := data[0]

	for _, row := range data[1:] {
		item := make(map[string]string)
		for i, value := range row {
			item[strcase.ToLowerCamel(headers[i])] = value
		}
		result = append(result, item)
	}

	return result
}
