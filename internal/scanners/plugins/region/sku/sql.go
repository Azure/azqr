// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package sku

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azqr/internal/az"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
)

// The SQL capabilities endpoint returns a nested object, not a standard
// { "value": [...] } envelope, so FetchPagedSKUs cannot be used here.
// Both SQL DB and SQL MI share the same endpoint and response shape.

// sqlCapabilities is the top-level response from
// /providers/Microsoft.Sql/locations/{loc}/capabilities
type sqlCapabilities struct {
	SupportedServerVersions []sqlServerVersion `json:"supportedServerVersions"`
	// SQL MI uses a different root key
	SupportedManagedInstanceVersions []sqlManagedInstanceVersion `json:"supportedManagedInstanceVersions"`
}

type sqlServerVersion struct {
	SupportedEditions []sqlEdition `json:"supportedEditions"`
}

type sqlEdition struct {
	SupportedServiceLevelObjectives []sqlSLO `json:"supportedServiceLevelObjectives"`
}

type sqlSLO struct {
	SKU    sqlSKU `json:"sku"`
	Status string `json:"status"` // "Available", "Default", "Disabled"
}

type sqlSKU struct {
	Name string `json:"name"`
	Tier string `json:"tier"`
}

// sqlManagedInstanceVersion is the root for SQL MI entries in the capabilities response.
type sqlManagedInstanceVersion struct {
	SupportedEditions []sqlMIEdition `json:"supportedEditions"`
}

// sqlMIEdition is one edition entry in the SQL MI capabilities response.
// Name is "GeneralPurpose" or "BusinessCritical" — it provides the tier prefix
// for constructing the composite sku.name key (e.g. "GP_Gen5", "BC_G8IM").
type sqlMIEdition struct {
	Name              string        `json:"name"` // e.g. "GeneralPurpose", "BusinessCritical"
	SupportedFamilies []sqlMIFamily `json:"supportedFamilies"`
}

type sqlMIFamily struct {
	Name string `json:"name"` // e.g. "Gen5", "Gen8im", "Gen8i"
}

// sqlMITierCode maps the SQL MI edition name (from capabilities response) to the
// abbreviated tier prefix used in sku.name by ARG and ARM templates.
var sqlMITierCode = map[string]string{
	"generalpurpose":     "GP",
	"businesscritical":   "BC",
	"nextgengeneralpurpose": "NGPGEN5", // future-proofing
}

// fetchSQLCapabilities fetches and unmarshals the SQL capabilities payload for a region.
func fetchSQLCapabilities(ctx context.Context, subID, region string, client *az.HttpClient) (*sqlCapabilities, error) {
	url := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/providers/Microsoft.Sql/locations/%s/capabilities?api-version=2021-11-01",
		subID, region,
	)
	body, err := client.Do(ctx, url)
	if err != nil {
		return nil, err
	}
	var caps sqlCapabilities
	if err := json.Unmarshal(body, &caps); err != nil {
		return nil, fmt.Errorf("sql: parse capabilities response: %w", err)
	}
	return &caps, nil
}

// ── SQL Database ─────────────────────────────────────────────────────────────

// SQLDatabaseProvider checks SQL Database SKU availability.
// Previously broken: the old JSON config used startPath traversal that was never
// implemented in Go, causing the endpoint to return 0 SKUs silently.
type SQLDatabaseProvider struct{}

func (SQLDatabaseProvider) ResourceType() string { return "microsoft.sql/servers/databases" }

func (SQLDatabaseProvider) FetchSKUs(ctx context.Context, subID, region string, client *az.HttpClient) (map[string]types.SKUAvailability, error) {
	caps, err := fetchSQLCapabilities(ctx, subID, region, client)
	if err != nil {
		return nil, err
	}

	result := make(map[string]types.SKUAvailability)
	for _, ver := range caps.SupportedServerVersions {
		for _, ed := range ver.SupportedEditions {
			for _, slo := range ed.SupportedServiceLevelObjectives {
				if slo.SKU.Name == "" {
					continue
				}
				avail := types.SKUAvailability{State: types.SKUAvailable}
				if strings.EqualFold(slo.Status, "Disabled") {
					avail.State = types.SKUUnavailable
				}
				result[strings.ToLower(slo.SKU.Name)] = avail
			}
		}
	}
	return result, nil
}

// ── SQL Managed Instance ─────────────────────────────────────────────────────

// SQLManagedInstanceProvider checks SQL Managed Instance SKU availability.
// Previously broken for the same reason as SQLDatabaseProvider.
type SQLManagedInstanceProvider struct{}

func (SQLManagedInstanceProvider) ResourceType() string { return "microsoft.sql/managedinstances" }

func (SQLManagedInstanceProvider) FetchSKUs(ctx context.Context, subID, region string, client *az.HttpClient) (map[string]types.SKUAvailability, error) {
	caps, err := fetchSQLCapabilities(ctx, subID, region, client)
	if err != nil {
		return nil, err
	}

	result := make(map[string]types.SKUAvailability)
	for _, ver := range caps.SupportedManagedInstanceVersions {
		for _, ed := range ver.SupportedEditions {
			// Derive the tier prefix ("GP", "BC") from the edition name.
			// Fall back to the raw name if not found in the map.
			tierCode, ok := sqlMITierCode[strings.ToLower(ed.Name)]
			if !ok {
				tierCode = ed.Name
			}
			for _, fam := range ed.SupportedFamilies {
				if fam.Name == "" {
					continue
				}
				// Construct the composite key matching ARG sku.name (e.g. "gp_gen5").
				key := strings.ToLower(tierCode + "_" + fam.Name)
				result[key] = types.SKUAvailability{State: types.SKUAvailable}
			}
		}
	}
	return result, nil
}

func init() {
	Register(SQLDatabaseProvider{})
	Register(SQLManagedInstanceProvider{})
}
