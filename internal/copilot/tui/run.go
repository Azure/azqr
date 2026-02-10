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
	m := NewModel(cfg, client, session, infoMessages)
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.program = p

	// Ensure alt screen is restored on SIGTERM / SIGHUP so the terminal is
	// not left in a broken state when the process is killed externally.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigs
		p.Kill()
	}()
	defer signal.Stop(sigs)

	finalModel, err := p.Run()

	// After altscreen exits, dump the full session transcript to stdout
	if fm, ok := finalModel.(*Model); ok && fm.persisted != "" {
		fmt.Print(fm.persisted)
	}

	return err
}
