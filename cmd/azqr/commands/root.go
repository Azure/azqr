// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"os"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/spf13/cobra"

	"github.com/Azure/azqr/internal/plugins"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "azqr",
	Short:   "Azure Quick Review (azqr) goal is to produce a high level assessment of an Azure Subscription or Resource Group",
	Long:    `Azure Quick Review (azqr) goal is to produce a high level assessment of an Azure Subscription or Resource Group`,
	Args:    cobra.NoArgs,
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize log level based on --debug flag
		// This runs before any command executes, making debug logging available globally
		debug, _ := cmd.Flags().GetBool("debug")
		InitializeLogLevel(debug)
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

func init() {
	// Add global --debug flag to root command (available to all subcommands)
	rootCmd.PersistentFlags().BoolP("debug", "", false, "Enable debug logging")
}

// InitializeLogLevel sets the global log level based on debug flag
// This is exported so it can be called from the root command to set logging early
func InitializeLogLevel(debug bool) {
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
		return
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func Execute() {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	// Load all YAML plugins after logger is configured
	if err := plugins.LoadAll(); err != nil {
		log.Warn().Err(err).Msg("Failed to load some plugins")
	}

	// Attach plugin commands as top-level commands and set their Run functions
	registry := plugins.GetRegistry()
	for _, plugin := range registry.List() {
		if plugin.Command != nil {
			pluginName := plugin.Metadata.Name
			// Capture pluginName in closure properly
			pName := pluginName
			// Set the Run function to enable only this plugin
			plugin.Command.Run = func(cmd *cobra.Command, args []string) {
				// Enable only this specific plugin
				// Note: We can't use Set() for StringArray flags, so we pass it directly to scan
				scannerKeys, _ := models.GetScanners()
				// Create a custom scan with this plugin enabled
				scanWithPlugin(cmd, scannerKeys, pName)
			}
			rootCmd.AddCommand(plugin.Command)
		}
	}

	cobra.CheckErr(rootCmd.Execute())
}
