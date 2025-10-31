// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azqr/internal/models"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// LoadYamlPlugin loads a YAML plugin from a file and converts queries to AprlRecommendation format
func LoadYamlPlugin(filePath string) (*Plugin, []models.AprlRecommendation, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read YAML plugin file: %w", err)
	}

	var config YamlPluginConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, nil, fmt.Errorf("failed to parse YAML plugin: %w", err)
	}

	// Validate plugin configuration
	if config.Name == "" {
		return nil, nil, fmt.Errorf("plugin name is required")
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}
	if len(config.Queries) == 0 {
		return nil, nil, fmt.Errorf("plugin must have at least one query")
	}

	// Load external query files if specified
	baseDir := filepath.Dir(filePath)
	for i := range config.Queries {
		if config.Queries[i].QueryFile != "" {
			queryPath := filepath.Join(baseDir, config.Queries[i].QueryFile)
			queryData, err := os.ReadFile(queryPath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read query file %s: %w", config.Queries[i].QueryFile, err)
			}
			config.Queries[i].Query = string(queryData)
		}

		// Validate query has either inline query or query file
		if config.Queries[i].Query == "" {
			return nil, nil, fmt.Errorf("query %s must have either 'query' or 'queryFile' specified", config.Queries[i].AprlGuid)
		}

		// Validate required fields
		if config.Queries[i].AprlGuid == "" {
			return nil, nil, fmt.Errorf("query missing required field 'aprlGuid'")
		}
		if config.Queries[i].Description == "" {
			return nil, nil, fmt.Errorf("query %s missing required field 'description'", config.Queries[i].AprlGuid)
		}
	}

	// Convert YamlPluginQuery to AprlRecommendation format
	recommendations := make([]models.AprlRecommendation, 0, len(config.Queries))
	for _, query := range config.Queries {
		automationAvailable := "false"
		if query.AutomationAvailable {
			automationAvailable = "true"
		}

		recommendation := models.AprlRecommendation{
			RecommendationID:    query.AprlGuid,
			ResourceType:        query.RecommendationResourceType,
			Recommendation:      query.Description,
			Category:            query.RecommendationControl,
			Impact:              query.RecommendationImpact,
			LearnMoreLink:       query.LearnMoreLink,
			GraphQuery:          query.Query,
			Source:              config.Name,
			MetadataState:       query.RecommendationMetadataState,
			LongDescription:     query.LongDescription,
			PotentialBenefits:   query.PotentialBenefits,
			PgVerified:          query.PgVerified,
			AutomationAvailable: automationAvailable,
			Tags:                query.Tags,
		}
		recommendations = append(recommendations, recommendation)
	}

	plugin := &Plugin{
		Metadata: PluginMetadata{
			Name:        config.Name,
			Version:     config.Version,
			Description: config.Description,
			Author:      config.Author,
			License:     config.License,
			Type:        PluginTypeYaml,
			CommandPath: filePath,
		},
		YamlRecommendations: recommendations,
	}

	return plugin, recommendations, nil
}

// discoverYamlPlugins searches for YAML plugins in configured directories
func discoverYamlPlugins(dirs []string) ([]*Plugin, error) {
	plugins := make([]*Plugin, 0)
	seen := make(map[string]bool)

	for _, dir := range dirs {
		// Check if directory exists
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to stat plugin directory %s: %w", dir, err)
		}

		if !info.IsDir() {
			continue
		}

		// Walk directory looking for .yaml or .yml files
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files we can't access
			}

			if info.IsDir() {
				return nil
			}

			// Only process .yaml and .yml files
			ext := filepath.Ext(path)
			if ext != ".yaml" && ext != ".yml" {
				return nil
			}

			// Try to load as plugin
			plugin, recommendations, err := LoadYamlPlugin(path)
			if err != nil {
				log.Debug().
					Err(err).
					Str("path", path).
					Msg("Skipping file - not a valid YAML plugin")
				return nil
			}

			// Skip duplicates (first found wins)
			if seen[plugin.Metadata.Name] {
				log.Debug().
					Str("plugin", plugin.Metadata.Name).
					Str("path", path).
					Msg("Skipping duplicate YAML plugin")
				return nil
			}
			seen[plugin.Metadata.Name] = true

			plugins = append(plugins, plugin)
			log.Debug().
				Str("plugin", plugin.Metadata.Name).
				Str("path", path).
				Int("queries", len(recommendations)).
				Msg("Discovered YAML plugin")

			return nil
		})

		if err != nil {
			log.Warn().
				Err(err).
				Str("dir", dir).
				Msg("Error walking plugin directory")
		}
	}

	return plugins, nil
}
