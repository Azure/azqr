// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package scanners

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Azure/azqr/internal/models"
)

func TestBuildResourcesAppliesTagFilters(t *testing.T) {
	data := []json.RawMessage{
		json.RawMessage(`{"id":"/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/prod","subscriptionId":"sub","resourceGroup":"rg","type":"Microsoft.Compute/virtualMachines","name":"prod","tags":{"Env":"prod"}}`),
		json.RawMessage(`{"id":"/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/dev","subscriptionId":"sub","resourceGroup":"rg","type":"Microsoft.Compute/virtualMachines","name":"dev","tags":{"env":"dev"}}`),
	}

	// LoadFilters normally initializes these maps and scanners. Populate the scanner
	// type through a filter file is unnecessary for this mapping-focused test.
	filters := initializeTestFilters(t)
	included, excluded := buildResources(data, filters)

	if len(included) != 1 || included[0].Name != "prod" {
		t.Fatalf("unexpected included resources: %+v", included)
	}
	if len(excluded) != 1 || excluded[0].Name != "dev" {
		t.Fatalf("unexpected excluded resources: %+v", excluded)
	}
}

func TestCountResourcesByTypeAndSubscription(t *testing.T) {
	resources := []*models.Resource{
		{SubscriptionID: "sub", Type: "Microsoft.Test/widgets"},
		{SubscriptionID: "sub", Type: "Microsoft.Test/widgets"},
	}
	got := CountResourcesByTypeAndSubscription(resources, map[string]string{"sub": "Subscription"})
	if len(got) != 1 || got[0].Count != 2 || got[0].Subscription != "Subscription" {
		t.Fatalf("unexpected counts: %+v", got)
	}
}

func initializeTestFilters(t *testing.T) *models.Filters {
	t.Helper()
	original, existed := models.ScannerList["vm"]
	models.ScannerList["vm"] = []models.IAzureScanner{testResourceScanner{}}
	t.Cleanup(func() {
		if existed {
			models.ScannerList["vm"] = original
		} else {
			delete(models.ScannerList, "vm")
		}
	})
	// NewFilters cannot expose its internal indexes to this package, so use a real
	// temporary YAML file through LoadFilters.
	path := t.TempDir() + "/filters.yml"
	content := []byte("azqr:\n  include:\n    tags:\n      env: prod\n")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	return models.LoadFilters(path, []string{"vm"})
}

type testResourceScanner struct{}

func (testResourceScanner) ServiceName() string { return "Test" }
func (testResourceScanner) ResourceTypes() []string {
	return []string{"Microsoft.Compute/virtualMachines"}
}
