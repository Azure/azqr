// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package excel

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/plugins"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/Azure/azqr/internal/scanners/plugins/region/config"
	"github.com/Azure/azqr/internal/scanners/plugins/region/types"
	"github.com/Azure/azqr/internal/skus"
)

// safeSheetName truncates s to the 31-character maximum allowed by Excel sheet names.
func safeSheetName(s string) string {
	const maxLen = 31
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}

// BuildSvcAvailSheets generates one ExternalPluginOutput per unique target region containing
// per-resource-type and per-SKU availability, mirroring the SvcAvail_<Region> sheets from
// the AzRegionSelection PowerShell toolkit.
func BuildSvcAvailSheets(allResults []types.RegionComparison, inventory *types.ResourceInventory) []plugins.ExternalPluginOutput {
	if inventory == nil || len(allResults) == 0 {
		return nil
	}

	// Collect unique target regions (preserve insertion order via a slice + set)
	seenTargets := map[string]bool{}
	targetRegions := []string{}
	for _, r := range allResults {
		if !seenTargets[r.TargetRegion] {
			seenTargets[r.TargetRegion] = true
			targetRegions = append(targetRegions, r.TargetRegion)
		}
	}
	sort.Strings(targetRegions)

	// Pre-compute: for each resource type, which source regions contain it?
	implRegionsByType := map[string][]string{}
	for region, types := range inventory.ResourceTypesByRegion {
		for rt := range types {
			implRegionsByType[rt] = append(implRegionsByType[rt], region)
		}
	}
	for rt := range implRegionsByType {
		sort.Strings(implRegionsByType[rt])
	}

	sheets := make([]plugins.ExternalPluginOutput, 0, len(targetRegions))

	for _, targetRegion := range targetRegions {
		// Aggregate missing resource types and SKUs across ALL comparisons to this target
		missingTypeSet := map[string]bool{}
		missingSKUSet := map[string]bool{}
		for _, comp := range allResults {
			if comp.TargetRegion != targetRegion {
				continue
			}
			for _, t := range comp.MissingResourceTypes {
				missingTypeSet[strings.ToLower(t)] = true
			}
			for _, s := range comp.MissingSKUs {
				missingSKUSet[strings.ToLower(s)] = true
			}
		}

		header := []string{
			"ResourceType",
			"ResourceCount",
			"ImplementedRegions",
			"SKUCount",
			"SKU",
			"SKU available",
			"Service available",
		}
		rows := [][]string{header}

		// Sort resource types for deterministic output
		sortedTypes := make([]string, 0, len(inventory.ResourceTypes))
		for rt := range inventory.ResourceTypes {
			sortedTypes = append(sortedTypes, rt)
		}
		sort.Strings(sortedTypes)

		for _, rt := range sortedTypes {
			count := inventory.ResourceTypes[rt]

			serviceAvail := "Available"
			if missingTypeSet[strings.ToLower(rt)] {
				serviceAvail = "Not available"
			}

			implRegions := implRegionsByType[strings.ToLower(rt)]
			implRegionsStr := strings.Join(implRegions, ", ")

			skus := inventory.SKUsByType[rt]
			skuCount := len(skus)

			if skuCount == 0 {
				rows = append(rows, []string{
					rt,
					strconv.Itoa(count),
					implRegionsStr,
					"0",
					"N/A",
					"N/A",
					serviceAvail,
				})
				continue
			}

			// One row per resource type; only missing SKUs are shown in the SKU column
			skuNames := make([]string, 0, skuCount)
			for s := range skus {
				skuNames = append(skuNames, s)
			}
			sort.Strings(skuNames)

			missingSKUNames := make([]string, 0)
			for _, skuName := range skuNames {
				if missingSKUSet[strings.ToLower(rt+":"+skuName)] {
					missingSKUNames = append(missingSKUNames, skuName)
				}
			}

			allSKUsAvail := "Available"
			skuDisplay := strings.Join(skuNames, ", ")
			if config.GetPropertyMapConfig(rt) == nil {
				// No SKU availability API configured for this type — cannot determine SKU status.
				// Show "N/A" so it is clear the column was not checked, not that SKUs are fine.
				allSKUsAvail = "N/A"
			} else if len(missingSKUNames) > 0 {
				allSKUsAvail = "Not available"
				// Only show the missing SKUs so the reader knows exactly what is unavailable
				skuDisplay = strings.Join(missingSKUNames, ", ")
			}

			rows = append(rows, []string{
				rt,
				strconv.Itoa(count),
				implRegionsStr,
				strconv.Itoa(skuCount),
				skuDisplay,
				allSKUsAvail,
				serviceAvail,
			})
		}

		sheets = append(sheets, plugins.ExternalPluginOutput{
			SheetName:   safeSheetName("Svc Avail " + targetRegion),
			Description: fmt.Sprintf("Service and SKU availability for target region: %s", targetRegion),
			Table:       rows,
		})
	}

	return sheets
}

// BuildCostComparisonSheet generates a single ExternalPluginOutput with per-meter retail pricing
// across all regions, mirroring the CostComparison sheet from the AzRegionSelection PowerShell toolkit.
func BuildCostComparisonSheet(costData *types.CostComparisonData) *plugins.ExternalPluginOutput {
	if costData == nil || len(costData.MeterInputs) == 0 {
		return nil
	}

	// Collect all regions present in the pricing map
	regionSet := map[string]bool{}
	for _, regionPricing := range costData.RegionPricing {
		for region := range regionPricing {
			regionSet[region] = true
		}
	}
	if len(regionSet) == 0 {
		return nil
	}
	regions := make([]string, 0, len(regionSet))
	for r := range regionSet {
		regions = append(regions, r)
	}
	sort.Strings(regions)

	// Build a lookup: meterID → (ServiceName, ProductName) from PriceItems
	type meterMeta struct {
		serviceName string
		productName string
	}
	metaMap := map[string]meterMeta{}
	for _, item := range costData.PriceItems {
		for _, meter := range costData.MeterInputs {
			if meter.MeterName == item.MeterName &&
				meter.ProductID == item.ProductID &&
				meter.SkuName == item.SkuName {
				if _, exists := metaMap[meter.MeterID]; !exists {
					metaMap[meter.MeterID] = meterMeta{
						serviceName: item.ServiceName,
						productName: item.ProductName,
					}
				}
				break
			}
		}
	}

	// Build header: fixed columns + one RetailPrice column per region
	header := []string{"MeterId", "ServiceName", "MeterName", "ProductName", "SKUName"}
	for _, r := range regions {
		header = append(header, r+"-RetailPrice")
	}

	rows := [][]string{header}

	// One row per meter (sorted by meterID for determinism)
	sortedMeters := make([]types.MeterCostData, len(costData.MeterInputs))
	copy(sortedMeters, costData.MeterInputs)
	sort.Slice(sortedMeters, func(i, j int) bool {
		return sortedMeters[i].MeterID < sortedMeters[j].MeterID
	})

	for _, meter := range sortedMeters {
		serviceName := ""
		productName := ""
		if meta, ok := metaMap[meter.MeterID]; ok {
			serviceName = meta.serviceName
			productName = meta.productName
		}

		row := []string{meter.MeterID, serviceName, meter.MeterName, productName, meter.SkuName}
		for _, region := range regions {
			price := ""
			if pricing, ok := costData.RegionPricing[meter.MeterID]; ok {
				if p, ok := pricing[region]; ok && p > 0 {
					price = fmt.Sprintf("%.4f", p)
				}
			}
			row = append(row, price)
		}
		rows = append(rows, row)
	}

	return &plugins.ExternalPluginOutput{
		SheetName:   "CostComparison",
		Description: "Retail price comparison for resource meters across Azure regions",
		Table:       rows,
	}
}

// BuildInventorySheet converts a resource list into an Inventory ExternalPluginOutput
// using the same column layout as the main scan's inventory.csv.
// When the main scan has already written an "Inventory" sheet to the workbook,
// renderExternalPlugins will detect the duplicate and skip this sheet.
func BuildInventorySheet(resources []*models.Resource, mask bool) plugins.ExternalPluginOutput {
	headers := []string{"Subscription Id", "Resource Group", "Location", "Resource Type", "Resource Name", "Sku Name", "Sku Tier", "Capacity", "Kind", "Resource Id"}
	rows := make([][]string, 0, len(resources)+1)
	rows = append(rows, headers)

	for _, r := range resources {
		capacity := ""
		if r.SkuCapacity > 0 {
			capacity = fmt.Sprint(r.SkuCapacity)
		} else if v := skus.Lookup(r.SkuName); v > 0 {
			capacity = fmt.Sprint(v)
		}

		rows = append(rows, []string{
			renderers.MaskSubscriptionID(r.SubscriptionID, mask),
			r.ResourceGroup,
			r.Location,
			r.Type,
			r.Name,
			r.SkuName,
			r.SkuTier,
			capacity,
			r.Kind,
			renderers.MaskSubscriptionIDInResourceID(r.ID, mask),
		})
	}

	return plugins.ExternalPluginOutput{
		SheetName:   "Inventory",
		Description: "Resource inventory collected during region selection analysis",
		Table:       rows,
	}
}
