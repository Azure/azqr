// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package config handles copilot configuration persistence.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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
		Model: "claude-sonnet-4.5",
		Mode:  "agent",
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	// Use XDG config directory if available
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	azqrDir := filepath.Join(configDir, "azqr")
	return filepath.Join(azqrDir, "copilot.yaml"), nil
}

// Load reads configuration from disk or returns defaults
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes configuration to disk
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// CycleMode cycles through available modes
func (c *Config) CycleMode() string {
	switch c.Mode {
	case "ask":
		c.Mode = "plan"
	case "plan":
		c.Mode = "agent"
	case "agent":
		c.Mode = "ask"
	default:
		c.Mode = "agent"
	}
	return c.Mode
}
