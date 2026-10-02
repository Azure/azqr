// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmmodernization

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/skus"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"gopkg.in/yaml.v3"
)

//go:embed lifecycle-pricing.yaml
var rawLifecycleYAML []byte

// lifecycleEntry is one row of the embedded, dated lifecycle/pricing reference
// dataset. It is kept separate from live Azure Resource Graph data (plan
// decision D2) so "what azqr asserts" stays auditable and independently
// updatable from "what azqr observes live".
type lifecycleEntry struct {
	Families []string `yaml:"families"`
	// ExcludeFamilies removes otherwise-matching families from this row's
	// match set, independent of whether Families is empty ("applies
	// broadly") or populated. This lets a broad row state "every family
	// except these" without an exhaustive explicit allow-list (plan decision
	// D1), for example the regional price-increase row excluding v7 (which
	// the field alert explicitly carves out of the regional uplift).
	ExcludeFamilies []string `yaml:"excludeFamilies"`
	Regions         []string `yaml:"regions"`
	Impact          string   `yaml:"impact"`
	// ImpactPercent is the machine-readable numeric form of a price-increase
	// row's prose Impact string (plan decision D1). It is left nil/unset for
	// rows that are not price increases (retirement, capacity-growth,
	// RI-purchase-restriction rows), so downstream monetary-impact code can
	// distinguish "no applicable price change" from a populated figure
	// without parsing the prose Impact text.
	ImpactPercent *float64 `yaml:"impactPercent"`
	EffectiveDate string   `yaml:"effectiveDate"`
}

var lifecycleEntries []lifecycleEntry

func init() {
	if err := yaml.Unmarshal(rawLifecycleYAML, &lifecycleEntries); err != nil {
		lifecycleEntries = []lifecycleEntry{}
	}
}

// isExemptCloud reports whether the active cloud configuration — as
// determined by az.GetCloudConfiguration(), itself driven by the
// AZURE_CLOUD/AZURE_AUTHORITY_HOST environment variables rather than any
// per-call parameter — is Azure Government or Azure China (via China's
// 21Vianet operator, which uses the same cloud.AzureChina configuration).
// The field alert states its VM lifecycle and pricing changes explicitly do
// not apply to these clouds (plan decision D3), so lifecycle/pricing notes
// are suppressed entirely for them rather than annotated as "not applicable".
func isExemptCloud() bool {
	cfg := az.GetCloudConfiguration()
	return cfg.ActiveDirectoryAuthorityHost == cloud.AzureGovernment.ActiveDirectoryAuthorityHost ||
		cfg.ActiveDirectoryAuthorityHost == cloud.AzureChina.ActiveDirectoryAuthorityHost
}

// matchLifecycleNotes returns every lifecycle/pricing entry whose family list
// matches the VM's current size (or that applies broadly via an empty family
// list) AND whose region list matches the VM's location (or that applies to
// all regions via an empty region list), concatenated into one string. A VM
// matching no entry returns an explicit "no notes" value rather than a blank
// string, so absence of risk is visibly distinct from an unpopulated column.
// On an exempt cloud (Azure Government/China), this returns an empty string
// instead — distinct from the "N/A" no-match value — because the notes are
// deliberately suppressed, not merely inapplicable to this particular VM.
func matchLifecycleNotes(vmSize, location string) string {
	if isExemptCloud() {
		return ""
	}

	familyNorm := normalizeFamily(vmSize)
	regionNorm := normalizeRegion(location)

	var notes []string
	for _, entry := range lifecycleEntries {
		familyMatches := (len(entry.Families) == 0 || matchesAnyFamily(familyNorm, entry.Families)) &&
			!matchesAnyFamily(familyNorm, entry.ExcludeFamilies)
		regionMatches := len(entry.Regions) == 0 || matchesAnyRegion(regionNorm, entry.Regions)
		if familyMatches && regionMatches {
			notes = append(notes, fmt.Sprintf("%s (effective %s)", entry.Impact, entry.EffectiveDate))
		}
	}

	if len(notes) == 0 {
		return "N/A"
	}
	return strings.Join(notes, "; ")
}

// matchImpactPercent returns the single applicable price-increase percentage
// for a VM's size and location, using the same family/region matching rules
// matchLifecycleNotes already applies, so a reader never sees prose and a
// cost figure disagree about which lifecycle/pricing row applies. Because the
// dataset's excludeFamilies design already guarantees the v1/v2 25 percent row
// and the five regional rows are mutually exclusive per VM, an unexpected
// multiple match is treated as "no applicable percent" rather than guessed or
// summed. Returns ok=false (no applicable percent) on an exempt cloud,
// mirroring matchLifecycleNotes' suppression.
func matchImpactPercent(vmSize, location string) (percent float64, ok bool) {
	if isExemptCloud() {
		return 0, false
	}

	familyNorm := normalizeFamily(vmSize)
	regionNorm := normalizeRegion(location)

	var matches []float64
	for _, entry := range lifecycleEntries {
		if entry.ImpactPercent == nil {
			continue
		}
		familyMatches := (len(entry.Families) == 0 || matchesAnyFamily(familyNorm, entry.Families)) &&
			!matchesAnyFamily(familyNorm, entry.ExcludeFamilies)
		regionMatches := len(entry.Regions) == 0 || matchesAnyRegion(regionNorm, entry.Regions)
		if familyMatches && regionMatches {
			matches = append(matches, *entry.ImpactPercent)
		}
	}

	if len(matches) != 1 {
		return 0, false
	}
	return matches[0], true
}

// normalizeFamily resolves the VM size's SKU.Family (via skus.Lookup, the same
// known_skus.yaml-backed accessor P03-T01 uses) and strips the "standard"
// prefix and "Family" suffix, returning a lowercase short tag (for example
// "standardDSv3Family" -> "dsv3"). Returns "" when the SKU is unknown; the
// caller then only matches family-agnostic (broadly applicable) entries.
func normalizeFamily(vmSize string) string {
	sku, ok := skus.Lookup(vmSize)
	if !ok {
		return ""
	}
	f := strings.ToLower(sku.Family)
	f = strings.TrimPrefix(f, "standard")
	f = strings.TrimSuffix(f, "family")
	return f
}

// normalizeRegion lowercases and strips spaces/hyphens from an Azure region
// value so it reliably matches both the ARG `location` field's already-
// normalized form (for example "westeurope") and any more human-readable
// region string that might appear in the dataset or a caller.
func normalizeRegion(location string) string {
	r := strings.ToLower(location)
	r = strings.ReplaceAll(r, " ", "")
	r = strings.ReplaceAll(r, "-", "")
	return r
}

// matchesAnyRegion reports whether the normalized region tag matches one of
// the dataset's region tags (each also normalized the same way).
func matchesAnyRegion(regionNorm string, regions []string) bool {
	if regionNorm == "" {
		return false
	}
	for _, r := range regions {
		if regionNorm == normalizeRegion(r) {
			return true
		}
	}
	return false
}

// matchesAnyFamily reports whether the normalized SKU family tag matches one of
// the dataset's family tags. It also tries a trailing "s" alias (for example
// family tag "hcs" matching dataset tag "HC") because some Azure family name
// strings carry a storage-support "S" suffix absent from the short marketing
// name the user-supplied table uses.
func matchesAnyFamily(familyNorm string, families []string) bool {
	if familyNorm == "" {
		return false
	}
	for _, f := range families {
		tag := strings.ToLower(f)
		if familyNorm == tag {
			return true
		}
		if strings.HasSuffix(familyNorm, "s") && strings.TrimSuffix(familyNorm, "s") == tag {
			return true
		}
	}
	return false
}
