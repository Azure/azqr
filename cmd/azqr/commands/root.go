// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"os"
	"time"

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
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

func Execute() {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	// Load all YAML plugins after logger is configured
	if err := plugins.LoadAll(); err != nil {
		log.Warn().Err(err).Msg("Failed to load some plugins")
	}

	cobra.CheckErr(rootCmd.Execute())
}
