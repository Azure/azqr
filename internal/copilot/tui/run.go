// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Azure/azqr/internal/copilot/config"
	tea "github.com/charmbracelet/bubbletea"
	copilot "github.com/github/copilot-sdk/go"
)

// Run starts the TUI application
func Run(cfg *config.Config, client *copilot.Client, session *copilot.Session, infoMessages []string) error {
	// Print welcome and info messages directly to terminal (no alt screen)
	styles := DefaultStyles()
	welcome := NewWelcomeBanner(styles)
	fmt.Println(welcome.Render())
	for _, msg := range infoMessages {
		fmt.Println(styles.InfoDot.Render("●") + " " + styles.InfoText.Render(msg))
	}

	m := NewModel(cfg, client, session, infoMessages)
	p := tea.NewProgram(m)
	m.program = p

	// Handle termination signals gracefully
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigs
		p.Kill()
	}()
	defer signal.Stop(sigs)

	finalModel, err := p.Run()

	// Dump the full session transcript to stdout on exit
	if fm, ok := finalModel.(*Model); ok && fm.persisted != "" {
		fmt.Print(fm.persisted)
	}

	return err
}
