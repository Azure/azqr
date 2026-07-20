// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sku

import (
	"strings"

	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

// SKURestriction mirrors the ARM restriction shape shared by Compute and Storage SKU responses.
type SKURestriction struct {
	Type            string `json:"type"`
	ReasonCode      string `json:"reasonCode"`
	RestrictionInfo struct {
		Zones []string `json:"zones"`
	} `json:"restrictionInfo"`
}

// SKUCapability mirrors the ARM capability shape (e.g. "available": "false").
type SKUCapability struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CheckStandardRestrictions interprets the ARM restrictions/capabilities fields
// that are common to Compute (VMs, VMSS, Disks) and Storage SKU responses.
//
// Priority order (highest severity first):
//  1. Location restriction → Unavailable (hard block) or Restricted (quota-liftable)
//  2. Zone restriction     → ZoneRestricted with blocked zones populated
//  3. Capability "available=false" → Unavailable
//  4. Default              → Available
func CheckStandardRestrictions(restrictions []SKURestriction, capabilities []SKUCapability) types.SKUAvailability {
	for _, r := range restrictions {
		if !strings.EqualFold(r.Type, "Location") {
			continue
		}
		if strings.EqualFold(r.ReasonCode, "NotAvailableForSubscription") {
			return types.SKUAvailability{State: types.SKURestricted}
		}
		return types.SKUAvailability{State: types.SKUUnavailable}
	}

	var blockedZones []string
	for _, r := range restrictions {
		if strings.EqualFold(r.Type, "Zone") {
			blockedZones = append(blockedZones, r.RestrictionInfo.Zones...)
		}
	}
	if len(blockedZones) > 0 {
		return types.SKUAvailability{State: types.SKUZoneRestricted, BlockedZones: blockedZones}
	}

	for _, c := range capabilities {
		if strings.EqualFold(c.Name, "available") && !strings.EqualFold(c.Value, "true") {
			return types.SKUAvailability{State: types.SKUUnavailable}
		}
	}

	return types.SKUAvailability{State: types.SKUAvailable}
}
