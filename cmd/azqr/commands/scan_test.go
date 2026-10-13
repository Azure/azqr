// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/Azure/azqr/internal/store"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// newScanTestCommand builds a bare *cobra.Command with the --store-path flag runPersistedOrPlain
// reads, mirroring the scan command's flag without invoking the CLI end-to-end.
func newScanTestCommand(t *testing.T, storePath string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "scan"}
	cmd.Flags().String("store-path", storePath, "")
	return cmd
}

// captureLogOutput redirects the global zerolog logger to a buffer for the duration of the
// calling test, restoring it on cleanup.
func captureLogOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	original := log.Logger
	log.Logger = zerolog.New(&buf)
	t.Cleanup(func() { log.Logger = original })
	return &buf
}

func TestRunPersistedOrPlainLogsSessionIDWhenSessionCreated(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "azqr.db")
	cmd := newScanTestCommand(t, storePath)
	buf := captureLogOutput(t)

	stubSummary := &findings.Summary{}
	summary, err := runPersistedOrPlain(cmd, store.SessionKindFull, store.SessionTriggerCLI, &models.ScanParams{
		Subscriptions: []string{"sub-1"},
	}, func(sc *pipeline.Scanner) (*findings.Summary, error) {
		return stubSummary, nil
	})

	if err != nil {
		t.Fatalf("runPersistedOrPlain() error = %v", err)
	}
	if summary != stubSummary {
		t.Fatalf("runPersistedOrPlain() returned summary %v, want the stub summary unchanged", summary)
	}

	output := buf.String()
	if !strings.Contains(output, "sessionId") {
		t.Errorf("expected log output to contain the session ID, got: %s", output)
	}
	if !strings.Contains(output, "Scan session recorded") {
		t.Errorf("expected log output to contain the session-recorded message, got: %s", output)
	}
}

func TestRunPersistedOrPlainLogsNothingWhenStoreUnavailable(t *testing.T) {
	// An unwritable/nonexistent directory makes store.OpenForScan return nil, exercising the
	// existing unpersisted-scan fallback.
	cmd := newScanTestCommand(t, filepath.Join(t.TempDir(), "nonexistent-dir", "azqr.db"))
	buf := captureLogOutput(t)

	stubSummary := &findings.Summary{}
	summary, err := runPersistedOrPlain(cmd, store.SessionKindFull, store.SessionTriggerCLI, &models.ScanParams{
		Subscriptions: []string{"sub-1"},
	}, func(sc *pipeline.Scanner) (*findings.Summary, error) {
		return stubSummary, nil
	})

	if err != nil {
		t.Fatalf("runPersistedOrPlain() error = %v", err)
	}
	if summary != stubSummary {
		t.Fatalf("runPersistedOrPlain() returned summary %v, want the stub summary unchanged", summary)
	}

	if strings.Contains(buf.String(), "sessionId") {
		t.Errorf("expected no session-id log line when the store is unavailable, got: %s", buf.String())
	}
}

func TestScanCommandHasNoNewFlags(t *testing.T) {
	scanCmd := findCommand(rootCmd, "scan")
	if scanCmd == nil {
		t.Fatal("scan command should exist")
	}

	// This list mirrors the flags already registered in scan.go's init() (profiling flags are
	// conditional and excluded); the session-ID logging change in runPersistedOrPlain must not
	// add, remove, or rename any persistent flag on the scan command.
	expectedFlags := []string{
		"management-group-id", "subscription-id", "resource-group", "stages", "stage-param",
		"plugin", "xlsx", "json", "csv", "output-name", "mask", "filters", "fail-on", "sarif",
	}
	for _, name := range expectedFlags {
		if scanCmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("expected flag %q not found in scan command", name)
		}
	}

	usage := scanCmd.PersistentFlags().FlagUsages()
	actualCount := len(flagUsagePattern.FindAllString(usage, -1))
	if actualCount != len(expectedFlags) {
		t.Errorf("expected scan command to have exactly %d persistent flags, found %d:\n%s",
			len(expectedFlags), actualCount, usage)
	}
}

var flagUsagePattern = regexp.MustCompile(`(?m)^\s*(-\w, )?--[\w-]+`)
