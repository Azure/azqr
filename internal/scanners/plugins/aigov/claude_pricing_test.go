// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import "testing"

func TestResolveClaudePrice_KnownModel(t *testing.T) {
	price, ok := resolveClaudePrice("claude-sonnet-4-5")
	if !ok {
		t.Fatal("expected claude-sonnet-4-5 to resolve from embedded snapshot")
	}
	if price.Source != "claude_ccu_equivalent" {
		t.Errorf("expected source claude_ccu_equivalent, got %q", price.Source)
	}
	if price.InputPerMTokUSD <= 0 || price.OutputPerMTokUSD <= 0 {
		t.Errorf("expected positive rates, got input=%v output=%v", price.InputPerMTokUSD, price.OutputPerMTokUSD)
	}
}

func TestResolveClaudePrice_UnknownModel(t *testing.T) {
	if _, ok := resolveClaudePrice("claude-nonexistent-model"); ok {
		t.Fatal("expected unknown model to not resolve (no family fallback)")
	}
}

func TestResolveClaudePrice_AliasMatch(t *testing.T) {
	price, ok := resolveClaudePrice("Claude_Sonnet 4-5")
	if !ok {
		t.Fatal("expected alias/casing variant to resolve via canonical name")
	}
	if price.InputPerMTokUSD <= 0 {
		t.Errorf("expected positive input rate, got %v", price.InputPerMTokUSD)
	}
}

func TestEstimatedCostUSD(t *testing.T) {
	price := priceResolution{InputPerMTokUSD: 3.0, OutputPerMTokUSD: 15.0}
	got := estimatedCostUSD(price, 1_000_000, 1_000_000)
	want := 18.0
	if got != want {
		t.Errorf("estimatedCostUSD() = %v, want %v", got, want)
	}
}
