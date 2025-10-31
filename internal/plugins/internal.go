// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"context"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
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

// RegisterInternalPlugin registers an internal plugin
func RegisterInternalPlugin(name string, scanner InternalPluginScanner) {
	internalPluginRegistry[name] = scanner
}

// GetInternalPlugin retrieves a registered internal plugin
func GetInternalPlugin(name string) (InternalPluginScanner, bool) {
	scanner, exists := internalPluginRegistry[name]
	return scanner, exists
}

// ListInternalPlugins returns all registered internal plugins
func ListInternalPlugins() []string {
	names := make([]string, 0, len(internalPluginRegistry))
	for name := range internalPluginRegistry {
		names = append(names, name)
	}
	return names
}

// discoverInternalPlugins converts internal plugins to Plugin objects
func discoverInternalPlugins() []*Plugin {
	plugins := make([]*Plugin, 0, len(internalPluginRegistry))

	for _, scanner := range internalPluginRegistry {
		metadata := scanner.GetMetadata()
		metadata.Type = PluginTypeInternal

		plugin := &Plugin{
			Metadata:        metadata,
			InternalScanner: scanner,
		}

		plugins = append(plugins, plugin)
	}

	return plugins
}
