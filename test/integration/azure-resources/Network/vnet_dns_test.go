//go:build integration

package network

import (
	"testing"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/test/integration/helpers"
	"github.com/stretchr/testify/require"
)

// TestVnetAzureProvidedDNSNoViolation verifies that a VNet using Azure-provided DNS
// (no custom DNS servers configured) does NOT trigger recommendation vnet-009.
//
// This is a regression test for https://github.com/Azure/azqr/issues/816.
// When properties.dhcpOptions.dnsServers is null (Azure-provided DNS), vnet-009
// was incorrectly flagged as a false positive.
func TestVnetAzureProvidedDNSNoViolation(t *testing.T) {
	subscriptionID := helpers.RequireEnvVar(t, "AZURE_SUBSCRIPTION_ID")

	fixtureDir := helpers.GetFixturePath(t, "scenarios/Network/vnet-azure-provided-dns")

	tf := helpers.NewTerraformHelper(t, fixtureDir)
	tf.InitAndApply(nil)

	vnetName := tf.RequireOutput("vnet_name")
	resourceGroupName := tf.RequireOutput("resource_group_name")

	t.Logf("Provisioned VNet with Azure-provided DNS: %s in RG: %s", vnetName, resourceGroupName)

	azqr := helpers.NewAZQRHelper(t)
	result := azqr.RunScan(models.ScanArgs{
		Subscriptions:  []string{subscriptionID},
		ResourceGroups: []string{resourceGroupName},
		Services:       []string{"vnet"},
		Stages:         []string{"-diagnostics", "-advisor", "-defender"},
	})

	require.True(t, result.Success, "Scan should succeed: %s", result.ErrorMessage)

	// vnet-009 must not be present regardless of ARG propagation state:
	// if dnsServers is null (not yet indexed) the fix also correctly skips it.
	for _, item := range azqr.FilterByResourceName(result.Impacted, vnetName) {
		t.Logf("Found recommendation: [%s] %s - %s", item.Impact, item.RecommendationID, item.Recommendation)
		require.NotEqual(t, "vnet-009", item.RecommendationID,
			"vnet-009 should NOT fire on a VNet using Azure-provided DNS (false positive regression test)")
	}

	t.Logf("✓ AZQR correctly skipped vnet-009 for a VNet using Azure-provided DNS")
}

// TestVnetSingleCustomDNSViolation verifies that a VNet with only one custom DNS server
// IS flagged by recommendation vnet-009 (needs at least two for high availability).
func TestVnetSingleCustomDNSViolation(t *testing.T) {
	subscriptionID := helpers.RequireEnvVar(t, "AZURE_SUBSCRIPTION_ID")

	fixtureDir := helpers.GetFixturePath(t, "scenarios/Network/vnet-single-custom-dns")

	tf := helpers.NewTerraformHelper(t, fixtureDir)
	tf.InitAndApply(nil)

	vnetName := tf.RequireOutput("vnet_name")
	resourceGroupName := tf.RequireOutput("resource_group_name")

	t.Logf("Provisioned VNet with single custom DNS: %s in RG: %s", vnetName, resourceGroupName)

	azqr := helpers.NewAZQRHelper(t)
	scanArgs := models.ScanArgs{
		Subscriptions:  []string{subscriptionID},
		ResourceGroups: []string{resourceGroupName},
		Services:       []string{"vnet"},
		Stages:         []string{"-diagnostics", "-advisor", "-defender"},
	}

	// Retry until vnet-009 appears in ARG results — ARG may not have indexed
	// properties.dhcpOptions.dnsServers yet immediately after resource creation.
	waitForARGResult(t, vnetName, scanArgs, azqr, func(impacted []*models.GraphResult) bool {
		for _, item := range impacted {
			if item.RecommendationID == "vnet-009" {
				return true
			}
		}
		return false
	})

	t.Logf("✓ AZQR correctly detected vnet-009 violation for a VNet with a single custom DNS server")
}

// waitForARGResult retries an AZQR scan until the condition is met or the timeout is reached.
// It polls every 30 seconds for up to 5 minutes to handle Azure Resource Graph propagation delays.
func waitForARGResult(t *testing.T, resourceName string, args models.ScanArgs, azqr *helpers.AZQRHelper, condition func([]*models.GraphResult) bool) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Minute)
	pollInterval := 30 * time.Second

	for attempt := 1; ; attempt++ {
		result := azqr.RunScan(args)
		require.True(t, result.Success, "Scan should succeed: %s", result.ErrorMessage)

		impacted := azqr.FilterByResourceName(result.Impacted, resourceName)
		if condition(impacted) {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("ARG condition not met for resource %q after 5 minutes (%d attempts)", resourceName, attempt)
		}

		t.Logf("Attempt %d: ARG condition not yet met for %q, retrying in %s...", attempt, resourceName, pollInterval)
		time.Sleep(pollInterval)
	}
}
