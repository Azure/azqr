// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import tea "github.com/charmbracelet/bubbletea"

// CmdHandler wraps any tea.Msg as a tea.Cmd.
// This is the canonical one-liner for dispatching a typed message
// without an explicit goroutine.
func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg { return msg }
}
