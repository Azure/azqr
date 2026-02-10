// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"strings"
)

// View implements tea.Model.
// Layout (top to bottom):
//
//	content   – conversation history printed directly
//	separator
//	input (textarea, 1–5 rows)
//	separator
//	status bar
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	separator := m.styles.InputSeparatorLine(m.width - 2)
	content := m.buildViewportContent()

	parts := []string{
		content,
		separator,
		m.input.View(),
	}

	if m.showDropdown && m.dropdown.Visible {
		if overlay := m.dropdown.Render(); overlay != "" {
			parts = append(parts, overlay)
		}
	}

	parts = append(parts,
		separator,
		m.statusBar.Render(m.config.Model, m.config.Mode, m.ctrlCPressed, "", m.statusText, m.statusLevel),
	)

	return strings.Join(parts, "\n")
}

// renderActivity builds the live thinking/streaming state.
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
