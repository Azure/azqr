// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap defines all keybindings used by the TUI.
type keyMap struct {
	ModeSwitch  key.Binding
	HistoryUp   key.Binding
	HistoryDown key.Binding
}

// defaultKeys returns the standard keybinding set.
func defaultKeys() keyMap {
	return keyMap{
		ModeSwitch: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "cycle mode"),
		),
		HistoryUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "previous prompt"),
		),
		HistoryDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "next prompt"),
		),
	}
}
