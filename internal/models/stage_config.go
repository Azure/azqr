package models

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// Stage name constants
const (
	StageNameGraph                   = "graph"
	StageNameAdvisor                 = "advisor"
	StageNameDefender                = "defender"
	StageNameDefenderRecommendations = "defender-recommendations"
	StageNameArc                     = "arc"
	StageNamePolicy                  = "policy"
	StageNameCost                    = "cost"
	StageNameDiagnostics             = "diagnostics"
)

// allStages defines all available stages and whether they are enabled by default
// This map serves as both the registry of valid stages and their default configuration
var allStages = map[string]bool{
	StageNameGraph:                   true,  // Always enabled (mandatory)
	StageNameDiagnostics:             true,  // Enabled by default
	StageNameAdvisor:                 true,  // Enabled by default
	StageNameDefender:                true,  // Enabled by default
	StageNameDefenderRecommendations: false, // Disabled by default
	StageNameArc:                     false, // Disabled by default
	StageNamePolicy:                  false, // Disabled by default
	StageNameCost:                    false, // Disabled by default
}

// BaseStageConfig contains configuration for a stage
type BaseStageConfig struct {
	Enabled bool
}

// StageConfigs manages all stage configurations using a map-based approach
type StageConfigs struct {
	stages map[string]*BaseStageConfig
}

// NewStageConfigs creates an empty StageConfigs instance
func NewStageConfigs() *StageConfigs {
	return &StageConfigs{
		stages: map[string]*BaseStageConfig{},
	}
}

// NewStageConfigsWithDefaults creates StageConfigs with backwards-compatible defaults
func NewStageConfigsWithDefaults() *StageConfigs {
	configs := &StageConfigs{
		stages: make(map[string]*BaseStageConfig),
	}
	for stageName, enabled := range allStages {
		configs.stages[stageName] = &BaseStageConfig{Enabled: enabled}
	}
	return configs
}

// IsStageEnabled checks if a stage is enabled by name
func (sc *StageConfigs) IsStageEnabled(stageName string) bool {
	cfg, exists := sc.stages[stageName]
	return exists && cfg != nil && cfg.Enabled
}

// EnableStage enables a stage by name and creates the config if it doesn't exist
func (sc *StageConfigs) EnableStage(stageName string) error {
	// Validate stage name
	if !isValidStageName(stageName) {
		return fmt.Errorf("unknown stage name: %s", stageName)
	}

	if sc.stages[stageName] == nil {
		sc.stages[stageName] = &BaseStageConfig{}
	}
	sc.stages[stageName].Enabled = true
	return nil
}

// DisableStage disables a stage by name
func (sc *StageConfigs) DisableStage(stageName string) error {
	// Validate stage name
	if !isValidStageName(stageName) {
		return fmt.Errorf("unknown stage name: %s", stageName)
	}

	if sc.stages[stageName] != nil {
		sc.stages[stageName].Enabled = false
	}
	return nil
}

// GetEnabledStages returns a list of enabled stage names
func (sc *StageConfigs) GetEnabledStages() []string {
	var enabled []string
	for stageName := range sc.stages {
		if sc.IsStageEnabled(stageName) {
			enabled = append(enabled, stageName)
		}
	}
	return enabled
}

// ValidateGraphStageEnabled ensures the graph stage is enabled for regular scans
// Returns an error if the graph stage is disabled
func (sc *StageConfigs) ValidateGraphStageEnabled() error {
	if !sc.IsStageEnabled(StageNameGraph) {
		return fmt.Errorf("graph stage is mandatory for regular scans and cannot be disabled")
	}
	return nil
}

func (sc *StageConfigs) ConfigureStages(stageNames []string) {
	// If stages are explicitly specified, determine the model to use
	if len(stageNames) > 0 {
		// Parse and enable/disable each specified stage
		for _, stageName := range stageNames {
			// Handle comma-separated values within a single flag value
			stages := strings.Split(stageName, ",")
			for _, s := range stages {
				s = strings.TrimSpace(strings.ToLower(s))
				if s == "" {
					continue
				}

				// Check if stage should be disabled (prefix with '-')
				if strings.HasPrefix(s, "-") {
					// Remove the '-' prefix and disable the stage
					stageName := strings.TrimPrefix(s, "-")
					if err := sc.DisableStage(stageName); err != nil {
						log.Fatal().Err(err).Msgf("Invalid stage name: %s", stageName)
					}
				} else {
					// Enable the stage
					if err := sc.EnableStage(s); err != nil {
						log.Fatal().Err(err).Msgf("Invalid stage name: %s", s)
					}
				}
			}
		}
	}
}

// isValidStageName checks if a stage name is valid
func isValidStageName(name string) bool {
	_, exists := allStages[name]
	return exists
}
