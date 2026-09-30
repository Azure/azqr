// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package modelinfo

import "testing"

func TestDeploymentModeForSKU(t *testing.T) {
	tests := map[string]string{
		"GlobalStandard":             ModeGlobal,
		"DataZoneStandard":           ModeDataZone,
		"Standard":                   ModeRegional,
		"GlobalProvisionedManaged":   ModeGlobalProvisioned,
		"DataZoneProvisionedManaged": ModeDataZoneProvisioned,
		"ProvisionedManaged":         ModeRegionalProvisioned,
		"GlobalBatch":                modeBatch,
		"DataZoneBatch":              modeBatch,
		"global-standard":            ModeGlobal,
		"Global Standard":            ModeGlobal,
		"":                           ModeUnknown,
		"Mystery":                    ModeUnknown,
	}
	for sku, want := range tests {
		if got := DeploymentModeForSKU(sku); got != want {
			t.Errorf("deploymentModeForSKU(%q) = %q, want %q", sku, got, want)
		}
	}
}

func TestProviderFamilyForModel(t *testing.T) {
	tests := map[string]string{
		"gpt-4o":          FamilyAzureOpenAI,
		"claude-sonnet-4": familyClaude,
		"mistral-large":   familyPartner,
	}
	for model, want := range tests {
		if got := ProviderFamilyForModel(model); got != want {
			t.Errorf("providerFamilyForModel(%q) = %q, want %q", model, got, want)
		}
	}
}
