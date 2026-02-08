//go:build integration

package helpers

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	loggerInitialized bool
	loggerMutex       sync.Mutex
)

// AZQRHelper wraps AZQR pipeline execution for integration tests
// Uses direct package imports instead of subprocess execution for better performance and debugging
type AZQRHelper struct {
	t       *testing.T
	scanner *pipeline.Scanner
}

// ScanParams defines parameters for running an AZQR scan
type ScanParams struct {
	SubscriptionID string
	ResourceGroup  string
	Services       []string // Service keys to scan (e.g., "st", "aks", "vm")
	Stages         []string // Stages to run (e.g., "diagnostics", "advisor", "defender", "costs")
	Filters        []string
}

// ScanResult represents the results from an AZQR scan
type ScanResult struct {
	Impacted     []*models.GraphResult
	Success      bool
	ErrorMessage string
}

// NewAZQRHelper creates a new AZQR helper
func NewAZQRHelper(t *testing.T) *AZQRHelper {
	t.Helper()

	// Initialize logger once for all tests to enable streaming output
	initializeLogger()

	return &AZQRHelper{
		t:       t,
		scanner: &pipeline.Scanner{},
	}
}

// initializeLogger sets up zerolog for test output with streaming (not buffered)
// This ensures logs appear in real-time during test execution instead of being buffered
func initializeLogger() {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	if loggerInitialized {
		return
	}

	// Configure zerolog with ConsoleWriter for human-readable streaming output
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,    // Write to stderr for proper test output
		TimeFormat: time.RFC3339, // Human-readable timestamps
		NoColor:    false,        // Enable colors for better readability
	}

	// Set global logger with immediate output (no buffering)
	log.Logger = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel) // Enable info logs for integration tests

	loggerInitialized = true
}

// RunScan executes an AZQR scan with the given parameters using direct package imports
func (h *AZQRHelper) RunScan(params ScanParams) *ScanResult {
	h.t.Helper()

	// Create scan parameters
	scanParams := models.NewScanParams()

	if params.SubscriptionID != "" {
		scanParams.Subscriptions = []string{params.SubscriptionID}
	}

	if params.ResourceGroup != "" {
		scanParams.ResourceGroups = []string{params.ResourceGroup}
	}

	// Load filters with scanner keys to properly initialize resource type filtering
	// This is critical - LoadFilters populates the filter's resource type map from scanners
	if len(params.Services) > 0 {
		scanParams.ScannerKeys = params.Services
		scanParams.Filters = models.LoadFilters("", params.Services)
	}

	// Configure stages (enable/disable specific stages)
	if len(params.Stages) > 0 {
		scanParams.Stages.ConfigureStages(params.Stages)
	}

	scanParams.Mask = true

	h.t.Logf("Running AZQR scan with subscriptions: %v, resource groups: %v, services: %v, enabled stages: %v",
		scanParams.Subscriptions, scanParams.ResourceGroups, scanParams.ScannerKeys, scanParams.Stages.GetEnabledStages())

	// Build and execute the pipeline
	builder := pipeline.NewScanPipelineBuilder()
	scanCtx := pipeline.NewScanContext(scanParams)

	// Validate that graph stage is enabled for regular scans
	if err := scanParams.Stages.ValidateGraphStageEnabled(); err != nil {
		result := &ScanResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("configuration error: %v", err),
		}
		h.t.Logf("AZQR configuration error: %s", result.ErrorMessage)
		return result
	}

	pipe := builder.BuildDefault()
	err := pipe.Execute(scanCtx)

	result := &ScanResult{
		Success: err == nil,
	}

	if err != nil {
		result.ErrorMessage = fmt.Sprintf("scan failed: %v", err)
		h.t.Logf("AZQR scan error: %s", result.ErrorMessage)
		return result
	}

	// Extract results from ReportData
	result.Impacted = scanCtx.ReportData.Graph

	h.t.Logf("AZQR scan completed successfully with %d impacted resources", len(result.Impacted))

	return result
}

// FilterByResourceName filters impacted resources by resource name
func (h *AZQRHelper) FilterByResourceName(impacted []*models.GraphResult, resourceName string) []*models.GraphResult {
	h.t.Helper()

	filtered := make([]*models.GraphResult, 0)
	for _, item := range impacted {
		if strings.Contains(strings.ToLower(item.Name), strings.ToLower(resourceName)) {
			filtered = append(filtered, item)
		}
	}

	h.t.Logf("Filtered to %d impacted resources for resource name '%s'", len(filtered), resourceName)
	return filtered
}

// HasRecommendationID checks if a specific recommendation ID exists in the results
func (h *AZQRHelper) HasRecommendationID(impacted []*models.GraphResult, recommendationID string) bool {
	h.t.Helper()

	for _, item := range impacted {
		if item.RecommendationID == recommendationID {
			return true
		}
	}
	return false
}

// GetRecommendationIDs returns all unique recommendation IDs from the results
func (h *AZQRHelper) GetRecommendationIDs(impacted []*models.GraphResult) []string {
	h.t.Helper()

	ids := make(map[string]bool)
	for _, item := range impacted {
		ids[item.RecommendationID] = true
	}

	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}

	return result
}
