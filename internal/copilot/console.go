// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package copilot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/copilot/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	copilotSdk "github.com/github/copilot-sdk/go"
)

// styles used for the non-interactive (single-prompt) output path.
var (
	stylePrefix = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))    // blue
	styleTool   = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("214")) // amber
)

// RunInteractive starts a full Bubbletea TUI session for interactive use.
// progPtr is the double-pointer pre-allocated for the OnUserInputRequest handler;
// it is forwarded to tui.Run so both see the same *tea.Program.
func RunInteractive(session *copilotSdk.Session, model string, progPtr **tea.Program) error {
	return tui.Run(session, model, progPtr)
}

// RunSinglePrompt sends prompt to the agent and streams the response to stdout.
func RunSinglePrompt(session *copilotSdk.Session, prompt string) error {
	if err := streamResponse(session, prompt); err != nil {
		return err
	}
	fmt.Println()
	return nil
}

// streamResponse registers streaming event handlers and blocks until the full
// response is received, printing content directly to stdout as it arrives.
func streamResponse(session *copilotSdk.Session, prompt string) error {
	var (
		mu            sync.Mutex
		prefixPrinted bool
	)

	printPrefix := func() {
		if !prefixPrinted {
			fmt.Print(stylePrefix.Render("azqr") + " ")
			prefixPrinted = true
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	unsubscribe := session.On(func(event copilotSdk.SessionEvent) {
		switch event.Type {
		case "assistant.message_delta":
			if event.Data.DeltaContent != nil {
				mu.Lock()
				printPrefix()
				fmt.Print(*event.Data.DeltaContent)
				mu.Unlock()
			}
		case "tool.execution_start":
			if event.Data.ToolName != nil {
				mu.Lock()
				fmt.Println(styleTool.Render("  ⟳ " + *event.Data.ToolName + "…"))
				prefixPrinted = false
				mu.Unlock()
			}
		}
	})
	defer unsubscribe()

	go func() {
		<-ctx.Done()
		_ = session.Abort(ctx)
	}()

	_, err := session.SendAndWait(ctx, copilotSdk.MessageOptions{Prompt: prompt})
	return err
}
