// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// latency_gen fetches Azure network round-trip latency statistics and Azure region data
// from Microsoft Learn and writes a single generated Go source file used by the region-
// selection plugin.
//
// The generated file (zz_generated.latency.go) contains two package-level variables:
//   - azureRegionLatency: P50 RTT matrix (ms) between Azure regions
//   - regionCluster: mapping of each region's programmatic name to its geographic cluster
//
// Sources:
//   - Latency matrix: https://learn.microsoft.com/en-us/azure/networking/azure-network-latency
//     (full CSV block embedded at the bottom of the page)
//   - Region clusters: https://learn.microsoft.com/en-us/azure/reliability/regions-list
//     (programmatic names extracted per-tab: americas / europe / asia-pacific / middle-east / africa)
//
// Usage:
//
//	go run ./hack/code/latency_gen/main.go
//	go run ./hack/code/latency_gen/main.go --output ./internal/scanners/plugins/region/zz_generated.latency.go
package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultURL        = "https://learn.microsoft.com/en-us/azure/networking/azure-network-latency"
	defaultRegionsURL = "https://learn.microsoft.com/en-us/azure/reliability/regions-list?tabs=all"
	defaultOutput     = "internal/scanners/plugins/region/latency/zz_generated.latency.go"

	// csvHeaderPrefix is the start of the full combined CSV block present at the bottom of the page.
	// The tab-specific tables only show a subset of regions, so we look for the header that starts
	// with the first region in alphabetical order to identify the complete matrix.
	csvHeaderPrefix = "Source,Australia Central,"
)

// htmlTag matches any HTML tag so it can be stripped from raw HTML lines.
var htmlTag = regexp.MustCompile(`<[^>]+>`)

func main() {
	outputPath := flag.String("output", defaultOutput, "path to write the generated zz_generated.latency.go")
	sourceURL := flag.String("url", defaultURL, "URL of the Azure network latency statistics page")
	regionsURL := flag.String("regions-url", defaultRegionsURL, "URL of the Azure regions list page")
	flag.Parse()

	// Fetch and parse the latency matrix.
	log.Printf("fetching latency data from %s", *sourceURL)
	body, err := fetchPage(*sourceURL)
	if err != nil {
		log.Fatalf("fetching page: %v", err)
	}
	matrix, err := extractMatrix(body)
	if err != nil {
		log.Fatalf("extracting matrix: %v", err)
	}
	log.Printf("parsed %d source regions", len(matrix))

	// Fetch and parse the region→cluster assignments.
	log.Printf("fetching region list from %s", *regionsURL)
	regionsBody, err := fetchPage(*regionsURL)
	if err != nil {
		log.Fatalf("fetching regions page: %v", err)
	}
	clusters, err := extractRegionClusters(regionsBody)
	if err != nil {
		log.Fatalf("extracting region clusters: %v", err)
	}
	log.Printf("extracted %d region→cluster mappings", len(clusters))

	// Generate and write the combined Go source file.
	src := generateLatencyFile(matrix, clusters, *sourceURL, *regionsURL)
	if err := os.WriteFile(*outputPath, src, 0644); err != nil {
		log.Fatalf("writing %s: %v", *outputPath, err)
	}
	log.Printf("written to %s", *outputPath)
}

// fetchPage performs an HTTP GET and returns the response body.
func fetchPage(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "azqr-latency-gen/1.0 (github.com/Azure/azqr)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close response body: %v\n", cerr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d for %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// extractMatrix scans the raw HTML body for the full CSV block and builds the latency matrix.
//
// The Microsoft Learn page embeds a combined CSV code block at the bottom of the article
// that contains all region pairs. We find it by looking for a line that starts with
// csvHeaderPrefix after stripping HTML tags.
func extractMatrix(body []byte) (map[string]map[string]float64, error) {
	// Use a 1 MB per-line buffer; some HTML pages inline large scripts.
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var csvLines []string
	inCSV := false

	for scanner.Scan() {
		line := strings.TrimSpace(htmlTag.ReplaceAllString(scanner.Text(), ""))

		if !inCSV {
			if strings.HasPrefix(line, csvHeaderPrefix) {
				inCSV = true
				csvLines = append(csvLines, line)
			}
			continue
		}

		// Stop collecting on blank lines or structural markers that follow the CSV block.
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "---") {
			break
		}
		csvLines = append(csvLines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning body: %w", err)
	}
	if len(csvLines) < 2 {
		return nil, fmt.Errorf(
			"CSV block not found in page — expected a line starting with %q (found %d lines total); "+
				"the page structure may have changed",
			csvHeaderPrefix, len(csvLines),
		)
	}
	log.Printf("found CSV block: %d rows (1 header + %d data rows)", len(csvLines), len(csvLines)-1)

	r := csv.NewReader(strings.NewReader(strings.Join(csvLines, "\n")))
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing CSV: %w", err)
	}
	return buildMatrix(records)
}

// buildMatrix converts raw CSV records into a normalized latency map.
//
// The first record is the header row: ["Source", "Australia Central", ...].
// Subsequent records are data rows: ["Australia Central", "", "3", ...].
//
// Empty cells and cells with value ≤ 0 (used when no measurement is available,
// e.g. Qatar Central) are omitted so callers receive only confirmed latency values.
func buildMatrix(records [][]string) (map[string]map[string]float64, error) {
	if len(records) < 2 {
		return nil, fmt.Errorf("need at least 2 rows, got %d", len(records))
	}

	// Build the ordered list of target region identifiers from the header row.
	header := records[0]
	targets := make([]string, len(header)-1)
	for i, col := range header[1:] {
		targets[i] = normalizeRegion(col)
	}

	matrix := make(map[string]map[string]float64)
	skipped := 0
	for _, row := range records[1:] {
		if len(row) == 0 || strings.TrimSpace(row[0]) == "" {
			continue
		}
		src := normalizeRegion(row[0])
		for j, cell := range row[1:] {
			if j >= len(targets) {
				break
			}
			tgt := targets[j]
			if tgt == "" || src == tgt {
				continue
			}
			cell = strings.TrimSpace(cell)
			if cell == "" {
				skipped++
				continue
			}
			ms, err := strconv.ParseFloat(cell, 64)
			if err != nil || ms <= 0 {
				skipped++
				continue
			}
			if matrix[src] == nil {
				matrix[src] = make(map[string]float64)
			}
			matrix[src][tgt] = ms
		}
	}
	log.Printf("skipped %d empty/zero cells", skipped)
	return matrix, nil
}

// normalizeRegion converts a display name like "Australia Central" to "australiacentral".
// This mirrors the normalizeRegionName function in the region plugin (latency.go).
func normalizeRegion(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", ""))
}

// generateLatencyFile produces the combined Go source file containing both
// the azureRegionLatency matrix and the regionCluster map.
func generateLatencyFile(matrix map[string]map[string]float64, clusters map[string]string, latencyURL, regionsURL string) []byte {
	var b strings.Builder

	b.WriteString("// Code generated by hack/code/latency_gen; DO NOT EDIT.\n")
	b.WriteString("// Regenerate with: make latency\n")
	fmt.Fprintf(&b, "// Latency source: %s\n", latencyURL)
	fmt.Fprintf(&b, "// Regions source: %s\n", regionsURL)
	b.WriteString("\npackage latency\n")

	// --- azureRegionLatency ---
	b.WriteString("\n// azureRegionLatency contains P50 (median) round-trip time measurements in milliseconds\n")
	b.WriteString("// between Azure regions. Source: Azure Network Round-trip Latency Statistics (Microsoft).\n")
	b.WriteString("var azureRegionLatency = map[string]map[string]float64{\n")

	sources := make([]string, 0, len(matrix))
	for src := range matrix {
		sources = append(sources, src)
	}
	sort.Strings(sources)

	for _, src := range sources {
		inner := matrix[src]
		tgts := make([]string, 0, len(inner))
		for tgt := range inner {
			tgts = append(tgts, tgt)
		}
		sort.Strings(tgts)
		fmt.Fprintf(&b, "\t%q: {\n", src)
		for _, tgt := range tgts {
			fmt.Fprintf(&b, "\t\t%q: %g,\n", tgt, inner[tgt])
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n")

	// --- regionCluster ---
	clusterOrder := []string{"americas", "europe", "apac", "mea"}
	clusterComment := map[string]string{
		"americas": "Americas",
		"europe":   "Europe",
		"apac":     "Asia Pacific",
		"mea":      "Middle East & Africa",
	}
	byCluster := make(map[string][]string, len(clusterOrder))
	for region, cluster := range clusters {
		byCluster[cluster] = append(byCluster[cluster], region)
	}
	for _, regions := range byCluster {
		sort.Strings(regions)
	}

	b.WriteString("\n// regionCluster maps normalized Azure region identifiers to their geographic cluster.\n")
	b.WriteString("// Clusters align with Microsoft's own tab grouping on the regions list page.\n")
	b.WriteString("// Used to estimate latency for region pairs not in the measured matrix.\n")
	b.WriteString("var regionCluster = map[string]string{\n")
	for _, cluster := range clusterOrder {
		regions, ok := byCluster[cluster]
		if !ok || len(regions) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\t// %s\n", clusterComment[cluster])
		for _, region := range regions {
			fmt.Fprintf(&b, "\t%q: %q,\n", region, cluster)
		}
	}
	b.WriteString("}\n")

	return []byte(b.String())
}

// tabToCluster maps the data-tab attribute value from the Azure regions-list page
// to the four cluster identifiers used by the scoring algorithm.
// Microsoft Learn renders each geography as a separate tab section, e.g.
//
//	<section id="tabpanel_1_americas" role="tabpanel" data-tab="americas">
//
// Middle East and Africa share the "mea" cluster.
var tabToCluster = map[string]string{
	"americas":     "americas",
	"europe":       "europe",
	"asia-pacific": "apac",
	"middle-east":  "mea",
	"africa":       "mea",
}

// trRe and tdRe extract table row / cell content from raw HTML.
var trRe = regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
var tdRe = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)

// progNameRe matches a valid Azure programmatic region name (lowercase letters and digits, ≥3 chars).
var progNameRe = regexp.MustCompile(`^[a-z][a-z0-9]{2,}$`)

// extractRegionClusters parses each tab section from the Azure regions-list page and
// returns a map of normalized region identifier → cluster name.
//
// Each tab (<section data-tab="americas">, <section data-tab="europe">, etc.) lists the
// public regions in that geography. The programmatic name is always the last <td> in each row.
// No static geography-to-cluster mapping is needed: the tab ID is the cluster (with minor
// normalization for asia-pacific → apac and middle-east/africa → mea).
func extractRegionClusters(body []byte) (map[string]string, error) {
	clusters := make(map[string]string)

	for tabID, cluster := range tabToCluster {
		sectionRe := regexp.MustCompile(`(?s)<section[^>]+data-tab="` + tabID + `"[^>]*>(.*?)</section>`)
		m := sectionRe.FindSubmatch(body)
		if m == nil {
			log.Printf("warning: tab %q not found — skipping", tabID)
			continue
		}

		for _, trMatch := range trRe.FindAllSubmatch(m[1], -1) {
			tdMatches := tdRe.FindAllSubmatch(trMatch[1], -1)
			if len(tdMatches) == 0 {
				continue
			}
			// Programmatic name is always the last <td> in each data row.
			lastCell := tdMatches[len(tdMatches)-1][1]
			progName := strings.ToLower(strings.Join(strings.Fields(
				htmlTag.ReplaceAllString(string(lastCell), ""),
			), ""))

			if !progNameRe.MatchString(progName) {
				continue
			}
			clusters[progName] = cluster
		}
	}

	if len(clusters) == 0 {
		return nil, fmt.Errorf("no region clusters extracted — page structure may have changed")
	}
	return clusters, nil
}
