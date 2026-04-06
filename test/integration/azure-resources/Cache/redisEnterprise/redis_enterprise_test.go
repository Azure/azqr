//go:build integration

package redisEnterprise

import (
	"testing"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/test/integration/helpers"
	"github.com/stretchr/testify/require"
)

// TestAzureManagedRedisViolations provisions a single Azure Managed Redis (AMR) instance
// with multiple deliberate policy violations and verifies that AZQR detects each one.
//
// A single instance is shared across all sub-tests to minimise cost and provisioning time.
// The Balanced_B0 SKU is the smallest available AMR SKU (~10-20 min to provision).
//
// Rules verified (cluster-level, all queryable via ARG):
//   - redis-012: zone redundancy
//   - redis-013: high availability enabled
//   - redis-014: public network access disabled
//   - redis-015: private endpoints present
//   - redis-016: customer-managed key encryption
//
// Rules NOT tested (redis-017, redis-018, redis-019 — database-level):
//   Microsoft.Cache/redisEnterprise/databases is not indexed by Azure Resource Graph,
//   so those rules are documented as automationAvailable: false with no KQL file.
//
// Negative assertion:
//   - redis-010 must NOT fire because the Balanced_B0 SKU is not a deprecated Enterprise SKU.
func TestAzureManagedRedisViolations(t *testing.T) {
	subscriptionID := helpers.RequireEnvVar(t, "AZURE_SUBSCRIPTION_ID")

	fixtureDir := helpers.GetFixturePath(t, "scenarios/Cache/redisEnterprise/amr-violations")

	tf := helpers.NewTerraformHelper(t, fixtureDir)
	tf.InitAndApply(nil)

	clusterName := tf.RequireOutput("cluster_name")
	resourceGroupName := tf.RequireOutput("resource_group_name")
	haEnabled := tf.GetOutput("high_availability_enabled")

	t.Logf("Provisioned AMR instance: %s in RG: %s (HA: %s)", clusterName, resourceGroupName, haEnabled)

	require.Equal(t, "false", haEnabled, "fixture should have HA disabled for the violation scenario")

	azqr := helpers.NewAZQRHelper(t)
	result := azqr.RunScan(models.ScanArgs{
		Subscriptions:  []string{subscriptionID},
		ResourceGroups: []string{resourceGroupName},
		Services:       []string{"redis"},
		Stages:         []string{"-diagnostics", "-advisor", "-defender"},
	})

	require.True(t, result.Success, "AZQR scan should succeed: %s", result.ErrorMessage)
	require.NotEmpty(t, result.Impacted, "expected impacted resources from the non-compliant AMR instance")

	clusterImpacted := azqr.FilterByResourceName(result.Impacted, clusterName)

	for _, item := range result.Impacted {
		t.Logf("Impacted: [%s] %s — %s", item.RecommendationID, item.Name, item.Recommendation)
	}

	t.Run("ZoneRedundancy_redis012", func(t *testing.T) {
		require.True(t,
			azqr.HasRecommendationID(clusterImpacted, "redis-012"),
			"expected redis-012 (zone redundancy) for instance %q — no zones configured", clusterName,
		)
		t.Logf("✓ redis-012 correctly detected: instance has no availability zones")
	})

	t.Run("HighAvailability_redis013", func(t *testing.T) {
		require.True(t,
			azqr.HasRecommendationID(clusterImpacted, "redis-013"),
			"expected redis-013 (high availability) for instance %q — HA is disabled", clusterName,
		)
		t.Logf("✓ redis-013 correctly detected: high availability is disabled")
	})

	t.Run("PublicNetworkAccess_redis014", func(t *testing.T) {
		require.True(t,
			azqr.HasRecommendationID(clusterImpacted, "redis-014"),
			"expected redis-014 (public network access) for instance %q — public access is Enabled by default", clusterName,
		)
		t.Logf("✓ redis-014 correctly detected: public network access is enabled")
	})

	t.Run("NoPrivateEndpoints_redis015", func(t *testing.T) {
		require.True(t,
			azqr.HasRecommendationID(clusterImpacted, "redis-015"),
			"expected redis-015 (private endpoints) for instance %q — no private endpoints configured", clusterName,
		)
		t.Logf("✓ redis-015 correctly detected: no private endpoints")
	})

	t.Run("NoCustomerManagedKey_redis016", func(t *testing.T) {
		require.True(t,
			azqr.HasRecommendationID(clusterImpacted, "redis-016"),
			"expected redis-016 (customer-managed key) for instance %q — no CMK configured", clusterName,
		)
		t.Logf("✓ redis-016 correctly detected: no customer-managed key encryption")
	})

	// redis-010 flags deprecated Enterprise_* SKUs. Balanced_B0 is an AMR SKU so it must NOT be flagged.
	t.Run("NotDeprecatedSKU_redis010", func(t *testing.T) {
		require.False(t,
			azqr.HasRecommendationID(clusterImpacted, "redis-010"),
			"redis-010 (deprecated Enterprise SKU) should NOT fire for Balanced_B0 (AMR) instance %q", clusterName,
		)
		t.Logf("✓ redis-010 correctly absent: Balanced_B0 is not a deprecated Enterprise SKU")
	})
}



