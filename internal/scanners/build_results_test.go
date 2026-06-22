// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/azqr/internal/models"
)

// fakeScanner registers the resource types used by these tests so that
// LoadFilters marks them as included (IsServiceExcluded only includes
// resource types exposed by loaded scanners).
type fakeScanner struct{}

func (fakeScanner) ServiceName() string { return "fake" }
func (fakeScanner) ResourceTypes() []string {
	return []string{
		"Microsoft.Storage/storageAccounts",
		"Microsoft.AzureArcData/sqlServerInstances",
	}
}

func registerTestScanners() {
	models.ScannerList = map[string][]models.IAzureScanner{
		"fake": {fakeScanner{}},
	}
}

// includeAllFilters returns a filter that includes the registered test
// resource types and excludes nothing.
func includeAllFilters() *models.Filters {
	registerTestScanners()
	return models.LoadFilters("", []string{})
}

// filtersFromYAML writes a filter YAML to a temp file and loads it.
func filtersFromYAML(t *testing.T, yaml string) *models.Filters {
	t.Helper()
	registerTestScanners()
	path := filepath.Join(t.TempDir(), "filter.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatalf("failed to write filter file: %v", err)
	}
	return models.LoadFilters(path, []string{})
}

func rawRows(rows ...string) []json.RawMessage {
	data := make([]json.RawMessage, 0, len(rows))
	for _, r := range rows {
		data = append(data, json.RawMessage(r))
	}
	return data
}

const storageID = "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.Storage/storageAccounts/acct1"

func TestBuildAdvisorResults(t *testing.T) {
	filters := includeAllFilters()
	recTypes := map[string]string{"rt-1": "Use managed disks"}
	subscriptions := map[string]string{"sub1": "Sub One"}

	data := rawRows(
		`{"SubscriptionId":"sub1","ResourceId":"`+storageID+`","ImpactedValue":"acct1","Category":"Cost","Impact":"High","RecommendationTypeId":"rt-1"}`,
		`{bad json}`,
	)

	got := buildAdvisorResults(data, subscriptions, filters, recTypes)
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}

	r := got[0]
	if r.SubscriptionID != "sub1" || r.SubscriptionName != "Sub One" {
		t.Errorf("subscription mapping wrong: %+v", r)
	}
	if r.Name != "acct1" || r.Category != "Cost" || r.Impact != "High" {
		t.Errorf("field mapping wrong: %+v", r)
	}
	if r.RecommendationID != "rt-1" || r.Description != "Use managed disks" {
		t.Errorf("recommendation mapping wrong: %+v", r)
	}
	if r.ResourceID != storageID {
		t.Errorf("resource id mapping wrong: %+v", r)
	}
	if r.Type != models.GetResourceTypeFromResourceID(storageID) {
		t.Errorf("type mapping wrong: got %q", r.Type)
	}
}

func TestBuildDefenderResults(t *testing.T) {
	t.Run("maps rows and skips malformed", func(t *testing.T) {
		data := rawRows(
			`{"SubscriptionId":"sub1","SubscriptionName":"Sub One","Name":"StorageAccounts","Tier":"Standard"}`,
			`not-json`,
		)
		got := buildDefenderResults(data, includeAllFilters())
		if len(got) != 1 {
			t.Fatalf("expected 1 result, got %d", len(got))
		}
		r := got[0]
		if r.SubscriptionID != "sub1" || r.SubscriptionName != "Sub One" || r.Name != "StorageAccounts" || r.Tier != "Standard" {
			t.Errorf("field mapping wrong: %+v", r)
		}
	})

	t.Run("excluded subscription is filtered out", func(t *testing.T) {
		filters := filtersFromYAML(t, "azqr:\n  exclude:\n    subscriptions:\n      - sub1\n")
		data := rawRows(`{"SubscriptionId":"sub1","SubscriptionName":"Sub One","Name":"x","Tier":"Free"}`)
		got := buildDefenderResults(data, filters)
		if len(got) != 0 {
			t.Fatalf("expected excluded subscription to yield 0 results, got %d", len(got))
		}
	})
}

func TestBuildDefenderRecommendations(t *testing.T) {
	filters := includeAllFilters()
	subscriptions := map[string]string{"sub1": "Sub One"}

	row := `{"SubscriptionId":"sub1","ResourceGroupName":"rg1","ResourceType":"storageAccounts","ResourceName":"acct1","Category":"Security","RecommendationSeverity":"High","RecommendationName":"Encrypt data","ActionDescription":"action","RemediationDescription":"remediate","AzPortalLink":"portal.azure.com/x","ResourceId":"` + storageID + `"}`

	// Duplicate row to exercise dedup by (resourceID, category, recommendationName).
	data := rawRows(row, row, `{bad}`)

	got := buildDefenderRecommendations(data, subscriptions, filters)
	if len(got) != 1 {
		t.Fatalf("expected dedup to 1 result, got %d", len(got))
	}
	r := got[0]
	if r.SubscriptionId != "sub1" || r.SubscriptionName != "Sub One" {
		t.Errorf("subscription mapping wrong: %+v", r)
	}
	if r.AzPortalLink != "https://portal.azure.com/x" {
		t.Errorf("portal link should be prefixed with https://, got %q", r.AzPortalLink)
	}
	if r.RecommendationName != "Encrypt data" || r.RecommendationSeverity != "High" || r.Category != "Security" {
		t.Errorf("field mapping wrong: %+v", r)
	}
	if r.ResourceId != storageID {
		t.Errorf("resource id mapping wrong: %+v", r)
	}
}

func TestBuildAzurePolicyResults(t *testing.T) {
	filters := includeAllFilters()

	row := `{"subscriptionId":"sub1","subscriptionName":"Sub One","resourceId":"` + storageID + `","policyDefinitionDisplayName":"Audit storage","policyDescription":"desc","timestamp":"2024-01-01T00:00:00Z","policyDefinitionName":"def","policyDefinitionId":"def-1","policyAssignmentName":"assign","policyAssignmentId":"assign-1","complianceState":"NonCompliant"}`

	// Duplicate row to exercise dedup by (resourceID, policyDefinitionID).
	data := rawRows(row, row, `{invalid}`)

	got := buildAzurePolicyResults(data, filters)
	if len(got) != 1 {
		t.Fatalf("expected dedup to 1 result, got %d", len(got))
	}
	r := got[0]
	if r.SubscriptionID != "sub1" || r.SubscriptionName != "Sub One" {
		t.Errorf("subscription mapping wrong: %+v", r)
	}
	if r.PolicyDisplayName != "Audit storage" || r.PolicyDescription != "desc" {
		t.Errorf("policy display/description mapping wrong: %+v", r)
	}
	if r.PolicyDefinitionName != "def" || r.PolicyDefinitionID != "def-1" {
		t.Errorf("policy definition mapping wrong: %+v", r)
	}
	if r.PolicyAssignmentName != "assign" || r.PolicyAssignmentID != "assign-1" {
		t.Errorf("policy assignment mapping wrong: %+v", r)
	}
	if r.ComplianceState != "NonCompliant" || r.TimeStamp != "2024-01-01T00:00:00Z" {
		t.Errorf("compliance/timestamp mapping wrong: %+v", r)
	}
	if r.ResourceID != storageID {
		t.Errorf("resource id mapping wrong: %+v", r)
	}
}

func TestBuildArcSQLResults(t *testing.T) {
	filters := includeAllFilters()
	subscriptions := map[string]string{"sub1": "Sub One"}

	arcID := "/subscriptions/sub1/resourceGroups/rg1/providers/Microsoft.AzureArcData/sqlServerInstances/inst1"
	data := rawRows(
		`{"subscriptionId":"sub1","status":"Connected","AzureArcServer":"server1","SQLInstance":"`+arcID+`","resourceGroup":"rg1","version":"2019","Build":"15.0","patchLevel":"CU1","edition":"Standard","vcores":"4","License":"PAYG","DPSStatus":"OK","TELStatus":"__","DefenderStatus":"Protected"}`,
		`{nope}`,
	)

	got := buildArcSQLResults(data, subscriptions, filters)
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	r := got[0]
	if r.SubscriptionID != "sub1" || r.SubscriptionName != "Sub One" {
		t.Errorf("subscription mapping wrong: %+v", r)
	}
	if r.Status != "Connected" || r.AzureArcServer != "server1" || r.SQLInstance != arcID {
		t.Errorf("core field mapping wrong: %+v", r)
	}
	if r.Version != "2019" || r.Build != "15.0" || r.PatchLevel != "CU1" || r.Edition != "Standard" {
		t.Errorf("version/build mapping wrong: %+v", r)
	}
	if r.VCores != "4" || r.License != "PAYG" || r.DPSStatus != "OK" || r.TELStatus != "__" || r.DefenderStatus != "Protected" {
		t.Errorf("status mapping wrong: %+v", r)
	}
}

func TestBuildResources(t *testing.T) {
	row := `{"id":"` + storageID + `","subscriptionId":"sub1","resourceGroup":"rg1","location":"westus","type":"Microsoft.Storage/storageAccounts","name":"acct1","skuName":"Standard_LRS","skuTier":"Standard","skuFamily":"fam","skuCapacity":3,"kind":"StorageV2"}`

	t.Run("nil filter includes everything", func(t *testing.T) {
		included, excluded := buildResources(rawRows(row, `{bad}`), nil)
		if len(included) != 1 || len(excluded) != 0 {
			t.Fatalf("expected 1 included, 0 excluded; got %d/%d", len(included), len(excluded))
		}
		r := included[0]
		if r.ID != storageID || r.SubscriptionID != "sub1" || r.ResourceGroup != "rg1" || r.Location != "westus" {
			t.Errorf("core mapping wrong: %+v", r)
		}
		if r.Type != "Microsoft.Storage/storageAccounts" || r.Name != "acct1" || r.Kind != "StorageV2" {
			t.Errorf("type/name/kind mapping wrong: %+v", r)
		}
		if r.SkuName != "Standard_LRS" || r.SkuTier != "Standard" || r.SkuFamily != "fam" || r.SkuCapacity != 3 {
			t.Errorf("sku mapping wrong: %+v", r)
		}
	})

	t.Run("nil data returns empty slices", func(t *testing.T) {
		included, excluded := buildResources(nil, nil)
		if len(included) != 0 || len(excluded) != 0 {
			t.Fatalf("expected empty slices, got %d/%d", len(included), len(excluded))
		}
	})

	t.Run("excluded service goes to excluded slice", func(t *testing.T) {
		filters := filtersFromYAML(t, "azqr:\n  exclude:\n    services:\n      - "+storageID+"\n")
		included, excluded := buildResources(rawRows(row), filters)
		if len(included) != 0 || len(excluded) != 1 {
			t.Fatalf("expected 0 included, 1 excluded; got %d/%d", len(included), len(excluded))
		}
	})
}

func TestBuildResourceTypeCounts(t *testing.T) {
	filters := includeAllFilters()
	subscriptions := map[string]string{"sub1": "Sub One"}
	data := rawRows(
		`{"subscriptionId":"sub1","type":"Microsoft.Storage/storageAccounts","count_":5}`,
		`{bad}`,
	)
	got := buildResourceTypeCounts(data, subscriptions, filters)
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	r := got[0]
	if r.Subscription != "Sub One" || r.ResourceType != "Microsoft.Storage/storageAccounts" || r.Count != 5 {
		t.Errorf("count mapping wrong: %+v", r)
	}
}

func TestBuildResourceTypeCountMap(t *testing.T) {
	filters := includeAllFilters()
	data := rawRows(
		`{"type":"Microsoft.Storage/storageAccounts","count_":7}`,
		`{bad}`,
	)
	got := buildResourceTypeCountMap(data, filters)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got["Microsoft.Storage/storageAccounts"] != 7 {
		t.Errorf("count map wrong: %+v", got)
	}
}
