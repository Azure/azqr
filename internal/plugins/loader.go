// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// getPluginDirs returns the directories to search for YAML plugins
func getPluginDirs() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".azqr", "plugins"),
		"./plugins",
	}
}

// LoadAll discovers and loads all available plugins (internal and YAML)
func LoadAll() error {
	registry := GetRegistry()

	// Discover internal plugins (registered at init time)
	internalPlugins := discoverInternalPlugins()
	for _, plugin := range internalPlugins {
		if err := registry.Register(plugin); err != nil {
			log.Warn().
				Err(err).
				Str("plugin", plugin.Metadata.Name).
				Msg("Failed to register internal plugin")
		}
	}
	log.Info().Int("count", len(internalPlugins)).Msg("Internal plugins discovered")

	// Discover YAML plugins
	yamlPlugins, err := discoverYamlPlugins(getPluginDirs())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to discover YAML plugins")
	} else {
		for _, plugin := range yamlPlugins {
			if err := registry.Register(plugin); err != nil {
				log.Warn().
					Err(err).
					Str("plugin", plugin.Metadata.Name).
					Msg("Failed to register YAML plugin")
			}
		}
		log.Info().Int("count", len(yamlPlugins)).Msg("YAML plugins discovered")
	}

	return nil
}
