// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"fmt"
	"sort"
	"sync"

	"github.com/Azure/azqr/internal/models"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Registry manages all registered plugins
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
}

// globalRegistry is the singleton plugin registry
var globalRegistry = &Registry{
	plugins: make(map[string]*Plugin),
}

// GetRegistry returns the global plugin registry
func GetRegistry() *Registry {
	return globalRegistry
}

// Register registers a plugin with the registry
func (r *Registry) Register(plugin *Plugin) error {
	if plugin == nil {
		return fmt.Errorf("cannot register nil plugin")
	}
	if plugin.Metadata.Name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, exists := r.plugins[plugin.Metadata.Name]; exists {
		log.Warn().
			Str("plugin", plugin.Metadata.Name).
			Str("existing_version", existing.Metadata.Version).
			Str("new_version", plugin.Metadata.Version).
			Msg("Plugin already registered, replacing with new version")
	}

	r.plugins[plugin.Metadata.Name] = plugin

	return nil
}

// Unregister removes a plugin from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Remove from ScannerList
	delete(models.ScannerList, name)

	delete(r.plugins, name)

	log.Info().Str("plugin", name).Msg("Plugin unregistered")
	return nil
}

// Get retrieves a plugin by name
func (r *Registry) Get(name string) (*Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	return plugin, exists
}

// List returns all registered plugins sorted by name
func (r *Registry) List() []*Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	sort.Strings(names)

	plugins := make([]*Plugin, 0, len(names))
	for _, name := range names {
		plugins = append(plugins, r.plugins[name])
	}

	return plugins
}

// Count returns the number of registered plugins
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.plugins)
}

// AttachCommands attaches all plugin commands to the parent command
func (r *Registry) AttachCommands(parentCmd *cobra.Command) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	attached := 0
	for _, plugin := range r.plugins {
		if plugin.Command != nil {
			parentCmd.AddCommand(plugin.Command)
			attached++
		}
	}

	return attached
}

// pluginTypeString converts PluginType to string for logging
func pluginTypeString(t PluginType) string {
	switch t {
	case PluginTypeYaml:
		return "yaml"
	case PluginTypeInternal:
		return "internal"
	default:
		return "unknown"
	}
}
