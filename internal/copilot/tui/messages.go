// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

// glamourRender renders markdown content with syntax highlighting and formatting
// suitable for a dark terminal. The output is trimmed of leading/trailing blank
// lines that glamour adds by default.
func glamourRender(content string, width int) string {
	if width < 20 {
		width = 80
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width-4), // -4 for glamour's default left margin
	)
	if err != nil {
		return content
	}
	out, err := r.Render(content)
	if err != nil {
		return content
	}
	return strings.Trim(out, "\n")
}

// helpContent returns the help text to display in the viewport.
func helpContent() string {
	entries := []struct{ key, desc string }{
		{"Enter", "Send message"},
		{"↑ / ↓", "Navigate prompt history"},
		{"shift+tab", "Cycle mode (agent → ask)"},
		{"ctrl+l", "Clear screen"},
		{"ctrl+c", "Cancel / clear input"},
		{"ctrl+c ×2", "Exit"},
		{"ctrl+d", "Exit"},
		{"Esc", "Cancel current operation"},
	}
	slashes := []struct{ cmd, desc string }{
		{"/help", "Show this help"},
		{"/clear", "Clear conversation"},
		{"/model", "Show current model"},
		{"/new", "Start a new conversation"},
		{"/exit", "Exit azqr copilot"},
	}

	var sb strings.Builder
	sb.WriteString(" " + styleHelpTitle.Render("Keyboard shortcuts") + "\n\n")
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("  %s  %s\n",
			styleHelpKey.Render(fmt.Sprintf("%-15s", e.key)),
			styleHelpDesc.Render(e.desc)))
	}
	sb.WriteString("\n " + styleHelpTitle.Render("Slash commands") + "\n\n")
	for _, s := range slashes {
		sb.WriteString(fmt.Sprintf("  %s  %s\n",
			styleHelpKey.Render(fmt.Sprintf("%-15s", s.cmd)),
			styleHelpDesc.Render(s.desc)))
	}
	return sb.String()
}

// wordWrap wraps text at the given column width, preserving existing newlines.
func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, wrapLine(line, width))
	}
	return strings.Join(result, "\n")
}

func wrapLine(line string, width int) string {
	if len(line) <= width {
		return line
	}
	var sb strings.Builder
	col := 0
	for _, word := range strings.Fields(line) {
		wl := len(word)
		if col == 0 {
			sb.WriteString(word)
			col = wl
		} else if col+1+wl <= width {
			sb.WriteByte(' ')
			sb.WriteString(word)
			col += 1 + wl
		} else {
			sb.WriteByte('\n')
			sb.WriteString(word)
			col = wl
		}
	}
	return sb.String()
}
