// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/gate"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/Azure/azqr/internal/profiling"
	"github.com/Azure/azqr/internal/store"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

func init() {
	scanCmd.PersistentFlags().StringSliceP("management-group-id", "", []string{}, "Azure Management Group Id")
	scanCmd.PersistentFlags().StringSliceP("subscription-id", "s", []string{}, "Azure Subscription Id")
	scanCmd.PersistentFlags().StringSliceP("resource-group", "g", []string{}, "Azure Resource Group (Use with --subscription-id)")
	scanCmd.PersistentFlags().StringSliceP("stages", "", []string{}, "Control scan stages. Without this flag, defaults are used (enabled: diagnostics,advisor,defender). Specify stages to enable (e.g., --stages cost,policy) or prefix with '-' to disable (e.g., --stages -diagnostics). Available: advisor,defender,defender-recommendations,arc,policy,cost,diagnostics")
	scanCmd.PersistentFlags().StringArrayP("stage-param", "", []string{}, "Stage options in the form 'stage.key=value' (repeatable)")
	scanCmd.PersistentFlags().StringSliceP("plugin", "", []string{}, "Enable internal plugins (comma-separated or multiple flags)")
	scanCmd.PersistentFlags().BoolP("xlsx", "", true, "Create Excel report (default) (default true)")
	scanCmd.PersistentFlags().BoolP("json", "", false, "Create JSON report files")
	scanCmd.PersistentFlags().BoolP("csv", "", false, "Create CSV report files")
	scanCmd.PersistentFlags().StringP("output-name", "o", "", "Output file name without extension")
	scanCmd.PersistentFlags().BoolP("mask", "m", true, "Mask the subscription id in the report (default) (default true)")
	scanCmd.PersistentFlags().StringP("filters", "e", "", "Filters file (YAML format)")
	scanCmd.PersistentFlags().StringP("fail-on", "", "", "Fail when impacted recommendations meet or exceed this impact (High, Medium, Low)")
	scanCmd.PersistentFlags().Bool("sarif", false, "Create a SARIF 2.1.0 report")

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
	Use:          "scan",
	Short:        "Scan Azure Resources",
	Long:         "Scan Azure Resources",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		scannerKeys, _ := models.GetScanners()
		return scan(cmd, scannerKeys)
	},
}

func scan(cmd *cobra.Command, scannerKeys []string) error {
	managementGroups, _ := cmd.Flags().GetStringSlice("management-group-id")
	subscriptions, _ := cmd.Flags().GetStringSlice("subscription-id")
	resourceGroups, _ := cmd.Flags().GetStringSlice("resource-group")
	outputFileName, _ := cmd.Flags().GetString("output-name")
	stageNames, _ := cmd.Flags().GetStringSlice("stages")
	stageParams, _ := cmd.Flags().GetStringArray("stage-param")
	xlsx, _ := cmd.Flags().GetBool("xlsx")
	csv, _ := cmd.Flags().GetBool("csv")
	json, _ := cmd.Flags().GetBool("json")
	mask, _ := cmd.Flags().GetBool("mask")
	filtersFile, _ := cmd.Flags().GetString("filters")
	failOn, _ := cmd.Flags().GetString("fail-on")
	sarifEnabled, _ := cmd.Flags().GetBool("sarif")
	pluginNames, _ := cmd.Flags().GetStringSlice("plugin")

	var gateCriterion gate.Criterion
	var gateEnabled bool
	if failOn != "" {
		var err error
		gateCriterion, err = gate.Parse(failOn)
		if err != nil {
			return err
		}
		gateEnabled = true
	}

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

	if err := stageConfigs.ApplyStageParams(stageParams); err != nil {
		return fmt.Errorf("failed applying stage parameters: %w", err)
	}

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
		ScannerKeys:            scannerKeys,
		Filters:                filters,
		EnabledInternalPlugins: enabledInternalPlugins,
		CPUProfile:             cpuProfile,
		MemProfile:             memProfile,
		TraceProfile:           traceProfile,
		Sarif:                  sarifEnabled,
		Version:                version,
	}

	summary, err := runPersistedOrPlain(cmd, store.SessionKindFull, store.SessionTriggerCLI, &params, func(sc *pipeline.Scanner) (*findings.Summary, error) {
		return sc.Scan(&params)
	})
	if err != nil {
		return err
	}
	if gateEnabled {
		return checkGate(cmd, summary, gateCriterion)
	}
	return nil
}

func checkGate(cmd *cobra.Command, summary *findings.Summary, criterion gate.Criterion) error {
	err := gate.Check(summary, criterion)
	if err == nil {
		return nil
	}
	log.Error().Msg(err.Error())
	cmd.Root().SilenceErrors = true
	cmd.Root().SilenceUsage = true
	return err
}

// scanWithPlugin is a specialized version of scan that enables a specific plugin
// and forces plugin-only mode for faster execution by calling ScanPlugins directly
func scanWithPlugin(cmd *cobra.Command, scannerKeys []string, pluginName string) error {
	managementGroups, _ := cmd.Flags().GetStringSlice("management-group-id")
	subscriptions, _ := cmd.Flags().GetStringSlice("subscription-id")
	resourceGroups, _ := cmd.Flags().GetStringSlice("resource-group")
	outputFileName, _ := cmd.Flags().GetString("output-name")
	xlsx, _ := cmd.Flags().GetBool("xlsx")
	csv, _ := cmd.Flags().GetBool("csv")
	json, _ := cmd.Flags().GetBool("json")
	mask, _ := cmd.Flags().GetBool("mask")
	debug, _ := cmd.Flags().GetBool("debug")
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

	// Bridge plugin-specific flags into stage options so plugins can read them
	// via params.Stages.GetStageOptions without relying on package-level globals.
	if targetRegions, _ := cmd.Flags().GetStringSlice("target-regions"); len(targetRegions) > 0 {
		_ = stageConfigs.SetStageOptions(models.StageNamePlugin, map[string]any{
			"target-regions": strings.Join(targetRegions, ","),
		})
	}

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
		Debug:                  debug,
		ScannerKeys:            scannerKeys,
		Filters:                filters,
		EnabledInternalPlugins: enabledInternalPlugins,
		CPUProfile:             cpuProfile,
		MemProfile:             memProfile,
		TraceProfile:           traceProfile,
	}

	// Call ScanPlugins directly for optimized plugin-only execution
	_, err := runPersistedOrPlain(cmd, store.SessionKindPluginOnly, store.SessionTriggerCLI, &params, func(sc *pipeline.Scanner) (*findings.Summary, error) {
		return sc.ScanPlugins(&params)
	})
	return err
}

// runPersistedOrPlain opens the session store configured via --store-path (falling back to
// store.DefaultPath() when unset) and runs scanFn through store.RunPersistedScan so the scan
// becomes a persisted session. If the store cannot be opened, it logs a warning and runs scanFn
// against a plain *pipeline.Scanner instead, so store unavailability never blocks an otherwise
// successful scan and its existing export behavior.
func runPersistedOrPlain(
	cmd *cobra.Command,
	kind store.SessionKind,
	trigger store.SessionTrigger,
	params *models.ScanParams,
	scanFn func(*pipeline.Scanner) (*findings.Summary, error),
) (*findings.Summary, error) {
	storePath, _ := cmd.Flags().GetString("store-path")

	st := store.OpenForScan(storePath)
	if st == nil {
		return scanFn(&pipeline.Scanner{})
	}
	defer func() { _ = st.Close() }()

	var sess *store.Session
	summary, err := st.RunPersistedScan(kind, trigger, params, version, scanFn, func(s *store.Session) { sess = s })
	if sess != nil {
		log.Info().Str("sessionId", sess.ID).Str("scopeId", sess.ScopeID).Msg("Scan session recorded")
	}

	return summary, err
}
