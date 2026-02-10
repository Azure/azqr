// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"reflect"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap holds all key bindings for the TUI.
// Each field is exported so KeyMapToSlice can enumerate them via reflection.
type KeyMap struct {
	Send      key.Binding
	Newline   key.Binding
	Editor    key.Binding
	Quit      key.Binding
	Escape    key.Binding
	CycleMode key.Binding

	ScrollUp         key.Binding
	ScrollDown       key.Binding
	HalfPageUp       key.Binding
	HalfPageDown     key.Binding
	CtrlHalfPageUp   key.Binding
	CtrlHalfPageDown key.Binding
	GotoTop          key.Binding
	GotoBottom       key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send message"),
		),
		Newline: key.NewBinding(
			key.WithKeys("\\"),
			key.WithHelp("\\+enter", "insert newline"),
		),
		Editor: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "open $EDITOR"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit (press twice)"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel / clear"),
		),
		CycleMode: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "cycle mode"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "scroll down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "half page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "half page down"),
		),
		CtrlHalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "half page up"),
		),
		CtrlHalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "half page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("ctrl+home"),
			key.WithHelp("ctrl+home", "go to top"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("ctrl+end"),
			key.WithHelp("ctrl+end", "go to bottom"),
		),
	}
}

// KeyMapToSlice returns all key.Binding values from a KeyMap struct using reflection.
// This is used to enumerate bindings for the /help output.
func KeyMapToSlice(km KeyMap) []key.Binding {
	v := reflect.ValueOf(km)
	t := v.Type()
	var bindings []key.Binding
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		if b, ok := f.Interface().(key.Binding); ok {
			bindings = append(bindings, b)
		}
	}
	return bindings
}
