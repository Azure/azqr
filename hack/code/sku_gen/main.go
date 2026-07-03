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
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"gopkg.in/yaml.v3"
)

type skuEntry struct {
	Name         string `yaml:"name"`
	Family       string `yaml:"family"`
	VCPUs        int    `yaml:"vcpus"`
	DiscoveredOn string `yaml:"discoveredOn"`
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
			skuMap[name] = skuEntry{
				Name:         name,
				Family:       *sku.Family,
				VCPUs:        extractVCPUs(sku),
				DiscoveredOn: discoveredOn,
			}
		}
	}

	names := make([]string, 0, len(skuMap))
	for name := range skuMap {
		names = append(names, name)
	}
	sort.Strings(names)

	entries := make([]skuEntry, 0, len(names))
	for _, name := range names {
		entries = append(entries, skuMap[name])
	}
	log.Printf("fetched %d VM SKUs", len(entries))

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

// extractVCPUs reads the vCPUsAvailable capability (preferred) or vCPUs from the SKU capabilities list.
func extractVCPUs(sku *armcompute.ResourceSKU) int {
	var vCPUs, vCPUsAvailable int
	for _, cap := range sku.Capabilities {
		if cap.Name == nil || cap.Value == nil {
			continue
		}
		switch *cap.Name {
		case "vCPUsAvailable":
			vCPUsAvailable, _ = strconv.Atoi(*cap.Value)
		case "vCPUs":
			vCPUs, _ = strconv.Atoi(*cap.Value)
		}
	}
	if vCPUsAvailable > 0 {
		return vCPUsAvailable
	}
	return vCPUs
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
