---
name: tui-developer
description: Generate Go Bubble Tea TUI applications with chat-like interfaces. Use when the user asks to create terminal UIs, chat interfaces, interactive CLI, REPLs, or log viewers. Follow the specified architecture and implementation requirements for consistent results.
---

# TUI Developer Skill

Generate minimal, idiomatic Go Bubble Tea TUI applications with chat-like interfaces using **alt screen + internal viewport**.

## When to Use

Use this skill when the user asks to:
- Create a terminal UI with Bubble Tea
- Build a chat-like CLI interface
- Implement a REPL or interactive terminal
- Create a log viewer with input

## Architecture

The TUI uses **alt screen with an internal viewport** for scrollable conversation history. Completed output is accumulated in a `persisted` string that is written to stdout after the alt screen exits, giving the user a transcript in their terminal scrollback.

```
┌───────────────────────────────┐
│  viewport (scrollable)        │  <- persisted history + live activity
│  ❯ user message               │
│  ● assistant response         │
│  ● Thinking... (live)         │
├───────────────────────────────┤
│  ─────────────────────────    │  <- separator (with top margin)
│  ❯ input bar                  │
│  ─────────────────────────    │
│  status bar          ↕ model  │  <- scroll arrow + model/mode
└───────────────────────────────┘
```

### Key Design Decisions

- **`tea.WithAltScreen()`** — full-screen TUI; avoids inline rendering artifacts
- **`bubbles/viewport`** — scrollable content area; height = terminal height minus fixed chrome rows
- **`persisted string`** — accumulates all finalized output; written to stdout on exit so the user keeps a transcript
- **`buildViewportContent()`** — returns `persisted` + live activity (spinner/streaming) when a response is in progress
- **`appendToViewport(content)`** — appends to `persisted` and calls `updateViewport()` (SetContent + GotoBottom)
- **No `tea.WithMouseCellMotion()`** — mouse capture is omitted so the user can select text with the mouse
- **No `tea.Println`** — all output goes through the viewport

## Required Dependencies

```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/viewport"
    "github.com/charmbracelet/lipgloss"
)
```

## Implementation Requirements

### Model Structure

```go
type model struct {
    // Dimensions
    width  int
    height int

    // Core components
    input    textinput.Model
    spinner  spinner.Model
    viewport viewport.Model

    // Content
    persisted string // finalized transcript (also written to stdout on exit)

    // State
    processing bool
    quitting   bool
    program    *tea.Program
}
```

### Initialization

```go
func newModel() *model {
    return &model{
        input:    newTextInput(),
        spinner:  newSpinner(),
        viewport: viewport.New(80, 20),
        width:    80,
    }
}

func (m *model) Init() tea.Cmd {
    return tea.Batch(textinput.Blink, m.spinner.Tick)
}
```

### Viewport Height

Reserve 6 rows for chrome: two separators (each with a top margin counts as 2 rows) + input row + status row:

```go
func (m *model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
    m.width = msg.Width
    m.height = msg.Height
    m.input.Width = msg.Width - 4
    m.viewport.Width = msg.Width
    vpHeight := msg.Height - 6
    if vpHeight < 5 {
        vpHeight = 5
    }
    m.viewport.Height = vpHeight
    m.viewport.SetContent(m.buildViewportContent())
    return m, nil
}
```

### Viewport Content Helpers

```go
// appendToViewport adds finalized content and scrolls to bottom.
func (m *model) appendToViewport(content string) {
    m.persisted += content
    m.updateViewport()
}

// buildViewportContent returns persisted history plus any live activity.
func (m *model) buildViewportContent() string {
    if m.processing {
        if live := m.renderActivity(); live != "" {
            return m.persisted + live
        }
    }
    return m.persisted
}

// updateViewport refreshes content and scrolls to bottom.
func (m *model) updateViewport() {
    m.viewport.SetContent(m.buildViewportContent())
    m.viewport.GotoBottom()
}
```

### Keyboard Scrolling

```go
case tea.KeyPgUp:
    m.viewport.HalfViewUp()
    return m, nil
case tea.KeyPgDown:
    m.viewport.HalfViewDown()
    return m, nil
case tea.KeyUp:
    m.viewport.LineUp(1)
    return m, nil
case tea.KeyDown:
    m.viewport.LineDown(1)
    return m, nil
```

### Scroll Arrow in Status Bar

Display ↑/↓/↕ in the status bar to indicate scroll position:

```go
func (m *model) scrollArrow() string {
    if m.viewport.YOffset == 0 && m.viewport.AtBottom() {
        return "" // all content fits
    }
    canUp := m.viewport.YOffset > 0
    canDown := !m.viewport.AtBottom()
    switch {
    case canUp && canDown:
        return "↕"
    case canUp:
        return "↑"
    default:
        return "↓"
    }
}
```

### View Layout

`View()` renders the full screen: viewport + chrome. Return `""` when quitting so the alt screen clears cleanly.

```go
func (m *model) View() string {
    if m.quitting {
        return ""
    }
    separator := strings.Repeat("─", m.width-2)
    parts := []string{
        m.viewport.View(),
        separator,
        m.input.View(),
        separator,
        m.renderStatusBar(),
    }
    return strings.Join(parts, "\n")
}
```

### Program Entry and Exit Transcript

After `p.Run()` returns, print `persisted` to stdout so the user gets a transcript in their shell:

```go
func Run() error {
    m := newModel()
    p := tea.NewProgram(m, tea.WithAltScreen())
    m.program = p

    finalModel, err := p.Run()

    if fm, ok := finalModel.(*model); ok && fm.persisted != "" {
        fmt.Print(fm.persisted)
    }
    return err
}
```

### Ctrl+C Behaviour

First Ctrl+C clears the input (with a 2-second warning); second Ctrl+C within that window exits:

```go
func (m *model) handleCtrlC() (tea.Model, tea.Cmd) {
    now := time.Now()
    if m.ctrlCPressed && now.Sub(m.ctrlCTime) < 2*time.Second {
        m.quitting = true
        return m, tea.Quit
    }
    m.input.SetValue("")
    m.ctrlCPressed = true
    m.ctrlCTime = now
    return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
        return ctrlCClearMsg{}
    })
}
```

## Quality Checklist

Before returning generated code, verify:

- [ ] Uses `tea.WithAltScreen()` — full-screen, no inline artifacts
- [ ] Uses `bubbles/viewport` — scrollable content area
- [ ] Does NOT use `tea.Println` — all output goes through the viewport
- [ ] Does NOT use `tea.WithMouseCellMotion()` — preserves native text selection
- [ ] `persisted` string accumulates all finalized content
- [ ] `appendToViewport` = append to `persisted` + `updateViewport()`
- [ ] `buildViewportContent` = `persisted` + live activity when processing
- [ ] Viewport height = terminal height minus 6 fixed chrome rows
- [ ] `View()` returns `""` when quitting
- [ ] After `p.Run()`, prints `fm.persisted` to stdout for transcript
- [ ] PgUp/PgDn and Up/Down scroll the viewport
- [ ] Scroll arrow (↑/↓/↕) shown in status bar
- [ ] First Ctrl+C clears input; second Ctrl+C quits
- [ ] Model uses pointer receiver (`*model`) for all methods
- [ ] Update handlers are broken into focused methods
