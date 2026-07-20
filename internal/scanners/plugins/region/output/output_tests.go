// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package output

import (
	"strings"
	"testing"
)

func TestSafeSheetName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "short name unchanged",
			in:   "SvcAvail_eastus",
			want: "SvcAvail_eastus",
		},
		{
			name: "empty string unchanged",
			in:   "",
			want: "",
		},
		{
			name: "exactly 31 chars unchanged",
			in:   strings.Repeat("a", 31),
			want: strings.Repeat("a", 31),
		},
		{
			name: "32 chars truncated to 31",
			in:   strings.Repeat("a", 32),
			want: strings.Repeat("a", 31),
		},
		{
			name: "long name truncated to 31",
			in:   "SvcAvail_australiacentral_extra_region",
			want: "SvcAvail_australiacentral_extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeSheetName(tt.in)
			if got != tt.want {
				t.Errorf("safeSheetName(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if len([]rune(got)) > 31 {
				t.Errorf("safeSheetName(%q) returned %d runes, exceeds Excel limit of 31", tt.in, len([]rune(got)))
			}
		})
	}
}

func TestSafeSheetName_MultiByteRunes(t *testing.T) {
	// 40 multi-byte runes (each 'é' is 2 bytes in UTF-8) must be truncated by
	// rune count, not byte count, and must not split a rune.
	in := strings.Repeat("é", 40)
	got := safeSheetName(in)

	if rc := len([]rune(got)); rc != 31 {
		t.Errorf("expected 31 runes, got %d", rc)
	}
	if got != strings.Repeat("é", 31) {
		t.Errorf("unexpected truncation result: %q", got)
	}
}
