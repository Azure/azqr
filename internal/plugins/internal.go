// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// InternalPluginScanner defines the interface for internal plugins
type InternalPluginScanner interface {
	// Scan executes the plugin and returns table data
	Scan(ctx context.Context, cred azcore.TokenCredential, subscriptions map[string]string, filters *models.Filters) (*ExternalPluginOutput, error)

	// GetMetadata returns metadata about the plugin
	GetMetadata() PluginMetadata
}

// internalPluginRegistry holds all registered internal plugins
var internalPluginRegistry = make(map[string]InternalPluginScanner)

// RegisterInternalPlugin registers an internal plugin and creates its command
func RegisterInternalPlugin(name string, scanner InternalPluginScanner) {
	internalPluginRegistry[name] = scanner

	// Create a Cobra command for this plugin
	metadata := scanner.GetMetadata()
	cmd := createPluginCommand(name, metadata.Description)

	// Register the plugin with the global registry immediately
	plugin := &Plugin{
		Metadata:        metadata,
		InternalScanner: scanner,
		Command:         cmd,
	}

	// Register with global registry
	if err := GetRegistry().Register(plugin); err != nil {
		log.Fatal().Err(err).Msgf("Failed to register internal plugin: %s", name)
	}
}

// GetInternalPlugin retrieves a registered internal plugin
func GetInternalPlugin(name string) (InternalPluginScanner, bool) {
	scanner, exists := internalPluginRegistry[name]
	return scanner, exists
}

// createPluginCommand creates a Cobra command for a plugin
// The actual Run function will be set by the commands package to call scan infrastructure
func createPluginCommand(name, description string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: description,
		Long:  description,
		Args:  cobra.NoArgs,
	}

	// Add all the standard scan flags to the plugin command
	cmd.Flags().StringArrayP("management-group-id", "", []string{}, "Azure Management Group Id")
	cmd.Flags().StringArrayP("subscription-id", "s", []string{}, "Azure Subscription Id")
	cmd.Flags().StringArrayP("resource-group", "g", []string{}, "Azure Resource Group (Use with --subscription-id)")
	cmd.Flags().BoolP("xlsx", "", true, "Create Excel report (default) (default true)")
	cmd.Flags().BoolP("json", "", false, "Create JSON report files")
	cmd.Flags().BoolP("csv", "", false, "Create CSV report files")
	cmd.Flags().BoolP("stdout", "", false, "Write the JSON output to stdout")
	cmd.Flags().StringP("output-name", "o", "", "Output file name without extension")
	cmd.Flags().BoolP("mask", "m", true, "Mask the subscription id in the report (default) (default true)")
	cmd.Flags().StringP("filters", "e", "", "Filters file (YAML format)")
	cmd.Flags().BoolP("debug", "", false, "Set log level to debug")

	return cmd
}
