---
name: bubbletea
description: Browse Bubbletea TUI framework documentation and examples. Use when working with Bubbletea components, models, commands, or building terminal user interfaces in Go.
---

# Bubbletea Documentation

Bubbletea is a Go framework for building terminal user interfaces based on The Elm Architecture.

## Key Resources

When you need to understand Bubbletea patterns or find examples:

1. **Examples README** - Overview of all available examples:
   https://github.com/charmbracelet/bubbletea/blob/main/examples/README.md

2. **Examples Directory** - Full source code for all examples:
   https://github.com/charmbracelet/bubbletea/tree/main/examples

## How to Use

1. First, fetch the examples README to get an overview of available examples:
   ```
   WebFetch https://github.com/charmbracelet/bubbletea/blob/main/examples/README.md
   ```

2. Once you identify a relevant example, fetch its source code from the examples directory.

## Common Examples to Reference

- `list` - List component with filtering
- `table` - Table component
- `textinput` - Text input handling
- `textarea` - Multi-line text input
- `viewport` - Scrollable content
- `paginator` - Pagination
- `spinner` - Loading spinners
- `progress` - Progress bars
- `tabs` - Tab navigation
- `help` - Help text/keybindings display

## Core Concepts

- **Model**: Application state
- **Update**: Handles messages and returns updated model + commands
- **View**: Renders the model to a string
- **Cmd**: Side effects that produce messages
- **Msg**: Events that trigger updates

## Related Charm Libraries

- **Bubbles**: Pre-built components (github.com/charmbracelet/bubbles)
- **Lipgloss**: Styling and layout (github.com/charmbracelet/lipgloss)
- **Glamour**: Markdown rendering (github.com/charmbracelet/glamour)
