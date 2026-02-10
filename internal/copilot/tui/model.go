// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package tui provides a Bubbletea-based interactive terminal UI for the azqr copilot command.
package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	copilotSdk "github.com/github/copilot-sdk/go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ─── modes ────────────────────────────────────────────────────────────────────

type mode string

const (
	modeAgent mode = "agent"
	modeAsk   mode = "ask"
)

func (m mode) next() mode {
	switch m {
	case modeAgent:
		return modeAsk
	default:
		return modeAgent
	}
}

// ─── tea messages ─────────────────────────────────────────────────────────────

// streamDeltaMsg carries incremental text from the AI response stream.
type streamDeltaMsg struct{ content string }

// streamReasoningMsg carries incremental reasoning/thinking text.
type streamReasoningMsg struct{ content string }

// streamToolMsg signals a tool execution starting.
type streamToolMsg struct{ name string }

// streamDoneMsg signals the stream finished (with optional error).
type streamDoneMsg struct{ err error }

// ctrlCClearMsg is sent 2 seconds after the first Ctrl+C press to clear the warning.
type ctrlCClearMsg struct{}

// userInputRequestMsg is sent when the agent invokes ask_user.
// replyCh must receive exactly one string (the user’s answer) to unblock the SDK handler.
type userInputRequestMsg struct {
	question string
	choices  []string
	replyCh  chan string
}

// ─── model ────────────────────────────────────────────────────────────────────

// Model is the Bubbletea application model for the azqr copilot TUI.
type Model struct {
	// layout
	width, height int

	// sub-components
	input textarea.Model
	spin  spinner.Model
	keys  keyMap

	// session
	session *copilotSdk.Session
	model   string

	// state
	currentMode   mode
	liveContent   string   // current streaming/reasoning content shown in View()
	history       []string // prompt history (oldest → newest)
	historyIdx    int      // index into history; -1 = not browsing
	streaming     bool
	currentTool   string
	reasoningBuf  *strings.Builder // accumulates the current reasoning/thinking text
	streamBuf     *strings.Builder // accumulates the in-progress response
	lastInterrupt time.Time
	cwd           string // current working directory shown in header

	// userInputReplyCh is non-nil while the SDK is waiting for the user to
	// answer an ask_user question.  Enter sends the response and clears it.
	userInputReplyCh chan string

	// progPtr is a double-pointer shared between the local Model value in
	// Run() and bubbletea's internal value-copy.  tea.NewProgram copies the
	// Model by value, so a plain *tea.Program field would remain nil in that
	// copy; the extra level of indirection makes the assignment visible.
	progPtr **tea.Program
}

// New creates a new Model connected to the given Copilot session.
// Prefer newWithProgPtr when the double-pointer is managed externally.
func New(session *copilotSdk.Session, model string) Model {
	return newWithProgPtr(session, model, new(*tea.Program))
}

func newWithProgPtr(session *copilotSdk.Session, model string, pp **tea.Program) Model {
	ta := textarea.New()
	ta.Placeholder = "Ask anything about your Azure environment…"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.Prompt = ""
	ta.SetHeight(1)
	ta.SetWidth(76)
	// Enter submits; Alt+Enter / Shift+Enter inserts a newline
	ta.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("alt+enter", "shift+enter"))
	// Remove textarea's default borders — we control the look ourselves
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	_ = ta.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = styleStatusBar

	cwd, _ := os.Getwd()
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(cwd, home) {
		cwd = "~" + cwd[len(home):]
	}

	return Model{
		session:      session,
		model:        model,
		input:        ta,
		spin:         sp,
		keys:         defaultKeys(),
		currentMode:  modeAgent,
		historyIdx:   -1,
		cwd:          cwd,
		progPtr:      pp,
		streamBuf:    new(strings.Builder),
		reasoningBuf: new(strings.Builder),
	}
}

// BuildUserInputHandler returns a UserInputHandler that forwards ask_user
// requests from the agent to the running Bubbletea TUI and blocks until the
// user responds.  pp is the same double-pointer used by Run(); *pp is
// guaranteed non-nil by the time the handler fires.
func BuildUserInputHandler(pp **tea.Program) copilotSdk.UserInputHandler {
	return func(req copilotSdk.UserInputRequest, _ copilotSdk.UserInputInvocation) (copilotSdk.UserInputResponse, error) {
		prog := *pp
		if prog == nil {
			return copilotSdk.UserInputResponse{}, nil
		}
		replyCh := make(chan string, 1)
		prog.Send(userInputRequestMsg{
			question: req.Question,
			choices:  req.Choices,
			replyCh:  replyCh,
		})
		answer := <-replyCh
		return copilotSdk.UserInputResponse{Answer: answer, WasFreeform: true}, nil
	}
}

// Run starts the Bubbletea program and blocks until the user exits.
// Completed messages are printed above the TUI via tea.Println so they persist
// in the terminal's native scrollback buffer — enabling mouse wheel scroll and
// native text selection without any special terminal modes.
// An optional progPtr may be provided when the caller has already allocated it
// (e.g. to build a BuildUserInputHandler); if nil a new one is created internally.
func Run(session *copilotSdk.Session, model string, externalProgPtr ...**tea.Program) error {
	var pp **tea.Program
	if len(externalProgPtr) > 0 && externalProgPtr[0] != nil {
		pp = externalProgPtr[0]
	} else {
		pp = new(*tea.Program)
	}
	m := newWithProgPtr(session, model, pp)
	// Do NOT use WithAltScreen — that tears down the screen on exit and
	// erases everything the user saw.
	// Do NOT use WithMouseCellMotion — that intercepts mouse events and
	// prevents the terminal from doing native text selection / copy.
	p := tea.NewProgram(m)
	// Write the program pointer through the shared double-pointer so that
	// bubbletea's value-copy of the model can reach it from sendPrompt's
	// goroutine.  (tea.NewProgram copies the model by value, so a plain
	// *tea.Program field would still be nil in the copy.)
	*pp = p

	// Redirect zerolog through the TUI so that log lines (e.g. scan progress)
	// appear above the input chrome instead of bleeding through to the terminal.
	lw := &tuiLogWriter{program: p}
	prevLogger := log.Logger
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        lw,
		TimeFormat: "15:04:05",
		NoColor:    true,
	}).With().Timestamp().Logger()
	defer func() { log.Logger = prevLogger }()

	_, err := p.Run()
	return err
}

// ─── Init ─────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	// Print the welcome banner permanently above the TUI chrome before the
	// first render.  tea.Println queues a printLineMessage that the renderer
	// flushes to the terminal scrollback before drawing View().
	return tea.Batch(m.spin.Tick, tea.Println(bannerContent()))
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	// skipInput is set by the keyboard handler when a key has been fully
	// consumed and must not be forwarded to the textarea.
	var skipInput bool

	switch msg := msg.(type) {

	// ── window resize ──────────────────────────────────────────────────────
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// input width: full width minus the " › " prefix (3 chars)
		m.input.SetWidth(m.width - 3)
		// Recompute liveContent at the new width so word-wrap reflects the
		// current terminal width.
		if m.streaming {
			m.liveContent = m.computeLiveContent()
		}

	// ── streaming events ───────────────────────────────────────────────────
	case streamDeltaMsg:
		// When the first delta arrives, reasoning phase is over — discard it.
		if m.reasoningBuf.Len() > 0 {
			m.reasoningBuf.Reset()
		}
		m.streamBuf.WriteString(msg.content)
		m.liveContent = m.computeLiveContent()
		cmds = append(cmds, m.spin.Tick)

	case streamReasoningMsg:
		m.reasoningBuf.WriteString(msg.content)
		m.liveContent = m.computeLiveContent()
		cmds = append(cmds, m.spin.Tick)

	case streamToolMsg:
		m.currentTool = msg.name
		// Commit any in-progress LLM text before the tool entry.
		if m.streamBuf.Len() > 0 {
			content := strings.TrimRight(m.streamBuf.String(), " \n")
			m.streamBuf.Reset()
			m.reasoningBuf.Reset()
			m.liveContent = ""
			if content != "" {
				cmds = append(cmds, m.printAssistantMsg(content))
			}
		}
		toolLine := " " + styleToolName.Render("  ⟳ "+msg.name+"…")
		cmds = append(cmds, tea.Println(toolLine), m.spin.Tick)

	case logLineMsg:
		cmds = append(cmds, tea.Println(" "+styleLogMsg.Render(msg.line)))

	case streamDoneMsg:
		m.streaming = false
		m.currentTool = ""
		m.reasoningBuf.Reset()
		content := strings.TrimRight(m.streamBuf.String(), " \n")
		m.streamBuf.Reset()
		m.liveContent = ""
		if content != "" {
			cmds = append(cmds, m.printAssistantMsg(content))
		}
		if msg.err != nil {
			cmds = append(cmds, tea.Println(" "+styleError.Render("Error: "+msg.err.Error())))
		}

	// ── spinner tick ───────────────────────────────────────────────────────
	case spinner.TickMsg:
		if m.streaming {
			var cmd tea.Cmd
			m.spin, cmd = m.spin.Update(msg)
			cmds = append(cmds, cmd)
			// Spinner frame lives in renderStatus() inside View(), so we do NOT
			// call refreshViewport() here — that would rebuild all messages O(n)
			// on every animation frame. View() is called by bubbletea after every
			// Update, so the spinner frame is still updated correctly.
		}

	// ── user input request (ask_user) ─────────────────────────────────────
	case userInputRequestMsg:
		m.userInputReplyCh = msg.replyCh
		m.input.Placeholder = "Type your answer and press Enter…"
		prompt := styleHint.Render("Agent asks: ") + styleUserMsg.Render(msg.question)
		if len(msg.choices) > 0 {
			prompt += "\n" + styleHint.Render("Choices: "+strings.Join(msg.choices, " / "))
		}
		cmds = append(cmds, tea.Println(" "+prompt))

	// ── ctrl+c warning expiry ────────────────────────────────────────────
	case ctrlCClearMsg:
		m.lastInterrupt = time.Time{}

	// ── keyboard ───────────────────────────────────────────────────────────
	case tea.KeyMsg:
		switch {

		// Exit: Ctrl+D
		case msg.Type == tea.KeyCtrlD:
			return m, tea.Quit

		// Exit on second Ctrl+C within 2s; first press clears input or aborts stream
		case msg.Type == tea.KeyCtrlC:
			if m.streaming {
				// abort the running stream; commit any partial content first
				m.streaming = false
				m.currentTool = ""
				if content := strings.TrimRight(m.streamBuf.String(), " \n"); content != "" {
					cmds = append(cmds, m.printAssistantMsg(content))
				}
				m.streamBuf.Reset()
				m.reasoningBuf.Reset()
				m.liveContent = ""
				cmds = append(cmds, m.abortCmd())
				break
			}
			if !m.lastInterrupt.IsZero() && time.Since(m.lastInterrupt) < 2*time.Second {
				return m, tea.Quit
			}
			m.lastInterrupt = time.Now()
			if m.input.Value() != "" {
				m.input.SetValue("")
			}
			cmds = append(cmds, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return ctrlCClearMsg{} }))

		// Clear screen: Ctrl+L — print a visual separator into the scrollback
		case msg.Type == tea.KeyCtrlL:
			cmds = append(cmds, tea.Println(m.renderSeparator()))

		// Mode cycle: Shift+Tab
		case key.Matches(msg, m.keys.ModeSwitch):
			m.currentMode = m.currentMode.next()

		// Send message: Enter
		case msg.Type == tea.KeyEnter:
			if m.streaming {
				break // ignore enter while streaming
			}
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				break
			}
			m.input.Reset()
			m.historyIdx = -1
			m.lastInterrupt = time.Time{}

			// handle slash commands
			action, sysMsg := parseSlash(text)
			// If waiting for user-input answer, bypass slash commands and send directly.
			if m.userInputReplyCh != nil {
				ch := m.userInputReplyCh
				m.userInputReplyCh = nil
				m.input.Placeholder = "Ask anything about your Azure environment…"
				userLine := fmt.Sprintf(" %s › %s", styleUserLabel.Render("You"), styleUserMsg.Render(wordWrap(text, m.width-6)))
				cmds = append(cmds, tea.Println(userLine))
				// Send answer back to the blocking SDK handler.
				ch <- text
				break
			}
			switch action {
			case slashHelp:
				cmds = append(cmds, tea.Println(helpContent()))
			case slashClear:
				cmds = append(cmds, tea.Println(m.renderSeparator()))
			case slashExit:
				return m, tea.Quit
			case slashModel:
				cmds = append(cmds, tea.Println(fmt.Sprintf(" Current model: %s  |  mode: %s", m.model, m.currentMode)))
			case slashNew:
				cmds = append(cmds, tea.Println(" "+styleHint.Render("Starting a new conversation…")))
			case slashUnknown:
				cmds = append(cmds, tea.Println(" "+styleError.Render(sysMsg)))
			case slashNone:
				// normal message — add to history, print permanently, and stream
				m.history = append(m.history, text)
				userLine := fmt.Sprintf(" %s › %s", styleUserLabel.Render("You"), styleUserMsg.Render(wordWrap(text, m.width-6)))
				m.streamBuf.Reset()
				m.streaming = true
				m.liveContent = ""
				m.input.Reset()
				cmds = append(cmds, tea.Println(userLine), m.spin.Tick, m.sendPrompt(text))
			}

		// History: ↑/↓ navigation with context-sensitive behaviour.
		//
		//  • Multi-line textarea (shift+enter used): key moves cursor inside
		//    the textarea.
		//  • Single-line, input has text OR already mid-navigation: history
		//    recall replaces the input.
		//  • Single-line, input is empty, not mid-navigation: absorbed (no
		//    viewport to scroll in the new architecture).
		case key.Matches(msg, m.keys.HistoryUp):
			if m.input.LineCount() > 1 {
				// multi-line: let textarea handle cursor movement
			} else if m.historyIdx != -1 || m.input.Value() != "" {
				if len(m.history) == 0 {
					break
				}
				if m.historyIdx == -1 {
					m.historyIdx = len(m.history) - 1
				} else if m.historyIdx > 0 {
					m.historyIdx--
				}
				m.input.SetValue(m.history[m.historyIdx])
				m.input.CursorEnd()
				skipInput = true
			} else {
				// empty input, not browsing → absorb (nothing to scroll)
				skipInput = true
			}

		case key.Matches(msg, m.keys.HistoryDown):
			if m.input.LineCount() > 1 {
				// multi-line: let textarea handle cursor movement
			} else if m.historyIdx != -1 {
				if m.historyIdx < len(m.history)-1 {
					m.historyIdx++
					m.input.SetValue(m.history[m.historyIdx])
					m.input.CursorEnd()
				} else {
					m.historyIdx = -1
					m.input.Reset()
				}
				skipInput = true
			} else {
				// not browsing → absorb
				skipInput = true
			}

		// PgUp / PgDown / End: no viewport to scroll; absorb to prevent
		// textarea from receiving them.
		case msg.Type == tea.KeyPgUp, msg.Type == tea.KeyPgDown, msg.Type == tea.KeyEnd:
			skipInput = true

		default:
			m.lastInterrupt = time.Time{}
		}
	}

	// Propagate events to the textarea sub-component; skip Enter so it doesn't
	// insert a newline when the user submits.
	var inputCmd tea.Cmd
	if !skipInput {
		if keyMsg, ok := msg.(tea.KeyMsg); !ok || keyMsg.Type != tea.KeyEnter {
			m.input, inputCmd = m.input.Update(msg)
			// dynamically grow/shrink textarea with content
			lines := m.input.LineCount()
			if lines < 1 {
				lines = 1
			}
			if lines > 5 {
				lines = 5
			}
			if m.input.Height() != lines {
				m.input.SetHeight(lines)
			}
		}
	}
	cmds = append(cmds, inputCmd)

	return m, tea.Batch(cmds...)
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	parts := []string{}
	if m.liveContent != "" {
		parts = append(parts, m.liveContent)
	}
	parts = append(parts,
		m.renderCwd(),
		m.renderSeparator(),
		m.renderInput(),
		m.renderSeparator(),
		m.renderFooter(),
	)
	return strings.Join(parts, "\n")
}

// ─── render helpers ───────────────────────────────────────────────────────────

func (m Model) renderCwd() string {
	left := " " + styleHeaderCwd.Render(m.cwd)
	right := styleHeaderModel.Render(m.model + " ")
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) renderSeparator() string {
	return styleSeparator.Render(strings.Repeat("─", m.width))
}

func (m Model) renderInput() string {
	prompt := styleInputPrompt.Render("›")
	// Split the textarea view on newlines so that continuation lines (produced
	// by shift+enter for multi-line input) are indented to align with the first
	// line.  The prefix " › " is 3 characters wide; continuation lines get 3
	// spaces so the text column lines up.
	lines := strings.Split(m.input.View(), "\n")
	for i, line := range lines {
		if i == 0 {
			lines[i] = " " + prompt + " " + line
		} else {
			lines[i] = "    " + line // 4 chars = " " + "›" + " " + " " alignment
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderFooter() string {
	// ── left: Ctrl+C warning  OR  mode + hint ──────────────────────────────
	var left string
	if !m.lastInterrupt.IsZero() && time.Since(m.lastInterrupt) < 2*time.Second {
		left = styleFooterWarn.Render("Press ctrl+c again to exit")
	} else {
		left = styleFooterMode.Render(string(m.currentMode)) +
			styleFooterBar.Render("  shift+tab: switch mode")
	}

	// ── right: status ──────────────────────────────────────────────────────
	var statusPart string
	if m.streaming {
		if m.currentTool != "" {
			statusPart = m.spin.View() + " " + styleStatusBar.Render(m.currentTool+"…") + "  "
		} else {
			statusPart = m.spin.View() + " " + styleStatusBar.Render("thinking…") + "  "
		}
	} else if len(m.history) > 0 {
		statusPart = styleToolStatus.Render("● ready") + "  "
	}
	right := statusPart

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// ─── streaming ────────────────────────────────────────────────────────────────

// abortCmd returns a Cmd that calls session.Abort() to interrupt an in-flight
// LLM request.  The actual stream-done event arrives through the normal
// streamDoneMsg path, so this Cmd returns nil.
func (m *Model) abortCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = m.session.Abort(ctx)
		return nil
	}
}

// sendPrompt dispatches the user prompt to the Copilot session and returns a Cmd
// that feeds streaming events back into the Bubbletea program as messages.
func (m *Model) sendPrompt(prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// Capture the program pointer once — m.progPtr is a shared double-
		// pointer set by Run() after tea.NewProgram, so *m.progPtr is valid.
		prog := *m.progPtr

		// Send streamDoneMsg through prog.Send (same queue as tool/delta
		// messages) rather than returning it as the goroutine's return value.
		// This guarantees streamDoneMsg is processed AFTER all preceding
		// tool/delta messages, preventing the "ready" status from appearing
		// before the last tool event is rendered.
		doneCh := make(chan error, 1)
		unsubscribe := m.session.On(func(event copilotSdk.SessionEvent) {
			switch event.Type {
			case "assistant.message_delta":
				if event.Data.DeltaContent != nil && prog != nil {
					prog.Send(streamDeltaMsg{content: *event.Data.DeltaContent})
				}
			case "assistant.reasoning_delta":
				if event.Data.ReasoningText != nil && prog != nil {
					prog.Send(streamReasoningMsg{content: *event.Data.ReasoningText})
				}
			case "tool.execution_start":
				if event.Data.ToolName != nil && prog != nil {
					prog.Send(streamToolMsg{name: *event.Data.ToolName})
				}
			case "session.idle":
				// Enqueue done signal through the same channel so it lands
				// strictly after any tool/delta messages already in flight.
				if prog != nil {
					prog.Send(streamDoneMsg{})
				}
				select {
				case doneCh <- nil:
				default:
				}
			case "session.error":
				var errMsg string
				if event.Data.Message != nil {
					errMsg = *event.Data.Message
				} else {
					errMsg = "session error"
				}
				err := fmt.Errorf("%s", errMsg)
				if prog != nil {
					prog.Send(streamDoneMsg{err: err})
				}
				select {
				case doneCh <- err:
				default:
				}
			}
		})
		defer unsubscribe()

		if _, err := m.session.Send(ctx, copilotSdk.MessageOptions{Prompt: prompt}); err != nil {
			return streamDoneMsg{err: err}
		}
		// Wait for idle/error signal from the On handler (or context timeout).
		select {
		case <-doneCh:
		case <-ctx.Done():
			if prog != nil {
				prog.Send(streamDoneMsg{err: ctx.Err()})
			}
		}
		return nil
	}
}

// computeLiveContent builds the string shown in the live area of View() for
// in-progress streaming/reasoning.  It is bounded to the last N lines so the
// View() region never grows taller than the terminal.
func (m Model) computeLiveContent() string {
	var parts []string

	// Reasoning block (shown while the model is "thinking", before text flows)
	if r := m.reasoningBuf.String(); r != "" {
		label := " " + styleReasoningLabel.Render("  thinking") + styleFooterBar.Render(" ›")
		body := " " + styleReasoningMsg.Render(wordWrap(r, m.width-4))
		parts = append(parts, label+"\n"+body)
	}

	// Streaming response text
	if s := strings.TrimRight(m.streamBuf.String(), "\n"); s != "" {
		label := " " + styleAILabel.Render("azqr") + styleFooterBar.Render(" ›")
		body := " " + styleAIMsg.Render(wordWrap(s, m.width-4))
		parts = append(parts, label+"\n"+body)
	}

	if len(parts) == 0 {
		return ""
	}

	full := strings.Join(parts, "\n")

	// chrome = cwd(1) + sep(1) + input(inputH) + sep(1) + footer(1) = 4 + inputH
	inputH := m.input.Height()
	maxLines := m.height - 4 - inputH
	if maxLines < 3 {
		maxLines = 3
	}
	lines := strings.Split(full, "\n")
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return strings.Join(lines, "\n")
}

// printAssistantMsg returns a Cmd that prints a finalized assistant response
// permanently above the TUI chrome via tea.Println.
func (m Model) printAssistantMsg(content string) tea.Cmd {
	label := " " + styleAILabel.Render("azqr") + styleFooterBar.Render(" ›")
	return tea.Println(label + "\n" + glamourRender(content, m.width))
}

// ─── banner ───────────────────────────────────────────────────────────────────

func bannerContent() string {
	inner := strings.Join([]string{
		styleBannerTitle.Render("azqr copilot"),
		styleBannerSubtitle.Render("Describe a task to get started."),
		"",
		styleBannerHint.Render("Ask anything about your Azure resources, compliance, security,"),
		styleBannerHint.Render("or best practices. Type /help for available commands."),
	}, "\n")
	return "\n" + lipgloss.NewStyle().
		MarginLeft(2).
		Render(styleBannerBox.Render(inner)) + "\n"
}
