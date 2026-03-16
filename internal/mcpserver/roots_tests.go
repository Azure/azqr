// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import "testing"

func TestFileURIToPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// ── Unix paths ──────────────────────────────────────────────────────
		{
			name:  "unix three-slash URI",
			input: "file:///home/user/workspace",
			want:  "/home/user/workspace",
		},
		{
			name:  "unix three-slash URI with subdirectory",
			input: "file:///var/log/ghqr",
			want:  "/var/log/ghqr",
		},

		// ── Windows paths (uppercase drive letter) ──────────────────────────
		{
			name:  "windows three-slash URI uppercase drive",
			input: "file:///C:/Users/gh/workspace",
			want:  "C:/Users/gh/workspace",
		},
		{
			name:  "windows three-slash URI uppercase Z drive",
			input: "file:///Z:/Projects",
			want:  "Z:/Projects",
		},

		// ── Windows paths (lowercase drive letter) ──────────────────────────
		{
			name:  "windows three-slash URI lowercase drive",
			input: "file:///c:/Users/gh/workspace",
			want:  "c:/Users/gh/workspace",
		},

		// ── Windows paths with percent-encoded colon (%3A) ──────────────────
		// VS Code sends file:///c%3A/Users/... per Issue 1.
		{
			name:  "windows percent-encoded colon lowercase",
			input: "file:///c%3A/Users/gh/workspace",
			want:  "c:/Users/gh/workspace",
		},
		{
			name:  "windows percent-encoded colon uppercase",
			input: "file:///C%3A/Users/gh/workspace",
			want:  "C:/Users/gh/workspace",
		},

		// ── Paths that are not file URIs ──────────────────────────────────
		{
			name:  "plain path passthrough",
			input: "C:/Users/gh",
			want:  "C:/Users/gh",
		},
		{
			name:  "unix plain path passthrough",
			input: "/home/user",
			want:  "/home/user",
		},

		// ── Edge cases ───────────────────────────────────────────────────────
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "file URI with spaces (percent-encoded)",
			input: "file:///home/user/my%20project",
			want:  "/home/user/my project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileURIToPath(tt.input)
			if got != tt.want {
				t.Errorf("fileURIToPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsASCIILetter(t *testing.T) {
	for b := byte('a'); b <= 'z'; b++ {
		if !isASCIILetter(b) {
			t.Errorf("isASCIILetter(%q) = false, want true", b)
		}
	}

	for b := byte('A'); b <= 'Z'; b++ {
		if !isASCIILetter(b) {
			t.Errorf("isASCIILetter(%q) = false, want true", b)
		}
	}

	nonLetters := []byte{'0', '9', ':', '/', '\\', ' ', '-', '_'}
	for _, b := range nonLetters {
		if isASCIILetter(b) {
			t.Errorf("isASCIILetter(%q) = true, want false", b)
		}
	}
}
