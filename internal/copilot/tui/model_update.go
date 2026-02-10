// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleResize(msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case spinner.TickMsg:
		return m.handleSpinnerTick(msg)

	case activityMsg:
		return m.handleActivityMsg(msg)

	case responseMsg:
		return m.handleResponse()

	case editorDoneMsg:
		m.input.SetValue(msg.content)
		m.recalcViewportHeight()
		return m, nil

	case errorMsg:
		m.activity = Activity{}
		m.state = StateReady
		m.setStatus(statusError, msg.err.Error())
		return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg { return statusMsg{} })

	case statusMsg:
		if msg.text == "" {
			// Clear timer fired
			m.statusText = ""
		} else {
			m.setStatus(msg.level, msg.text)
			return m, tea.Tick(4*time.Second, func(time.Time) tea.Msg { return statusMsg{} })
		}
		return m, nil

	case ctrlCClearMsg:
		m.ctrlCPressed = false
		return m, nil
	}

	// Forward to textarea and dropdown
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.updateDropdown()

	return m, cmd
}

// handleResize adjusts input and viewport dimensions.
// The textarea starts at 1 row and can grow up to 5 rows.
// The viewport takes all available height minus the textarea rows and 4 chrome
// rows (two separators + status bar row + one margin row).
func (m *Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	m.input.SetWidth(msg.Width - 4)
	m.statusBar.SetWidth(msg.Width)

	m.viewport.Width = msg.Width
	m.recalcViewportHeight()
	// Refresh content without forcing a scroll jump
	m.viewport.SetContent(m.buildViewportContent())

	return m, nil
}

// recalcViewportHeight adjusts the viewport height to fill the screen around the textarea.
// It uses m.inputHeight which is kept in sync whenever the textarea height changes.
func (m *Model) recalcViewportHeight() {
	taHeight := m.inputHeight
	if taHeight < 1 {
		taHeight = 1
	}
	// 5 fixed rows: blank+top separator (MarginTop=1), blank+bottom separator (MarginTop=1), status bar
	chrome := taHeight + 5
	vpHeight := m.height - chrome
	if vpHeight < 5 {
		vpHeight = 5
	}
	m.viewport.Height = vpHeight
}

// handleSpinnerTick updates spinner during thinking/streaming
func (m *Model) handleSpinnerTick(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	if m.state != StateThinking && m.state != StateStreaming {
		return m, nil
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	m.updateViewport()
	return m, cmd
}

// handleActivityMsg processes streaming activity from copilot
func (m *Model) handleActivityMsg(msg activityMsg) (tea.Model, tea.Cmd) {
	// Append content in order as it arrives
	if msg.reasoning != "" {
		m.activity.AppendContent(ChunkReasoning, msg.reasoning)
	}
	if msg.assistant != "" {
		m.activity.AppendContent(ChunkAssistant, msg.assistant)
	}
	if msg.toolStart != "" {
		m.activity.CurrentTool = ToolInfo{Name: msg.toolStart, Details: msg.toolDetails}
	}
	if msg.toolDone && m.activity.CurrentTool.Name != "" {
		m.activity.AddTool(m.activity.CurrentTool)
		m.activity.CurrentTool = ToolInfo{}
	}
	m.updateViewport()
	return m, nil
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m.handleCtrlC()

	case tea.KeyEscape:
		return m.handleEscape()

	case tea.KeyShiftTab:
		m.config.CycleMode()
		_ = m.config.Save()
		return m, nil

	case tea.KeyCtrlE:
		if m.state == StateReady {
			return m, m.openEditor()
		}
		return m, nil

	// Scroll keys — do NOT forward to textarea; handle here and return early.
	case tea.KeyPgUp, tea.KeyCtrlU:
		m.viewport.HalfPageUp()
		return m, nil

	case tea.KeyPgDown, tea.KeyCtrlD:
		m.viewport.HalfPageDown()
		return m, nil

	case tea.KeyCtrlHome:
		m.viewport.GotoTop()
		return m, nil

	case tea.KeyCtrlEnd:
		m.viewport.GotoBottom()
		return m, nil

	case tea.KeyUp:
		if m.showDropdown && m.dropdown.Visible {
			m.dropdown.MoveUp()
			return m, nil
		}
		if m.input.Value() == "" && len(m.history) > 0 {
			return m.recallLastUserMessage()
		}
		m.viewport.ScrollUp(1)
		return m, nil

	case tea.KeyDown:
		if m.showDropdown && m.dropdown.Visible {
			m.dropdown.MoveDown()
			return m, nil
		}
		m.viewport.ScrollDown(1)
		return m, nil

	case tea.KeyTab:
		if m.showDropdown && m.dropdown.Visible {
			return m.handleDropdownNav()
		}

	case tea.KeyEnter:
		if m.showDropdown && m.dropdown.Visible {
			return m.handleDropdownNav()
		}
		// Backslash at end of current value → insert a real newline instead of submitting
		val := m.input.Value()
		if strings.HasSuffix(val, "\\") {
			m.input.SetValue(strings.TrimSuffix(val, "\\") + "\n")
			m.input.CursorEnd()
			m.recalcViewportHeight()
			return m, nil
		}
		return m.handleSubmit()
	}

	// Forward to textarea (handles typing, backspace, left/right movement, etc.)
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	m.updateDropdown()
	// Textarea may have grown (paste, home/end wrapping) — recalc height.
	m.recalcViewportHeight()

	return m, inputCmd
}

// cancelThinking cancels any in-progress thinking/streaming and returns to ready state
func (m *Model) cancelThinking() {
	if m.cancelThink != nil {
		m.cancelThink()
	}
	m.state = StateReady
}

// handleCtrlC handles double-press to exit or cancel thinking
func (m *Model) handleCtrlC() (tea.Model, tea.Cmd) {
	if m.state == StateThinking || m.state == StateStreaming {
		m.cancelThinking()
		return m, nil
	}

	now := time.Now()
	if m.ctrlCPressed && now.Sub(m.ctrlCTime) < 2*time.Second {
		m.appendToViewport(m.resumeMsg())
		m.quitting = true
		return m, tea.Quit
	}

	m.input.SetValue("")
	m.ctrlCPressed = true
	m.ctrlCTime = now
	return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return ctrlCClearMsg{}
	})
}

// handleEscape cancels thinking or clears dropdown/input
func (m *Model) handleEscape() (tea.Model, tea.Cmd) {
	if m.state == StateThinking || m.state == StateStreaming {
		m.cancelThinking()
		return m, nil
	}
	if m.showDropdown {
		m.showDropdown = false
		m.dropdown.Visible = false
		return m, nil
	}
	m.input.SetValue("")
	return m, nil
}

// handleDropdownNav confirms the currently selected dropdown item.
func (m *Model) handleDropdownNav() (tea.Model, tea.Cmd) {
	if item := m.dropdown.SelectedItem(); item != nil {
		m.input.SetValue("/" + item.Value + " ")
		m.showDropdown = false
		m.dropdown.Visible = false
	}
	return m, nil
}

// recallLastUserMessage puts the last user message into input
func (m *Model) recallLastUserMessage() (tea.Model, tea.Cmd) {
	for i := len(m.history) - 1; i >= 0; i-- {
		if m.history[i].Role == "user" {
			m.input.SetValue(m.history[i].Content)
			m.recalcViewportHeight()
			break
		}
	}
	return m, nil
}

func (m *Model) updateDropdown() {
	value := m.input.Value()

	if strings.HasPrefix(value, "/") && !strings.Contains(value, " ") {
		prefix := strings.TrimPrefix(value, "/")
		matches := m.commands.Match(prefix)

		var items []DropdownItem
		for _, cmd := range matches {
			items = append(items, DropdownItem{
				Label:       cmd.Name,
				Description: cmd.Description,
				Value:       cmd.Name,
			})
		}

		m.dropdown.SetItems(items)
		m.showDropdown = len(items) > 0
	} else {
		m.showDropdown = false
		m.dropdown.Visible = false
	}
}
