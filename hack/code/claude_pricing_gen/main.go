// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// claude_pricing_gen generates
// internal/scanners/plugins/aigov/zz_generated.claude_pricing.go from
// Anthropic's official published pricing page.
//
// This exists so the "ai-gov" plugin's Claude/Anthropic pricing never does a
// live scrape during a scan (per project decision): the page is parsed once,
// here, by a developer running this tool, and the result is committed to the
// repository as plain Go source, like hack/code/latency_gen already does for
// its own zz_generated.latency.go data file.
//
// Usage:
//
//	go run ./hack/code/claude_pricing_gen/main.go [--output ./internal/scanners/plugins/aigov/zz_generated.claude_pricing.go]
//
// The page is expected to publish one HTML table per model family with a
// header row containing "input", "output", and (optionally) cache-related
// columns, and cell values formatted as "$<amount>/MTok". This mirrors the
// parsing contract of TokenLens-for-Azure's pricing_sources/claude_docs.py,
// without the runtime fetch-and-cache machinery.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const claudePricingURL = "https://platform.claude.com/docs/en/about-claude/pricing"

var priceCellRe = regexp.MustCompile(`(?i)\$\s*([0-9]+(?:\.[0-9]+)?)\s*/\s*MTok`)

// modelPriceEntry mirrors the claudeModelPrice struct consumed by
// internal/scanners/plugins/aigov/claude_pricing.go.
type modelPriceEntry struct {
	Model            string
	Aliases          []string
	InputPerMTokUSD  float64
	OutputPerMTokUSD float64
}

func main() {
	outputPath := flag.String("output", "internal/scanners/plugins/aigov/zz_generated.claude_pricing.go", "output file path")
	sourceURL := flag.String("url", claudePricingURL, "pricing page URL")
	flag.Parse()

	body, err := fetch(*sourceURL)
	if err != nil {
		log.Fatalf("failed to fetch %s: %v", *sourceURL, err)
	}

	entries, err := parseClaudePricing(body)
	if err != nil {
		log.Fatalf("failed to parse pricing page: %v", err)
	}
	if len(entries) == 0 {
		log.Fatal("parsed zero pricing rows; the page layout may have changed (parser contract drift)")
	}

	if err := writeGoSource(*outputPath, *sourceURL, entries); err != nil {
		log.Fatalf("failed to write %s: %v", *outputPath, err)
	}

	fmt.Printf("wrote %d Claude model pricing entries to %s\n", len(entries), *outputPath)
}

func fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url) //nolint:gosec // URL is an explicit, developer-supplied flag, not user input
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// parseClaudePricing walks every HTML table on the page, and for any table
// whose header row contains recognizable "input"/"output" column labels,
// extracts one modelPriceEntry per subsequent row.
func parseClaudePricing(body []byte) ([]modelPriceEntry, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parsing html: %w", err)
	}

	var tables [][][]string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tables = append(tables, extractTableRows(n))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	var entries []modelPriceEntry
	seen := map[string]bool{}
	for _, rows := range tables {
		if len(rows) < 2 {
			continue
		}
		header := rows[0]
		inputCol := findColumn(header, "input")
		outputCol := findColumn(header, "output")
		if inputCol < 0 || outputCol < 0 {
			continue // not a pricing table we recognize
		}

		for _, row := range rows[1:] {
			if len(row) <= max(inputCol, outputCol) {
				continue
			}
			model := strings.TrimSpace(row[0])
			if model == "" || seen[model] {
				continue
			}
			inputPrice, ok1 := parsePriceCell(row[inputCol])
			outputPrice, ok2 := parsePriceCell(row[outputCol])
			if !ok1 || !ok2 {
				continue
			}
			seen[model] = true
			entries = append(entries, modelPriceEntry{
				Model:            model,
				InputPerMTokUSD:  inputPrice,
				OutputPerMTokUSD: outputPrice,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Model < entries[j].Model })
	return entries, nil
}

func extractTableRows(table *html.Node) [][]string {
	var rows [][]string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var cells []string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
					cells = append(cells, strings.Join(strings.Fields(textContent(c)), " "))
				}
			}
			if len(cells) > 0 {
				rows = append(rows, cells)
			}
			return // do not descend into nested tables
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(table)
	return rows
}

func textContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(textContent(c))
	}
	return sb.String()
}

func findColumn(header []string, label string) int {
	for i, h := range header {
		if strings.Contains(strings.ToLower(h), label) {
			return i
		}
	}
	return -1
}

func parsePriceCell(cell string) (float64, bool) {
	match := priceCellRe.FindStringSubmatch(cell)
	if match == nil {
		return 0, false
	}
	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

// writeGoSource renders entries as a Go source file defining
// claudePricingData, in the same "Code generated ... DO NOT EDIT" style as
// hack/code/latency_gen's zz_generated.latency.go, then gofmt-formats the
// result via go/format before writing it out.
func writeGoSource(path, sourceURL string, entries []modelPriceEntry) error {
	var b strings.Builder

	b.WriteString("// Code generated by hack/code/claude_pricing_gen; DO NOT EDIT.\n")
	b.WriteString("// Regenerate with: make claude-pricing\n")
	fmt.Fprintf(&b, "// Pricing source: %s\n", sourceURL)
	fmt.Fprintf(&b, "// Fetched on: %s\n", time.Now().UTC().Format("2006-01-02"))
	b.WriteString("//\n")
	b.WriteString("// Azure bills Claude-on-Foundry usage through Consumption Units (CCU) at a\n")
	b.WriteString("// fixed rate of 100 CCU = $1. The dollar rates below are the CCU-equivalent\n")
	b.WriteString("// cost per million tokens, labelled \"claude_ccu_equivalent\" (rather than\n")
	b.WriteString("// \"exact retail rate\") at resolution time.\n")
	b.WriteString("\npackage aigov\n")

	b.WriteString("\n// claudePricingData is the embedded Claude/Anthropic pricing snapshot used by\n")
	b.WriteString("// resolveClaudePrice. See claude_pricing.go for the lookup logic.\n")
	b.WriteString("var claudePricingData = []claudeModelPrice{\n")
	for _, e := range entries {
		fmt.Fprintf(&b, "\t{\n\t\tModel: %q,\n", e.Model)
		if len(e.Aliases) > 0 {
			b.WriteString("\t\tAliases: []string{")
			for i, alias := range e.Aliases {
				if i > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(&b, "%q", alias)
			}
			b.WriteString("},\n")
		}
		fmt.Fprintf(&b, "\t\tInputPerMTokUSD: %g,\n", e.InputPerMTokUSD)
		fmt.Fprintf(&b, "\t\tOutputPerMTokUSD: %g,\n", e.OutputPerMTokUSD)
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n")

	formatted, err := format.Source([]byte(b.String()))
	if err != nil {
		return fmt.Errorf("formatting generated source: %w", err)
	}

	return os.WriteFile(path, formatted, 0o644)
}
