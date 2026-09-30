// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// ptu_capacity_gen generates
// internal/scanners/plugins/aigov/ptu/zz_generated.ptu_capacity.go from the
// Microsoft Learn PTU sizing page.
//
// The page is parsed once, here, and the result is committed as plain Go
// source so scans never fetch it. Models with a numeric output-to-input ratio
// are emitted as documented; normalized-token models (GPT-6 family) use the
// short-context output weight as an approximate ratio (see
// normalizedTokenModels). Image models are skipped.
//
// Usage:
//
//	go run ./hack/code/ptu_capacity_gen/main.go [--output <path>] [--url <url>]
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
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const sizingURL = "https://learn.microsoft.com/en-us/azure/foundry/openai/how-to/provisioned-throughput-sizing"

type capacity struct {
	InputTPMPerPTU  int
	OutputRatio     int
	GlobalMin       int
	GlobalIncrement int
	RegionalMin     int
	RegionalIncrem  int
}

type table struct {
	header []string
	rows   [][]string
}

func main() {
	output := flag.String("output", "internal/scanners/plugins/aigov/ptu/zz_generated.ptu_capacity.go", "output file path")
	url := flag.String("url", sizingURL, "PTU sizing page URL")
	flag.Parse()

	body, err := fetch(*url)
	if err != nil {
		log.Fatalf("failed to fetch %s: %v", *url, err)
	}
	models, err := parseCapacity(body)
	if err != nil {
		log.Fatalf("failed to parse page: %v", err)
	}
	if len(models) == 0 {
		log.Fatal("parsed zero models; the page layout may have changed (parser contract drift)")
	}
	if err := write(*output, *url, models); err != nil {
		log.Fatalf("failed to write %s: %v", *output, err)
	}
	fmt.Printf("wrote %d PTU capacity entries to %s\n", len(models), *output)
}

func fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url) //nolint:gosec // developer-supplied flag, not user input
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func parseCapacity(body []byte) (map[string]capacity, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var tables []table
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tables = append(tables, extractTable(n))
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	normalizedRatio, normalized := normalizedTokenModels(tables)

	out := map[string]capacity{}
	for _, t := range tables {
		if len(t.header) < 2 || !strings.EqualFold(t.header[0], "Topic") {
			continue
		}
		for col := 1; col < len(t.header); col++ {
			name := strings.ToLower(modelName(t.header[col]))
			if name == "" {
				continue
			}
			c, ok := columnCapacity(t, col, normalized[name], normalizedRatio)
			if ok {
				out[name] = c
			}
		}
	}
	return out, nil
}

// normalizedTokenModels finds models sized by normalized tokens (the GPT-6
// family) and the weight of one short-context output token. The sizing
// tables use per-token-type weights instead of a flat output ratio, so the
// short-context output weight is used as the ratio. Cached input and
// long-context weights are not modeled: cached input is counted at full
// weight (overestimates PTUs) and long-context at the short-context weight
// (underestimates).
func normalizedTokenModels(tables []table) (ratio int, models map[string]bool) {
	models = map[string]bool{}
	for _, t := range tables {
		if len(t.header) < 2 {
			continue
		}
		switch {
		case strings.EqualFold(t.header[0], "Model") && strings.Contains(strings.ToLower(strings.Join(t.header, " ")), "normalized tpm per ptu"):
			for _, r := range t.rows {
				if len(r) > 0 {
					models[strings.ToLower(strings.TrimSpace(r[0]))] = true
				}
			}
		case strings.EqualFold(t.header[0], "Token type"):
			for _, r := range t.rows {
				if len(r) > 1 && strings.EqualFold(r[0], "Short-context output") {
					if f, err := strconv.ParseFloat(strings.TrimSpace(r[len(r)-1]), 64); err == nil {
						ratio = int(f + 0.5)
					}
				}
			}
		}
	}
	return ratio, models
}

// columnCapacity reads one model column; ok is false when a numeric value the
// capacity shape requires (such as the output ratio) is missing.
func columnCapacity(t table, col int, normalized bool, normalizedRatio int) (capacity, bool) {
	vals := map[string]string{}
	for _, r := range t.rows {
		if len(r) > col {
			vals[strings.ReplaceAll(strings.ToLower(r[0]), "&", "and")] = r[col]
		}
	}
	get := func(label string) (int, bool) {
		for k, v := range vals {
			if strings.HasPrefix(k, label) {
				return parseInt(v)
			}
		}
		return 0, false
	}
	var c capacity
	var ok1, ok2, ok3, ok4 bool
	c.InputTPMPerPTU, ok1 = get("input tpm per ptu")
	c.OutputRatio, ok2 = get("output-to-input ratio")
	if !ok2 && normalized && normalizedRatio > 0 {
		c.OutputRatio, ok2 = normalizedRatio, true
	}
	c.GlobalMin, ok3 = get("global and data zone provisioned minimum")
	c.GlobalIncrement, ok4 = get("global and data zone provisioned scale")
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return capacity{}, false
	}
	// Regional values may be "NA" (global/data zone only models); keep 0.
	c.RegionalMin, _ = get("regional provisioned minimum")
	c.RegionalIncrem, _ = get("regional provisioned scale")
	return c, true
}

func parseInt(s string) (int, bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", ""))
	n, err := strconv.Atoi(s)
	return n, err == nil
}

// modelName returns the first line of a header cell: the page renders "name,
// version" with the version after the first comma or line break.
func modelName(cell string) string {
	name := strings.TrimSpace(strings.SplitN(cell, ",", 2)[0])
	return strings.TrimSpace(strings.SplitN(name, "\n", 2)[0])
}

func extractTable(n *html.Node) table {
	var t table
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var cells []string
			isHeader := false
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type != html.ElementNode || (c.Data != "td" && c.Data != "th") {
					continue
				}
				if c.Data == "th" {
					isHeader = true
				}
				cells = append(cells, cellText(c))
			}
			if isHeader && t.header == nil {
				t.header = cells
			} else if len(cells) > 0 {
				t.rows = append(t.rows, cells)
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return t
}

// cellText flattens a cell, rendering <br> as a comma so "name<br>date"
// splits predictably in modelName.
func cellText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		switch {
		case n.Type == html.ElementNode && n.Data == "sup":
			return // footnote markers would corrupt adjacent numbers
		case n.Type == html.TextNode:
			b.WriteString(n.Data)
		case n.Type == html.ElementNode && n.Data == "br":
			b.WriteString(",")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

func write(path, url string, models map[string]capacity) error {
	names := make([]string, 0, len(models))
	for n := range models {
		names = append(names, n)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("// Code generated by hack/code/ptu_capacity_gen; DO NOT EDIT.\n")
	b.WriteString("// Regenerate with: make ptu-capacity\n")
	fmt.Fprintf(&b, "// Source: %s\n\n", url)
	b.WriteString("package ptu\n\n")
	b.WriteString("// generatedPTUCapacity holds the documented PTU sizing parameters per\n")
	b.WriteString("// canonical model name. A regional minimum of 0 means regional\n")
	b.WriteString("// provisioned is not offered for the model.\n")
	b.WriteString("var generatedPTUCapacity = map[string]sizing{\n")
	for _, n := range names {
		c := models[n]
		fmt.Fprintf(&b, "\t%q: {%d, %d, %d, %d, %d, %d},\n", n,
			c.InputTPMPerPTU, c.OutputRatio, c.GlobalMin, c.GlobalIncrement, c.RegionalMin, c.RegionalIncrem)
	}
	b.WriteString("}\n")

	formatted, err := format.Source([]byte(b.String()))
	if err != nil {
		return fmt.Errorf("formatting generated source: %w", err)
	}
	return os.WriteFile(path, formatted, 0o644)
}
