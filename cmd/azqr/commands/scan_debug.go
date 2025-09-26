// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

//go:build debug

package commands

import (
	"os"
	"runtime/pprof"
	"runtime/trace"

	"github.com/Azure/azqr/internal"
	"github.com/Azure/azqr/internal/models"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.PersistentFlags().StringArrayP("management-group-id", "", []string{}, "Azure Management Group Id")
	scanCmd.PersistentFlags().StringArrayP("subscription-id", "s", []string{}, "Azure Subscription Id")
	scanCmd.PersistentFlags().StringArrayP("resource-group", "g", []string{}, "Azure Resource Group (Use with --subscription-id)")
	scanCmd.PersistentFlags().BoolP("defender", "d", true, "Scan Defender Status (default) (default true)")
	scanCmd.PersistentFlags().BoolP("advisor", "a", true, "Scan Azure Advisor Recommendations (default) (default true)")
	scanCmd.PersistentFlags().BoolP("costs", "c", true, "Scan Azure Costs (default) (default true)")
	scanCmd.PersistentFlags().BoolP("xlsx", "", true, "Create Excel report (default) (default true)")
	scanCmd.PersistentFlags().BoolP("json", "", false, "Create JSON report files")
	scanCmd.PersistentFlags().BoolP("csv", "", false, "Create CSV report files")
	scanCmd.PersistentFlags().BoolP("stdout", "", false, "Create CSV report files")
	scanCmd.PersistentFlags().StringP("output-name", "o", "", "Output file name without extension")
	scanCmd.PersistentFlags().BoolP("mask", "m", true, "Mask the subscription id in the report (default) (default true)")
	scanCmd.PersistentFlags().StringP("filters", "e", "", "Filters file (YAML format)")
	scanCmd.PersistentFlags().BoolP("azqr", "", true, "Scan Azure Quick Review Recommendations (default) (default true)")
	scanCmd.PersistentFlags().BoolP("debug", "", false, "Set log level to debug")

	// Profiling flags (only available in debug builds)
	scanCmd.PersistentFlags().StringP("cpu-profile", "", "", "Write CPU profile to file (debug build only)")
	scanCmd.PersistentFlags().StringP("mem-profile", "", "", "Write memory profile to file (debug build only)")
	scanCmd.PersistentFlags().StringP("trace-profile", "", "", "Write execution trace to file (debug build only)")

	rootCmd.AddCommand(scanCmd)
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan Azure Resources",
	Long:  "Scan Azure Resources (debug build with profiling support)",
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
	xlsx, _ := cmd.Flags().GetBool("xlsx")
	csv, _ := cmd.Flags().GetBool("csv")
	json, _ := cmd.Flags().GetBool("json")
	mask, _ := cmd.Flags().GetBool("mask")
	debug, _ := cmd.Flags().GetBool("debug")
	stdout, _ := cmd.Flags().GetBool("stdout")
	filtersFile, _ := cmd.Flags().GetString("filters")
	useAzqr, _ := cmd.Flags().GetBool("azqr")

	// Get profiling flags (only available in debug builds)
	cpuProfile, _ := cmd.Flags().GetString("cpu-profile")
	memProfile, _ := cmd.Flags().GetString("mem-profile")
	traceProfile, _ := cmd.Flags().GetString("trace-profile")

	// Start CPU profiling if requested
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create CPU profile file")
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close CPU profile file")
			}
		}()

		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal().Err(err).Msg("Failed to start CPU profiling")
		}
		defer pprof.StopCPUProfile()
		log.Info().Msgf("CPU profiling enabled, writing to: %s", cpuProfile)
	}

	// Start execution trace if requested
	if traceProfile != "" {
		f, err := os.Create(traceProfile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create trace profile file")
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close trace profile file")
			}
		}()

		if err := trace.Start(f); err != nil {
			log.Fatal().Err(err).Msg("Failed to start execution trace")
		}
		defer trace.Stop()
		log.Info().Msgf("Execution trace enabled, writing to: %s", traceProfile)
	}

	// load filters
	filters := models.LoadFilters(filtersFile, scannerKeys)

	params := internal.ScanParams{
		ManagementGroups:       managementGroups,
		Subscriptions:          subscriptions,
		ResourceGroups:         resourceGroups,
		OutputName:             outputFileName,
		Defender:               defender,
		Advisor:                advisor,
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
	}

	scanner := internal.Scanner{}
	scanner.Scan(&params)

	// Write memory profile if requested
	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create memory profile file")
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close memory profile file")
			}
		}()

		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal().Err(err).Msg("Failed to write memory profile")
		}
		log.Info().Msgf("Memory profile written to: %s", memProfile)
	}
}
