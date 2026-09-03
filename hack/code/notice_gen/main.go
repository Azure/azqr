// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// notice_gen reads raw per-dependency license records emitted by
// "go-licenses report" (via a delimiter template) and writes a compact
// NOTICE.md, grouping dependencies that share identical license text under
// a single heading.
//
// Usage:
//
//	go run ./hack/code/notice_gen/main.go <raw-records-file> <output-file>
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	// recordSep separates dependency records in the raw input.
	recordSep = "\x1e"
	// fieldSep separates fields (name, version, license name, license text)
	// within a single dependency record.
	fieldSep = "\x1f"
)

const header = `# NOTICES AND INFORMATION

Do Not Translate or Localize

Azure Quick Review (azqr) incorporates material from the third-party open
source components listed below, each distributed under a permissive license
(MIT, Apache-2.0, BSD-3-Clause, or BSD-2-Clause). No GPL/LGPL-licensed
components are used. The full text of each component's license is reproduced
below as required by its terms. Components that share identical license text
are grouped together.

## Third-Party Components
`

type dependency struct {
	name    string
	version string
}

type licenseGroup struct {
	licenseName string
	licenseText string
	members     []dependency
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <raw-records-file> <output-file>\n", os.Args[0])
		os.Exit(1)
	}
	rawPath, outputPath := os.Args[1], os.Args[2]

	raw, err := os.ReadFile(rawPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", rawPath, err)
		os.Exit(1)
	}

	groups, total, err := groupRecords(string(raw))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing license records: %v\n", err)
		os.Exit(1)
	}

	out := render(groups, total)

	if err := os.WriteFile(outputPath, []byte(out), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outputPath, err)
		os.Exit(1)
	}
}

// groupRecords parses the raw delimiter-separated records and groups
// dependencies that share identical (license name, license text) pairs,
// preserving first-seen order for stable, deterministic output.
func groupRecords(raw string) ([]*licenseGroup, int, error) {
	records := strings.Split(raw, recordSep)

	index := map[string]*licenseGroup{}
	var groups []*licenseGroup
	total := 0

	for _, rec := range records {
		if rec == "" {
			continue
		}
		fields := strings.Split(rec, fieldSep)
		if len(fields) != 4 {
			return nil, 0, fmt.Errorf("malformed record (expected 4 fields, got %d): %q", len(fields), rec)
		}
		name, version, licenseName, licenseText := fields[0], fields[1], fields[2], strings.TrimSpace(fields[3])
		total++

		key := licenseName + "\x00" + licenseText
		g, ok := index[key]
		if !ok {
			g = &licenseGroup{licenseName: licenseName, licenseText: licenseText}
			index[key] = g
			groups = append(groups, g)
		}
		g.members = append(g.members, dependency{name: name, version: version})
	}

	return groups, total, nil
}

// render formats the grouped license data as Markdown. When multiple
// distinct license texts share the same license name (e.g. "MIT" with
// different copyright holders), each variant gets a distinguishing suffix.
func render(groups []*licenseGroup, total int) string {
	nameCounts := map[string]int{}
	for _, g := range groups {
		nameCounts[g.licenseName]++
	}
	seenNameIdx := map[string]int{}

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	fmt.Fprintf(&b, "%d third-party components are included below, grouped into %d distinct license texts.\n\n", total, len(groups))

	for _, g := range groups {
		members := append([]dependency(nil), g.members...)
		sort.Slice(members, func(i, j int) bool { return members[i].name < members[j].name })

		heading := g.licenseName
		if nameCounts[g.licenseName] > 1 {
			seenNameIdx[g.licenseName]++
			heading = fmt.Sprintf("%s (variant %d)", g.licenseName, seenNameIdx[g.licenseName])
		}

		fmt.Fprintf(&b, "### %s\n\n", heading)
		b.WriteString("Applies to:\n\n")
		for _, m := range members {
			fmt.Fprintf(&b, "- `%s` %s\n", m.name, m.version)
		}
		b.WriteString("\n```\n")
		b.WriteString(g.licenseText)
		b.WriteString("\n```\n\n")
	}

	return b.String()
}
