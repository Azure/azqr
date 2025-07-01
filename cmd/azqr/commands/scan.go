// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/models"

	"github.com/spf13/cobra"
)

func init() {
	scanCmd.PersistentFlags().StringArrayP("management-group-id", "", []string{}, "Azure Management Group Id")
	scanCmd.PersistentFlags().StringArrayP("subscription-id", "s", []string{}, "Azure Subscription Id")
	scanCmd.PersistentFlags().StringArrayP("resource-group", "g", []string{}, "Azure Resource Group (Use with --subscription-id)")
	scanCmd.PersistentFlags().BoolP("defender", "d", true, "Scan Defender Status (default)")
	scanCmd.PersistentFlags().BoolP("advisor", "a", true, "Scan Azure Advisor Recommendations (default)")
	scanCmd.PersistentFlags().BoolP("costs", "c", true, "Scan Azure Costs (default)")
	scanCmd.PersistentFlags().BoolP("xslx", "", true, "Create Excel report (default)")
	scanCmd.PersistentFlags().BoolP("json", "", false, "Create JSON report files")
	scanCmd.PersistentFlags().BoolP("csv", "", false, "Create CSV report files")
	scanCmd.PersistentFlags().StringP("output-name", "o", "", "Output file name without extension")
	scanCmd.PersistentFlags().BoolP("mask", "m", true, "Mask the subscription id in the report (default)")

	scanCmd.PersistentFlags().BoolP("debug", "", false, "Set log level to debug")
	scanCmd.PersistentFlags().StringP("filters", "e", "", "Filters file (YAML format)")
	scanCmd.PersistentFlags().BoolP("azqr", "", true, "Scan Azure Quick Review Recommendations (default)")

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
	xlsx, _ := cmd.Flags().GetBool("xslx")
	csv, _ := cmd.Flags().GetBool("csv")
	json, _ := cmd.Flags().GetBool("json")
	mask, _ := cmd.Flags().GetBool("mask")
	debug, _ := cmd.Flags().GetBool("debug")

	filtersFile, _ := cmd.Flags().GetString("filters")
	useAzqr, _ := cmd.Flags().GetBool("azqr")

	// load filters
	filters := models.LoadFilters(filtersFile, scannerKeys)

	params := internal.ScanParams{
		ManagementGroups:        managementGroups,
		Subscriptions:           subscriptions,
		ResourceGroups:          resourceGroups,
		OutputName:              outputFileName,
		Defender:                defender,
		Advisor:                 advisor,
		Xlsx:                    xlsx,
		Cost:                    cost,
		Csv:                     csv,
		Json:                    json,
		Mask:                    mask,
		Debug:                   debug,
		ScannerKeys:             scannerKeys,
		Filters:                 filters,
		UseAzqrRecommendations:  useAzqr,
	}

	scanner := internal.Scanner{}
	scanner.Scan(&params)
}
