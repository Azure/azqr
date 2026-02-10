// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

// Version is set at build time
var Version = "dev"

// WelcomeBanner generates the styled welcome message
type WelcomeBanner struct {
	styles *Styles
}

// NewWelcomeBanner creates a new welcome banner
func NewWelcomeBanner(styles *Styles) *WelcomeBanner {
	return &WelcomeBanner{styles: styles}
}

// Render returns the full welcome banner
func (wb *WelcomeBanner) Render() string {
	title := wb.styles.WelcomeTitle.Render("Azure Quick Review") + " " +
		wb.styles.WelcomeVersion.Render("v"+Version)
	content := title + "\n" +
		"Describe a task to get started.\n\n" +
		"Pick a model with " + wb.styles.CommandName.Render("/model") +
		". Copilot uses AI, so always check for mistakes."
	return wb.styles.WelcomeBox.Width(75).Render(content)
}
