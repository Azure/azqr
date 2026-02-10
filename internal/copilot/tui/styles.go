// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Header bar
	styleHeaderCwd   = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	styleHeaderModel = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// Separator bar
	styleSeparator = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("33"))

	// Footer bar
	styleFooterBar  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleFooterMode = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	styleFooterWarn = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// Messages
	styleUserLabel = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("76"))
	styleUserMsg   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	styleAILabel   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	styleAIMsg     = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	// Tool execution
	styleToolName   = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("33"))
	styleToolStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))

	// Status bar
	styleStatusBar = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))

	// Input area
	styleInputPrompt = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("76"))

	// Misc
	styleError  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	styleHint   = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("245"))
	styleLogMsg = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("238"))

	// Reasoning (thinking)
	styleReasoningLabel = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("39"))
	styleReasoningMsg   = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("252"))

	// Help overlay
	styleHelpKey   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	styleHelpDesc  = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	styleHelpTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Underline(true)

	// Welcome banner
	styleBannerBox      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("33")).Padding(1, 2)
	styleBannerTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	styleBannerSubtitle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	styleBannerHint     = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("245"))
)
