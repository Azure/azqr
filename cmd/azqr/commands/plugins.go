// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Azure/azqr/internal/plugins"
	"github.com/spf13/cobra"
)

func init() {
	pluginsCmd.AddCommand(pluginsListCmd)
	pluginsCmd.AddCommand(pluginsInfoCmd)
	rootCmd.AddCommand(pluginsCmd)
}

var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Manage azqr plugins",
	Long:  "List, inspect, and manage azqr plugins for extending functionality",
	Args:  cobra.NoArgs,
}

var pluginsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered plugins",
	Long:  "List all registered plugins including built-in and external command plugins",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		registry := plugins.GetRegistry()
		pluginList := registry.List()

		if len(pluginList) == 0 {
			fmt.Println("No plugins registered")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tVERSION\tTYPE\tDESCRIPTION")
		_, _ = fmt.Fprintln(w, "----\t-------\t----\t-----------")

		for _, p := range pluginList {
			pluginType := "yaml"
			switch p.Metadata.Type {
			case plugins.PluginTypeInternal:
				pluginType = "internal"
			case plugins.PluginTypeYaml:
				pluginType = "yaml"
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				p.Metadata.Name,
				p.Metadata.Version,
				pluginType,
				p.Metadata.Description,
			)
		}

		_ = w.Flush()
	},
}

var pluginsInfoCmd = &cobra.Command{
	Use:   "info <plugin-name>",
	Short: "Show detailed information about a plugin",
	Long:  "Show detailed information about a specific plugin including metadata and capabilities",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pluginName := args[0]
		registry := plugins.GetRegistry()

		plugin, exists := registry.Get(pluginName)
		if !exists {
			fmt.Printf("Plugin '%s' not found\n", pluginName)
			os.Exit(1)
		}

		fmt.Printf("Plugin: %s\n", plugin.Metadata.Name)
		fmt.Printf("Version: %s\n", plugin.Metadata.Version)
		fmt.Printf("Description: %s\n", plugin.Metadata.Description)

		pluginType := "yaml"
		switch plugin.Metadata.Type {
		case plugins.PluginTypeInternal:
			pluginType = "internal"
		case plugins.PluginTypeYaml:
			pluginType = "yaml"
		}
		fmt.Printf("Type: %s\n", pluginType)

		if plugin.Metadata.Author != "" {
			fmt.Printf("Author: %s\n", plugin.Metadata.Author)
		}
		if plugin.Metadata.License != "" {
			fmt.Printf("License: %s\n", plugin.Metadata.License)
		}
		if plugin.Metadata.CommandPath != "" {
			fmt.Printf("Command Path: %s\n", plugin.Metadata.CommandPath)
		}

		// Show YAML plugin recommendations
		if len(plugin.YamlRecommendations) > 0 {
			fmt.Printf("\nYAML Plugin Information:\n")
			fmt.Printf("  Recommendations: %d\n", len(plugin.YamlRecommendations))
			// Collect unique resource types
			resourceTypeSet := make(map[string]bool)
			for _, rec := range plugin.YamlRecommendations {
				resourceTypeSet[rec.ResourceType] = true
			}
			resourceTypes := make([]string, 0, len(resourceTypeSet))
			for rt := range resourceTypeSet {
				resourceTypes = append(resourceTypes, rt)
			}
			fmt.Printf("  Resource Types: %v\n", resourceTypes)
		}

		// Show internal plugin scanner info
		if plugin.InternalScanner != nil {
			fmt.Printf("\nInternal Plugin Information:\n")
			fmt.Printf("  Scanner available\n")
		}
	},
}
