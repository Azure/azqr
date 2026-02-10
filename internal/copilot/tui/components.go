// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders the bottom status bar
type StatusBar struct {
	styles *Styles
	width  int
}

// NewStatusBar creates a new status bar
func NewStatusBar(styles *Styles, width int) *StatusBar {
	return &StatusBar{
		styles: styles,
		width:  width,
	}
}

// Render returns the rendered status bar.
// statusText is an optional transient notice that replaces the hint when non-empty.
// scrollArrow is an optional styled scroll indicator string (e.g. ↑, ↓, ↕).
func (sb *StatusBar) Render(model, mode string, ctrlCPressed bool, scrollArrow string, statusText string, level statusLevel) string {
	// Left side: transient status > ctrl+c warning > normal hints
	var leftHints string
	switch {
	case statusText != "":
		switch level {
		case statusError:
			leftHints = sb.styles.ErrorText.Render("✗ " + statusText)
		case statusWarn:
			leftHints = sb.styles.WarningText.Render("⚠ " + statusText)
		default:
			leftHints = sb.styles.InfoText.Render("● " + statusText)
		}
	case ctrlCPressed:
		leftHints = "press ctrl+c again to exit"
	default:
		leftHints = sb.styles.StatusHint.Render("shift+tab cycle mode  ctrl+e editor")
	}

	// Right side: optional scroll arrow + model + mode
	rightInfo := ""
	if scrollArrow != "" {
		rightInfo = scrollArrow + " "
	}
	rightInfo += sb.styles.StatusModel.Render(model) +
		sb.styles.HintText.Render(" (") +
		sb.styles.StatusMode.Render(mode) +
		sb.styles.HintText.Render(")")

	// Calculate padding
	leftLen := lipgloss.Width(leftHints)
	rightLen := lipgloss.Width(rightInfo)
	padding := sb.width - leftLen - rightLen - 2
	if padding < 0 {
		padding = 0
	}

	return leftHints + strings.Repeat(" ", padding) + rightInfo
}

// SetWidth updates the status bar width
func (sb *StatusBar) SetWidth(width int) {
	sb.width = width
}

// Dropdown represents an autocomplete dropdown
type Dropdown struct {
	styles   *Styles
	Items    []DropdownItem
	Selected int
	Visible  bool
	MaxItems int
}

// DropdownItem is a single item in the dropdown
type DropdownItem struct {
	Label       string
	Description string
	Value       string
}

// NewDropdown creates a new dropdown
func NewDropdown(styles *Styles) *Dropdown {
	return &Dropdown{
		styles:   styles,
		MaxItems: 15,
	}
}

// SetItems updates the dropdown items
func (d *Dropdown) SetItems(items []DropdownItem) {
	// Only reset selection if the items actually changed
	if !d.itemsEqual(items) {
		d.Items = items
		d.Selected = 0
	}
	d.Visible = len(items) > 0
}

// itemsEqual checks if the new items are the same as current items
func (d *Dropdown) itemsEqual(items []DropdownItem) bool {
	if len(d.Items) != len(items) {
		return false
	}
	for i, item := range items {
		if d.Items[i].Value != item.Value {
			return false
		}
	}
	return true
}

// MoveUp moves selection up
func (d *Dropdown) MoveUp() {
	if d.Selected > 0 {
		d.Selected--
	}
}

// MoveDown moves selection down
func (d *Dropdown) MoveDown() {
	if d.Selected < len(d.Items)-1 {
		d.Selected++
	}
}

// SelectedItem returns the currently selected item
func (d *Dropdown) SelectedItem() *DropdownItem {
	if d.Selected >= 0 && d.Selected < len(d.Items) {
		return &d.Items[d.Selected]
	}
	return nil
}

// Render returns the rendered dropdown
func (d *Dropdown) Render() string {
	if !d.Visible || len(d.Items) == 0 {
		return ""
	}

	maxToShow := min(d.MaxItems, len(d.Items))

	// Find the longest command name for alignment
	maxNameLen := 0
	for i := 0; i < maxToShow; i++ {
		nameLen := len(d.Items[i].Label) + 1 // +1 for the slash
		if nameLen > maxNameLen {
			maxNameLen = nameLen
		}
	}

	var lines []string
	for i := 0; i < maxToShow; i++ {
		item := d.Items[i]
		name := "/" + item.Label

		// Pad name to align descriptions
		padding := maxNameLen - len(name) + 4
		if padding < 2 {
			padding = 2
		}
		paddedName := name + strings.Repeat(" ", padding)

		var line string
		if i == d.Selected {
			line = d.styles.CommandName.Render(paddedName)
		} else {
			line = d.styles.CommandSlash.Render(paddedName)
		}

		if item.Description != "" {
			line += d.styles.CommandDesc.Render(item.Description)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// HistoryEntry represents a single conversation turn
type HistoryEntry struct {
	Role    string // "user" or "assistant"
	Content string
}
