//go:build integration

package resources

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/test/integration/helpers"
	"github.com/stretchr/testify/require"
)

// TestResourceNoTagsViolation verifies that AZQR detects resources without tags (resources-001)
func TestResourceNoTagsViolation(t *testing.T) {
	subscriptionID := helpers.RequireEnvVar(t, "AZURE_SUBSCRIPTION_ID")

	// Get path to Terraform fixture
	fixtureDir := helpers.GetFixturePath(t, "scenarios/Resources/resource-no-tags")

	// Provision a storage account without any tags (violation)
	tf := helpers.NewTerraformHelper(t, fixtureDir)
	tf.InitAndApply(nil)

	// Get provisioned resources
	storageAccountName := tf.RequireOutput("storage_account_name")
	resourceGroupName := tf.RequireOutput("resource_group_name")

	t.Logf("Provisioned storage account: %s in RG: %s (no tags)",
		storageAccountName, resourceGroupName)

	// Run AZQR scan for storage accounts only
	azqr := helpers.NewAZQRHelper(t)
	result := azqr.RunScan(models.ScanArgs{
		Subscriptions:  []string{subscriptionID},
		ResourceGroups: []string{resourceGroupName},
		Services:       []string{"st"},                                    // Scan resources and storage accounts
		Stages:         []string{"-diagnostics", "-advisor", "-defender"}, // Skip stages
	})

	// Assert scan succeeded
	require.True(t, result.Success, "Scan should succeed: %s", result.ErrorMessage)

	// Filter impacted resources for this storage account
	storageImpacted := azqr.FilterByResourceName(result.Impacted, storageAccountName)

	// Should have at least one recommendation
	require.NotEmpty(t, storageImpacted, "Expected recommendations for storage account without tags")

	// Check for specific tags recommendation (resources-001)
	found := false
	for _, item := range storageImpacted {
		t.Logf("Found recommendation: [%s] %s - %s", item.Impact, item.RecommendationID, item.Recommendation)
		if item.RecommendationID == "resources-001" {
			found = true
			require.Contains(t, item.Recommendation, "tags", "Recommendation should mention tags")
		}
	}

	require.True(t, found, "Expected to find recommendation resources-001 (resource should have tags) but it was not present")

	t.Logf("✓ AZQR correctly detected missing tags violation (recommendation resources-001)")
}
