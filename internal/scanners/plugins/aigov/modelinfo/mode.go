// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package modelinfo

import (
	"regexp"
	"strings"
)

// Canonical deployment modes derived from the exact deployment SKU name,
// following TokenLens's foundry_workflow/discovery.py mapping.
const (
	ModeGlobal              = "global"
	ModeDataZone            = "data_zone"
	ModeRegional            = "regional"
	modeBatch               = "batch"
	ModeGlobalProvisioned   = "global_provisioned"
	ModeDataZoneProvisioned = "data_zone_provisioned"
	ModeRegionalProvisioned = "regional_provisioned"
	ModeUnknown             = "unknown"
)

// Provider families used to decide which purchasing programs apply.
const (
	FamilyAzureOpenAI = "azure_openai"
	familyClaude      = "claude_foundry"
	familyPartner     = "partner_model"
)

type skuMode struct {
	sku  string
	mode string
}

// skuModes is ordered: exact matches are tried first, then prefix/suffix
// matches in this order, so more specific names must precede shorter ones.
var skuModes = []skuMode{
	{"globalprovisionedmanaged", ModeGlobalProvisioned},
	{"datazoneprovisionedmanaged", ModeDataZoneProvisioned},
	{"provisionedmanaged", ModeRegionalProvisioned},
	{"globalbatch", modeBatch},
	{"datazonebatch", modeBatch},
	{"globalstandard", ModeGlobal},
	{"datazonestandard", ModeDataZone},
	{"standard", ModeRegional},
	{"batch", modeBatch},
}

var skuSeparatorRe = regexp.MustCompile(`[\s_-]+`)

// DeploymentModeForSKU maps an exact deployment SKU name to a canonical mode.
// A missing or unrecognized SKU is reported as modeUnknown rather than
// falling back to Global, because pricing and PTU sizing differ per mode.
func DeploymentModeForSKU(sku string) string {
	normalized := skuSeparatorRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(sku)), "")
	if normalized == "" {
		return ModeUnknown
	}
	for _, c := range skuModes {
		if normalized == c.sku {
			return c.mode
		}
	}
	for _, c := range skuModes {
		if strings.HasPrefix(normalized, c.sku) || strings.HasSuffix(normalized, c.sku) {
			return c.mode
		}
	}
	return ModeUnknown
}

// IsProvisionedMode reports whether the deployment is already billed per PTU.
func IsProvisionedMode(mode string) bool {
	return mode == ModeGlobalProvisioned || mode == ModeDataZoneProvisioned || mode == ModeRegionalProvisioned
}

// IsBatchMode reports whether the deployment is billed with batch meters.
func IsBatchMode(mode string) bool {
	return mode == modeBatch
}

// ProviderFamilyForModel classifies a model as first-party Azure OpenAI,
// Claude on Foundry, or another partner model.
func ProviderFamilyForModel(modelName string) string {
	switch InferPublisher(modelName) {
	case "":
		return FamilyAzureOpenAI
	case "anthropic":
		return familyClaude
	default:
		return familyPartner
	}
}
