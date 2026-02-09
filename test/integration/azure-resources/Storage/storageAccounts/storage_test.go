//go:build integration

package storageAccounts

import (
	"testing"

	"github.com/Azure/azqr/test/integration/helpers"
	"github.com/stretchr/testify/require"
)

// TestStorageAccountHTTPSViolation verifies that AZQR detects storage accounts without HTTPS enforcement
func TestStorageAccountHTTPSViolation(t *testing.T) {
	subscriptionID := helpers.RequireEnvVar(t, "AZURE_SUBSCRIPTION_ID")

	// Get path to Terraform fixture
	fixtureDir := helpers.GetFixturePath(t, "scenarios/Storage/storageAccounts/storage-no-https")

	// Provision infrastructure with HTTPS disabled (violation)
	tf := helpers.NewTerraformHelper(t, fixtureDir)
	tf.InitAndApply(nil)

	// Get provisioned resources
	storageAccountName := tf.RequireOutput("storage_account_name")
	resourceGroupName := tf.RequireOutput("resource_group_name")
	httpsEnabled := tf.GetOutput("https_enabled")

	t.Logf("Provisioned storage account: %s in RG: %s (HTTPS enabled: %s)",
		storageAccountName, resourceGroupName, httpsEnabled)

	// Verify HTTPS is actually disabled
	require.Equal(t, "false", httpsEnabled, "HTTPS should be disabled for this test")

	// Run AZQR scan for storage accounts only
	azqr := helpers.NewAZQRHelper(t)
	result := azqr.RunScan(helpers.ScanParams{
		SubscriptionID: subscriptionID,
		ResourceGroup:  resourceGroupName,
		Services:       []string{"st"},                                    // Only scan storage accounts
		Stages:         []string{"-diagnostics", "-advisor", "-defender"}, // Skip stages
	})

	// Assert scan succeeded
	require.True(t, result.Success, "Scan should succeed: %s", result.ErrorMessage)

	// Filter impacted resources for this storage account
	storageImpacted := azqr.FilterByResourceName(result.Impacted, storageAccountName)

	// Should have at least one recommendation
	require.NotEmpty(t, storageImpacted, "Expected recommendations for non-compliant storage account")

	// Check for specific HTTPS recommendation (st-007)
	found := false
	for _, item := range storageImpacted {
		t.Logf("Found recommendation: [%s] %s - %s", item.Impact, item.RecommendationID, item.Recommendation)
		if item.RecommendationID == "st-007" {
			found = true
			require.Contains(t, item.Recommendation, "HTTPS", "Recommendation should mention HTTPS")
		}
	}

	require.True(t, found, "Expected to find recommendation st-007 (HTTPS only) but it was not present")

	t.Logf("✓ AZQR correctly detected HTTPS violation (recommendation st-007)")
}

// TestStorageAccountTLSViolation verifies that AZQR detects storage accounts with old TLS versions
func TestStorageAccountTLSViolation(t *testing.T) {
	subscriptionID := helpers.RequireEnvVar(t, "AZURE_SUBSCRIPTION_ID")

	// Get path to Terraform fixture
	fixtureDir := helpers.GetFixturePath(t, "scenarios/Storage/storageAccounts/storage-old-tls")

	// Provision infrastructure with old TLS version (violation)
	tf := helpers.NewTerraformHelper(t, fixtureDir)
	tf.InitAndApply(nil)

	// Get provisioned resources
	storageAccountName := tf.RequireOutput("storage_account_name")
	resourceGroupName := tf.RequireOutput("resource_group_name")
	tlsVersion := tf.GetOutput("min_tls_version")

	t.Logf("Provisioned storage account: %s in RG: %s (TLS version: %s)",
		storageAccountName, resourceGroupName, tlsVersion)

	// Verify TLS version is old
	require.Equal(t, "TLS1_0", tlsVersion, "TLS version should be TLS1_0 for this test")

	// Run AZQR scan for storage accounts only
	azqr := helpers.NewAZQRHelper(t)
	result := azqr.RunScan(helpers.ScanParams{
		SubscriptionID: subscriptionID,
		ResourceGroup:  resourceGroupName,
		Services:       []string{"st"},                                    // Only scan storage accounts
		Stages:         []string{"-diagnostics", "-advisor", "-defender"}, // Skip stages
	})

	// Assert scan succeeded
	require.True(t, result.Success, "Scan should succeed: %s", result.ErrorMessage)

	// Filter impacted resources for this storage account
	storageImpacted := azqr.FilterByResourceName(result.Impacted, storageAccountName)

	// Should have at least one recommendation
	require.NotEmpty(t, storageImpacted, "Expected recommendations for non-compliant storage account")

	// Check for specific TLS recommendation (st-009)
	found := false
	for _, item := range storageImpacted {
		t.Logf("Found recommendation: [%s] %s - %s", item.Impact, item.RecommendationID, item.Recommendation)
		if item.RecommendationID == "st-009" {
			found = true
			require.Contains(t, item.Recommendation, "TLS", "Recommendation should mention TLS")
		}
	}

	require.True(t, found, "Expected to find recommendation st-009 (TLS >= 1.2) but it was not present")

	t.Logf("✓ AZQR correctly detected TLS violation (recommendation st-009)")
}

// TestStorageAccountImmutableVersioningViolation verifies that AZQR detects storage accounts without immutable storage versioning
func TestStorageAccountImmutableVersioningViolation(t *testing.T) {
	subscriptionID := helpers.RequireEnvVar(t, "AZURE_SUBSCRIPTION_ID")

	// Get path to Terraform fixture
	fixtureDir := helpers.GetFixturePath(t, "scenarios/Storage/storageAccounts/storage-no-immutable-versioning")

	// Provision infrastructure without immutable storage versioning (violation)
	tf := helpers.NewTerraformHelper(t, fixtureDir)
	tf.InitAndApply(nil)

	// Get provisioned resources
	storageAccountName := tf.RequireOutput("storage_account_name")
	resourceGroupName := tf.RequireOutput("resource_group_name")

	t.Logf("Provisioned storage account: %s in RG: %s (immutable versioning: disabled)",
		storageAccountName, resourceGroupName)

	// Run AZQR scan for storage accounts only
	azqr := helpers.NewAZQRHelper(t)
	result := azqr.RunScan(helpers.ScanParams{
		SubscriptionID: subscriptionID,
		ResourceGroup:  resourceGroupName,
		Services:       []string{"st"},                                    // Only scan storage accounts
		Stages:         []string{"-diagnostics", "-advisor", "-defender"}, // Skip stages
	})

	// Assert scan succeeded
	require.True(t, result.Success, "Scan should succeed: %s", result.ErrorMessage)

	// Filter impacted resources for this storage account
	storageImpacted := azqr.FilterByResourceName(result.Impacted, storageAccountName)

	// Should have at least one recommendation
	require.NotEmpty(t, storageImpacted, "Expected recommendations for non-compliant storage account")

	// Check for specific immutable versioning recommendation (st-010)
	found := false
	for _, item := range storageImpacted {
		t.Logf("Found recommendation: [%s] %s - %s", item.Impact, item.RecommendationID, item.Recommendation)
		if item.RecommendationID == "st-010" {
			found = true
			require.Contains(t, item.Recommendation, "immutable", "Recommendation should mention immutable storage")
		}
	}

	require.True(t, found, "Expected to find recommendation st-010 (immutable storage versioning) but it was not present")

	t.Logf("✓ AZQR correctly detected immutable versioning violation (recommendation st-010)")
}
