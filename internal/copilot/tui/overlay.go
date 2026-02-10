// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
//
// PlaceOverlay is adapted from the opencode-ai/opencode project
// (https://github.com/opencode-ai/opencode, now charmbracelet/crush)
// which itself derives from a lipgloss PR by meowgorithm.
// Used under the MIT License.

package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PlaceOverlay renders the fg string on top of the bg string at position (x, y).
// Coordinates are 0-based column/row offsets from the top-left of bg.
// If the overlay would exceed the bg bounds it is clamped.
func PlaceOverlay(x, y int, fg, bg string) string {
	fgLines := strings.Split(fg, "\n")
	bgLines := strings.Split(bg, "\n")

	fgHeight := len(fgLines)
	bgHeight := len(bgLines)

	// Measure widths
	fgWidth := 0
	for _, l := range fgLines {
		if w := lipgloss.Width(l); w > fgWidth {
			fgWidth = w
		}
	}
	bgWidth := 0
	for _, l := range bgLines {
		if w := lipgloss.Width(l); w > bgWidth {
			bgWidth = w
		}
	}

	// Clamp overlay position so it fits within bg
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+fgWidth > bgWidth {
		x = bgWidth - fgWidth
		if x < 0 {
			x = 0
		}
	}
	if y+fgHeight > bgHeight {
		y = bgHeight - fgHeight
		if y < 0 {
			y = 0
		}
	}

	result := make([]string, bgHeight)
	copy(result, bgLines)

	for row, fgLine := range fgLines {
		bgRow := y + row
		if bgRow >= bgHeight {
			break
		}
		bg := result[bgRow]
		// Pad bg line to at least x columns
		bgActualWidth := lipgloss.Width(bg)
		if bgActualWidth < x {
			bg += strings.Repeat(" ", x-bgActualWidth)
		}
		// Splice: left portion of bg, then the full fg line, then right of bg
		left := visibleLeft(bg, x)
		right := visibleRight(bg, x+fgWidth)
		result[bgRow] = left + fgLine + right
	}

	return strings.Join(result, "\n")
}

// visibleLeft returns the first n visible columns of s (ANSI-aware, best-effort).
func visibleLeft(s string, n int) string {
	if n <= 0 {
		return ""
	}
	col := 0
	inEscape := false
	for i, r := range s {
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if col >= n {
			return s[:i]
		}
		col++
	}
	return s
}

// visibleRight returns the substring of s starting at visual column n (ANSI-aware, best-effort).
func visibleRight(s string, n int) string {
	col := 0
	inEscape := false
	for i, r := range s {
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if col >= n {
			return s[i:]
		}
		col++
	}
	return ""
}
