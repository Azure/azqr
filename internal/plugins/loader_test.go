// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPluginDirs(t *testing.T) {
	dirs := getPluginDirs()

	assert.Len(t, dirs, 2)
	assert.Equal(t, "./plugins", dirs[1])
	assert.True(t, filepath.IsAbs(dirs[0]) || dirs[0] != "",
		"first plugin dir should resolve from the user home directory")
	assert.Equal(t, filepath.Join(".azqr", "plugins"), filepath.Join(filepath.Base(filepath.Dir(dirs[0])), filepath.Base(dirs[0])))
}

func TestLoadAll_NoPluginDirsReturnsNil(t *testing.T) {
	// Isolate from any real plugin directories.
	t.Setenv("HOME", t.TempDir())
	t.Chdir(t.TempDir())

	err := LoadAll()
	assert.NoError(t, err)
}

func TestLoadAll_RegistersDiscoveredPlugin(t *testing.T) {
	// Isolate home so ~/.azqr/plugins cannot interfere.
	t.Setenv("HOME", t.TempDir())

	// Create a ./plugins directory with a valid YAML plugin relative to the cwd.
	workDir := t.TempDir()
	pluginDir := filepath.Join(workDir, "plugins")
	if err := os.MkdirAll(pluginDir, 0750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	const pluginName = "loader-test-plugin"
	yamlContent := `---
name: ` + pluginName + `
version: 2.1.0
description: Loader discovery test plugin
queries:
  - aprlGuid: loader-guid-001
    description: Loader test recommendation
    query: |
      resources
      | project id, name
`
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte(yamlContent), 0600); err != nil {
		t.Fatalf("failed to write plugin file: %v", err)
	}

	t.Chdir(workDir)

	if err := LoadAll(); err != nil {
		t.Fatalf("LoadAll returned error: %v", err)
	}

	plugin, ok := GetRegistry().Get(pluginName)
	assert.True(t, ok, "discovered plugin should be registered")
	assert.Equal(t, "2.1.0", plugin.Metadata.Version)
	assert.Equal(t, PluginTypeYaml, plugin.Metadata.Type)
}
