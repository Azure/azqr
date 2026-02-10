// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"strings"
)

// View implements tea.Model.
// Layout (top to bottom):
//
//	viewport  – scrollable conversation history (fills most of the terminal)
//	separator
//	input (textarea, 1–5 rows)
//	separator
//	status bar
//
// The command dropdown is rendered as a floating overlay anchored just above
// the separator so it does not displace the status bar.
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	separator := m.styles.InputSeparatorLine(m.width - 2)

	base := strings.Join([]string{
		m.viewport.View(),
		separator,
		m.input.View(),
		separator,
		m.statusBar.Render(m.config.Model, m.config.Mode, m.ctrlCPressed, m.scrollArrow(), m.statusText, m.statusLevel),
	}, "\n")

	if m.showDropdown && m.dropdown.Visible {
		overlay := m.dropdown.Render()
		if overlay != "" {
			overlayHeight := strings.Count(overlay, "\n") + 1
			// Anchor: place overlay just above the first separator.
			// Row = viewport height + 1 (separator) - overlayHeight - 1 (zero-based)
			row := m.viewport.Height - overlayHeight
			if row < 0 {
				row = 0
			}
			base = PlaceOverlay(2, row, overlay, base)
		}
	}

	return base
}

// scrollArrow returns a styled arrow indicating scroll position, or "" when
// all content fits in the viewport.
func (m *Model) scrollArrow() string {
	if m.viewport.YOffset == 0 && m.viewport.AtBottom() {
		return ""
	}
	canUp := m.viewport.YOffset > 0
	canDown := !m.viewport.AtBottom()
	switch {
	case canUp && canDown:
		return m.styles.ThinkingText.Render("↕")
	case canUp:
		return m.styles.ThinkingText.Render("↑")
	default:
		return m.styles.ThinkingText.Render("↓")
	}
}

// renderActivity builds the live thinking/streaming state shown at the bottom
// of the viewport while a response is in progress.
func (m *Model) renderActivity() string {
	if m.state != StateThinking && m.state != StateStreaming {
		return ""
	}
	return m.renderChunks() + m.renderSpinner()
}

// renderChunks renders all activity chunks in order
func (m *Model) renderChunks() string {
	var b strings.Builder
	for i, chunk := range m.activity.Chunks {
		if i > 0 {
			b.WriteString("\n\n")
		}
		switch chunk.Type {
		case ChunkReasoning:
			b.WriteString(m.formatReasoning(chunk.Content))
		case ChunkAssistant:
			b.WriteString(m.styles.ThinkingText.Render("● "))
			b.WriteString(WrapText(chunk.Content, m.width))
		case ChunkTool:
			toolText := formatToolDisplay(chunk.Tool)
			b.WriteString(m.styles.SuccessText.Render("● ") + WrapText(toolText, m.width))
		}
	}
	return b.String()
}

// renderSpinner renders the spinner with current tool info
func (m *Model) renderSpinner() string {
	prefix := ""
	if len(m.activity.Chunks) > 0 {
		prefix = "\n"
	}
	spinnerText := "Thinking"
	if m.activity.CurrentTool.Name != "" {
		spinnerText = WrapText(formatToolDisplay(m.activity.CurrentTool), m.width)
	}
	return prefix + m.spinner.View() + " " + m.styles.ThinkingText.Render(spinnerText) + "\n"
}
