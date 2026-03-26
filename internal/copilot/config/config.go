// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package config handles copilot configuration persistence.
package config

// Config holds user preferences for the copilot TUI
type Config struct {
	// Model is the AI model to use
	Model string `yaml:"model"`

	// Mode is the interaction mode (ask, plan, agent)
	Mode string `yaml:"mode"`
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Model: "claude-sonnet-4.6",
		Mode:  "agent",
	}
}
