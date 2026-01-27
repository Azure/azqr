// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

//go:build !debug

package commands

import (
	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/profiling"

	"github.com/spf13/cobra"
)

func init() {
	scanCmd.PersistentFlags().StringArrayP("management-group-id", "", []string{}, "Azure Management Group Id")
	scanCmd.PersistentFlags().StringArrayP("subscription-id", "s", []string{}, "Azure Subscription Id")
	scanCmd.PersistentFlags().StringArrayP("resource-group", "g", []string{}, "Azure Resource Group (Use with --subscription-id)")
	scanCmd.PersistentFlags().StringArrayP("stages", "", []string{}, "Control scan stages. Without this flag, defaults are used (enabled: diagnostics,advisor,defender). Specify stages to enable (e.g., --stages cost,policy) or prefix with '-' to disable (e.g., --stages -diagnostics). Available: advisor,defender,defender-recommendations,arc,policy,cost,diagnostics")
	scanCmd.PersistentFlags().StringArrayP("plugin", "", []string{}, "Enable internal plugins (comma-separated or multiple flags)")
	scanCmd.PersistentFlags().BoolP("xlsx", "", true, "Create Excel report (default) (default true)")
	scanCmd.PersistentFlags().BoolP("json", "", false, "Create JSON report files")
	scanCmd.PersistentFlags().BoolP("csv", "", false, "Create CSV report files")
	scanCmd.PersistentFlags().BoolP("stdout", "", false, "Write the JSON output to stdout")
	scanCmd.PersistentFlags().StringP("output-name", "o", "", "Output file name without extension")
	scanCmd.PersistentFlags().BoolP("mask", "m", true, "Mask the subscription id in the report (default) (default true)")
	scanCmd.PersistentFlags().StringP("filters", "e", "", "Filters file (YAML format)")

	// Conditionally add profiling flags if profiling is available and enabled via environment
	// Build with -tags debug to enable profiling features
	if profiling.IsProfilingAvailable() {
		scanCmd.PersistentFlags().StringP("cpu-profile", "", "", "Write CPU profile to file (requires debug build or AZQR_ENABLE_PROFILING=1)")
		scanCmd.PersistentFlags().StringP("mem-profile", "", "", "Write memory profile to file (requires debug build or AZQR_ENABLE_PROFILING=1)")
		scanCmd.PersistentFlags().StringP("trace-profile", "", "", "Write execution trace to file (requires debug build or AZQR_ENABLE_PROFILING=1)")
	}

	rootCmd.AddCommand(scanCmd)
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan Azure Resources",
	Long:  "Scan Azure Resources",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scannerKeys, _ := models.GetScanners()
		scan(cmd, scannerKeys)
	},
}

func scan(cmd *cobra.Command, scannerKeys []string) {
	managementGroups, _ := cmd.Flags().GetStringArray("management-group-id")
	subscriptions, _ := cmd.Flags().GetStringArray("subscription-id")
	resourceGroups, _ := cmd.Flags().GetStringArray("resource-group")
	outputFileName, _ := cmd.Flags().GetString("output-name")
	stageNames, _ := cmd.Flags().GetStringArray("stages")
	xlsx, _ := cmd.Flags().GetBool("xlsx")
	csv, _ := cmd.Flags().GetBool("csv")
	json, _ := cmd.Flags().GetBool("json")
	mask, _ := cmd.Flags().GetBool("mask")
	stdout, _ := cmd.Flags().GetBool("stdout")
	filtersFile, _ := cmd.Flags().GetString("filters")
	pluginNames, _ := cmd.Flags().GetStringArray("plugin")

	// Get profiling flags if available
	var cpuProfile, memProfile, traceProfile string
	if profiling.IsProfilingAvailable() {
		cpuProfile, _ = cmd.Flags().GetString("cpu-profile")
		memProfile, _ = cmd.Flags().GetString("mem-profile")
		traceProfile, _ = cmd.Flags().GetString("trace-profile")
	}

	// load filters
	filters := models.LoadFilters(filtersFile, scannerKeys)

	// Build enabled plugins map from --plugin flag
	enabledInternalPlugins := map[string]bool{}
	for _, pluginName := range pluginNames {
		enabledInternalPlugins[pluginName] = true
	}

	// Initialize stage configs
	stageConfigs := models.NewStageConfigsWithDefaults()
	stageConfigs.ConfigureStages(stageNames)

	params := models.ScanParams{
		ManagementGroups:       managementGroups,
		Subscriptions:          subscriptions,
		ResourceGroups:         resourceGroups,
		OutputName:             outputFileName,
		Stages:                 stageConfigs,
		Xlsx:                   xlsx,
		Csv:                    csv,
		Json:                   json,
		Mask:                   mask,
		Stdout:                 stdout,
		ScannerKeys:            scannerKeys,
		Filters:                filters,
		EnabledInternalPlugins: enabledInternalPlugins,
		CPUProfile:             cpuProfile,
		MemProfile:             memProfile,
		TraceProfile:           traceProfile,
	}

	scanner := internal.Scanner{}
	scanner.Scan(&params)
}

// scanWithPlugin is a specialized version of scan that enables a specific plugin
// and forces plugin-only mode for faster execution by calling ScanPlugins directly
func scanWithPlugin(cmd *cobra.Command, scannerKeys []string, pluginName string) {
	managementGroups, _ := cmd.Flags().GetStringArray("management-group-id")
	subscriptions, _ := cmd.Flags().GetStringArray("subscription-id")
	resourceGroups, _ := cmd.Flags().GetStringArray("resource-group")
	outputFileName, _ := cmd.Flags().GetString("output-name")
	xlsx, _ := cmd.Flags().GetBool("xlsx")
	csv, _ := cmd.Flags().GetBool("csv")
	json, _ := cmd.Flags().GetBool("json")
	mask, _ := cmd.Flags().GetBool("mask")
	debug, _ := cmd.Flags().GetBool("debug")
	stdout, _ := cmd.Flags().GetBool("stdout")
	filtersFile, _ := cmd.Flags().GetString("filters")

	// Get profiling flags if available
	var cpuProfile, memProfile, traceProfile string
	if profiling.IsProfilingAvailable() {
		cpuProfile, _ = cmd.Flags().GetString("cpu-profile")
		memProfile, _ = cmd.Flags().GetString("mem-profile")
		traceProfile, _ = cmd.Flags().GetString("trace-profile")
	}

	// load filters
	filters := models.LoadFilters(filtersFile, scannerKeys)

	// Enable only the specified plugin
	enabledInternalPlugins := map[string]bool{
		pluginName: true,
	}

	stageConfigs := models.NewStageConfigs()

	params := models.ScanParams{
		ManagementGroups:       managementGroups,
		Subscriptions:          subscriptions,
		ResourceGroups:         resourceGroups,
		OutputName:             outputFileName,
		Stages:                 stageConfigs,
		Xlsx:                   xlsx,
		Csv:                    csv,
		Json:                   json,
		Mask:                   mask,
		Stdout:                 stdout,
		Debug:                  debug,
		ScannerKeys:            scannerKeys,
		Filters:                filters,
		EnabledInternalPlugins: enabledInternalPlugins,
		CPUProfile:             cpuProfile,
		MemProfile:             memProfile,
		TraceProfile:           traceProfile,
	}

	scanner := internal.Scanner{}
	// Call ScanPlugins directly for optimized plugin-only execution
	scanner.ScanPlugins(&params)
}
