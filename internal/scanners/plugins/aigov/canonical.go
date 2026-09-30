// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"regexp"
	"strings"
)

var (
	whitespaceOrUnderscoreRe = regexp.MustCompile(`[\s_]+`)
	repeatedDashRe           = regexp.MustCompile(`-+`)
)

// canonicalModelName normalizes harmless naming variants (case, whitespace,
// underscores) into a single canonical form used to key pricing and PTU
// capacity lookups, without applying any model-family fallback.
func canonicalModelName(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = whitespaceOrUnderscoreRe.ReplaceAllString(normalized, "-")
	normalized = repeatedDashRe.ReplaceAllString(normalized, "-")
	return normalized
}

// publisherPrefix maps a canonical model name prefix to its publisher.
type publisherPrefix struct {
	prefix    string
	publisher string
}

// publisherPrefixes is a best-effort classification table. It never changes
// pricing resolution (no family fallback); it only helps decide which
// purchasing program (for example Azure PTU) plausibly applies.
var publisherPrefixes = []publisherPrefix{
	{"claude-", "anthropic"},
	{"ministral-", "mistral"},
	{"mistral-", "mistral"},
	{"llama-", "meta"},
	{"meta-llama", "meta"},
	{"deepseek-", "deepseek"},
	{"gemini-", "google"},
}

// inferPublisher best-effort infers the publisher of a model from its
// canonical name. It returns "" when no known prefix matches.
func inferPublisher(modelName string) string {
	canonical := canonicalModelName(modelName)
	for _, p := range publisherPrefixes {
		if strings.HasPrefix(canonical, p.prefix) {
			return p.publisher
		}
	}
	return ""
}
