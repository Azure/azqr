// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

//go:build !debug

package commands

import (
	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"

	"github.com/spf13/cobra"
)

func init() {
	scanCmd.PersistentFlags().StringArrayP("management-group-id", "", []string{}, "Azure Management Group Id")
	scanCmd.PersistentFlags().StringArrayP("subscription-id", "s", []string{}, "Azure Subscription Id")
	scanCmd.PersistentFlags().StringArrayP("resource-group", "g", []string{}, "Azure Resource Group (Use with --subscription-id)")
	scanCmd.PersistentFlags().BoolP("defender", "d", true, "Scan Defender Status (default) (default true)")
	scanCmd.PersistentFlags().BoolP("advisor", "a", true, "Scan Azure Advisor Recommendations (default) (default true)")
	scanCmd.PersistentFlags().BoolP("costs", "c", true, "Scan Azure Costs (default) (default true)")
	// ...existing code...
	// Add flags for all internal plugins, default to false
	for _, pluginName := range requireInternalPluginsList() {
		desc := "Enable internal plugin: " + pluginName
		scanCmd.PersistentFlags().Bool(pluginName, false, desc)
	}

	scanCmd.PersistentFlags().BoolP("arc", "", true, "Scan Azure Arc-enabled resources (default) (default true)")
	scanCmd.PersistentFlags().BoolP("xlsx", "", true, "Create Excel report (default) (default true)")
	scanCmd.PersistentFlags().BoolP("json", "", false, "Create JSON report files")
	scanCmd.PersistentFlags().BoolP("csv", "", false, "Create CSV report files")
	scanCmd.PersistentFlags().BoolP("stdout", "", false, "Create CSV report files")
	scanCmd.PersistentFlags().StringP("output-name", "o", "", "Output file name without extension")
	scanCmd.PersistentFlags().BoolP("mask", "m", true, "Mask the subscription id in the report (default) (default true)")
	scanCmd.PersistentFlags().StringP("filters", "e", "", "Filters file (YAML format)")
	scanCmd.PersistentFlags().BoolP("azqr", "", true, "Scan Azure Quick Review Recommendations (default) (default true)")
	scanCmd.PersistentFlags().BoolP("debug", "", false, "Set log level to debug")

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
	defender, _ := cmd.Flags().GetBool("defender")
	advisor, _ := cmd.Flags().GetBool("advisor")
	cost, _ := cmd.Flags().GetBool("costs")
	arc, _ := cmd.Flags().GetBool("arc")
	xlsx, _ := cmd.Flags().GetBool("xlsx")
	csv, _ := cmd.Flags().GetBool("csv")
	json, _ := cmd.Flags().GetBool("json")
	mask, _ := cmd.Flags().GetBool("mask")
	debug, _ := cmd.Flags().GetBool("debug")
	stdout, _ := cmd.Flags().GetBool("stdout")
	filtersFile, _ := cmd.Flags().GetString("filters")
	useAzqr, _ := cmd.Flags().GetBool("azqr")

	// load filters
	filters := models.LoadFilters(filtersFile, scannerKeys)

	// Read enabled internal plugin flags
	enabledInternalPlugins := map[string]bool{}
	for _, pluginName := range requireInternalPluginsList() {
		val, _ := cmd.Flags().GetBool(pluginName)
		enabledInternalPlugins[pluginName] = val
	}

	params := internal.ScanParams{
		ManagementGroups:       managementGroups,
		Subscriptions:          subscriptions,
		ResourceGroups:         resourceGroups,
		OutputName:             outputFileName,
		Defender:               defender,
		Advisor:                advisor,
		Arc:                    arc,
		Xlsx:                   xlsx,
		Cost:                   cost,
		Csv:                    csv,
		Json:                   json,
		Mask:                   mask,
		Stdout:                 stdout,
		Debug:                  debug,
		ScannerKeys:            scannerKeys,
		Filters:                filters,
		UseAzqrRecommendations: useAzqr,
		EnabledInternalPlugins: enabledInternalPlugins,
	}

	scanner := internal.Scanner{}
	scanner.Scan(&params)
}

// requireInternalPluginsList returns the list of internal plugin names
func requireInternalPluginsList() []string {
	return plugins.ListInternalPlugins()
}
