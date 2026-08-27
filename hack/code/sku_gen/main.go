// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// sku_gen generates data/known_skus.yaml with Azure VM SKU names, families, and vCPU counts.
//
// Subscription resolution order:
//  1. AZURE_SUBSCRIPTION_ID environment variable
//  2. Active Azure CLI account (az account show)
//  3. Panic with a descriptive message if neither is available
//
// Usage:
//
//	go run ./hack/code/sku_gen/main.go [--output ./internal/skus/known_skus.yaml]
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"gopkg.in/yaml.v3"
)

type skuEntry struct {
	Name                  string  `yaml:"name"`
	Family                string  `yaml:"family"`
	VCPUs                 int     `yaml:"vcpus"`
	MemoryGB              float64 `yaml:"memoryGb"`
	GPUCount              float64 `yaml:"gpuCount"`
	MaxDataDisks          int     `yaml:"maxDataDisks"`
	AcceleratedNetworking bool    `yaml:"acceleratedNetworking"`
	DiscoveredOn          string  `yaml:"discoveredOn"`
}

func main() {
	outputPath := flag.String("output", "internal/skus/known_skus.yaml", "output file path")
	flag.Parse()

	sub := resolveSubscription()
	log.Printf("using subscription: %s", sub)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to create credential: %v", err)
	}

	client, err := armcompute.NewResourceSKUsClient(sub, cred, nil)
	if err != nil {
		log.Fatalf("failed to create ResourceSKUs client: %v", err)
	}

	existing := readExisting(*outputPath)
	today := time.Now().UTC().Format("2006-01-02")

	log.Println("fetching VM SKUs from Azure...")

	skuMap := make(map[string]skuEntry)
	filter := "resourceType eq 'virtualMachines'"
	pager := client.NewListPager(&armcompute.ResourceSKUsClientListOptions{
		Filter: &filter,
	})

	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to list SKUs: %v", err)
		}
		for _, sku := range page.Value {
			if sku.Name == nil || sku.Family == nil {
				continue
			}
			name := *sku.Name
			if _, exists := skuMap[name]; exists {
				continue
			}
			discoveredOn := today
			if prev, ok := existing[name]; ok && prev.DiscoveredOn != "" {
				discoveredOn = prev.DiscoveredOn
			}
			caps := extractCapabilities(sku)
			caps.Name = name
			caps.Family = *sku.Family
			caps.DiscoveredOn = discoveredOn
			skuMap[name] = caps
		}
	}

	names := make([]string, 0, len(skuMap))
	for name := range skuMap {
		names = append(names, name)
	}
	sort.Strings(names)
	log.Printf("fetched %d VM SKUs", len(names))

	// Second pass: enrich GPU counts for N-series families the API does not populate.
	enrichMissingGPUCounts(skuMap)

	entries := make([]skuEntry, 0, len(names))
	for _, name := range names {
		entries = append(entries, skuMap[name])
	}

	out, err := yaml.Marshal(entries)
	if err != nil {
		log.Fatalf("failed to marshal YAML: %v", err)
	}

	header := []byte("# This file is auto-generated. DO NOT EDIT.\n")
	if err := os.WriteFile(*outputPath, append(header, out...), 0644); err != nil {
		log.Fatalf("failed to write %s: %v", *outputPath, err)
	}
	log.Printf("written to %s", *outputPath)
}

// resolveSubscription returns the Azure subscription ID to use.
// It checks AZURE_SUBSCRIPTION_ID first, then falls back to the active Azure CLI account.
func resolveSubscription() string {
	if sub := strings.TrimSpace(os.Getenv("AZURE_SUBSCRIPTION_ID")); sub != "" {
		return sub
	}
	out, err := exec.Command("az", "account", "show", "--query", "id", "--output", "tsv").Output()
	if err == nil {
		if sub := strings.TrimSpace(string(out)); sub != "" {
			return sub
		}
	}
	panic("no Azure subscription found: set AZURE_SUBSCRIPTION_ID or log in with 'az login'")
}

// gpuFamilyRe matches N-series GPU VM SKU names (NV=visualization, NC=compute, ND=distributed).
var gpuFamilyRe = regexp.MustCompile(`(?i)^Standard_N[VCD]\d`)

// familyVersionRe matches the trailing version suffix in a family page name (e.g. "v5").
var familyVersionRe = regexp.MustCompile(`(v\d+)$`)

// allUpperBodyRe matches family body strings that contain only uppercase letters and digits
// (i.e. no lowercase letters). Used to detect families with the ISR-style naming convention.
var allUpperBodyRe = regexp.MustCompile(`^[A-Z0-9]+$`)

// gpuFamilyBodyRe decomposes all-uppercase N-series family bodies into their components.
// Groups: 1=vm-series (ND/NV/NC), 2=attribute block (ISR etc., discarded),
// 3=GPU model (GB300/A100/H100/T4), 4=version (V6/V5), 5=trailing suffix (NDR/empty).
// e.g. "NDISRGB300V6" → ["ND", "ISR", "GB300", "V6", ""]
// e.g. "NDISRGB200V6NDR" → ["ND", "ISR", "GB200", "V6", "NDR"]
var gpuFamilyBodyRe = regexp.MustCompile(`^(N[VCD])([A-Z]*?)([A-Z]{1,2}\d+)(V\d+)([A-Z]*)$`)

// tableRowRe matches a Markdown table row whose first cell is a Standard_N VM SKU name
// and whose second cell is a GPU accelerator quantity (fraction or integer).
// e.g. "| Standard_NV4ads_V710_v5 | 1/6 | 4 |"
var tableRowRe = regexp.MustCompile(`\|\s*(Standard_N\w+)\s*\|\s*([0-9/]+)\s*\|`)

// fracRe matches fraction strings such as "1/6".
var fracRe = regexp.MustCompile(`^(\d+)/(\d+)$`)

// familyToDocPage derives the Azure docs page name from a VM family name.
//
// Two naming conventions exist:
//   - CamelCase: "StandardNVadsV710v5Family" → "nvadsv710-v5-series"
//   - All-uppercase body: "standardNDISRGB300V6Family" → "nd-gb300-v6-series"
//     (Azure omits the ISR-style attribute block from the docs URL)
func familyToDocPage(family string) string {
	// Strip "Standard"/"standard" prefix and "Family"/"family" suffix.
	s := family
	for _, p := range []string{"Standard", "standard"} {
		if strings.HasPrefix(s, p) {
			s = s[len(p):]
			break
		}
	}
	for _, p := range []string{"Family", "family"} {
		if strings.HasSuffix(s, p) {
			s = s[:len(s)-len(p)]
			break
		}
	}

	// All-uppercase body: extract vm-series, GPU model, version, and optional trailing suffix.
	// e.g. "NDISRGB300V6" → nd-gb300-v6, "NDISRGB200V6NDR" → nd-gb200-v6-ndr
	if allUpperBodyRe.MatchString(s) {
		if m := gpuFamilyBodyRe.FindStringSubmatch(s); m != nil {
			// Azure drops both the middle attribute block (m[2]: ISR) and the trailing
			// hardware suffix (m[5]: NDR) from docs URLs — only vm-series, GPU model,
			// and version are retained.
			// e.g. NDISRGB300V6 → nd-gb300-v6, NDISRGB200V6NDR → nd-gb200-v6
			parts := []string{strings.ToLower(m[1]), strings.ToLower(m[3]), strings.ToLower(m[4])}
			return strings.Join(parts, "-") + "-series"
		}
	}

	// CamelCase fallback: insert a dash before the trailing version suffix.
	// e.g. "NVadsV710v5" → "nvadsv710-v5-series"
	s = familyVersionRe.ReplaceAllString(s, "-$1")
	return strings.ToLower(s) + "-series"
}

// parseFraction converts "1/6" → ~0.1667 and "1" → 1.0.
func parseFraction(s string) float64 {
	s = strings.TrimSpace(s)
	if m := fracRe.FindStringSubmatch(s); m != nil {
		num, _ := strconv.ParseFloat(m[1], 64)
		den, _ := strconv.ParseFloat(m[2], 64)
		if den > 0 {
			return num / den
		}
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// fetchAcceleratorsFromDocs fetches the Accelerators table from the public Azure docs
// GitHub repository for the given page name and returns a map of SKU name → GPU count.
func fetchAcceleratorsFromDocs(page string) (map[string]float64, error) {
	url := fmt.Sprintf(
		"https://raw.githubusercontent.com/MicrosoftDocs/azure-compute-docs/main/articles/virtual-machines/sizes/gpu-accelerated/%s.md",
		page,
	)
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Markdown escapes underscores in table cells (Standard\_NV4ads\_V710\_v5); normalise.
	content := strings.ReplaceAll(string(body), `\_`, "_")

	// The page has multiple tables (vCPUs, memory, network, accelerators).
	// Narrow to the Accelerators tab section to avoid picking up vCPU counts.
	if idx := strings.Index(content, "tab/sizeaccelerators"); idx != -1 {
		content = content[idx:]
	} else if idx := strings.Index(content, "Accelerators (Qty.)"); idx != -1 {
		content = content[idx:]
	}

	result := make(map[string]float64)
	for _, m := range tableRowRe.FindAllStringSubmatch(content, -1) {
		count := parseFraction(m[2])
		if count > 0 {
			result[strings.TrimSpace(m[1])] = count
		}
	}
	return result, nil
}

// enrichMissingGPUCounts performs a second pass over the SKU map. For every N-series
// GPU family where all SKUs still report gpuCount=0 (meaning the Azure ResourceSKUs API
// omits GPU metadata entirely), it fetches the Accelerators table from the Azure docs
// and applies the correct fractional GPU counts.
func enrichMissingGPUCounts(skuMap map[string]skuEntry) {
	// Count total and zero-GPU SKUs per N-series family.
	total := make(map[string]int)
	zeros := make(map[string]int)
	for _, e := range skuMap {
		if gpuFamilyRe.MatchString(e.Name) {
			total[e.Family]++
			if e.GPUCount == 0 {
				zeros[e.Family]++
			}
		}
	}
	for family, t := range total {
		if zeros[family] < t {
			continue // API already returned GPU data for this family
		}
		page := familyToDocPage(family)
		log.Printf("enriching GPU counts for %s from docs (%s)", family, page)
		counts, err := fetchAcceleratorsFromDocs(page)
		if err != nil {
			log.Printf("warning: could not enrich GPU counts for %s: %v", family, err)
			continue
		}
		for name, count := range counts {
			if e, ok := skuMap[name]; ok && e.GPUCount == 0 {
				e.GPUCount = count
				skuMap[name] = e
				log.Printf("  %s gpuCount=%.4f", name, count)
			}
		}
	}
}

// extractCapabilities reads relevant capabilities from a SKU into a skuEntry.
func extractCapabilities(sku *armcompute.ResourceSKU) skuEntry {
	var e skuEntry
	var vCPUs, vCPUsAvailable int
	var gpus float64
	for _, cap := range sku.Capabilities {
		if cap.Name == nil || cap.Value == nil {
			continue
		}
		switch *cap.Name {
		case "vCPUsAvailable":
			vCPUsAvailable, _ = strconv.Atoi(*cap.Value)
		case "vCPUs":
			vCPUs, _ = strconv.Atoi(*cap.Value)
		case "MemoryGB":
			e.MemoryGB, _ = strconv.ParseFloat(*cap.Value, 64)
		case "GPUs":
			gpus, _ = strconv.ParseFloat(*cap.Value, 64)
		case "MaxDataDiskCount":
			e.MaxDataDisks, _ = strconv.Atoi(*cap.Value)
		case "AcceleratedNetworkingEnabled":
			e.AcceleratedNetworking = strings.EqualFold(*cap.Value, "true")
		}
	}
	if vCPUsAvailable > 0 {
		e.VCPUs = vCPUsAvailable
	} else {
		e.VCPUs = vCPUs
	}
	e.GPUCount = gpus
	return e
}

// readExisting loads the existing SKU file so that discoveredOn dates are preserved on re-generation.
func readExisting(path string) map[string]skuEntry {
	existing := make(map[string]skuEntry)
	data, err := os.ReadFile(path)
	if err != nil {
		return existing
	}
	var entries []skuEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return existing
	}
	for _, e := range entries {
		existing[e.Name] = e
	}
	return existing
}
