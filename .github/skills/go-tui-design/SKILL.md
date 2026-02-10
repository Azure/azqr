---
name: go-tui-design
description: Create distinctive, production-grade Go terminal UIs with Bubbletea, Lipgloss, and Bubbles. Use when building TUI apps, CLI tools, interactive sidebars, dashboards, or any terminal interface in Go.
---

# Go TUI Design

Create distinctive, production-grade terminal user interfaces in Go using the Charmbracelet ecosystem (Bubbletea, Lipgloss, Bubbles). This skill covers both visual design craft and architectural patterns for building polished, maintainable TUI applications.

Use this skill when building Go CLI tools, TUI applications, interactive terminal sidebars, dashboards, or any terminal-based interface using Bubbletea.

---

## Design Thinking

Before writing code, commit to a clear aesthetic direction:

1. **Purpose**: What problem does this solve? Who uses it? What's the workflow? How much screen real estate do you have?
2. **Tone**: Choose an intentional aesthetic — not a default. Examples: minimalist utility, dense monitoring dashboard, retro-CRT amber, cyberpunk neon, monochrome brutalist, warm terminal, cool nord, playful whimsical, military tactical, zen single-focus. The key is *intentionality*, not intensity.
3. **Constraints**: Terminal width, color support (ANSI 16 / 256 / true color), whether it runs in a split pane or full screen, target terminals (Ghostty, iTerm2, basic xterm).
4. **Differentiation**: What's the one thing someone will remember about this interface? A great TUI has a signature element — a distinctive status bar, an elegant transition, a perfect information density.

Match implementation complexity to the aesthetic vision. A dense dashboard needs elaborate panels. A minimal sidebar needs restraint, precision, and perfect alignment. Elegance comes from executing the vision well, not from adding more decoration.

---

## Architecture — Bubbletea Patterns

### Model-View-Update (MVU)

Bubbletea uses the Elm Architecture. Every program has three parts:

- **Model**: A struct holding all application state
- **Update**: Receives messages, returns updated model + optional commands
- **View**: Pure render function — model in, string out

```go
type Model struct {
    items    []Item
    cursor   int
    width    int
    height   int
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* handle messages */ }
func (m Model) View() string { /* render UI */ }
```

**Rules:**
- Keep `Update()` and `View()` fast — never block. Offload I/O to `tea.Cmd` functions.
- All state lives in the model. No globals, no goroutines mutating state.
- `View()` must be deterministic — same model, same output.

### Component Composition

Any non-trivial program outgrows a single model. Structure as a tree:

```go
type AppModel struct {
    sidebar  SidebarModel
    content  ContentModel
    active   string // "sidebar" or "content"
    width    int
    height   int
}
```

**Message routing pattern:**
- **Global keys** (quit, help): Handle at the root
- **Context-specific input**: Route to the active child model
- **System messages** (tea.WindowSizeMsg): Broadcast to all children

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // Broadcast to children
        m.sidebar, _ = m.sidebar.Update(msg)
        m.content, _ = m.content.Update(msg)
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
        // Route to active child
        switch m.active {
        case "sidebar":
            var cmd tea.Cmd
            m.sidebar, cmd = m.sidebar.Update(msg)
            return m, cmd
        }
    }
    return m, nil
}
```

### Command Patterns

```go
// Concurrent — independent operations
return m, tea.Batch(fetchData, startTimer, checkStatus)

// Sequential — ordered dependencies
return m, tea.Sequence(loadConfig, applyConfig)
```

**Custom message types** for domain logic:

```go
type dataLoadedMsg struct{ items []Item }
type errMsg struct{ err error }
type tickMsg time.Time
```

### State Machines

Use explicit states for complex flows:

```go
type appState int
const (
    stateLoading appState = iota
    stateReady
    stateEditing
    stateError
)
```

Route both `Update()` and `View()` through the current state. This prevents impossible state combinations and makes the UI predictable.

### Polling & Live Data

Use `tea.Tick` for periodic refresh (notifications, external state):

```go
func tickCmd() tea.Cmd {
    return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

Return `tickCmd()` from both `Init()` and the tick handler in `Update()` to keep the loop running.

---

## Bubbles Components

Use the `bubbles` library for pre-built components. Each is a model that follows MVU.

**Integration pattern** (same for all components):
1. Embed the component model in your parent model
2. Forward messages via the component's `Update()`
3. Call the component's `View()` in your render

### Viewport (scrollable content)

```go
vp := viewport.New(width, height)
vp.SetContent(longContent)
// In Update: vp, cmd = vp.Update(msg)
// In View: vp.View()
```

Use for any content that might exceed terminal height. Always recalculate dimensions on `tea.WindowSizeMsg`.

### List (filterable, paginated)

Built-in fuzzy filtering, pagination, help text, and status messages. Use for item selection with many entries.

### Table (structured data)

Includes its own viewport for scrolling. Set columns with widths, populate rows as `[]string`.

### Text Input

For inline editing, renaming, search. Call `.Focus()` to activate, delegate messages while focused.

### Spinner

For async operations. Multiple styles: `spinner.Dot`, `spinner.Line`, `spinner.MiniDot`, etc.

---

## Layout — Lipgloss

### Core Concepts

Lipgloss is purely functional — styles are immutable values, not mutable objects.

```go
style := lipgloss.NewStyle().
    Foreground(lipgloss.Color("15")).
    Background(lipgloss.Color("4")).
    Bold(true).
    Padding(0, 1).
    MarginBottom(1)

rendered := style.Render("Hello")
```

### Composition Functions

```go
// Side by side (position: lipgloss.Top, Center, Bottom)
lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

// Stacked (position: lipgloss.Left, Center, Right)
lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

// Place text in whitespace
lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
```

### Responsive Layout

**Never hardcode dimensions.** Always derive from `tea.WindowSizeMsg`:

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    sidebarWidth := 30
    contentWidth := m.width - sidebarWidth - 1 // -1 for border
    headerHeight := lipgloss.Height(m.renderHeader())
    bodyHeight := m.height - headerHeight - footerHeight
```

Use `lipgloss.Width()` and `lipgloss.Height()` to measure rendered strings. Never assume — measure.

### Borders

```go
style := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).  // ╭─╮│╰╯
    BorderForeground(lipgloss.Color("8")).
    Width(30).
    Padding(1, 2)
```

Border types and when to use them:
- `NormalBorder()` — `┌─┐│└┘` — Clean, default
- `RoundedBorder()` — `╭─╮│╰╯` — Soft, modern, friendly
- `DoubleBorder()` — `╔═╗║╚╝` — Bold, formal, retro-mainframe
- `ThickBorder()` — `┏━┓┃┗┛` — Strong, industrial
- `HiddenBorder()` — Reserve space without drawing (alignment tool)
- Custom borders via `lipgloss.Border{}` struct for unique aesthetics

### Width Control

```go
// Fixed width — text wraps or truncates
style.Width(40)

// Max width — natural size up to limit
style.MaxWidth(60)

// Inline padding to simulate fixed width
style.Padding(0, 2)
```

---

## Color & Theme

### Color Systems

```go
// ANSI 16 — universal, respects user's terminal theme
lipgloss.Color("1")  // red (adapts to light/dark themes)

// ANSI 256 — richer palette, still widely supported
lipgloss.Color("205")  // hot pink

// True color — full spectrum, Ghostty/iTerm2/modern terminals
lipgloss.Color("#ff6b6b")

// Adaptive — different colors for light vs dark backgrounds
lipgloss.AdaptiveColor{Light: "0", Dark: "15"}
```

**Prefer ANSI 16 for utility tools** — they adapt to the user's theme. Use 256/true color when the aesthetic demands it.

### Building a Palette

Define a cohesive palette as package-level constants:

```go
var (
    colorPrimary   = lipgloss.Color("4")   // Blue — active elements
    colorSecondary = lipgloss.Color("6")   // Cyan — headers, accents
    colorMuted     = lipgloss.Color("8")   // Gray — secondary info
    colorText      = lipgloss.Color("7")   // Light gray — body text
    colorBright    = lipgloss.Color("15")  // White — emphasis
    colorAlert     = lipgloss.Color("3")   // Yellow — notifications
    colorSuccess   = lipgloss.Color("2")   // Green — confirmations
    colorDanger    = lipgloss.Color("1")   // Red — destructive actions
)
```

### Atmosphere Techniques

- **Dim text** (`lipgloss.Color("8")`) for secondary information, timestamps, paths
- **Bold + bright foreground** for primary actions, selected items, headers
- **Background highlight** for cursor/selection — full-width bar effect
- **Reverse video** (swap fg/bg) for maximum emphasis
- **Gradient fills** using block characters: `░▒▓█`
- **Color-coded semantics** — but be intentional, not cliched

### Example Palettes

```go
// Warm amber (vintage CRT)
colorAmber     = lipgloss.Color("#ffb000")
colorAmberDim  = lipgloss.Color("#806000")
colorAmberBg   = lipgloss.Color("#1a1000")

// Cool nord
colorNordFrost = lipgloss.Color("#88c0d0")
colorNordSnow  = lipgloss.Color("#eceff4")
colorNordNight = lipgloss.Color("#2e3440")

// Cyberpunk
colorNeon      = lipgloss.Color("#ff00ff")
colorCyan      = lipgloss.Color("#00ffff")
colorDeepBg    = lipgloss.Color("#1a0a2e")

// Agent status (semantic, warm/cool contrast)
colorWaiting   = lipgloss.Color("#d7875f") // warm amber — needs attention
colorWorking   = lipgloss.Color("#5f87af") // steel blue — in progress
colorDone      = lipgloss.Color("#5faf5f") // soft green — complete
colorMuted     = lipgloss.Color("#585858") // dim gray — secondary info
colorBright    = lipgloss.Color("#e4e4e4") // near white — emphasis
```

---

## Typography & Text

The terminal is ALL typography. Make it count.

### Visual Hierarchy

```go
headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorSecondary)
bodyStyle    = lipgloss.NewStyle().Foreground(colorText)
mutedStyle   = lipgloss.NewStyle().Foreground(colorMuted)
```

Three levels of emphasis is usually enough: **primary** (bold/bright), normal, and **muted** (dim/gray).

### Unicode Symbols

Enrich your UI beyond ASCII:

```
Bullets:     ▸ › ◆ ● ○ ◉ ⬢
Status:      ● active  ○ empty  ◐ partial  ✓ done  ✗ failed
Arrows:      → ← ↳ ⟶
Indicators:  ▸ cursor  ↳ nested detail  › sub-item
Separators:  ─── ═══ ··· ░░░
Sparklines:  ▁▂▃▄▅▆▇█
Progress:    ████░░░░ or ◉◉◉○○○
```

### Section Headers

```go
// Minimal
" WORKSPACES"

// With divider
" WORKSPACES\n" + " " + strings.Repeat("─", width-2)

// Bracketed
"[ WORKSPACES ]"

// Decorated
"◆ WORKSPACES"

// Centered rule
"─── WORKSPACES ───"
```

Choose ONE style and use it consistently.

---

## Box Drawing & Decorative Elements

### Border Styles by Aesthetic

| Style | Characters | Mood |
|-------|-----------|------|
| Single | `┌─┐│└┘` | Clean, modern |
| Rounded | `╭─╮│╰╯` | Soft, friendly |
| Double | `╔═╗║╚╝` | Bold, formal, retro |
| Heavy | `┏━┓┃┗┛` | Industrial, strong |
| Dashed | `┄ ┆` | Light, informal |
| ASCII | `+-+\|` | Universal compat |
| Block | `█▀▄▌▐` | Brutalist, chunky |

### Dividers

```go
// Simple
strings.Repeat("─", width)

// Dotted
strings.Repeat("·", width)

// Gradient
strings.Repeat("░", width)

// With title
fmt.Sprintf("── %s %s", title, strings.Repeat("─", width-len(title)-4))
```

### Custom Decorations

Mix Unicode for unique frames: `◢◣◤◥`, `◆◈✦⬡`, `⌘λ∴≡`

Use sparingly. A single distinctive element (a custom bullet, a unique divider) has more impact than decorating everything.

---

## Interaction Design

Visual design only matters if the interaction is solid. These patterns make or break a TUI.

### Popup / Launcher Pattern

For ephemeral UIs (switchers, pickers, command palettes), use `tmux display-popup` instead of persistent panes:

- **Size to content**: Don't use 100% width/height. 40-50% is usually right for a launcher.
- **Two-line cards**: Name + status on line 1, path/detail indented below. Gives breathing room without wasting space.
- **Auto-dismiss**: Quit on action (Enter to select, then `tea.Quit`). The popup is a verb, not a noun.
- **Vim command mode**: Support `:q` for muscle-memory users. Simple state machine: `cmdMode` bool + `cmdBuf` string.
- **Summary bar**: Show aggregate status at the top with a separator — lets users glance without scanning every entry.

```go
// Popup from Go — called by a tmux keybind
tmux.Run("display-popup", "-E",
    "-w", "45%", "-h", "40%",
    "-b", "rounded",
    "-S", "fg=#c0c0c0",
    "-T", " ⬡ title ",
    "myapp", "popup")
```

### Cursor & Selection

Always make two things visually distinct — use **different symbols and different colors**:
1. **Where the cursor is** (navigating) — `❯` prefix in bright white
2. **What is currently active** (selected/current) — `◆` prefix in accent color

```go
if isCursor {
    // Full-width background bar
    line = cursorStyle.Width(m.width).Render("▸ " + name)
} else if isCurrent {
    line = currentStyle.Render("  " + name)
} else {
    line = normalStyle.Render("  " + name)
}
```

### Keyboard Conventions

Follow vim/TUI conventions users expect:

| Key | Action |
|-----|--------|
| `j/k` or `↑/↓` | Navigate |
| `Enter` | Select / confirm |
| `q` / `Esc` | Quit / back |
| `g` / `G` | Jump to top / bottom |
| `/` | Search / filter |
| `n` | New / create |
| `r` | Rename |
| `x` / `d` | Delete (with confirmation) |
| `Tab` | Switch focus between panels |
| `?` | Help |

### Focus Management

When you have multiple panels, make focus obvious:

```go
func (m Model) View() string {
    sidebarBorder := lipgloss.NormalBorder()
    sidebarColor := colorMuted

    if m.active == "sidebar" {
        sidebarBorder = lipgloss.RoundedBorder()
        sidebarColor = colorSecondary
    }

    sidebarStyle := lipgloss.NewStyle().
        Border(sidebarBorder).
        BorderForeground(sidebarColor)
}
```

### Confirmations for Destructive Actions

Never delete on a single keypress. Use inline confirmation:

```go
// State: confirmingDelete
// View: "Delete session 'api-server'? y/n"
// Update: only 'y' proceeds, anything else cancels
```

### Empty States

Don't show a blank screen. Always have a message:

```go
if len(m.items) == 0 {
    return mutedStyle.Render("  No sessions running. Press n to create one.")
}
```

### Loading & Async Feedback

Show spinners for operations that take >200ms:

```go
if m.loading {
    return spinner.View() + " Loading sessions..."
}
```

---

## Animation & Dynamic Content

### Spinners

Bubbletea's spinner component offers multiple styles:

```go
s := spinner.New()
s.Spinner = spinner.Dot      // ⣾⣽⣻⢿⡿⣟⣯⣷
s.Spinner = spinner.Line     // |/-\
s.Spinner = spinner.MiniDot  // ⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏
s.Style = lipgloss.NewStyle().Foreground(colorSecondary)
```

### Progress Indicators

```go
// Block-based bar
filled := int(float64(width) * percent)
bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

// With percentage
fmt.Sprintf("%s %3.0f%%", bar, percent*100)
```

### Live Data

For streaming data, sparklines, or real-time charts:

```go
sparkline := ""
chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
for _, v := range values {
    idx := int(v / maxVal * float64(len(chars)-1))
    sparkline += string(chars[idx])
}
```

---

## Testing

### Unit Test Update/View Directly

```go
func TestCursorNavigation(t *testing.T) {
    m := NewModel()
    m.entries = []entry{{Name: "a"}, {Name: "b"}, {Name: "c"}}

    // Press j
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
    if m.(Model).cursor != 1 {
        t.Errorf("expected cursor=1, got %d", m.(Model).cursor)
    }
}
```

### Golden File Testing for View

Snapshot `View()` output and diff against expected files. Catch visual regressions.

### teatest for Integration

```go
tm := teatest.NewTestModel(t, NewModel())
tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
out := tm.FinalOutput(t)
// Assert output contains expected content
```

---

## Debugging

### Message Logging

Dump messages to a file and `tail -f` in another terminal:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.debugLog != nil {
        fmt.Fprintf(m.debugLog, "%T: %+v\n", msg, msg)
    }
    // ...
}
```

### Common Pitfalls

- **Blocking in Update/View**: Offload I/O to commands. A slow `View()` freezes the entire UI.
- **Hardcoded dimensions**: Use `lipgloss.Width()`/`Height()` and `tea.WindowSizeMsg`. Test with different terminal sizes.
- **Race conditions**: Never modify model state from goroutines. Use `tea.Cmd` or `p.Send()`.
- **Layout arithmetic off by one**: Borders add 2 to width/height. Padding adds to both sides. Always account for chrome.
- **Forgotten nil checks**: `tea.WindowSizeMsg` arrives after `Init()`. Guard against zero width/height in `View()`.
- **Panics in commands**: Don't crash the terminal. Wrap command functions with recover, or return `errMsg`.
- **Circular rendering in chrome measurement**: Never call `lipgloss.Height(m.renderHeader())` inside a function that `renderHeader()` also depends on — this can cause runaway memory allocation and SIGKILL (OOM). Use fixed constants for chrome height instead.
- **Viewport initialization timing**: The viewport is not ready until the first `tea.WindowSizeMsg`. Guard all `viewport.SetContent()` calls behind a `ready` flag. Return empty string from `View()` if not ready.

### tmux Integration

When building TUI apps that run inside tmux panes or popups:

- **Prefer `display-popup` over `split-window`** for ephemeral UIs (switchers, launchers, pickers). Popups float over panes, don't steal screen real estate, and auto-close with `-E` when the command exits. Use `split-window` only for persistent panels.
- **`display-popup` styling**: Use `-b rounded` for modern borders, `-S "fg=#color"` for border/title color (note: title and border share the same style), `-T " title "` with padding for a clean look.
- **Don't use shell wrappers**: `sh -c "cmd1; cmd2"` breaks TTY control. Bubbletea needs direct TTY access. Run the TUI binary directly in `split-window` or `display-popup`.
- **Pane detection**: Don't use `pgrep -f` to find TUI panes — it's fragile and can match the wrong process (including the caller). Use `tmux select-pane -T <title>` to tag panes, then detect with `list-panes -F '#{pane_title}'`.
- **"No space for new pane"**: tmux refuses splits when the target pane is too narrow. Consider targeting the largest pane explicitly, or gracefully handle this error.
- **Alt-screen cleanup**: Always use `tea.WithAltScreen()` so the TUI cleans up properly when the pane is killed externally.
- **macOS binary quarantine**: `cp` adds `com.apple.provenance` xattr which can cause SIGKILL. Build directly to the install path (`go build -o ~/.local/bin/myapp .`) instead of building then copying.

---

## Anti-Patterns

NEVER produce:
- Plain unformatted text output with no styling
- Default colors without an intentional palette
- Inconsistent spacing and alignment
- Walls of unstructured text
- Generic `[INFO]` / `[ERROR]` prefixes without styling
- Simple `----` dividers when a proper `─────` is available
- Hardcoded widths that break on resize
- Components that don't respond to `tea.WindowSizeMsg`
- Blocking operations in `Update()` or `View()`
- Single-keypress destructive actions without confirmation
- Empty screens with no guidance
- Decoration without purpose — every visual element should earn its place

---

## Quick Reference

### Charmbracelet Ecosystem

| Package | Purpose |
|---------|---------|
| `bubbletea` | TUI framework (MVU architecture) |
| `lipgloss` | Styling and layout composition |
| `bubbles` | Pre-built components (viewport, list, table, textinput, spinner, progress, paginator) |
| `log` | Styled logging |
| `gum` | Shell script TUI utilities |
| `vhs` | Record terminal GIFs from scripts |
| `wish` | SSH server for TUI apps |

### ANSI Quick Ref

```
\x1b[1m  Bold       \x1b[2m  Dim        \x1b[3m  Italic
\x1b[4m  Underline  \x1b[7m  Reverse    \x1b[0m  Reset
\x1b[31m Red fg     \x1b[42m Green bg   \x1b[38;2;R;G;Bm True color
```

(Prefer lipgloss over raw ANSI — but useful for debugging and understanding what lipgloss generates.)
