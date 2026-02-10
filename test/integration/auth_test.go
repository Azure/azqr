//go:build integration

package integration

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/test/integration/helpers"
	"github.com/stretchr/testify/require"
)

// TestAuthentication verifies that AZQR can authenticate to Azure
func TestAuthentication(t *testing.T) {
	// Require Azure credentials
	subscriptionID := helpers.RequireEnvVar(t, "AZURE_SUBSCRIPTION_ID")

	t.Logf("Testing authentication with subscription: %s", subscriptionID)

	azqr := helpers.NewAZQRHelper(t)

	// Run a simple scan to verify authentication works
	result := azqr.RunScan(models.ScanArgs{
		Subscriptions: []string{subscriptionID},
		Services:      []string{"rg"},
		Stages:        []string{"-diagnostics", "-advisor", "-defender"}, // Skip stages to speed up test
	})

	require.True(t, result.Success, "Authentication failed: %s", result.ErrorMessage)
	t.Log("✓ Authentication successful")
}

// TestSubscriptionDiscovery verifies that AZQR can discover subscriptions
func TestSubscriptionDiscovery(t *testing.T) {
	subscriptionID := helpers.RequireEnvVar(t, "AZURE_SUBSCRIPTION_ID")

	t.Logf("Testing subscription discovery for: %s", subscriptionID)

	azqr := helpers.NewAZQRHelper(t)

	// Run scan on the subscription
	result := azqr.RunScan(models.ScanArgs{
		Subscriptions: []string{subscriptionID},
		Services:      []string{"rg"},
		Stages:        []string{"-diagnostics", "-advisor", "-defender"}, // Skip stages to speed up test
	})

	require.True(t, result.Success, "Subscription discovery failed: %s", result.ErrorMessage)

	t.Logf("✓ Successfully discovered subscription with %d resources", len(result.Impacted))
}
