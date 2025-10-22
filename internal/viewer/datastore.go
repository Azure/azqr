// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package viewer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Dataset constants as produced by the JSON renderer.
const (
	DataSetRecommendations         = "recommendations"
	DataSetImpacted                = "impacted"
	DataSetResourceType            = "resourceType"
	DataSetInventory               = "inventory"
	DataSetAdvisor                 = "advisor"
	DataSetAzurePolicy             = "azurePolicy"
	DataSetArcSQL                  = "arcSQL"
	DataSetDefender                = "defender"
	DataSetDefenderRecommendations = "defenderRecommendations"
	DataSetCosts                   = "costs"
	DataSetOutOfScope              = "outOfScope"
)

// DataStore holds all report datasets in memory.
type DataStore struct {
	Data map[string][]map[string]string
}

// LoadDataStore loads a consolidated azqr JSON or Excel report.
func LoadDataStore(path string) (*DataStore, error) {
	// Determine file type by extension
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".xlsx", ".xls":
		return ExcelToDataStore(path)
	case ".json":
		return loadJSONDataStore(path)
	default:
		// Try Excel first (default azqr output format)
		if ds, err := ExcelToDataStore(path); err == nil {
			return ds, nil
		}
		// If Excel fails, try JSON for backward compatibility
		if ds, err := loadJSONDataStore(path); err == nil {
			return ds, nil
		}
		return nil, fmt.Errorf("unsupported file format: %s (supported: .xlsx, .xls, .json)", ext)
	}
}

// loadJSONDataStore loads a JSON format report (original implementation).
func loadJSONDataStore(path string) (*DataStore, error) {
	// Clean and normalize the path for cross-platform compatibility
	cleanPath := filepath.Clean(path)
	f, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("open json file: %w", err)
	}
	defer func() { _ = f.Close() }()

	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read json file: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}

	ds := &DataStore{Data: map[string][]map[string]string{}}
	for k, v := range raw {
		arr, ok := v.([]interface{})
		if !ok {
			continue
		}
		records := make([]map[string]string, 0, len(arr))
		for _, row := range arr {
			obj, ok := row.(map[string]interface{})
			if !ok {
				continue
			}
			rec := map[string]string{}
			for key, val := range obj {
				rec[key] = toString(val)
			}
			records = append(records, rec)
		}
		ds.Data[k] = records
	}
	return ds, nil
}

// Get dataset by name.
func (ds *DataStore) Get(name string) []map[string]string { return ds.Data[name] }

// ListDataSets returns sorted dataset names.
func (ds *DataStore) ListDataSets() []string {
	out := make([]string, 0, len(ds.Data))
	for k := range ds.Data {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Filter dataset with query parameters.
func (ds *DataStore) Filter(name string, params map[string][]string) ([]map[string]string, error) {
	records := ds.Get(name)
	if records == nil {
		// Return empty array instead of error for datasets that don't exist
		// This handles cases where optional datasets (like arcSQL) have no data
		return []map[string]string{}, nil
	}

	global := strings.ToLower(first(params, "q"))
	limitStr := first(params, "limit")
	limit := -1
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l >= 0 {
			limit = l
		}
	}

	fieldFilters := map[string][]string{}
	for k, vs := range params {
		if k == "q" || k == "limit" {
			continue
		}
		filtered := make([]string, 0, len(vs))
		for _, v := range vs {
			v = strings.TrimSpace(v)
			if v != "" {
				filtered = append(filtered, strings.ToLower(v))
			}
		}
		if len(filtered) > 0 {
			fieldFilters[k] = filtered
		}
	}

	out := make([]map[string]string, 0, len(records))
	for _, r := range records {
		if global != "" && !matchesGlobal(r, global) {
			continue
		}
		if !matchesFields(r, fieldFilters) {
			continue
		}
		out = append(out, r)
		if limit >= 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// Summary metrics for dashboard.
func (ds *DataStore) Summary() map[string]interface{} {
	recs := ds.Get(DataSetRecommendations)
	impacted := ds.Get(DataSetImpacted)
	resourceTypes := ds.Get(DataSetResourceType)
	inventory := ds.Get(DataSetInventory)
	advisor := ds.Get(DataSetAdvisor)
	policy := ds.Get(DataSetAzurePolicy)
	arcSQL := ds.Get(DataSetArcSQL)
	defender := ds.Get(DataSetDefender)
	defRec := ds.Get(DataSetDefenderRecommendations)
	costs := ds.Get(DataSetCosts)
	outOfScope := ds.Get(DataSetOutOfScope)

	implemented, notImplemented := 0, 0
	for _, r := range recs {
		switch strings.ToLower(r["implemented"]) {
		case "true":
			implemented++
		case "false":
			notImplemented++
		}
	}

	totalCost := 0.0
	for _, c := range costs {
		if v, err := strconv.ParseFloat(c["value"], 64); err == nil {
			totalCost += v
		}
	}

	nonCompliantPolicy := 0
	for _, p := range policy {
		if !strings.EqualFold(p["complianceState"], "Compliant") && p["complianceState"] != "" {
			nonCompliantPolicy++
		}
	}

	return map[string]interface{}{
		"recommendationsTotal":          len(recs),
		"recommendationsImplemented":    implemented,
		"recommendationsNotImplemented": notImplemented,
		"impactedCount":                 len(impacted),
		"resourceTypeCount":             len(resourceTypes),
		"inventoryCount":                len(inventory),
		"advisorCount":                  len(advisor),
		"azurePolicyCount":              len(policy),
		"azurePolicyNonCompliant":       nonCompliantPolicy,
		"arcSQLCount":                   len(arcSQL),
		"defenderCount":                 len(defender),
		"defenderRecommendationsCount":  len(defRec),
		"costItems":                     len(costs),
		"totalCost":                     totalCost,
		"outOfScopeCount":               len(outOfScope),
	}
}

// Analytics returns extended CTO-level metrics for richer dashboards.
// Structure groups: implementation, impact, categories, resourceTypes, policy, defender, cost, hotspot, sla.
func (ds *DataStore) Analytics() map[string]interface{} {
	recs := ds.Get(DataSetRecommendations)
	impacted := ds.Get(DataSetImpacted)
	policy := ds.Get(DataSetAzurePolicy)
	defenderRec := ds.Get(DataSetDefenderRecommendations)
	inventory := ds.Get(DataSetInventory)
	resourceTypes := ds.Get(DataSetResourceType)
	costs := ds.Get(DataSetCosts)

	// Implementation metrics
	implemented, notImplemented := 0, 0
	highImpactNotImplemented := 0
	for _, r := range recs {
		impl := strings.ToLower(r["implemented"])
		impact := strings.ToLower(r["impact"])
		switch impl {
		case "true":
			implemented++
		case "false":
			notImplemented++
			if impact == "high" {
				highImpactNotImplemented++
			}
		}
	}
	// Deployed recommendations: those that are relevant (implemented true/false) i.e. exclude N/A
	deployed := implemented + notImplemented
	implRate := pct(implemented, deployed)

	// Impact distribution across impacted resources
	impactCounts := map[string]int{"high": 0, "medium": 0, "low": 0}
	for _, ir := range impacted {
		ic := strings.ToLower(ir["impact"])
		if _, ok := impactCounts[ic]; ok {
			impactCounts[ic]++
		}
	}
	highPct := pct(impactCounts["high"], len(impacted))

	// Category metrics
	categoryStats := map[string]struct {
		total  int
		high   int
		medium int
		low    int
	}{}
	for _, ir := range impacted {
		cat := ir["category"]
		s := categoryStats[cat]
		s.total++
		impact := strings.ToLower(ir["impact"])
		switch impact {
		case "high":
			s.high++
		case "medium":
			s.medium++
		case "low":
			s.low++
		}
		categoryStats[cat] = s
	}
	categoryList := make([]map[string]interface{}, 0, len(categoryStats))
	for cat, v := range categoryStats {
		categoryList = append(categoryList, map[string]interface{}{
			"category":      cat,
			"impactedTotal": v.total,
			"highImpact":    v.high,
			"mediumImpact":  v.medium,
			"lowImpact":     v.low,
		})
	}
	sort.Slice(categoryList, func(i, j int) bool {
		return categoryList[i]["impactedTotal"].(int) > categoryList[j]["impactedTotal"].(int)
	})
	topCategories := firstN(categoryList, 5)

	// Resource type impact (from impacted dataset resourceType field)
	rtImpact := map[string]int{}
	for _, ir := range impacted {
		rtImpact[ir["resourceType"]]++
	}
	rtImpactList := make([]map[string]interface{}, 0, len(rtImpact))
	for rt, c := range rtImpact {
		rtImpactList = append(rtImpactList, map[string]interface{}{"resourceType": rt, "impactedCount": c})
	}
	sort.Slice(rtImpactList, func(i, j int) bool {
		return rtImpactList[i]["impactedCount"].(int) > rtImpactList[j]["impactedCount"].(int)
	})
	topRtImpact := firstN(rtImpactList, 10)

	// Deployed resource types with zero impact (potential coverage gaps)
	impactedSet := map[string]bool{}
	for rt := range rtImpact {
		impactedSet[strings.ToLower(rt)] = true
	}
	deployedNoImpact := []string{}
	for _, rt := range resourceTypes {
		rtt := strings.ToLower(rt["resourceType"])
		if rtt == "" {
			continue
		}
		if !impactedSet[rtt] {
			deployedNoImpact = append(deployedNoImpact, rt["resourceType"])
		}
	}
	sort.Strings(deployedNoImpact)
	if len(deployedNoImpact) > 20 {
		deployedNoImpact = deployedNoImpact[:20]
	}

	// Policy compliance
	nonCompliant := 0
	for _, p := range policy {
		if !strings.EqualFold(p["complianceState"], "Compliant") && p["complianceState"] != "" {
			nonCompliant++
		}
	}
	policyRate := pct(nonCompliant, len(policy))

	// Defender severity breakdown
	severity := map[string]int{"high": 0, "medium": 0, "low": 0, "unknown": 0}
	for _, d := range defenderRec {
		sev := strings.ToLower(d["recommendationSeverity"])
		if _, ok := severity[sev]; !ok {
			sev = "unknown"
		}
		severity[sev]++
	}

	// Costs
	totalCost := 0.0
	for _, c := range costs {
		if v, err := strconv.ParseFloat(c["value"], 64); err == nil {
			totalCost += v
		}
	}
	costPerImpacted := 0.0
	if len(impacted) > 0 {
		costPerImpacted = totalCost / float64(len(impacted))
	}

	// Hotspot score
	hotspot := impactCounts["high"]*3 + impactCounts["medium"]*2 + impactCounts["low"]

	// SLA coverage (from inventory 'sla' field if present)
	withSLA := 0
	for _, inv := range inventory {
		if strings.TrimSpace(inv["sla"]) != "" {
			withSLA++
		}
	}
	slaPct := pct(withSLA, len(inventory))

	return map[string]interface{}{
		"implementation": map[string]interface{}{
			"implemented":              implemented,
			"notImplemented":           notImplemented,
			"deployedRecommendations":  deployed,
			"implementationRate":       implRate,
			"highImpactNotImplemented": highImpactNotImplemented,
		},
		"impact": map[string]interface{}{
			"distribution":      impactCounts,
			"highImpactPercent": highPct,
		},
		"categories": map[string]interface{}{
			"top": topCategories,
			"all": categoryList,
		},
		"resourceTypes": map[string]interface{}{
			"topImpacted":      topRtImpact,
			"deployedNoImpact": deployedNoImpact,
		},
		"policy": map[string]interface{}{
			"total":             len(policy),
			"nonCompliant":      nonCompliant,
			"nonComplianceRate": policyRate,
		},
		"defender": map[string]interface{}{
			"recommendations": len(defenderRec),
			"severity":        severity,
		},
		"cost": map[string]interface{}{
			"totalCost":       totalCost,
			"costPerImpacted": costPerImpacted,
		},
		"hotspot": map[string]interface{}{
			"score":   hotspot,
			"formula": "(high*3 + medium*2 + low)",
		},
		"sla": map[string]interface{}{
			"resourcesWithSLA": withSLA,
			"totalResources":   len(inventory),
			"coveragePercent":  slaPct,
		},
	}
}

func pct(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(part) / float64(total)) * 100
}
func firstN(list []map[string]interface{}, n int) []map[string]interface{} {
	if len(list) <= n {
		return list
	}
	return list[:n]
}

func matchesGlobal(rec map[string]string, global string) bool {
	g := strings.ToLower(global)
	for _, v := range rec {
		if strings.Contains(strings.ToLower(v), g) {
			return true
		}
	}
	return false
}
func matchesFields(rec map[string]string, filters map[string][]string) bool {
	for field, values := range filters {
		rv, ok := rec[field]
		if !ok {
			return false
		}
		rvLower := strings.ToLower(rv)
		matched := false
		for _, fv := range values {
			if strings.Contains(rvLower, fv) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}
func first(m map[string][]string, key string) string {
	if vs, ok := m[key]; ok && len(vs) > 0 {
		return vs[0]
	}
	return ""
}
func toString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%f", t)
	case bool:
		return fmt.Sprintf("%t", t)
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}
