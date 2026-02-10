// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	copilot "github.com/github/copilot-sdk/go"
)

func (m *Model) handleSubmit() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.input.Value())
	if value == "" {
		return m, nil
	}

	m.input.SetValue("")
	m.recalcViewportHeight()
	m.showDropdown = false
	m.dropdown.Visible = false

	// Handle slash commands
	if strings.HasPrefix(value, "/") {
		result := m.commands.Execute(m, value)
		if result.Exit {
			m.appendToViewport(m.resumeMsg())
			m.quitting = true
			return m, tea.Quit
		}
		if result.ClearChat {
			m.history = nil
			m.conversationTurns = 0
		}
		if result.Error != nil {
			m.setStatus(statusError, result.Error.Error())
		}
		if result.EditorCmd != nil {
			return m, result.EditorCmd
		}
		if result.Output != "" {
			m.appendToViewport(m.styles.ThinkingText.Render("● ") + result.Output + "\n\n")
		}
		if m.pendingPrompt != "" {
			value = m.pendingPrompt
			m.pendingPrompt = ""
		} else {
			return m, nil
		}
	}

	// Print user message into viewport
	m.appendToViewport(m.styles.UserLabel.Render("❯ ") + m.styles.UserText.Render(value) + "\n\n")

	m.history = append(m.history, HistoryEntry{
		Role:    "user",
		Content: value,
	})

	m.state = StateThinking
	m.activity = Activity{}

	// 10 minutes to accommodate long-running tools
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	m.cancelThink = cancel

	return m, tea.Batch(m.sendPrompt(ctx, value), m.spinner.Tick)
}

func (m *Model) sendPrompt(ctx context.Context, prompt string) tea.Cmd {
	return func() (msg tea.Msg) {
		// Recover from any panic so the alt screen is always restored and the
		// error is surfaced in the TUI rather than crashing the process.
		defer func() {
			if r := recover(); r != nil {
				msg = errorMsg{err: fmt.Errorf("internal error: %v", r)}
			}
		}()
		if m.session == nil {
			return errorMsg{err: errors.New("copilot session not initialized")}
		}

		unsubscribe := m.session.On(func(event copilot.SessionEvent) {
			if m.program == nil {
				return
			}
			switch event.Type {
			case "assistant.message_delta":
				if event.Data.DeltaContent != nil {
					m.program.Send(activityMsg{assistant: *event.Data.DeltaContent})
				}
			case "assistant.reasoning_delta":
				if event.Data.DeltaContent != nil {
					m.program.Send(activityMsg{reasoning: *event.Data.DeltaContent})
				}
			case "tool.execution_start":
				if event.Data.ToolName != nil {
					m.program.Send(activityMsg{toolStart: *event.Data.ToolName, toolDetails: formatToolArgs(event.Data.Arguments)})
				}
			case "tool.execution_complete":
				m.program.Send(activityMsg{toolDone: true})
			}
		})
		defer unsubscribe()

		go func() {
			<-ctx.Done()
			_ = m.session.Abort(ctx)
		}()

		if _, err := m.session.SendAndWait(ctx, copilot.MessageOptions{Prompt: prompt}); err != nil {
			return errorMsg{err: err}
		}
		return responseMsg{}
	}
}

func (m *Model) handleResponse() (tea.Model, tea.Cmd) {
	m.state = StateReady
	m.conversationTurns++

	// Build output from all chunks in order
	var outputParts []string
	var assistantContent string
	for _, chunk := range m.activity.Chunks {
		switch chunk.Type {
		case ChunkReasoning:
			outputParts = append(outputParts, m.formatReasoning(chunk.Content))
		case ChunkAssistant:
			outputParts = append(outputParts, m.formatAssistant(chunk.Content))
			assistantContent += chunk.Content
		case ChunkTool:
			outputParts = append(outputParts, m.styles.SuccessText.Render("● ")+WrapText(formatToolDisplay(chunk.Tool), m.width))
		}
	}

	m.activity = Activity{}

	if assistantContent != "" {
		m.history = append(m.history, HistoryEntry{
			Role:    "assistant",
			Content: assistantContent,
		})
	}

	if len(outputParts) > 0 {
		output := strings.Join(outputParts, "\n\n")
		m.appendToViewport(output + "\n\n")
	} else {
		// Clear live activity from viewport even if there was no printable output
		m.updateViewport()
	}

	return m, nil
}

// formatReasoning formats reasoning text for output history
func (m *Model) formatReasoning(text string) string {
	var lines []string
	for _, line := range strings.Split(WrapText(text, m.width), "\n") {
		if line = strings.TrimSpace(line); line == "" {
			continue
		}
		if len(lines) == 0 {
			lines = append(lines, m.styles.ThinkingText.Render("● ")+line)
		} else {
			lines = append(lines, "  "+line)
		}
	}
	return strings.Join(lines, "\n")
}

// formatAssistant formats assistant text for output history
func (m *Model) formatAssistant(text string) string {
	return m.styles.ThinkingText.Render("● ") + WrapText(text, m.width)
}

// setStatus sets a transient status bar message.
func (m *Model) setStatus(level statusLevel, text string) {
	m.statusText = text
	m.statusLevel = level
}

// resumeMsg returns the styled exit message with the current session ID.
func (m *Model) resumeMsg() string {
	cmd := m.styles.WelcomeTitle.Render("azqr copilot --resume \"" + m.session.SessionID + "\"")
	return "Resume any session with " + cmd + "\n"
}

// openEditor suspends the alt screen, opens $EDITOR with a temp file, and
// sets the textarea value to the saved content when the editor exits.
func (m *Model) openEditor() tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	tmp, err := os.CreateTemp("", "azqr-msg-*.md")
	if err != nil {
		m.setStatus(statusError, "could not create temp file: "+err.Error())
		return nil
	}
	// Pre-populate with current textarea content if any
	if v := m.input.Value(); v != "" {
		_, _ = tmp.WriteString(v)
	}
	tmpName := tmp.Name()
	if err := tmp.Close(); err != nil {
		return nil
	}

	c := exec.Command(editor, tmpName) //nolint:gosec
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return tea.ExecProcess(c, func(err error) tea.Msg {
		defer func() { _ = os.Remove(tmpName) }()
		if err != nil {
			return statusMsg{level: statusError, text: "editor exited with error: " + err.Error()}
		}
		content, rerr := os.ReadFile(tmpName) //nolint:gosec
		if rerr != nil {
			return statusMsg{level: statusError, text: "could not read temp file: " + rerr.Error()}
		}
		trimmed := strings.TrimSpace(string(content))
		if trimmed == "" {
			return statusMsg{level: statusWarn, text: "editor returned empty content"}
		}
		return editorDoneMsg{content: trimmed}
	})
}

// editorDoneMsg carries the content written by the external editor.
type editorDoneMsg struct{ content string }

func (m *Model) appendToViewport(content string) {
	m.persisted += content
	m.viewport.SetContent(m.buildViewportContent())
	m.viewport.GotoBottom()
}

// buildViewportContent combines persisted history with any live activity.
func (m *Model) buildViewportContent() string {
	if m.state == StateThinking || m.state == StateStreaming {
		if live := m.renderActivity(); live != "" {
			// Strip trailing newlines from persisted so we control the
			// exact spacing before the live activity (one blank line).
			return strings.TrimRight(m.persisted, "\n") + "\n\n" + live
		}
	}
	return m.persisted
}

// updateViewport refreshes viewport content. Only scrolls to bottom if the user
// is already there — preserves manual scroll position during streaming.
func (m *Model) updateViewport() {
	atBottom := m.viewport.AtBottom()
	m.viewport.SetContent(m.buildViewportContent())
	if atBottom {
		m.viewport.GotoBottom()
	}
}
