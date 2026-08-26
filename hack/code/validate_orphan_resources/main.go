// validate_orphan_resources checks that every query listed in the Azure Orphaned
// Resources workbook has a corresponding implementation under
// internal/graph/azure-orphan-resources.
//
// The expected query list is generated at runtime by downloading and parsing:
// https://github.com/dolevshor/azure-orphan-resources/blob/main/Queries/orphan-resources-queries.md
//
// Usage:
//
//	go run ./hack/code/validate_orphan_resources [path-to-azure-orphan-resources-dir]
//
// The default path is internal/graph/azure-orphan-resources relative to the
// current working directory.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const markdownURL = "https://raw.githubusercontent.com/dolevshor/azure-orphan-resources/main/Queries/orphan-resources-queries.md"

// expectedQuery describes one orphan-resources query from the upstream workbook.
type expectedQuery struct {
	resourceType string
	description  string
	section      string
	kqlText      string // raw KQL from the upstream Markdown
}

// headingOverrides maps #### heading text to the correct recommendationResourceType
// for cases where the KQL query operates on a parent resource type rather than the
// logical sub-resource (e.g. Subnets are queried via virtualnetworks), or where
// ARG uses a longer type path that differs from the ARM resource type used locally.
var headingOverrides = map[string]string{
	"Subnets":         "Microsoft.Network/virtualNetworks/subnets",
	"Resource Groups": "Microsoft.Resources/resourceGroups",
}

// typeExtractRe matches the first "type" filter in a KQL block:
//
//	| where type == "..."
//	| where type =~ "..."
//	| where type has "..."
var typeExtractRe = regexp.MustCompile(`(?i)\|\s*where\s+type\s*(?:==|=~|has)\s*['"]([^'"]+)['"]`)

// fetchExpectedQueries downloads the upstream Markdown, parses every ####
// sub-section, extracts the resource type from the embedded KQL, and returns
// the full list of expected queries.
func fetchExpectedQueries(url string) ([]expectedQuery, error) {
	resp, err := http.Get(url) //nolint:gosec,noctx // url is a compile-time constant
	if err != nil {
		return nil, fmt.Errorf("fetching markdown: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %s fetching %s", resp.Status, url)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return parseMarkdown(string(body))
}

// parseMarkdown extracts expectedQuery entries from the Markdown content.
func parseMarkdown(md string) ([]expectedQuery, error) {
	var queries []expectedQuery

	currentSection := "" // top-level ## heading
	currentHeading := "" // #### sub-heading
	inKQL := false
	var kqlBuf strings.Builder

	for _, line := range strings.Split(md, "\n") {
		// Track top-level (##) sections.
		if strings.HasPrefix(line, "## ") {
			currentSection = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			currentHeading = ""
			continue
		}

		// Track #### sub-headings — each one is one expected query.
		if strings.HasPrefix(line, "#### ") {
			// Flush any previous KQL buffer (shouldn't have one here, but be safe).
			inKQL = false
			kqlBuf.Reset()
			currentHeading = strings.TrimSpace(strings.TrimPrefix(line, "#### "))
			continue
		}

		// KQL code fence detection.
		trimmed := strings.TrimSpace(line)
		if (trimmed == "```kql" || trimmed == "``` kql") && !inKQL {
			inKQL = true
			kqlBuf.Reset()
			continue
		}
		if trimmed == "```" && inKQL {
			inKQL = false
			// We have a complete KQL block — derive the resource type.
			if currentHeading != "" {
				kql := kqlBuf.String()
				rt, err := resourceTypeFromKQL(currentHeading, kql)
				if err != nil {
					return nil, fmt.Errorf("section %q: %w", currentHeading, err)
				}
				queries = append(queries, expectedQuery{
					resourceType: rt,
					description:  currentHeading,
					section:      currentSection,
					kqlText:      strings.TrimSpace(kql),
				})
			}
			kqlBuf.Reset()
			continue
		}

		if inKQL {
			kqlBuf.WriteString(line)
			kqlBuf.WriteByte('\n')
		}
	}

	if len(queries) == 0 {
		return nil, fmt.Errorf("no queries parsed from markdown — check the URL or the document format")
	}
	return queries, nil
}

// resourceTypeFromKQL derives the recommendationResourceType for a query section.
// It first checks headingOverrides for known edge-cases, then falls back to
// extracting the first type filter from the KQL text.
func resourceTypeFromKQL(heading, kql string) (string, error) {
	if override, ok := headingOverrides[heading]; ok {
		return override, nil
	}
	m := typeExtractRe.FindStringSubmatch(kql)
	if m == nil {
		return "", fmt.Errorf("cannot extract resource type from KQL")
	}
	return m[1], nil
}

// recommendation mirrors the relevant fields from a queries.yaml entry.
type recommendation struct {
	Description                string `yaml:"description"`
	AprlGuid                   string `yaml:"aprlGuid"`
	RecommendationResourceType string `yaml:"recommendationResourceType"`
}

func main() {
	baseDir := filepath.Join("internal", "graph", "azure-orphan-resources")
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}

	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving path %s: %v\n", baseDir, err)
		os.Exit(1)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "directory not found: %s\n", absDir)
		os.Exit(1)
	}

	fmt.Printf("Fetching upstream queries from:\n  %s\n\n", markdownURL)
	expectedQueries, err := fetchExpectedQueries(markdownURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching expected queries: %v\n", err)
		os.Exit(1)
	}

	// Collect all recommendations from YAML files.
	// key: lowercase resource type, value: recommendation
	implemented := map[string]recommendation{}
	// key: aprlGuid, value: directory of the yaml file
	guidToDir := map[string]string{}

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}
		recs, loadErr := loadYAML(path)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, loadErr)
			return nil
		}
		for _, r := range recs {
			if r.RecommendationResourceType != "" {
				implemented[strings.ToLower(r.RecommendationResourceType)] = r
			}
			if r.AprlGuid != "" {
				guidToDir[r.AprlGuid] = filepath.Dir(path)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error scanning directory: %v\n", err)
		os.Exit(1)
	}

	exitCode := 0

	fmt.Printf("Orphan resource query coverage check\n")
	fmt.Printf("Directory: %s\n\n", absDir)
	fmt.Printf("Expected queries : %d\n", len(expectedQueries))
	fmt.Printf("Implemented types: %d\n\n", len(implemented))

	// Build markdown table: | Type | OK | Source KQL (GitHub) | Target KQL (local) |
	fmt.Println("| Resource Type | OK | Source Query (GitHub) | Target Query (local .kql) |")
	fmt.Println("|---|:---:|---|---|")

	missing := []expectedQuery{}
	for _, eq := range expectedQueries {
		rec, ok := implemented[strings.ToLower(eq.resourceType)]

		check := "✅"
		if !ok {
			check = "❌"
			missing = append(missing, eq)
		}

		sourceCell := mdCell(eq.kqlText)

		targetCell := "*(not found)*"
		if ok && rec.AprlGuid != "" {
			if dir, exists := guidToDir[rec.AprlGuid]; exists {
				kqlPath := filepath.Join(dir, "kql", rec.AprlGuid+".kql")
				if data, err := os.ReadFile(filepath.Clean(kqlPath)); err == nil { //nolint:gosec
					targetCell = mdCell(string(data))
				}
			}
		}

		fmt.Printf("| `%s` | %s | %s | %s |\n", eq.resourceType, check, sourceCell, targetCell)
	}
	fmt.Println()

	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "❌ Missing queries (%d):\n", len(missing))
		currentSection := ""
		for _, m := range missing {
			if m.section != currentSection {
				fmt.Fprintf(os.Stderr, "\n  [%s]\n", m.section)
				currentSection = m.section
			}
			fmt.Fprintf(os.Stderr, "    - %-60s %s\n", m.resourceType, m.description)
		}
		fmt.Fprintln(os.Stderr)
		exitCode = 1
	}

	// 2. Verify that every YAML entry with an aprlGuid has a matching KQL file.
	missingKQL := []string{}
	for guid, dir := range guidToDir {
		kqlPath := filepath.Join(dir, "kql", guid+".kql")
		if _, err := os.Stat(kqlPath); os.IsNotExist(err) {
			missingKQL = append(missingKQL, kqlPath)
		}
	}

	if len(missingKQL) > 0 {
		fmt.Fprintf(os.Stderr, "❌ YAML entries without a corresponding KQL file (%d):\n", len(missingKQL))
		for _, p := range missingKQL {
			fmt.Fprintf(os.Stderr, "    - %s\n", p)
		}
		fmt.Fprintln(os.Stderr)
		exitCode = 1
	}

	// 3. Detect KQL files that have no matching YAML entry.
	orphanedKQL := []string{}
	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".kql") {
			return nil
		}
		guid := strings.TrimSuffix(filepath.Base(path), ".kql")
		if _, ok := guidToDir[guid]; !ok {
			orphanedKQL = append(orphanedKQL, path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error scanning for KQL files: %v\n", err)
		os.Exit(1)
	}

	if len(orphanedKQL) > 0 {
		fmt.Fprintf(os.Stderr, "⚠️  KQL files without a matching YAML entry (%d):\n", len(orphanedKQL))
		for _, p := range orphanedKQL {
			fmt.Fprintf(os.Stderr, "    - %s\n", p)
		}
		fmt.Fprintln(os.Stderr)
		exitCode = 1
	}

	if exitCode == 0 {
		fmt.Printf("✅ All %d orphan resource queries are implemented and every KQL file is referenced!\n", len(expectedQueries))
	}

	os.Exit(exitCode)
}

// mdCell formats a KQL string as a single markdown table cell by collapsing
// lines into a space-separated sequence (skipping blank/comment lines),
// escaping pipe characters, and truncating to 120 characters.
func mdCell(kql string) string {
	var parts []string
	for _, line := range strings.Split(kql, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "//") {
			continue
		}
		parts = append(parts, t)
	}
	result := strings.ReplaceAll(strings.Join(parts, " "), "|", "\\|")
	return result
}

// loadYAML reads a queries.yaml file and returns the list of recommendations.
func loadYAML(path string) ([]recommendation, error) {
	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath) //nolint:gosec // path comes from filepath.Walk in a controlled directory
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	var recs []recommendation
	if err := yaml.Unmarshal(data, &recs); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}
	return recs, nil
}
