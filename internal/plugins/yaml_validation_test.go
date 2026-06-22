// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeYamlPlugin writes content to a temp file and returns its path.
func writeYamlPlugin(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "plugin.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write yaml plugin: %v", err)
	}
	return path
}

func TestLoadYamlPlugin_MissingQuery(t *testing.T) {
	// Query has neither inline 'query' nor 'queryFile'.
	path := writeYamlPlugin(t, `---
name: test-plugin
queries:
  - aprlGuid: guid-001
    description: A recommendation
`)
	_, _, err := LoadYamlPlugin(path)
	if err == nil || !strings.Contains(err.Error(), "must have either 'query' or 'queryFile'") {
		t.Fatalf("expected query-required error, got %v", err)
	}
}

func TestLoadYamlPlugin_MissingAprlGuid(t *testing.T) {
	path := writeYamlPlugin(t, `---
name: test-plugin
queries:
  - description: A recommendation
    query: |
      resources | project id
`)
	_, _, err := LoadYamlPlugin(path)
	if err == nil || !strings.Contains(err.Error(), "aprlGuid") {
		t.Fatalf("expected missing aprlGuid error, got %v", err)
	}
}

func TestLoadYamlPlugin_MissingDescription(t *testing.T) {
	path := writeYamlPlugin(t, `---
name: test-plugin
queries:
  - aprlGuid: guid-001
    query: |
      resources | project id
`)
	_, _, err := LoadYamlPlugin(path)
	if err == nil || !strings.Contains(err.Error(), "description") {
		t.Fatalf("expected missing description error, got %v", err)
	}
}

func TestLoadYamlPlugin_MissingQueryFile(t *testing.T) {
	// queryFile references a file that does not exist.
	dir := t.TempDir()
	path := filepath.Join(dir, "plugin.yaml")
	content := `---
name: test-plugin
queries:
  - aprlGuid: guid-001
    description: A recommendation
    queryFile: missing.kql
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write yaml plugin: %v", err)
	}

	_, _, err := LoadYamlPlugin(path)
	if err == nil || !strings.Contains(err.Error(), "failed to read query file") {
		t.Fatalf("expected query file read error, got %v", err)
	}
}

func TestLoadYamlPlugin_ExternalQueryFileLoaded(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "query.kql"), []byte("resources | project id, name"), 0600); err != nil {
		t.Fatalf("failed to write kql file: %v", err)
	}
	path := filepath.Join(dir, "plugin.yaml")
	content := `---
name: test-plugin
queries:
  - aprlGuid: guid-001
    description: A recommendation
    queryFile: query.kql
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write yaml plugin: %v", err)
	}

	_, recs, err := LoadYamlPlugin(path)
	if err != nil {
		t.Fatalf("LoadYamlPlugin failed: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}
	if !strings.Contains(recs[0].GraphQuery, "project id, name") {
		t.Errorf("expected external kql content to populate GraphQuery, got %q", recs[0].GraphQuery)
	}
}

func TestLoadYamlPlugin_VersionDefaulting(t *testing.T) {
	// Version omitted should default to 1.0.0.
	path := writeYamlPlugin(t, `---
name: test-plugin
queries:
  - aprlGuid: guid-001
    description: A recommendation
    query: |
      resources | project id
`)
	plugin, _, err := LoadYamlPlugin(path)
	if err != nil {
		t.Fatalf("LoadYamlPlugin failed: %v", err)
	}
	if plugin.Metadata.Version != "1.0.0" {
		t.Errorf("expected default version 1.0.0, got %q", plugin.Metadata.Version)
	}
}
