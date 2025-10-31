// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetRegistry(t *testing.T) {
	registry := GetRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.plugins)
}

func TestRegistryRegister(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	// Initialize models.ScannerList
	models.ScannerList = make(map[string][]models.IAzureScanner)

	plugin := &Plugin{
		Metadata: PluginMetadata{
			Name:        "test1",
			Version:     "1.0.0",
			Description: "Test Plugin 1",
			Type:        PluginTypeYaml,
		},
	}

	err := registry.Register(plugin)
	assert.NoError(t, err)
	assert.Equal(t, 1, registry.Count())

	// Verify plugin was registered
	retrieved, exists := registry.Get("test1")
	assert.True(t, exists)
	assert.Equal(t, "test1", retrieved.Metadata.Name)
	assert.Equal(t, "1.0.0", retrieved.Metadata.Version)
}

func TestRegistryRegisterNilPlugin(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	err := registry.Register(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot register nil plugin")
}

func TestRegistryRegisterEmptyName(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	plugin := &Plugin{
		Metadata: PluginMetadata{
			Name: "",
		},
	}

	err := registry.Register(plugin)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin name cannot be empty")
}

func TestRegistryRegisterDuplicate(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	plugin1 := &Plugin{
		Metadata: PluginMetadata{
			Name:    "test",
			Version: "1.0.0",
		},
		YamlRecommendations: []models.AprlRecommendation{},
	}

	plugin2 := &Plugin{
		Metadata: PluginMetadata{
			Name:    "test",
			Version: "2.0.0",
		},
		YamlRecommendations: []models.AprlRecommendation{},
	}

	err := registry.Register(plugin1)
	assert.NoError(t, err)

	// Registering duplicate should replace
	err = registry.Register(plugin2)
	assert.NoError(t, err)
	assert.Equal(t, 1, registry.Count())

	retrieved, exists := registry.Get("test")
	assert.True(t, exists)
	assert.Equal(t, "2.0.0", retrieved.Metadata.Version)
}

func TestRegistryUnregister(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	plugin := &Plugin{
		Metadata: PluginMetadata{
			Name:    "test",
			Version: "1.0.0",
		},
		YamlRecommendations: []models.AprlRecommendation{},
	}

	_ = registry.Register(plugin)
	assert.Equal(t, 1, registry.Count())

	err := registry.Unregister("test")
	assert.NoError(t, err)
	assert.Equal(t, 0, registry.Count())

	_, exists := registry.Get("test")
	assert.False(t, exists)
}

func TestRegistryUnregisterNonExistent(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	err := registry.Unregister("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin nonexistent not found")
}

func TestRegistryGet(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	plugin := &Plugin{
		Metadata: PluginMetadata{
			Name:    "test",
			Version: "1.0.0",
		},
	}

	_ = registry.Register(plugin)

	retrieved, exists := registry.Get("test")
	assert.True(t, exists)
	assert.Equal(t, "test", retrieved.Metadata.Name)

	_, exists = registry.Get("nonexistent")
	assert.False(t, exists)
}

func TestRegistryList(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	// Register multiple plugins
	plugins := []*Plugin{
		{Metadata: PluginMetadata{Name: "zebra", Version: "1.0.0"}},
		{Metadata: PluginMetadata{Name: "alpha", Version: "1.0.0"}},
		{Metadata: PluginMetadata{Name: "beta", Version: "1.0.0"}},
	}

	for _, p := range plugins {
		_ = registry.Register(p)
	}

	list := registry.List()
	assert.Equal(t, 3, len(list))

	// Verify alphabetical order
	assert.Equal(t, "alpha", list[0].Metadata.Name)
	assert.Equal(t, "beta", list[1].Metadata.Name)
	assert.Equal(t, "zebra", list[2].Metadata.Name)
}

func TestRegistryListEmpty(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	list := registry.List()
	assert.NotNil(t, list)
	assert.Equal(t, 0, len(list))
}

func TestRegistryCount(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	assert.Equal(t, 0, registry.Count())

	plugin1 := &Plugin{Metadata: PluginMetadata{Name: "test1", Version: "1.0.0"}}
	plugin2 := &Plugin{Metadata: PluginMetadata{Name: "test2", Version: "1.0.0"}}

	_ = registry.Register(plugin1)
	assert.Equal(t, 1, registry.Count())

	_ = registry.Register(plugin2)
	assert.Equal(t, 2, registry.Count())

	_ = registry.Unregister("test1")
	assert.Equal(t, 1, registry.Count())
}

func TestAttachCommands(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	rootCmd := &cobra.Command{Use: "root"}

	plugin1 := &Plugin{
		Metadata: PluginMetadata{Name: "test1"},
		Command:  &cobra.Command{Use: "test1"},
	}
	plugin2 := &Plugin{
		Metadata: PluginMetadata{Name: "test2"},
		Command:  &cobra.Command{Use: "test2"},
	}
	plugin3 := &Plugin{
		Metadata: PluginMetadata{Name: "test3"},
		Command:  nil, // No command
	}

	_ = registry.Register(plugin1)
	_ = registry.Register(plugin2)
	_ = registry.Register(plugin3)

	attached := registry.AttachCommands(rootCmd)
	assert.Equal(t, 2, attached) // Only 2 plugins have commands

	// Verify commands were attached
	assert.True(t, rootCmd.HasSubCommands())
	subcommands := rootCmd.Commands()
	assert.Equal(t, 2, len(subcommands))
}

func TestPluginTypeString(t *testing.T) {
	assert.Equal(t, "yaml", pluginTypeString(PluginTypeYaml))
	assert.Equal(t, "unknown", pluginTypeString(PluginType(999)))
}

func TestConcurrentAccess(t *testing.T) {
	registry := &Registry{
		plugins: make(map[string]*Plugin),
	}

	// Test concurrent registration
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			plugin := &Plugin{
				Metadata: PluginMetadata{
					Name:    string(rune('a' + id)),
					Version: "1.0.0",
				},
			}
			_ = registry.Register(plugin)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(t, 10, registry.Count())
}
