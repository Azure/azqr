// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package models

import (
	"fmt"
	"strconv"
	"strings"
)

// OptionSpec defines the schema for a stage option
type OptionSpec struct {
	Type        string // "bool", "int", "float64", "string"
	Default     any
	Description string
}

// StageOptionRegistry defines allowed options for each stage
var stageOptionRegistry = map[string]map[string]OptionSpec{
	StageNameCost: {
		"previousMonth": {
			Type:        "bool",
			Default:     false,
			Description: "Scan costs for the previous calendar month (UTC) instead of default 3-month period",
		},
	},
}

// ParseAndValidateStageParams parses and validates stage parameters against the registry.
// Returns a map of stage -> options with typed values according to the schema.
// Errors on unknown stage, unknown key for a stage, or type mismatch.
func ParseAndValidateStageParams(params []string) (map[string]map[string]any, error) {
	options := make(map[string]map[string]any)

	for _, param := range params {
		param = strings.TrimSpace(param)
		if param == "" {
			continue
		}

		stageKey, value, ok := strings.Cut(param, "=")
		if !ok {
			return nil, fmt.Errorf("stage param must be in the form stage.key=value: %s", param)
		}

		stage, key, ok := strings.Cut(stageKey, ".")
		if !ok || stage == "" || key == "" {
			return nil, fmt.Errorf("stage param must be in the form stage.key=value: %s", param)
		}

		// Validate stage exists in registry
		stageSpecs, exists := stageOptionRegistry[stage]
		if !exists {
			return nil, fmt.Errorf("unknown stage: %s", stage)
		}

		// Validate key exists for this stage
		spec, exists := stageSpecs[key]
		if !exists {
			return nil, fmt.Errorf("unknown option %q for stage %q", key, stage)
		}

		// Parse value according to spec type
		parsed, err := parseValue(value, spec.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s.%s: %w", stage, key, err)
		}

		if _, exists := options[stage]; !exists {
			options[stage] = make(map[string]any)
		}
		options[stage][key] = parsed
	}

	return options, nil
}

// ApplyStageParams parses, validates and applies stage parameters.
func (sc *StageConfigs) ApplyStageParams(params []string) error {
	options, err := ParseAndValidateStageParams(params)
	if err != nil {
		return err
	}

	for stageName, stageOptions := range options {
		if err := sc.SetStageOptions(stageName, stageOptions); err != nil {
			return err
		}
	}
	return nil
}

// SetStageOptions sets or merges options for a stage.
func (sc *StageConfigs) SetStageOptions(stageName string, stageOptions map[string]any) error {
	if !isValidStageName(stageName) {
		return fmt.Errorf("unknown stage name: %s", stageName)
	}

	cfg := sc.stages[stageName]
	if cfg == nil {
		cfg = &BaseStageConfig{}
		sc.stages[stageName] = cfg
	}

	if cfg.Options == nil {
		cfg.Options = make(map[string]any)
	}

	for key, value := range stageOptions {
		cfg.Options[key] = value
	}

	return nil
}

// GetStageOptions returns the raw options for a stage.
func (sc *StageConfigs) GetStageOptions(stageName string) map[string]any {
	cfg, exists := sc.stages[stageName]
	if !exists || cfg == nil {
		return nil
	}
	return cfg.Options
}

func parseValue(value string, typeStr string) (any, error) {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"'")

	switch typeStr {
	case "bool":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("expected bool, got %q", value)
		}
		return parsed, nil

	case "int":
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("expected int, got %q", value)
		}
		return parsed, nil

	case "float64":
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("expected float64, got %q", value)
		}
		return parsed, nil

	case "string":
		return value, nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", typeStr)
	}
}
