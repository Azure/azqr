//go:build integration

package storageAccounts

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/test/integration/helpers"
	"github.com/stretchr/testify/require"
)

// TestStorageAccountDiagnosticSettings verifies that AZQR correctly detects
// storage accounts without diagnostic settings (st-001) and does not flag
// storage accounts that have diagnostic settings enabled.
func TestStorageAccountDiagnosticSettings(t *testing.T) {
	subscriptionID := helpers.RequireEnvVar(t, "AZURE_SUBSCRIPTION_ID")

	// Get path to Terraform fixture that deploys two storage accounts:
	// one with diagnostic settings and one without
	fixtureDir := helpers.GetFixturePath(t, "scenarios/Storage/storageAccounts/storage-diagnostic-settings")

	// Provision infrastructure
	tf := helpers.NewTerraformHelper(t, fixtureDir)
	tf.InitAndApply(nil)

	// Get provisioned resource names
	storageWithDiag := tf.RequireOutput("storage_account_with_diag_name")
	storageWithoutDiag := tf.RequireOutput("storage_account_without_diag_name")
	resourceGroupName := tf.RequireOutput("resource_group_name")

	t.Logf("Provisioned storage account WITH diagnostic settings: %s", storageWithDiag)
	t.Logf("Provisioned storage account WITHOUT diagnostic settings: %s", storageWithoutDiag)
	t.Logf("Resource group: %s", resourceGroupName)

	// Run AZQR scan with diagnostics stage enabled (only skip advisor and defender)
	azqr := helpers.NewAZQRHelper(t)
	result := azqr.RunScan(models.ScanArgs{
		Subscriptions:  []string{subscriptionID},
		ResourceGroups: []string{resourceGroupName},
		Services:       []string{"st"},                    // Only scan storage accounts
		Stages:         []string{"-advisor", "-defender"}, // Keep diagnostics stage enabled
	})

	// Assert scan succeeded
	require.True(t, result.Success, "Scan should succeed: %s", result.ErrorMessage)

	// --- Validate storage account WITHOUT diagnostic settings triggers st-001 ---
	withoutDiagImpacted := azqr.FilterByResourceName(result.Impacted, storageWithoutDiag)

	foundSt001 := false
	for _, item := range withoutDiagImpacted {
		t.Logf("Recommendation for %s: [%s] %s - %s", storageWithoutDiag, item.Impact, item.RecommendationID, item.Recommendation)
		if item.RecommendationID == "st-001" {
			foundSt001 = true
		}
	}

	require.True(t, foundSt001,
		"Expected recommendation st-001 (diagnostic settings) for storage account '%s' without diagnostic settings",
		storageWithoutDiag)

	t.Logf("✓ AZQR correctly detected missing diagnostic settings for '%s' (recommendation st-001)", storageWithoutDiag)

	// --- Validate storage account WITH diagnostic settings does NOT trigger st-001 ---
	withDiagImpacted := azqr.FilterByResourceName(result.Impacted, storageWithDiag)

	foundSt001ForCompliant := false
	for _, item := range withDiagImpacted {
		t.Logf("Recommendation for %s: [%s] %s - %s", storageWithDiag, item.Impact, item.RecommendationID, item.Recommendation)
		if item.RecommendationID == "st-001" {
			foundSt001ForCompliant = true
		}
	}

	require.False(t, foundSt001ForCompliant,
		"Storage account '%s' has diagnostic settings enabled and should NOT have recommendation st-001",
		storageWithDiag)

	t.Logf("✓ AZQR correctly skipped st-001 for '%s' (diagnostic settings enabled)", storageWithDiag)
}
