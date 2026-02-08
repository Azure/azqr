package helpers

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// RequireEnvVar retrieves an environment variable and fails the test if it's not set.
// This is used to enforce that required Azure credentials are configured before running integration tests.
func RequireEnvVar(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	if value == "" {
		t.Skipf("Environment variable %s not set - skipping integration test", key)
	}
	return value
}

// GetFixturePath returns the absolute path to a Terraform fixture directory.
// It constructs the path relative to the test/fixtures/terraform directory.
//
// Example:
//
//	GetFixturePath(t, "baseline/storage-account-compliant")
//	// Returns: /path/to/azqr/test/fixtures/terraform/baseline/storage-account-compliant
func GetFixturePath(t *testing.T, relativePath string) string {
	t.Helper()

	// Get the directory of the current test file
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("Failed to get caller information")
	}

	// Navigate from test/integration/* to test/fixtures/terraform
	testDir := filepath.Dir(filename)
	for filepath.Base(testDir) != "integration" && testDir != "/" {
		testDir = filepath.Dir(testDir)
	}

	fixturesDir := filepath.Join(filepath.Dir(testDir), "fixtures", "terraform", relativePath)
	return fixturesDir
}
