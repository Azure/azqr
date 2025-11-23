// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"github.com/Azure/azqr/internal/models"
	"github.com/spf13/cobra"
)

// PluginType represents the type of plugin
type PluginType int

const (
	// PluginTypeYaml represents a YAML-based plugin with graph queries
	PluginTypeYaml PluginType = iota
	// PluginTypeInternal represents an internal Go-based plugin
	PluginTypeInternal
)

// PluginMetadata contains information about a plugin
type PluginMetadata struct {
	// Name is the unique identifier for the plugin (e.g., "myservice")
	Name string
	// Version is the semantic version of the plugin
	Version string
	// Description is a short description of what the plugin does
	Description string
	// Author is the plugin author/maintainer
	Author string
	// License is the plugin license (e.g., "MIT", "Apache-2.0")
	License string
	// Type indicates if this is a built-in or external plugin
	Type PluginType
	// Command is the full path to the external command (for external plugins only)
	CommandPath string
	// ColumnMetadata defines the columns and their filter types for the viewer
	ColumnMetadata []ColumnMetadata
}

// Plugin represents a loaded plugin with its scanner and command
type Plugin struct {
	// Metadata contains information about the plugin
	Metadata PluginMetadata
	// InternalScanner is the internal plugin scanner (for internal plugins)
	InternalScanner InternalPluginScanner
	// YamlRecommendations holds APRL recommendations for YAML plugins
	YamlRecommendations []models.GraphRecommendation
	// Command is the Cobra command for this plugin (optional)
	Command *cobra.Command
}

// YamlPluginQuery represents a single query definition in a YAML plugin
type YamlPluginQuery struct {
	// Description of what this query checks
	Description string `yaml:"description"`
	// AprlGuid is a unique identifier for the recommendation
	AprlGuid string `yaml:"aprlGuid"`
	// RecommendationTypeId from Azure Advisor (optional)
	RecommendationTypeId *string `yaml:"recommendationTypeId"`
	// RecommendationControl category (e.g., HighAvailability, Security)
	RecommendationControl string `yaml:"recommendationControl"`
	// RecommendationImpact level (High, Medium, Low)
	RecommendationImpact string `yaml:"recommendationImpact"`
	// RecommendationResourceType Azure resource type
	RecommendationResourceType string `yaml:"recommendationResourceType"`
	// RecommendationMetadataState (Active, Preview, Deprecated)
	RecommendationMetadataState string `yaml:"recommendationMetadataState"`
	// LongDescription detailed explanation
	LongDescription string `yaml:"longDescription"`
	// PotentialBenefits of implementing the recommendation
	PotentialBenefits string `yaml:"potentialBenefits"`
	// PgVerified indicates if verified by product group
	PgVerified bool `yaml:"pgVerified"`
	// AutomationAvailable indicates if automation is available
	AutomationAvailable bool `yaml:"automationAvailable"`
	// Tags for categorization
	Tags []string `yaml:"tags"`
	// LearnMoreLink for additional information
	LearnMoreLink []struct {
		Name string `yaml:"name"`
		Url  string `yaml:"url"`
	} `yaml:"learnMoreLink"`
	// Query is the inline KQL query (optional if using external .kql file)
	Query string `yaml:"query,omitempty"`
	// QueryFile is the path to external .kql file (optional if using inline query)
	QueryFile string `yaml:"queryFile,omitempty"`
}

// YamlPluginConfig represents the structure of a YAML plugin file
type YamlPluginConfig struct {
	// Name of the plugin
	Name string `yaml:"name"`
	// Version of the plugin
	Version string `yaml:"version"`
	// Description of the plugin
	Description string `yaml:"description"`
	// Author of the plugin
	Author string `yaml:"author,omitempty"`
	// License of the plugin
	License string `yaml:"license,omitempty"`
	// Queries list of graph queries to execute
	Queries []YamlPluginQuery `yaml:"queries,omitempty"`
}

// FilterType represents the type of filter for a column
type FilterType string

const (
	// FilterTypeNone indicates no filtering is available for this column
	FilterTypeNone FilterType = "none"
	// FilterTypeDropdown indicates a dropdown filter with distinct values (≤20 unique values)
	FilterTypeDropdown FilterType = "dropdown"
	// FilterTypeSearch indicates a search/text filter (>20 unique values)
	FilterTypeSearch FilterType = "search"
)

// FlagProvider is an optional interface that plugins can implement to register
// plugin-specific CLI flags onto their cobra command.
type FlagProvider interface {
	RegisterFlags(cmd *cobra.Command)
}

// ColumnMetadata defines filtering behavior for a column in the viewer
type ColumnMetadata struct {
	Name       string     `json:"name"`       // Display name (e.g., "Latest Month Emissions")
	DataKey    string     `json:"dataKey"`    // JSON data key (e.g., "latestMonthEmissions")
	FilterType FilterType `json:"filterType"` // "none", "dropdown", or "search"
}

// ExternalPluginOutput represents the output from a plugin execution
type ExternalPluginOutput struct {
	Metadata    PluginMetadata `json:"metadata"`
	SheetName   string         `json:"sheet_name"`      // Name for Excel sheet
	Description string         `json:"description"`     // Description of data
	Table       [][]string     `json:"table"`           // Headers + data rows
	Error       string         `json:"error,omitempty"` // Error if failed
}
