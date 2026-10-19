// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmmodernization

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/graph"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TestArgBool_UnmarshalJSON locks in the fix for a real production error:
// Azure Resource Graph serializes computed boolean expressions
// (isnotnull/isnotempty/case/iff — used by HasScaleSet, IsAKSManaged,
// AzureDiskEncryption, IsMarketplaceVM in discover-assess.kql) as the JSON
// number 0 or 1, not a JSON true/false literal, confirmed via a live query
// against the Resource Graph REST API. A plain encoding/json bool target
// fails with "cannot unmarshal number into Go struct field ... of type bool".
func TestArgBool_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json string
		want bool
	}{
		{"0", false},
		{"1", true},
		{"true", true},
		{"false", false},
		{"null", false},
	}
	for _, tt := range tests {
		var b argBool
		if err := json.Unmarshal([]byte(tt.json), &b); err != nil {
			t.Fatalf("argBool.UnmarshalJSON(%q) returned error: %v", tt.json, err)
		}
		if bool(b) != tt.want {
			t.Errorf("argBool.UnmarshalJSON(%q) = %v, want %v", tt.json, bool(b), tt.want)
		}
	}
}

// TestVMModernizationRow_UnmarshalFromARGShape reproduces the exact row shape
// Azure Resource Graph returns (booleans as 0/1 numbers) and confirms the full
// row struct decodes without the "cannot unmarshal number into Go struct
// field" error.
func TestVMModernizationRow_UnmarshalFromARGShape(t *testing.T) {
	const sample = `{
		"Name": "vm1", "Id": "/subscriptions/x/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
		"ResourceGroup": "rg1", "SubscriptionId": "sub1", "Location": "eastus",
		"VMSize": "Standard_D4s_v3", "HyperVGeneration": "V2", "DiskControllerType": "SCSI",
		"HasScaleSet": 0, "IsAKSManaged": 1, "AzureDiskEncryption": 0,
		"IsDatabricksManaged": 0, "IsMarketplaceVM": 1
	}`
	var row vmModernizationRow
	if err := json.Unmarshal([]byte(sample), &row); err != nil {
		t.Fatalf("vmModernizationRow failed to unmarshal an ARG-shaped row: %v", err)
	}
	if bool(row.HasScaleSet) != false || bool(row.IsAKSManaged) != true || bool(row.AzureDiskEncryption) != false || bool(row.IsDatabricksManaged) != false || bool(row.IsMarketplaceVM) != true {
		t.Errorf("unexpected decoded values: %+v", row)
	}
}

// TestVMModernizationRow_UnmarshalFromARGShape_Uniform reproduces the row
// shape discover-assess.kql now produces for a Uniform-orchestration VMSS
// member instance (P01-T01): a
// Microsoft.Compute/virtualMachineScaleSets/virtualMachines resource ID,
// HasScaleSet forced to true, MembershipModel reading "VMSS (Uniform)", and
// the two pool-level fields (P02) populated from the VMSS-parent join.
func TestVMModernizationRow_UnmarshalFromARGShape_Uniform(t *testing.T) {
	const sample = `{
		"Name": "vmss1_0", "Id": "/subscriptions/x/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachineScaleSets/vmss1/virtualMachines/0",
		"ResourceGroup": "rg1", "SubscriptionId": "sub1", "Location": "eastus",
		"VMSize": "Standard_D4s_v3", "HyperVGeneration": "V2", "DiskControllerType": "SCSI",
		"HasScaleSet": 1, "IsAKSManaged": 1, "AzureDiskEncryption": 0,
		"IsDatabricksManaged": 0, "IsMarketplaceVM": 0,
		"MembershipModel": "VMSS (Uniform)",
		"OrchestrationMode": "Uniform", "PoolModelVMSize": "Standard_D4s_v5"
	}`
	var row vmModernizationRow
	if err := json.Unmarshal([]byte(sample), &row); err != nil {
		t.Fatalf("vmModernizationRow failed to unmarshal a Uniform-mode ARG-shaped row: %v", err)
	}
	if !bool(row.HasScaleSet) {
		t.Errorf("HasScaleSet = false, want true for a Uniform-mode member instance")
	}
	if row.MembershipModel != "VMSS (Uniform)" {
		t.Errorf("MembershipModel = %q, want %q", row.MembershipModel, "VMSS (Uniform)")
	}
	if row.OrchestrationMode != "Uniform" || row.PoolModelVMSize != "Standard_D4s_v5" {
		t.Errorf("unexpected pool-level fields: OrchestrationMode=%q PoolModelVMSize=%q", row.OrchestrationMode, row.PoolModelVMSize)
	}
}

func TestClassifyWorkloadPattern(t *testing.T) {
	tests := []struct {
		name                string
		hasScaleSet         bool
		isAKSManaged        bool
		isDatabricksManaged bool
		avdHostPoolType     string
		imagePublisher      string
		imageOffer          string
		wantPatternCode     string
		wantPoolMembership  string
	}{
		{"AKS node pool", true, true, false, "", "", "", patternComputePool, "AKS Node Pool"},
		{"generic VMSS", true, false, false, "", "", "", patternComputePool, "VMSS"},
		{"standalone VM", false, false, false, "", "", "", patternCustomerManaged, "Standalone"},
		{"certified ISV appliance", false, false, false, "", "paloaltonetworks", "vmseries-flex", patternCertifiedAppliance, "Palo Alto Networks"},
		{"non-appliance image stays customer-managed", false, false, false, "", "contoso", "contoso-app", patternCustomerManaged, "Standalone"},
		{"Databricks-managed node", false, false, true, "", "", "", patternPlatformManaged, "Azure Databricks"},
		{"AVD pooled host pool", false, false, false, "Pooled", "", "", patternAVDPooled, "AVD Pooled Host Pool"},
		{"AVD pooled host pool is case-insensitive", false, false, false, "pooled", "", "", patternAVDPooled, "AVD Pooled Host Pool"},
		{"AVD personal host pool falls through to customer-managed", false, false, false, "Personal", "", "", patternCustomerManaged, "Standalone"},
		{"absent AVD host-pool type falls through to customer-managed", false, false, false, "", "", "", patternCustomerManaged, "Standalone"},
		// A-over-G overlap: a pool/AKS member whose image also matches a known
		// appliance vendor still resolves to pattern A, since pool membership
		// is checked before the appliance-vendor match (NFR-003 precedence order).
		{"A-over-G: AKS-managed VM with an appliance image stays pattern A", false, true, false, "", "paloaltonetworks", "vmseries-flex", patternComputePool, "AKS Node Pool"},
		{"G-over-VMSS: VMSS member with an appliance image is pattern G", true, false, false, "", "fortinet", "fortigate-vm", patternCertifiedAppliance, "Fortinet"},
		// C-over-A overlap: a Databricks-tagged VM that also reports
		// HasScaleSet=true or IsAKSManaged=true still resolves to pattern C,
		// since the Databricks tag check runs before the pool-membership
		// checks (NFR-003 precedence order).
		{"C-over-A: Databricks-tagged VM also reporting HasScaleSet stays pattern C", true, false, true, "", "", "", patternPlatformManaged, "Azure Databricks"},
		{"C-over-A: Databricks-tagged VM also reporting IsAKSManaged stays pattern C", false, true, true, "", "", "", patternPlatformManaged, "Azure Databricks"},
		// B-excluded-by-A overlap: a VM that is both AVD-pooled and a
		// VMSS/AKS pool member resolves to pattern A, not B, consistent with
		// the canonical order placing A before B (NFR-003 precedence order).
		{"B-excluded-by-A: AVD-pooled VM also reporting IsAKSManaged stays pattern A", false, true, false, "Pooled", "", "", patternComputePool, "AKS Node Pool"},
		{"B-over-VMSS: AVD-pooled VM also reporting HasScaleSet is pattern B", true, false, false, "Pooled", "", "", patternAVDPooled, "AVD Pooled Host Pool"},
		// G-over-E baseline: an appliance-image VM with no pool, Databricks,
		// or AVD signal resolves to G, not the E default.
		{"G-over-E: appliance-image VM with no other signal resolves to G", false, false, false, "", "checkpoint", "vmseries", patternCertifiedAppliance, "Check Point"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, gotPool := classifyWorkloadPattern(tt.hasScaleSet, tt.isAKSManaged, tt.isDatabricksManaged, tt.avdHostPoolType, tt.imagePublisher, tt.imageOffer, false)
			if gotCode != tt.wantPatternCode || gotPool != tt.wantPoolMembership {
				t.Errorf("classifyWorkloadPattern(%v, %v, %v, %q, %q, %q) = (%q, %q), want (%q, %q)",
					tt.hasScaleSet, tt.isAKSManaged, tt.isDatabricksManaged, tt.avdHostPoolType, tt.imagePublisher, tt.imageOffer, gotCode, gotPool, tt.wantPatternCode, tt.wantPoolMembership)
			}
		})
	}

	aroTests := []struct {
		name        string
		row         vmModernizationRow
		wantPattern string
		wantPool    string
	}{
		{"ARO node", vmModernizationRow{}, patternComputePool, "ARO Cluster"},
		{"ARO before VMSS", vmModernizationRow{HasScaleSet: true}, patternComputePool, "ARO Cluster"},
		{"ARO before AKS", vmModernizationRow{IsAKSManaged: true}, patternComputePool, "ARO Cluster"},
		{"ARO before AVD", vmModernizationRow{AVDHostPoolType: "Pooled"}, patternComputePool, "ARO Cluster"},
		{"ARO before appliance", vmModernizationRow{ImagePublisher: "fortinet"}, patternComputePool, "ARO Cluster"},
		{"Databricks before ARO", vmModernizationRow{IsDatabricksManaged: true}, patternPlatformManaged, "Azure Databricks"},
	}
	for _, tt := range aroTests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.row
			pattern, pool := classifyWorkloadPattern(bool(r.HasScaleSet), bool(r.IsAKSManaged), bool(r.IsDatabricksManaged), r.AVDHostPoolType, r.ImagePublisher, r.ImageOffer, true)
			if pattern != tt.wantPattern || pool != tt.wantPool {
				t.Errorf("ARO classification = (%q, %q), want (%q, %q)", pattern, pool, tt.wantPattern, tt.wantPool)
			}
		})
	}
}

func TestClassifySeriesGeneration(t *testing.T) {
	tests := []struct {
		vmSize               string
		wantLabel            string
		wantTargetGeneration bool
	}{
		{"Standard_D4s_v3", "v3", false},
		{"Standard_D4s_v6", "v6", true},
		{"Standard_D4ads_v7", "v7", true},
		{"Standard_D4", "Unknown (no _vN suffix)", false},
	}
	for _, tt := range tests {
		t.Run(tt.vmSize, func(t *testing.T) {
			gotLabel, gotTarget := classifySeriesGeneration(tt.vmSize)
			if gotLabel != tt.wantLabel || gotTarget != tt.wantTargetGeneration {
				t.Errorf("classifySeriesGeneration(%q) = (%q, %v), want (%q, %v)",
					tt.vmSize, gotLabel, gotTarget, tt.wantLabel, tt.wantTargetGeneration)
			}
		})
	}
}

func TestRecommendExecutionMethod(t *testing.T) {
	tests := []struct {
		name             string
		patternCode      string
		hyperVGeneration string
		isTargetGen      bool
		wantContains     string
	}{
		{"pool-managed VM", patternComputePool, "V2", false, "pool/cluster"},
		{"already on target", patternCustomerManaged, "V2", true, "No action needed"},
		{"Gen1 customer-managed", patternCustomerManaged, "V1", false, manualReviewRequired},
		{"Gen2 customer-managed", patternCustomerManaged, "V2", false, "Redeploy"},
		{"unknown Hyper-V generation", patternCustomerManaged, "", false, manualReviewRequired},
		{"certified ISV appliance is actionable like a customer-managed VM", patternCertifiedAppliance, "V2", false, "only after vendor certification is confirmed"},
		{"customer-managed redeploy is conditional on the gates", patternCustomerManaged, "V2", false, "conditional on the Temp Disk Dependency and SAP/ISV Certification Gate review"},
		{"AVD pooled host pool still needs golden-image checks", patternAVDPooled, "V2", false, "golden image still needs Assess/Plan checks"},
		{"Databricks-managed node is non-actionable like a pool-managed VM", patternPlatformManaged, "V2", false, "pool/cluster"},
		{"AVD pooled host pool is non-actionable like a pool-managed VM", patternAVDPooled, "V2", false, "pool/cluster"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := recommendExecutionMethod(tt.patternCode, tt.hyperVGeneration, tt.isTargetGen)
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("recommendExecutionMethod(%q, %q, %v) = %q, want substring %q",
					tt.patternCode, tt.hyperVGeneration, tt.isTargetGen, got, tt.wantContains)
			}
		})
	}
}

func TestRecommendHardGates(t *testing.T) {
	tests := []struct {
		name                    string
		patternCode             string
		wantTempDiskDependency  string
		wantSAPISVCertification string
	}{
		{"compute pool (A) is not subject to either gate", patternComputePool, "N/A (pool/platform-managed)", "N/A (pool/platform-managed)"},
		{"AVD pooled host pool (B) needs both gates reviewed", patternAVDPooled, manualReviewRequired, manualReviewRequired},
		{"platform-managed e.g. Databricks (C) is not subject to either gate", patternPlatformManaged, "N/A (pool/platform-managed)", "N/A (pool/platform-managed)"},
		{"customer-managed VM (E) needs both gates reviewed", patternCustomerManaged, manualReviewRequired, manualReviewRequired},
		{"certified ISV appliance (G) needs both gates reviewed", patternCertifiedAppliance, manualReviewRequired, manualReviewRequired},
		{"SAP workload (F, not yet emitted) needs both gates reviewed for forward-compatibility", patternSAPWorkload, manualReviewRequired, manualReviewRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTempDisk, gotSAPISV := recommendHardGates(tt.patternCode)
			if gotTempDisk != tt.wantTempDiskDependency {
				t.Errorf("recommendHardGates(%q) tempDiskDependency = %q, want %q", tt.patternCode, gotTempDisk, tt.wantTempDiskDependency)
			}
			if gotSAPISV != tt.wantSAPISVCertification {
				t.Errorf("recommendHardGates(%q) sapISVCertificationGate = %q, want %q", tt.patternCode, gotSAPISV, tt.wantSAPISVCertification)
			}
		})
	}
}

func TestIsV6V7Family(t *testing.T) {
	tests := []struct {
		family string
		want   bool
	}{
		{"standardDsv6Family", true},
		{"StandardDadsv7Family", true},
		{"standardDv3Family", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isV6V7Family(tt.family); got != tt.want {
			t.Errorf("isV6V7Family(%q) = %v, want %v", tt.family, got, tt.want)
		}
	}
}

func TestSuggestTargetSKU_UnknownSKU(t *testing.T) {
	if _, _, ok := suggestTargetSKU("Not_A_Real_SKU"); ok {
		t.Error("suggestTargetSKU on an unknown SKU name should return ok=false")
	}
}

func TestMatchLifecycleNotes(t *testing.T) {
	// Pin AZURE_CLOUD to a non-exempt value so this test's assertions do not
	// depend on ambient environment state: matchLifecycleNotes suppresses all
	// output on Azure Government/China (see TestIsExemptCloud), and this test
	// expects lifecycle notes to be present throughout (RV-002).
	t.Setenv("AZURE_CLOUD", "")

	// Dv3 is explicitly listed in the embedded dataset's retirement row, which
	// is scoped to a disjoint family list from the v1-v2 price-increase row,
	// so a Dv3 SKU must match the former and not the latter.
	got := matchLifecycleNotes("Standard_D4_v3", "eastus")
	if !strings.Contains(got, "retirement date") {
		t.Errorf("matchLifecycleNotes(Standard_D4_v3, eastus) = %q, want it to mention the retirement", got)
	}
	if strings.Contains(got, "price increase on v1-v2 VM series") {
		t.Errorf("matchLifecycleNotes(Standard_D4_v3, eastus) = %q, should not match the v1-v2 price-increase row (Dv3 is not in its family list)", got)
	}

	// Dv2 matches the v1-v2 price-increase row and its own 2028-05-01
	// retirement row, but not the 2029-11-15 Dv3/Dsv3/Ev3/Esv3 row.
	got = matchLifecycleNotes("Standard_D4_v2", "eastus")
	if !strings.Contains(got, "price increase on v1-v2 VM series") {
		t.Errorf("matchLifecycleNotes(Standard_D4_v2, eastus) = %q, want the v1-v2 price-increase note", got)
	}
	if !strings.Contains(got, "effective 2028-05-01") || strings.Contains(got, "effective 2029-11-15") {
		t.Errorf("matchLifecycleNotes(Standard_D4_v2, eastus) = %q, want only the 2028-05-01 retirement", got)
	}

	for _, tc := range []struct{ sku, date string }{
		{"Standard_NP10s", "2027-05-31"},
		{"Standard_HC44rs", "2027-05-31"},
		{"Standard_DS2_v2", "2028-05-01"},
		{"Standard_L8s", "2028-05-01"},
		{"Standard_A2_v2", "2028-11-15"},
		{"Standard_B2s", "2028-11-15"},
		{"Standard_F4s_v2", "2028-11-15"},
		{"Standard_GS2", "2028-11-15"},
		{"Standard_L8s_v2", "2028-11-15"},
	} {
		if got := matchLifecycleNotes(tc.sku, "eastus"); !strings.Contains(got, "effective "+tc.date) {
			t.Errorf("matchLifecycleNotes(%s, eastus) = %q, want retirement effective %s", tc.sku, got, tc.date)
		}
	}

	// A v6 SKU matches no family-scoped row, and outside the region-scoped
	// row's listed regions, has no applicable notes at all.
	got = matchLifecycleNotes("Standard_D4s_v6", "eastus")
	if got != "N/A" {
		t.Errorf("matchLifecycleNotes(Standard_D4s_v6, eastus) = %q, want \"N/A\" (no family or region match)", got)
	}

	// The regional price increase rows are family-agnostic but region-scoped,
	// each stating its own exact percentage: a non-v1/v2/v7 SKU in an affected
	// region must see that region's precise figure, and must see nothing
	// regional outside the listed regions.
	got = matchLifecycleNotes("Standard_D4s_v6", "westeurope")
	if !strings.Contains(got, "Regional price increase of 9 percent") {
		t.Errorf("matchLifecycleNotes(Standard_D4s_v6, westeurope) = %q, want the West Europe 9 percent regional price increase note", got)
	}
	got = matchLifecycleNotes("Standard_D4s_v6", "northeurope")
	if !strings.Contains(got, "Regional price increase of 17 percent") {
		t.Errorf("matchLifecycleNotes(Standard_D4s_v6, northeurope) = %q, want the North Europe 17 percent regional price increase note", got)
	}
	got = matchLifecycleNotes("Standard_D4s_v6", "eastus")
	if strings.Contains(got, "Regional price increase") {
		t.Errorf("matchLifecycleNotes(Standard_D4s_v6, eastus) = %q, should not match any regional price increase row outside its listed regions", got)
	}

	// RV-001: a v1/v2 family in an affected region must show only the 25
	// percent v1-v2 note, never also a regional percentage note — the field
	// alert states v1/v2 are excluded from the regional increase and take the
	// 25 percent change instead ("do not double-count"), and the regional
	// rows' excludeFamilies now carries the v1/v2 family set alongside v7.
	got = matchLifecycleNotes("Standard_D4_v2", "westeurope")
	if !strings.Contains(got, "price increase on v1-v2 VM series") {
		t.Errorf("matchLifecycleNotes(Standard_D4_v2, westeurope) = %q, want the v1-v2 price-increase note", got)
	}
	if strings.Contains(got, "Regional price increase") {
		t.Errorf("matchLifecycleNotes(Standard_D4_v2, westeurope) = %q, should not also match the regional price increase row (v1/v2 take the 25 percent change instead, per the field alert's do-not-double-count rule)", got)
	}
	got = matchLifecycleNotes("Standard_D4", "northeurope")
	if !strings.Contains(got, "price increase on v1-v2 VM series") {
		t.Errorf("matchLifecycleNotes(Standard_D4, northeurope) = %q, want the v1-v2 price-increase note", got)
	}
	if strings.Contains(got, "Regional price increase") {
		t.Errorf("matchLifecycleNotes(Standard_D4, northeurope) = %q, should not also match the regional price increase row (v1/v2 take the 25 percent change instead, per the field alert's do-not-double-count rule)", got)
	}

	// Decision D1/Finding 1: a v7 family must not match the region-scoped
	// price-increase rows (excludeFamilies carves it out), even though those
	// rows have an empty Families list that would otherwise match every SKU.
	got = matchLifecycleNotes("Standard_D4s_v7", "westeurope")
	if strings.Contains(got, "Regional price increase") {
		t.Errorf("matchLifecycleNotes(Standard_D4s_v7, westeurope) = %q, should not match the regional price increase row (v7 is excluded)", got)
	}
	got = matchLifecycleNotes("Standard_E4s_v7", "northeurope")
	if strings.Contains(got, "Regional price increase") {
		t.Errorf("matchLifecycleNotes(Standard_E4s_v7, northeurope) = %q, should not match the regional price increase row (v7 is excluded)", got)
	}

	// Finding 3: a VM on a no-new-RI-restricted family shows that note
	// regardless of region.
	got = matchLifecycleNotes("Standard_D4_v2", "eastus")
	if !strings.Contains(got, "No new 1-year Azure Reserved VM Instance") {
		t.Errorf("matchLifecycleNotes(Standard_D4_v2, eastus) = %q, want the no-new-1-year-RI note for Dv2", got)
	}
	got = matchLifecycleNotes("Standard_D4_v3", "eastus")
	if !strings.Contains(got, "No new 1-year or 3-year Azure Reserved VM Instance") {
		t.Errorf("matchLifecycleNotes(Standard_D4_v3, eastus) = %q, want the no-new-1-or-3-year-RI note for Dv3", got)
	}
}

// TestIsExemptCloud locks in Decision D3: Azure Government and Azure China are
// exempt from the field alert's lifecycle/pricing changes, so matchLifecycleNotes
// must suppress its output entirely (distinct from the "N/A" no-match value) on
// those clouds, while Azure Public (the default, no AZURE_CLOUD set) is
// unaffected.
func TestIsExemptCloud(t *testing.T) {
	tests := []struct {
		name       string
		azureCloud string
		want       bool
	}{
		{"default (Azure Public)", "", false},
		{"Azure Government", "AzureGovernment", true},
		{"Azure China", "AzureChina", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AZURE_CLOUD", tt.azureCloud)
			if got := isExemptCloud(); got != tt.want {
				t.Errorf("isExemptCloud() with AZURE_CLOUD=%q = %v, want %v", tt.azureCloud, got, tt.want)
			}
		})
	}
}

// TestMatchLifecycleNotes_SuppressedOnExemptCloud confirms matchLifecycleNotes
// returns an empty string (not "N/A") for a VM that would otherwise match
// lifecycle/pricing entries, once the active cloud is exempt.
func TestMatchLifecycleNotes_SuppressedOnExemptCloud(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "AzureGovernment")
	got := matchLifecycleNotes("Standard_D4_v3", "eastus")
	if got != "" {
		t.Errorf("matchLifecycleNotes(Standard_D4_v3, eastus) on an exempt cloud = %q, want empty string (suppressed)", got)
	}
}

// TestMatchImpactPercent covers P01-T02's numeric matcher, mirroring
// TestMatchLifecycleNotes' family/region coverage: a v1/v2-family VM, a
// regionally-affected VM, a VM matching no price-increase row, a VM matching
// only a non-monetary row (confirming retirement rows never leak into the
// cost calculation), and the exempt-cloud suppression case.
func TestMatchImpactPercent(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "")

	if percent, ok := matchImpactPercent("Standard_D4_v2", "eastus"); !ok || percent != 25.0 {
		t.Errorf("matchImpactPercent(Standard_D4_v2, eastus) = (%v, %v), want (25, true)", percent, ok)
	}

	if percent, ok := matchImpactPercent("Standard_D4s_v6", "westeurope"); !ok || percent != 9.0 {
		t.Errorf("matchImpactPercent(Standard_D4s_v6, westeurope) = (%v, %v), want (9, true)", percent, ok)
	}

	if percent, ok := matchImpactPercent("Standard_D4s_v6", "northeurope"); !ok || percent != 17.0 {
		t.Errorf("matchImpactPercent(Standard_D4s_v6, northeurope) = (%v, %v), want (17, true)", percent, ok)
	}

	if _, ok := matchImpactPercent("Standard_D4s_v6", "eastus"); ok {
		t.Error("matchImpactPercent(Standard_D4s_v6, eastus) should report not-ok (no applicable price-increase row)")
	}

	// Dv3 matches only the retirement row (non-monetary, no ImpactPercent),
	// so it must report not-ok rather than leaking a retirement row into a
	// cost calculation.
	if _, ok := matchImpactPercent("Standard_D4_v3", "eastus"); ok {
		t.Error("matchImpactPercent(Standard_D4_v3, eastus) should report not-ok (only the non-monetary retirement row matches)")
	}

	// v7 is excluded from the regional price-increase rows.
	if _, ok := matchImpactPercent("Standard_D4s_v7", "westeurope"); ok {
		t.Error("matchImpactPercent(Standard_D4s_v7, westeurope) should report not-ok (v7 is excluded from the regional row)")
	}
}

// TestMatchImpactPercent_SuppressedOnExemptCloud confirms matchImpactPercent
// reports not-ok on an exempt cloud for a VM that would otherwise match a
// price-increase row, mirroring TestMatchLifecycleNotes_SuppressedOnExemptCloud.
func TestMatchImpactPercent_SuppressedOnExemptCloud(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "AzureGovernment")
	if _, ok := matchImpactPercent("Standard_D4_v2", "eastus"); ok {
		t.Error("matchImpactPercent(Standard_D4_v2, eastus) on an exempt cloud should report not-ok (suppressed)")
	}
}

func TestMatchApplianceVendor(t *testing.T) {
	tests := []struct {
		name           string
		imagePublisher string
		imageOffer     string
		wantVendor     string
		wantOK         bool
	}{
		{"known publisher, no offer filter", "PaloAltoNetworks", "vmseries-flex", "Palo Alto Networks", true},
		{"known publisher, case-insensitive", "FORTINET", "fortigate-vm", "Fortinet", true},
		{"publisher with offer filter, matching offer", "citrix", "netscaler-adc-vpx", "Citrix ADC", true},
		{"publisher with offer filter, non-matching offer", "citrix", "unrelated-offer", "", false},
		{"unknown publisher", "contoso", "contoso-app", "", false},
		{"empty publisher", "", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVendor, gotOK := matchApplianceVendor(tt.imagePublisher, tt.imageOffer)
			if gotVendor != tt.wantVendor || gotOK != tt.wantOK {
				t.Errorf("matchApplianceVendor(%q, %q) = (%q, %v), want (%q, %v)",
					tt.imagePublisher, tt.imageOffer, gotVendor, gotOK, tt.wantVendor, tt.wantOK)
			}
		})
	}
}

func TestOrNA(t *testing.T) {
	if got := orNA(""); got != "N/A" {
		t.Errorf("orNA(\"\") = %q, want \"N/A\"", got)
	}
	if got := orNA("Zonal"); got != "Zonal" {
		t.Errorf("orNA(\"Zonal\") = %q, want \"Zonal\"", got)
	}
}

// TestPoolLevelColumns locks in the three pool-level columns' (P02) decode/
// output behavior: populated and reusing classifySeriesGeneration's verdict
// for a VMSS-backed row, notPoolManaged for a non-pool row, and the narrower
// "N/A" fallback for a pool-managed row whose VMSS-parent join did not
// resolve a model VM size.
func TestPoolLevelColumns(t *testing.T) {
	tests := []struct {
		name                  string
		hasScaleSet           bool
		orchestrationMode     string
		poolModelVMSize       string
		wantOrchestrationMode string
		wantPoolModelVMSize   string
		wantSeriesGeneration  string
	}{
		{"standalone VM reads notPoolManaged for all three", false, "", "", notPoolManaged, notPoolManaged, notPoolManaged},
		{"VMSS Flex row already on target generation", true, "Flexible", "Standard_D4s_v6", "Flexible", "Standard_D4s_v6", "v6"},
		{"Uniform row not yet on target generation", true, "Uniform", "Standard_D4s_v3", "Uniform", "Standard_D4s_v3", "v3"},
		{"pool-managed row whose join did not resolve a model VM size", true, "Uniform", "", "Uniform", "N/A", "N/A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := vmModernizationRow{
				HasScaleSet:       argBool(tt.hasScaleSet),
				OrchestrationMode: tt.orchestrationMode,
				PoolModelVMSize:   tt.poolModelVMSize,
			}
			gotMode, gotSize, gotSeries := r.poolLevelColumns()
			if gotMode != tt.wantOrchestrationMode || gotSize != tt.wantPoolModelVMSize || gotSeries != tt.wantSeriesGeneration {
				t.Errorf("poolLevelColumns() = (%q, %q, %q), want (%q, %q, %q)",
					gotMode, gotSize, gotSeries, tt.wantOrchestrationMode, tt.wantPoolModelVMSize, tt.wantSeriesGeneration)
			}
		})
	}
}

// TestMergeAVDHostPoolType locks in mergeAVDHostPoolType's (P04-T02) by-VM-ID
// matching behavior: an exact-case match, a match that only resolves after
// lowercasing (mirroring the AVD query's own tolower()-derived vmId values),
// and a row with no corresponding AVD entry, which must keep AVDHostPoolType
// empty rather than erroring or panicking.
func TestMergeAVDHostPoolType(t *testing.T) {
	tests := []struct {
		name            string
		rowID           string
		avdHostPoolByID map[string]string
		wantHostPool    string
	}{
		{
			name:            "exact-case match",
			rowID:           "/subscriptions/x/resourcegroups/rg1/providers/microsoft.compute/virtualmachines/vm1",
			avdHostPoolByID: map[string]string{"/subscriptions/x/resourcegroups/rg1/providers/microsoft.compute/virtualmachines/vm1": "Pooled"},
			wantHostPool:    "Pooled",
		},
		{
			name:            "case-differing match resolves via lowercasing",
			rowID:           "/subscriptions/X/resourceGroups/RG1/providers/Microsoft.Compute/virtualMachines/VM1",
			avdHostPoolByID: map[string]string{"/subscriptions/x/resourcegroups/rg1/providers/microsoft.compute/virtualmachines/vm1": "Personal"},
			wantHostPool:    "Personal",
		},
		{
			name:            "no matching AVD entry leaves AVDHostPoolType empty",
			rowID:           "/subscriptions/x/resourcegroups/rg1/providers/microsoft.compute/virtualmachines/vm2",
			avdHostPoolByID: map[string]string{"/subscriptions/x/resourcegroups/rg1/providers/microsoft.compute/virtualmachines/vm1": "Pooled"},
			wantHostPool:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := []vmModernizationRow{{ID: tt.rowID}}
			mergeAVDHostPoolType(rows, tt.avdHostPoolByID)
			if rows[0].AVDHostPoolType != tt.wantHostPool {
				t.Errorf("AVDHostPoolType = %q, want %q", rows[0].AVDHostPoolType, tt.wantHostPool)
			}
		})
	}
}

func TestLoadAROClusters(t *testing.T) {
	cluster := aroClusterRow{
		AROClusterID:       "/subscriptions/sub1/resourceGroups/parent/providers/Microsoft.RedHatOpenShift/openShiftClusters/cluster1",
		AROClusterName:     "cluster1",
		AROResourceGroupID: "/subscriptions/sub1/resourceGroups/aro-nodes",
	}
	mixedCaseCluster := cluster
	mixedCaseCluster.AROClusterID = strings.ToUpper(cluster.AROClusterID)
	mixedCaseCluster.AROResourceGroupID = strings.ToUpper(cluster.AROResourceGroupID)
	tests := []struct {
		name      string
		rows      []aroClusterRow
		raw       []json.RawMessage
		queryErr  error
		nilResult bool
		wantMatch bool
		wantLog   string
	}{
		{name: "successful lookup", rows: []aroClusterRow{cluster}, wantMatch: true},
		{name: "mixed-case ownership", rows: []aroClusterRow{mixedCaseCluster}, wantMatch: true},
		{name: "no clusters", rows: []aroClusterRow{}},
		{name: "nil data"},
		{name: "query failure", queryErr: errors.New("query unavailable"), wantLog: "ARO workload-pattern classification will be unavailable"},
		{name: "nil response", nilResult: true, wantLog: "ARO workload-pattern classification will be unavailable"},
		{name: "malformed row does not discard valid row", rows: []aroClusterRow{cluster},
			raw: []json.RawMessage{json.RawMessage(`{"AROClusterId":7}`)}, wantMatch: true, wantLog: "Skipping malformed ARO Cluster row"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			original := log.Logger
			log.Logger = zerolog.New(&output)
			t.Cleanup(func() { log.Logger = original })

			data := append([]json.RawMessage(nil), tt.raw...)
			for _, row := range tt.rows {
				raw, err := json.Marshal(row)
				if err != nil {
					t.Fatal(err)
				}
				data = append(data, raw)
			}
			ctx := context.Background()
			subscriptions := map[string]string{"sub1": "Subscription 1"}
			calls := 0
			query := func(gotCtx context.Context, kql string, gotSubs map[string]string, opts ...graph.QueryOptions) (*graph.GraphResult, error) {
				calls++
				if gotCtx != ctx || kql != discoverAssessAROQuery || !reflect.DeepEqual(gotSubs, subscriptions) || len(opts) != 0 {
					t.Fatal("ARO query did not preserve context, embedded query, subscriptions, and default options")
				}
				if tt.nilResult || tt.queryErr != nil {
					return nil, tt.queryErr
				}
				return &graph.GraphResult{Data: data}, nil
			}
			got := loadAROClusters(ctx, query, subscriptions)
			if calls != 1 {
				t.Fatalf("query called %d times, want one batched lookup", calls)
			}
			if tt.wantMatch {
				if len(got) != 1 {
					t.Fatalf("got %d ownership groups, want 1; logs: %s", len(got), output.String())
				}
				row, ok := got[strings.ToLower(cluster.AROResourceGroupID)]
				if !ok || row.AROClusterID != strings.ToLower(cluster.AROClusterID) || row.AROClusterName != cluster.AROClusterName {
					t.Errorf("unexpected normalized ownership: %+v", got)
				}
			} else if len(got) != 0 {
				t.Errorf("got ownership for absent data: %+v", got)
			}
			if tt.wantLog != "" && !strings.Contains(output.String(), tt.wantLog) {
				t.Errorf("logs %q do not contain %q", output.String(), tt.wantLog)
			}
			if tt.wantLog == "" && strings.Contains(output.String(), `"level":"warn"`) {
				t.Errorf("unexpected warning: %s", output.String())
			}
		})
	}
}

func TestMergeAROClusters(t *testing.T) {
	cluster := aroClusterRow{
		AROClusterID:       "/subscriptions/sub1/resourcegroups/parent/providers/microsoft.redhatopenshift/openshiftclusters/cluster1",
		AROClusterName:     "cluster1",
		AROResourceGroupID: "/subscriptions/sub1/resourcegroups/aro-nodes",
	}
	tests := []struct {
		name     string
		clusters map[string]aroClusterRow
		wantARO  bool
	}{
		{"matching cluster", map[string]aroClusterRow{cluster.AROResourceGroupID: cluster}, true},
		{"nil lookup", nil, false},
		{"empty lookup", map[string]aroClusterRow{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := []vmModernizationRow{
				{Name: "master", SubscriptionID: "SUB1", ResourceGroup: "ARO-Nodes"},
				{Name: "worker", SubscriptionID: "sub1", ResourceGroup: "aro-nodes"},
				{Name: "other-subscription", SubscriptionID: "sub2", ResourceGroup: "aro-nodes"},
				{Name: "other-group", SubscriptionID: "sub1", ResourceGroup: "other"},
				{Name: "missing-subscription", ResourceGroup: "aro-nodes"},
				{Name: "missing-group", SubscriptionID: "sub1"},
			}
			before := append([]vmModernizationRow(nil), rows...)
			mergeAROClusters(rows, tt.clusters)
			for i, row := range rows {
				want := tt.wantARO && i < 2
				if row.IsAROManaged != want {
					t.Errorf("row %q IsAROManaged = %v, want %v", row.Name, row.IsAROManaged, want)
				}
				row.IsAROManaged = false
				if !reflect.DeepEqual(row, before[i]) {
					t.Errorf("merge changed inventory or order at row %d", i)
				}
			}
			mergeAROClusters(rows, nil)
			for _, row := range rows {
				if row.IsAROManaged {
					t.Errorf("empty lookup retained stale ownership for %q", row.Name)
				}
			}
		})
	}
}

func TestToRecord_ColumnCountMatchesMetadata(t *testing.T) {
	// Pin AZURE_CLOUD to a non-exempt value (RV-002): toRecord ultimately
	// calls matchLifecycleNotes, which suppresses its output on Azure
	// Government/China, so this test must not depend on ambient environment
	// state to see the lifecycle/pricing column populated as expected.
	t.Setenv("AZURE_CLOUD", "")

	// The record toRecord() produces must have exactly one value per
	// ColumnMetadata entry, in the same order, including the broadened raw
	// inventory signal columns appended after Lifecycle/Pricing Notes.
	s := &Scanner{}
	meta := s.GetMetadata()

	row := vmModernizationRow{
		Name:            "vm1",
		ResourceGroup:   "rg1",
		Location:        "eastus",
		VMSize:          "Standard_D4s_v3",
		MembershipModel: "Standalone",
	}
	record := row.toRecord("sub1", &vmPriceCache{})

	if len(record) != len(meta.ColumnMetadata) {
		t.Fatalf("toRecord() produced %d columns, ColumnMetadata declares %d — they must match",
			len(record), len(meta.ColumnMetadata))
	}
}

func TestToRecord_AROOwnership(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "")
	const sku = "Standard_D4_v2"
	cache := &vmPriceCache{rows: map[skuRegionPair][]vmRetailPriceItem{
		{ArmSkuName: sku, ArmRegionName: "eastus"}: {{
			ArmSkuName: sku, ArmRegionName: "eastus", RetailPrice: 0.20,
			Type: "Consumption", MeterName: sku, ProductName: "Virtual Machines", SkuName: sku,
			EffectiveStartDate: "2020-01-01T00:00:00Z",
		}},
	}}
	header := (&Scanner{}).GetMetadata().HeaderRow()
	wantHeader := []string{
		"Subscription",
		"Resource Group",
		"Name",
		"Location",
		"Current VM Size",
		"Series Generation",
		"Workload Pattern",
		"Pool Membership",
		"Membership Model",
		"Orchestration Mode",
		"VMSS Name",
		"Pool Model VM Size",
		"Pool Model Series Generation",
		"Availability Set Name",
		"Hyper-V Generation",
		"Disk Controller Type",
		"Azure Disk Encryption",
		"Temp Disk Dependency",
		"SAP/ISV Certification Gate",
		"VM Power State",
		"Instance View Captured",
		"Hibernation Enabled",
		"Recommended Execution Method",
		"Suggested Target SKU",
		"Target SKU Compatibility Score",
		"Lifecycle/Pricing Notes",
		"Current Monthly Cost (USD)",
		"Projected Monthly Cost After Increase (USD)",
		"Monthly Cost Impact (USD)",
		"Zone Placement",
		"Availability Zone",
		"Platform Fault Domain",
		"OS Type",
		"OS Version",
		"Computer Name",
		"Is Marketplace VM",
		"Image Publisher",
		"Image Offer",
		"Image SKU",
		"Image Version",
		"OS Disk Type",
		"OS Disk Size (GB)",
		"Ephemeral OS Disk",
		"Data Disk Count",
		"ADE OS Disk (VM Property)",
		"Ultra SSD Enabled",
		"License Type",
		"NIC Count",
		"Accelerated Networking",
		"IP Forwarding",
		"Security Type",
		"Secure Boot Enabled",
		"vTPM Enabled",
	}
	if !reflect.DeepEqual(header, wantHeader) {
		t.Fatalf("report schema changed: got %v, want %v", header, wantHeader)
	}
	tests := []struct {
		name string
		row  vmModernizationRow
	}{
		{"standalone placement", vmModernizationRow{MembershipModel: "Standalone"}},
		{"availability set placement", vmModernizationRow{MembershipModel: "Availability Set", AvailabilitySetName: "as1"}},
		{"VMSS placement", vmModernizationRow{
			HasScaleSet: true, MembershipModel: "VMSS (Uniform)", VMSSName: "vmss1",
			OrchestrationMode: "Uniform", PoolModelVMSize: "Standard_D4s_v5",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := tt.row
			row.Name, row.ResourceGroup, row.SubscriptionID = "node1", "aro-nodes", "sub1"
			row.VMSize, row.Location, row.OSType = sku, "eastus", "Linux"
			row.HyperVGeneration, row.DiskControllerType = "V2", "SCSI"
			baseline := row.toRecord("Subscription 1", cache)
			rows := []vmModernizationRow{row}
			cluster := aroClusterRow{AROClusterID: "/subscriptions/sub1/resourcegroups/parent/providers/microsoft.redhatopenshift/openshiftclusters/cluster1", AROClusterName: "cluster1"}
			mergeAROClusters(rows, map[string]aroClusterRow{"/subscriptions/sub1/resourcegroups/aro-nodes": cluster})
			got := rows[0].toRecord("Subscription 1", cache)
			if len(got) != len(wantHeader) {
				t.Fatalf("got %d columns, want %d", len(got), len(wantHeader))
			}
			wantChanges := map[string]string{
				"Workload Pattern":               patternComputePool,
				"Pool Membership":                "ARO Cluster",
				"Temp Disk Dependency":           "N/A (pool/platform-managed)",
				"SAP/ISV Certification Gate":     "N/A (pool/platform-managed)",
				"Recommended Execution Method":   "Managed by pool/cluster — evaluate at the pool level",
				"Suggested Target SKU":           "N/A (pool-managed)",
				"Target SKU Compatibility Score": "N/A",
			}
			for i, column := range wantHeader {
				want, changed := wantChanges[column]
				if !changed {
					want = baseline[i]
				}
				if got[i] != want {
					t.Errorf("%s = %q, want %q", column, got[i], want)
				}
			}
			for column, want := range map[string]string{
				"Current Monthly Cost (USD)":                  formatUSD(0.20 * 730),
				"Projected Monthly Cost After Increase (USD)": formatUSD(0.20 * 730 * 1.25),
				"Monthly Cost Impact (USD)":                   formatUSD(0.20 * 730 * 0.25),
			} {
				for i, name := range wantHeader {
					if name == column && got[i] != want {
						t.Errorf("%s = %q, want %q", column, got[i], want)
					}
				}
			}
		})
	}
}

// TestToRecord_MonetaryImpactColumns locks in the confirmed fallback
// precedence for the three monetary-impact columns (current monthly cost,
// projected monthly cost, monthly impact): exempt cloud -> deallocated
// VM -> retail price lookup failure -> no applicable impactPercent ->
// full calculation. Pool-managed rows are no longer a separate fallback step:
// they compute the same per-instance cost as any other row (see P01-T01 of
// the pool-monetary-impact fix).
func TestToRecord_MonetaryImpactColumns(t *testing.T) {
	const (
		region   = "eastus"
		osType   = "Linux"
		priceUSD = 0.20
	)

	buildCache := func(sku string) *vmPriceCache {
		return &vmPriceCache{
			rows: map[skuRegionPair][]vmRetailPriceItem{
				{ArmSkuName: sku, ArmRegionName: region}: {
					{
						ArmSkuName:         sku,
						ArmRegionName:      region,
						RetailPrice:        priceUSD,
						Type:               "Consumption",
						MeterName:          sku,
						ProductName:        "Virtual Machines",
						SkuName:            sku,
						EffectiveStartDate: "2020-01-01T00:00:00Z",
					},
				},
			},
		}
	}
	emptyCache := &vmPriceCache{}

	baseRow := func(sku string) vmModernizationRow {
		return vmModernizationRow{
			Name:          "vm1",
			ResourceGroup: "rg1",
			Location:      region,
			VMSize:        sku,
			OSType:        osType,
		}
	}

	tests := []struct {
		name          string
		azureCloud    string
		row           vmModernizationRow
		cache         *vmPriceCache
		wantCurrent   string
		wantProjected string
		wantMonthly   string
	}{
		{
			name:          "exempt cloud suppresses all three columns",
			azureCloud:    "AzureUSGovernment",
			row:           baseRow("Standard_D4_v2"),
			cache:         buildCache("Standard_D4_v2"),
			wantCurrent:   "",
			wantProjected: "",
			wantMonthly:   "",
		},
		{
			name:       "deallocated VM",
			azureCloud: "",
			row: func() vmModernizationRow {
				r := baseRow("Standard_D4_v2")
				r.VMPowerState = "VM deallocated"
				return r
			}(),
			cache:         buildCache("Standard_D4_v2"),
			wantCurrent:   "N/A (VM deallocated)",
			wantProjected: "N/A (VM deallocated)",
			wantMonthly:   "N/A (VM deallocated)",
		},
		{
			name:       "stopped but allocated VM is still billed (case-insensitive)",
			azureCloud: "",
			row: func() vmModernizationRow {
				r := baseRow("Standard_D4_v2")
				r.VMPowerState = "VM STOPPED"
				return r
			}(),
			cache:         buildCache("Standard_D4_v2"),
			wantCurrent:   formatUSD(priceUSD * 730),
			wantProjected: formatUSD(priceUSD * 730 * 1.25),
			wantMonthly:   formatUSD(priceUSD * 730 * 0.25),
		},
		{
			// Pool-managed rows (here, an AVD pooled host pool, pattern B)
			// now compute the same per-instance cost as any other row using
			// its own current SKU — the isActionable gate no longer
			// suppresses this calculation (see P01-T01); only the unrelated
			// target-SKU suggestion still checks that gate. Standard_D4_v2 in
			// eastus matches the same global 25 percent v1/v2 price-increase
			// row as the "full calculation" case below.
			name:       "pool-managed VM (AVD pooled host pool) computes the same per-instance cost as any other row",
			azureCloud: "",
			row: func() vmModernizationRow {
				r := baseRow("Standard_D4_v2")
				r.AVDHostPoolType = "Pooled"
				return r
			}(),
			cache:         buildCache("Standard_D4_v2"),
			wantCurrent:   formatUSD(priceUSD * 730),
			wantProjected: formatUSD(priceUSD * 730 * 1.25),
			wantMonthly:   formatUSD(priceUSD * 730 * 0.25),
		},
		{
			name:          "retail price lookup failure (empty cache)",
			azureCloud:    "",
			row:           baseRow("Standard_D4_v2"),
			cache:         emptyCache,
			wantCurrent:   manualReviewRequired,
			wantProjected: manualReviewRequired,
			wantMonthly:   manualReviewRequired,
		},
		{
			// Standard_D4_v3 in eastus matches only the non-monetary
			// retirement row (see TestMatchImpactPercent), so it must report
			// the current cost but no price-change figures.
			name:          "no applicable impactPercent (Dv3 matches only the retirement row)",
			azureCloud:    "",
			row:           baseRow("Standard_D4_v3"),
			cache:         buildCache("Standard_D4_v3"),
			wantCurrent:   formatUSD(priceUSD * 730),
			wantProjected: formatUSD(priceUSD * 730),
			wantMonthly:   formatUSD(0),
		},
		{
			// Standard_D4_v2 in eastus matches the global 25 percent v1/v2
			// price-increase row (see TestMatchImpactPercent).
			name:          "full calculation (worked example, 25 percent impact)",
			azureCloud:    "",
			row:           baseRow("Standard_D4_v2"),
			cache:         buildCache("Standard_D4_v2"),
			wantCurrent:   formatUSD(priceUSD * 730),
			wantProjected: formatUSD(priceUSD * 730 * 1.25),
			wantMonthly:   formatUSD(priceUSD * 730 * 0.25),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AZURE_CLOUD", tt.azureCloud)
			current, projected, monthly := tt.row.computeMonetaryImpact(tt.cache)
			if current != tt.wantCurrent || projected != tt.wantProjected || monthly != tt.wantMonthly {
				t.Errorf("computeMonetaryImpact() = (%q, %q, %q), want (%q, %q, %q)",
					current, projected, monthly,
					tt.wantCurrent, tt.wantProjected, tt.wantMonthly)
			}
		})
	}
}

// TestBuildGeneralNoticesSheet locks in Decision D4/Finding 6: account-wide
// notices are reported once per run on their own sheet (not Metadata-bearing,
// following the region-selection plugin's additional-sheet precedent), and
// suppressed entirely on an exempt cloud per Decision D3.
func TestBuildGeneralNoticesSheet(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "")
	sheet := buildGeneralNoticesSheet()
	if sheet == nil {
		t.Fatal("buildGeneralNoticesSheet() = nil on Azure Public, want a non-nil sheet")
	}
	if sheet.Metadata.ColumnMetadata != nil {
		t.Errorf("buildGeneralNoticesSheet().Metadata.ColumnMetadata = %v, want zero value (no per-VM metadata)", sheet.Metadata.ColumnMetadata)
	}
	if len(sheet.Table) != len(generalNotices)+1 {
		t.Errorf("buildGeneralNoticesSheet().Table has %d rows, want %d (1 header + %d notices)",
			len(sheet.Table), len(generalNotices)+1, len(generalNotices))
	}
	if len(sheet.SheetName) == 0 || len([]rune(sheet.SheetName)) > 31 {
		t.Errorf("buildGeneralNoticesSheet().SheetName = %q (%d runes), want 1-31 runes for Excel", sheet.SheetName, len([]rune(sheet.SheetName)))
	}
}

func TestBuildGeneralNoticesSheet_SuppressedOnExemptCloud(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "AzureChina")
	if sheet := buildGeneralNoticesSheet(); sheet != nil {
		t.Errorf("buildGeneralNoticesSheet() on Azure China = %+v, want nil (suppressed)", sheet)
	}
}

func TestMatchLifecycleNotes_BSeriesV1(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "")

	for _, region := range []string{"eastus", "westeurope"} {
		got := matchLifecycleNotes("Standard_B2s", region)
		if !strings.Contains(got, "25 percent price increase on v1-v2 VM series") {
			t.Errorf("matchLifecycleNotes(Standard_B2s, %s) = %q, want the v1-v2 price-increase note", region, got)
		}
	}
	if pct, ok := matchImpactPercent("Standard_B2s", "westeurope"); !ok || pct != 25 {
		t.Errorf("matchImpactPercent(Standard_B2s, westeurope) = %v, %v, want 25, true", pct, ok)
	}
}

func TestMatchImpactPercent_NPHCNotDoubleCounted(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "")

	for _, sku := range []string{"Standard_NP10s", "Standard_HC44rs"} {
		for _, region := range []string{"eastus", "westeurope", "northeurope", "francecentral"} {
			pct, ok := matchImpactPercent(sku, region)
			if !ok || pct != 25 {
				t.Errorf("matchImpactPercent(%s, %s) = %v, %v, want 25, true", sku, region, pct, ok)
			}
			if got := matchLifecycleNotes(sku, region); strings.Contains(got, "Regional price increase") {
				t.Errorf("matchLifecycleNotes(%s, %s) = %q, should not include a regional increase", sku, region, got)
			}
		}
	}
}

func TestBuildAVDHostPoolByVMId_KeepsEveryHostInPool(t *testing.T) {
	rows := []avdHostPoolRow{
		{HostPoolArmID: "/subscriptions/s/resourcegroups/rg/providers/microsoft.desktopvirtualization/hostpools/hp1", AVDHostPoolType: "Pooled"},
		{HostPoolArmID: "/subscriptions/s/resourcegroups/rg/providers/microsoft.desktopvirtualization/hostpools/hp1", VMId: "/subscriptions/s/resourcegroups/rg/providers/microsoft.compute/virtualmachines/vm1"},
		{HostPoolArmID: "/subscriptions/s/resourcegroups/rg/providers/microsoft.desktopvirtualization/hostpools/hp1", VMId: "/subscriptions/s/resourcegroups/rg/providers/microsoft.compute/virtualmachines/vm2"},
		{HostPoolArmID: "/subscriptions/s/resourcegroups/rg/providers/microsoft.desktopvirtualization/hostpools/hp2", AVDHostPoolType: "Personal"},
		{HostPoolArmID: "/subscriptions/s/resourcegroups/rg/providers/microsoft.desktopvirtualization/hostpools/hp2", VMId: "/subscriptions/s/resourcegroups/rg/providers/microsoft.compute/virtualmachines/vm3"},
		{HostPoolArmID: "/subscriptions/s/resourcegroups/rg/providers/microsoft.desktopvirtualization/hostpools/missing", VMId: "/subscriptions/s/resourcegroups/rg/providers/microsoft.compute/virtualmachines/vm4"},
	}
	got := buildAVDHostPoolByVMId(rows)
	want := map[string]string{
		"/subscriptions/s/resourcegroups/rg/providers/microsoft.compute/virtualmachines/vm1": "Pooled",
		"/subscriptions/s/resourcegroups/rg/providers/microsoft.compute/virtualmachines/vm2": "Pooled",
		"/subscriptions/s/resourcegroups/rg/providers/microsoft.compute/virtualmachines/vm3": "Personal",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildAVDHostPoolByVMId = %v, want %v", got, want)
	}
}

func TestMatchImpactPercentAt_EffectiveDate(t *testing.T) {
	t.Setenv("AZURE_CLOUD", "")

	before := time.Date(2027, 1, 31, 23, 59, 59, 0, time.UTC)
	onDate := time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)

	if pct, ok := matchImpactPercentAt("Standard_D4_v2", "eastus", before); !ok || pct != 25 {
		t.Errorf("before effective date = %v, %v, want 25, true", pct, ok)
	}
	if pct, ok := matchImpactPercentAt("Standard_D4_v2", "eastus", onDate); ok {
		t.Errorf("on effective date = %v, %v, want no pending percent (already in current rates)", pct, ok)
	}
}

func TestGenericVMSSRecommendsManualValidation(t *testing.T) {
	meta := (&Scanner{}).GetMetadata().ColumnMetadata
	col := func(rec []string, name string) string {
		for i, c := range meta {
			if c.Name == name {
				return rec[i]
			}
		}
		return ""
	}

	vmss := vmModernizationRow{HasScaleSet: true, VMSize: "Standard_D4_v3", HyperVGeneration: "V2"}.toRecord("sub", &vmPriceCache{})
	if got := col(vmss, "Recommended Execution Method"); !strings.HasPrefix(got, manualReviewRequired) {
		t.Errorf("generic VMSS execution method = %q, want Manual Review Required", got)
	}
	for _, name := range []string{"Temp Disk Dependency", "SAP/ISV Certification Gate"} {
		if got := col(vmss, name); got != manualReviewRequired {
			t.Errorf("generic VMSS %s = %q, want %q", name, got, manualReviewRequired)
		}
	}

	aks := vmModernizationRow{HasScaleSet: true, IsAKSManaged: true, VMSize: "Standard_D4_v3", HyperVGeneration: "V2"}.toRecord("sub", &vmPriceCache{})
	if got := col(aks, "Recommended Execution Method"); strings.HasPrefix(got, manualReviewRequired) {
		t.Errorf("AKS node execution method = %q, want pool-level text", got)
	}
	if got := col(aks, "SAP/ISV Certification Gate"); got != "N/A (pool/platform-managed)" {
		t.Errorf("AKS node SAP/ISV Certification Gate = %q, want N/A (pool/platform-managed)", got)
	}
}

func TestMergeVMSSPools(t *testing.T) {
	rows := []vmModernizationRow{{VMSSID: "/subscriptions/s/resourcegroups/rg/providers/microsoft.compute/virtualmachinescalesets/a"}, {}}
	mergeVMSSPools(rows, map[string]vmssPoolRow{
		"/subscriptions/s/resourcegroups/rg/providers/microsoft.compute/virtualmachinescalesets/a": {OrchestrationMode: "Flexible", PoolModelVMSize: "Standard_D4s_v5"},
	})
	if rows[0].OrchestrationMode != "Flexible" || rows[0].PoolModelVMSize != "Standard_D4s_v5" {
		t.Errorf("pool fields not merged: %+v", rows[0])
	}
	if rows[1].OrchestrationMode != "" || rows[1].PoolModelVMSize != "" {
		t.Errorf("row without VMSSId must stay empty: %+v", rows[1])
	}
}
