// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package vmmodernization

import (
	_ "embed"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed appliance-vendors.yaml
var rawApplianceVendorsYAML []byte

// applianceVendorEntry is one row of the embedded, auditable certified-ISV-
// appliance vendor dataset (Discover article pattern G). It is kept separate
// from classifyWorkloadPattern's Go code so the vendor list stays reviewable
// and independently updatable, the same rationale lifecycleEntry already
// applies to the lifecycle/pricing dataset.
type applianceVendorEntry struct {
	Publisher     string   `yaml:"publisher"`
	OfferContains []string `yaml:"offerContains"`
	Vendor        string   `yaml:"vendor"`
}

var applianceVendorEntries []applianceVendorEntry

func init() {
	if err := yaml.Unmarshal(rawApplianceVendorsYAML, &applianceVendorEntries); err != nil {
		applianceVendorEntries = []applianceVendorEntry{}
	}
}

// matchApplianceVendor reports whether imagePublisher (matched exactly,
// case-insensitive) identifies a known certified-ISV-appliance vendor from the
// embedded dataset. When the matched entry also declares offerContains
// substrings, imageOffer must contain at least one of them (case-insensitive)
// for the match to hold; an entry with no offerContains matches any offer from
// that publisher. Returns ok=false when no entry matches — callers must not
// force pattern G on a non-match.
func matchApplianceVendor(imagePublisher, imageOffer string) (vendor string, ok bool) {
	publisherNorm := strings.ToLower(strings.TrimSpace(imagePublisher))
	if publisherNorm == "" {
		return "", false
	}
	offerNorm := strings.ToLower(imageOffer)

	for _, entry := range applianceVendorEntries {
		if strings.ToLower(entry.Publisher) != publisherNorm {
			continue
		}
		if len(entry.OfferContains) == 0 {
			return entry.Vendor, true
		}
		for _, substr := range entry.OfferContains {
			if strings.Contains(offerNorm, strings.ToLower(substr)) {
				return entry.Vendor, true
			}
		}
	}
	return "", false
}
