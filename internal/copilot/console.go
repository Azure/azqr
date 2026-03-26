// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package copilot

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/chzyer/readline"
	copilotSdk "github.com/github/copilot-sdk/go"
)

// styles used throughout the interactive console.
var (
	styleHeader  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))    // bright blue
	styleModel   = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("245"))  // grey
	styleDivider = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("238"))  // dark grey
	stylePrompt  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("76"))    // green
	stylePrefix  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))    // blue
	styleTool    = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("214")) // amber
	styleError   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))   // red
	styleHint    = lipgloss.NewStyle().Faint(true)
)

// RunInteractive starts an interactive console session with styled output.
//   - Ctrl+C once  → clears the current line; hints to press again to exit
//   - Ctrl+C twice → exits (second press within 2 seconds)
func RunInteractive(session *copilotSdk.Session, model string) error {
	printHeader(model)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          stylePrompt.Render("›") + " ",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return err
	}
	defer func() { _ = rl.Close() }()

	var lastInterrupt time.Time

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			// Ctrl+C: readline already cleared the line.
			// Second press within 2 seconds exits.
			if !lastInterrupt.IsZero() && time.Since(lastInterrupt) < 2*time.Second {
				fmt.Println(styleHint.Render("Goodbye."))
				return nil
			}
			lastInterrupt = time.Now()
			fmt.Println(styleHint.Render("Press Ctrl+C again to exit."))
			continue
		}
		if err == io.EOF {
			fmt.Println()
			return nil
		}
		if err != nil {
			return err
		}

		// Any successful input resets the interrupt timer.
		lastInterrupt = time.Time{}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			fmt.Println(styleHint.Render("Goodbye."))
			return nil
		}

		fmt.Println()
		if err := streamResponse(session, line); err != nil {
			fmt.Fprintln(os.Stderr, styleError.Render("error: "+err.Error()))
		}
		fmt.Println()
		fmt.Println()
	}
}

// RunSinglePrompt sends prompt to the agent and streams the response to stdout.
func RunSinglePrompt(session *copilotSdk.Session, prompt string) error {
	if err := streamResponse(session, prompt); err != nil {
		return err
	}
	fmt.Println()
	return nil
}

// printHeader renders the welcome banner.
func printHeader(model string) {
	title := styleHeader.Render("azqr copilot")
	modelLabel := styleModel.Render("model: " + model)
	hint := styleHint.Render("Ctrl+C twice to exit")
	divider := styleDivider.Render(strings.Repeat("─", 50))

	fmt.Println()
	fmt.Println(" " + title + "  " + modelLabel)
	fmt.Println(" " + hint)
	fmt.Println(" " + divider)
	fmt.Println()
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
