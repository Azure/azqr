// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"context"
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// mockInternalScanner is a minimal InternalPluginScanner implementation for tests.
type mockInternalScanner struct {
	metadata PluginMetadata
}

func (m *mockInternalScanner) Scan(_ context.Context, _ azcore.TokenCredential, _ map[string]string, _ *models.ScanParams) ([]ExternalPluginOutput, error) {
	return nil, nil
}

func (m *mockInternalScanner) GetMetadata() PluginMetadata {
	return m.metadata
}

// mockFlagProviderScanner additionally implements FlagProvider.
type mockFlagProviderScanner struct {
	mockInternalScanner
	flagRegistered bool
}

func (m *mockFlagProviderScanner) RegisterFlags(cmd *cobra.Command) {
	m.flagRegistered = true
	cmd.Flags().String("custom-flag", "", "a plugin-specific flag")
}

func TestRegisterInternalPlugin(t *testing.T) {
	const name = "test-internal-register"
	scanner := &mockInternalScanner{
		metadata: PluginMetadata{
			Name:        name,
			Version:     "1.0.0",
			Description: "internal test plugin",
			Type:        PluginTypeInternal,
		},
	}

	RegisterInternalPlugin(name, scanner)

	// Retrievable via the internal plugin registry.
	got, exists := GetInternalPlugin(name)
	assert.True(t, exists)
	assert.Equal(t, scanner, got)

	// Registered with the global plugin registry, with a command attached.
	plugin, ok := GetRegistry().Get(name)
	assert.True(t, ok)
	assert.NotNil(t, plugin.Command)
	assert.Equal(t, scanner, plugin.InternalScanner)
	assert.Equal(t, "internal test plugin", plugin.Metadata.Description)
}

func TestRegisterInternalPlugin_FlagProvider(t *testing.T) {
	const name = "test-internal-flagprovider"
	scanner := &mockFlagProviderScanner{
		mockInternalScanner: mockInternalScanner{
			metadata: PluginMetadata{
				Name:        name,
				Version:     "1.0.0",
				Description: "flag provider test plugin",
				Type:        PluginTypeInternal,
			},
		},
	}

	RegisterInternalPlugin(name, scanner)

	assert.True(t, scanner.flagRegistered, "RegisterFlags should be invoked for FlagProvider scanners")

	plugin, ok := GetRegistry().Get(name)
	assert.True(t, ok)
	assert.NotNil(t, plugin.Command.Flags().Lookup("custom-flag"))
}

func TestGetInternalPlugin_NotFound(t *testing.T) {
	got, exists := GetInternalPlugin("does-not-exist")
	assert.False(t, exists)
	assert.Nil(t, got)
}

func TestCreatePluginCommand(t *testing.T) {
	cmd := createPluginCommand("sample", "sample description")

	assert.Equal(t, "sample", cmd.Use)
	assert.Equal(t, "sample description", cmd.Short)

	// All standard scan flags must be present.
	standardFlags := []string{
		"management-group-id",
		"subscription-id",
		"resource-group",
		"xlsx",
		"json",
		"csv",
		"stdout",
		"output-name",
		"mask",
		"filters",
	}
	for _, f := range standardFlags {
		assert.NotNilf(t, cmd.Flags().Lookup(f), "expected flag %q to be registered", f)
	}
}
