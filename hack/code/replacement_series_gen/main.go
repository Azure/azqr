// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// replacement_series_gen fetches the "Recommended Replacement VM Series" table from
// Microsoft Learn and writes a generated Go source file used by the vm-modernization
// plugin.
//
// The generated file (zz_generated.replacement_series.go) contains one package-level
// variable, documentedReplacements, with one entry per table row:
//   - Sources: the current VM series covered by the row
//   - Targets: the documented target VM series lines
//   - Differences: the documented specification differences in the target VM
//
// Source:
//   - https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/lifecycle/retirement/retired-sizes-modernization-guide
//     (read from the page's Markdown source in the MicrosoftDocs/azure-compute-docs repository)
//
// Usage:
//
//	go run ./hack/code/replacement_series_gen/main.go
//	go run ./hack/code/replacement_series_gen/main.go --output ./internal/scanners/plugins/vmmodernization/zz_generated.replacement_series.go
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	defaultURL = "https://raw.githubusercontent.com/MicrosoftDocs/azure-compute-docs/main/articles/virtual-machines/sizes/lifecycle/retirement/retired-sizes-modernization-guide.md"
	pageURL    = "https://learn.microsoft.com/en-us/azure/virtual-machines/sizes/lifecycle/retirement/retired-sizes-modernization-guide"

	defaultOutput = "internal/scanners/plugins/vmmodernization/zz_generated.replacement_series.go"

	sectionHeading = "## Recommended Replacement VM Series"
)

var (
	lineBreak     = regexp.MustCompile(`(?i)<br\s*/?>`)
	markdownURL   = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	seriesSuffix  = regexp.MustCompile(`(?i)-series$`)
	parenthetical = regexp.MustCompile(`\([^)]*\)`)
)

// replacement is one parsed table row.
type replacement struct {
	Sources     []string
	Targets     []string
	Differences []string
}

func main() {
	outputPath := flag.String("output", defaultOutput, "path to write the generated zz_generated.replacement_series.go")
	sourceURL := flag.String("url", defaultURL, "URL of the Markdown source of the Learn page")
	flag.Parse()

	body, err := fetch(*sourceURL)
	if err != nil {
		log.Fatalf("fetching %s: %v", *sourceURL, err)
	}

	rows, err := parseReplacements(string(body))
	if err != nil {
		log.Fatalf("parsing replacement table: %v", err)
	}

	src, err := render(rows)
	if err != nil {
		log.Fatalf("rendering: %v", err)
	}

	if err := os.WriteFile(*outputPath, src, 0o644); err != nil {
		log.Fatalf("writing %s: %v", *outputPath, err)
	}
	fmt.Printf("wrote %d replacement rows to %s\n", len(rows), *outputPath)
}

func fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// parseReplacements reads the first Markdown table under the replacement series
// heading. Rows that do not have three populated cells are an error so that a
// layout change on the page fails the generation instead of producing bad data.
func parseReplacements(markdown string) ([]replacement, error) {
	var rows []replacement
	inSection, inTable := false, false

	scanner := bufio.NewScanner(strings.NewReader(markdown))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if !inSection {
			inSection = strings.EqualFold(line, sectionHeading)
			continue
		}
		if !strings.HasPrefix(line, "|") {
			if inTable {
				break
			}
			continue
		}
		if !inTable {
			// The first table line is the header; the next is the separator.
			inTable = true
			continue
		}
		if strings.Trim(line, "|-: ") == "" {
			continue
		}

		row, err := parseRow(line)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows found under %q", sectionHeading)
	}
	return rows, nil
}

func parseRow(line string) (replacement, error) {
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	cells := strings.Split(line, "|")
	if len(cells) != 3 {
		return replacement{}, fmt.Errorf("expected 3 cells, got %d in %q", len(cells), line)
	}

	row := replacement{
		Sources:     cleanSources(cells[0]),
		Targets:     cleanLines(cells[1]),
		Differences: cleanLines(cells[2]),
	}
	if len(row.Sources) == 0 || len(row.Targets) == 0 || len(row.Differences) == 0 {
		return replacement{}, fmt.Errorf("empty cell in %q", line)
	}
	return row, nil
}

// cleanLines splits a cell on <br>, strips Markdown links and drops empty lines.
func cleanLines(cell string) []string {
	cell = markdownURL.ReplaceAllString(cell, "$1")
	cell = strings.ReplaceAll(cell, `\_`, "_")

	var lines []string
	for _, part := range lineBreak.Split(cell, -1) {
		if part = strings.Join(strings.Fields(part), " "); part != "" {
			lines = append(lines, part)
		}
	}
	return lines
}

// cleanSources keeps the series names of the first cell, dropping parenthetical
// size lists and the "-series" suffix.
func cleanSources(cell string) []string {
	var sources []string
	for _, line := range cleanLines(parenthetical.ReplaceAllString(cell, "")) {
		sources = append(sources, seriesSuffix.ReplaceAllString(line, ""))
	}
	return sources
}

func render(rows []replacement) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("// Copyright (c) Microsoft Corporation.\n// Licensed under the MIT License.\n\n")
	b.WriteString("// Code generated by hack/code/replacement_series_gen; DO NOT EDIT.\n")
	fmt.Fprintf(&b, "// Source: %s\n\n", pageURL)
	b.WriteString("package vmmodernization\n\n")
	b.WriteString("// documentedReplacements is the Microsoft-documented replacement VM series table.\n")
	b.WriteString("var documentedReplacements = []documentedReplacement{\n")
	for _, r := range rows {
		b.WriteString("\t{\n")
		writeField(&b, "Sources", r.Sources)
		writeField(&b, "Targets", r.Targets)
		writeField(&b, "Differences", r.Differences)
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n")
	return format.Source(b.Bytes())
}

func writeField(b *bytes.Buffer, name string, values []string) {
	fmt.Fprintf(b, "\t\t%s: []string{\n", name)
	for _, v := range values {
		fmt.Fprintf(b, "\t\t\t%q,\n", v)
	}
	b.WriteString("\t\t},\n")
}
