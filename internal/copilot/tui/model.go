// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"context"
	"time"

	"github.com/Azure/azqr/internal/copilot/config"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	copilot "github.com/github/copilot-sdk/go"
)

// Model is the main Bubble Tea model for the copilot TUI
type Model struct {
	// Dimensions
	width int

	// Core components
	input    textarea.Model
	spinner  spinner.Model
	markdown *glamour.TermRenderer

	// UI components
	dropdown  *Dropdown
	statusBar *StatusBar
	welcome   *WelcomeBanner
	commands  *CommandRegistry
	styles    *Styles
	keyMap    KeyMap

	// Configuration
	config *config.Config

	// Conversation state
	history           []HistoryEntry
	conversationTurns int

	// UI state
	state        State
	showDropdown bool
	infoMessages []string
	inputHeight  int // current textarea display height in rows

	// Transient status bar message (replaces hint text until TTL expires)
	statusText  string
	statusLevel statusLevel

	// Viewport content
	persisted string // finalized conversation content (also written to stdout on exit)

	// Copilot integration
	client        *copilot.Client
	session       *copilot.Session
	program       *tea.Program
	cancelThink   context.CancelFunc
	pendingPrompt string
	activity      Activity

	// Exit handling
	ctrlCPressed bool
	ctrlCTime    time.Time
	quitting     bool
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config, client *copilot.Client, session *copilot.Session, infoMessages []string) *Model {
	styles := DefaultStyles()
	return &Model{
		input:        newTextArea(styles),
		spinner:      newSpinner(),
		inputHeight:  1,
		markdown:     newMarkdownRenderer(80),
		dropdown:     NewDropdown(styles),
		statusBar:    NewStatusBar(styles, 80),
		welcome:      NewWelcomeBanner(styles),
		commands:     NewCommandRegistry(),
		styles:       styles,
		keyMap:       DefaultKeyMap(),
		config:       cfg,
		width:        80,
		infoMessages: infoMessages,
		client:       client,
		session:      session,
	}
}

func newTextArea(styles *Styles) textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "Type / for commands"
	ta.CharLimit = -1 // unlimited
	ta.ShowLineNumbers = false
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = styles.InputStyle
	ta.BlurredStyle.Base = styles.InputStyle
	ta.FocusedStyle.Placeholder = styles.PlaceholderStyle
	ta.BlurredStyle.Placeholder = styles.PlaceholderStyle
	ta.Prompt = "❯ "
	ta.Focus()
	return ta
}

func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{"●", " "},
		FPS:    500 * time.Millisecond,
	}
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#d946ef"))
	return s
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	m.persisted = ""
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

// newMarkdownRenderer creates a glamour renderer matching the terminal background.
func newMarkdownRenderer(width int) *glamour.TermRenderer {
	style := "dark"
	if !lipgloss.HasDarkBackground() {
		style = "light"
	}
	md, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	return md
}
