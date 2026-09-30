// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import "testing"

func TestCanonicalModelName(t *testing.T) {
	tests := map[string]string{
		"GPT-4o":        "gpt-4o",
		"gpt_4o mini":   "gpt-4o-mini",
		"  Claude-Opus": "claude-opus",
		"a---b":         "a-b",
	}
	for input, want := range tests {
		if got := canonicalModelName(input); got != want {
			t.Errorf("canonicalModelName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestInferPublisher(t *testing.T) {
	tests := map[string]string{
		"claude-opus-4.5": "anthropic",
		"Mistral-Large":   "mistral",
		"gpt-4o":          "",
		"llama-3-70b":     "meta",
	}
	for input, want := range tests {
		if got := inferPublisher(input); got != want {
			t.Errorf("inferPublisher(%q) = %q, want %q", input, got, want)
		}
	}
}
