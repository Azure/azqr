// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package tui provides the terminal user interface for the azqr copilot.
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette matching GitHub Copilot CLI style
var (
	// Primary colors
	colorMagenta = lipgloss.Color("#d946ef") // Bright magenta for borders/accents
	colorCyan    = lipgloss.Color("#22d3ee") // Cyan for commands and highlights
	colorGreen   = lipgloss.Color("#22c55e") // Green for success
	colorYellow  = lipgloss.Color("#eab308") // Yellow for warnings
	colorRed     = lipgloss.Color("#ef4444") // Red for errors
	colorGray    = lipgloss.Color("#6b7280") // Gray for muted text
	colorWhite   = lipgloss.Color("#f9fafb") // White for regular text
	colorDim     = lipgloss.Color("#4b5563") // Dim gray for hints
)

// Styles contains all the lipgloss styles for the TUI
type Styles struct {
	// Welcome box styles
	WelcomeBox     lipgloss.Style
	WelcomeTitle   lipgloss.Style
	WelcomeVersion lipgloss.Style

	// Prompt styles
	PromptIcon       lipgloss.Style
	InputStyle       lipgloss.Style
	PlaceholderStyle lipgloss.Style

	// Command styles
	CommandSlash lipgloss.Style
	CommandName  lipgloss.Style
	CommandDesc  lipgloss.Style

	// Status bar styles
	StatusModel lipgloss.Style
	StatusMode  lipgloss.Style
	StatusHint  lipgloss.Style

	// Message styles
	UserLabel    lipgloss.Style
	UserText     lipgloss.Style
	ErrorText    lipgloss.Style
	WarningText  lipgloss.Style
	SuccessText  lipgloss.Style
	ThinkingText lipgloss.Style

	// Mode indicator styles
	ModeAsk   lipgloss.Style
	ModePlan  lipgloss.Style
	ModeAgent lipgloss.Style

	// Hint styles
	HintText lipgloss.Style

	// Info message styles
	InfoDot  lipgloss.Style
	InfoText lipgloss.Style

	// Input area styles
	InputSeparator lipgloss.Style
}

// DefaultStyles returns the default styled configuration
func DefaultStyles() *Styles {
	return &Styles{
		// Welcome box - magenta border like Copilot
		WelcomeBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMagenta).
			Padding(0, 1).
			MarginBottom(1),

		WelcomeTitle: lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true),

		WelcomeVersion: lipgloss.NewStyle().
			Foreground(colorGray),

		// Prompt - cyan arrow like Copilot
		PromptIcon: lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true),

		InputStyle: lipgloss.NewStyle().
			Foreground(colorWhite),

		PlaceholderStyle: lipgloss.NewStyle().
			Foreground(colorGray),

		// Commands
		CommandSlash: lipgloss.NewStyle().
			Foreground(colorCyan),

		CommandName: lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true),

		CommandDesc: lipgloss.NewStyle().
			Foreground(colorGray),

		// Status bar
		StatusModel: lipgloss.NewStyle().
			Foreground(colorCyan),

		StatusMode: lipgloss.NewStyle().
			Foreground(colorMagenta).
			Bold(true),

		StatusHint: lipgloss.NewStyle().
			Foreground(colorDim),

		// Messages
		UserLabel: lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true),

		UserText: lipgloss.NewStyle().
			Foreground(colorWhite),

		ErrorText: lipgloss.NewStyle().
			Foreground(colorRed),

		WarningText: lipgloss.NewStyle().
			Foreground(colorYellow),

		SuccessText: lipgloss.NewStyle().
			Foreground(colorGreen),

		ThinkingText: lipgloss.NewStyle().
			Foreground(colorMagenta),

		// Mode indicators
		ModeAsk: lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true),

		ModePlan: lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true),

		ModeAgent: lipgloss.NewStyle().
			Foreground(colorMagenta).
			Bold(true),

		// Hints
		HintText: lipgloss.NewStyle().
			Foreground(colorDim),

		// Info messages (blue dot + gray text)
		InfoDot: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")),

		InfoText: lipgloss.NewStyle().
			Foreground(colorGray),

		// Input area
		InputSeparator: lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1),
	}
}

// Emoji constants for consistent usage
const (
	EmojiCheck = "✓"
	EmojiCross = "❌"
)

// FormatError formats an error message
func (s *Styles) FormatError(msg string) string {
	return s.ErrorText.Render(EmojiCross + " " + msg)
}

// FormatSuccess formats a success message
func (s *Styles) FormatSuccess(msg string) string {
	return s.SuccessText.Render(EmojiCheck + " " + msg)
}

// InputSeparatorLine returns a horizontal separator line for the input area
func (s *Styles) InputSeparatorLine(width int) string {
	if width <= 0 {
		width = 60
	}
	line := strings.Repeat("─", width)
	return s.InputSeparator.Render(line)
}

// ModeIndicator returns the styled mode indicator
func (s *Styles) ModeIndicator(mode string) string {
	switch mode {
	case "plan":
		return s.ModePlan.Render("Plan")
	case "agent":
		return s.ModeAgent.Render("Agent")
	default:
		return s.ModeAsk.Render("Ask")
	}
}
