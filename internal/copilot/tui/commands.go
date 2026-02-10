// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Command represents a slash command
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Handler     func(m *Model, args []string) CommandResult
}

// CommandResult is returned by command handlers
type CommandResult struct {
	Output    string
	ClearChat bool
	Exit      bool
	Error     error
	Raw       bool    // If true, output is pre-styled and should not be markdown-rendered
	EditorCmd tea.Cmd // If set, caller should return this as the tea.Cmd
}

// CommandRegistry holds all registered commands
type CommandRegistry struct {
	commands map[string]*Command
	aliases  map[string]string
}

// NewCommandRegistry creates a new command registry with default commands
func NewCommandRegistry() *CommandRegistry {
	r := &CommandRegistry{
		commands: make(map[string]*Command),
		aliases:  make(map[string]string),
	}
	r.registerDefaults()
	return r
}

func (r *CommandRegistry) registerDefaults() {
	// Help command
	r.Register(&Command{Name: "help", Description: "Show available commands", Handler: cmdHelp})
	r.Register(&Command{Name: "exit", Aliases: []string{"quit", "q"}, Description: "Exit the session", Handler: cmdExit})
	r.Register(&Command{Name: "clear", Aliases: []string{"new"}, Description: "Clear the conversation history", Handler: cmdClear})
	r.Register(&Command{Name: "editor", Aliases: []string{"e"}, Description: "Open $EDITOR to compose a message", Handler: cmdEditor})
	r.Register(&Command{Name: "model", Description: "Show or change the AI model", Handler: cmdModel})
	r.Register(&Command{Name: "context", Description: "Show context window token usage", Handler: cmdContext})
	r.Register(&Command{Name: "compact", Description: "Summarize history to reduce context usage", Handler: cmdCompact})
	r.Register(&Command{Name: "scan", Description: "Run Azure resource compliance scan", Handler: cmdScan})
	r.Register(&Command{Name: "rules", Description: "Show recommendations catalog", Handler: cmdRules})
	r.Register(&Command{Name: "services", Description: "List supported Azure services", Handler: cmdServices})
	r.Register(&Command{Name: "mode", Description: "Switch interaction mode (ask/plan/agent)", Handler: cmdMode})
	r.Register(&Command{Name: "session", Description: "Show current session info or start new session", Handler: cmdSession})
	r.Register(&Command{Name: "diff", Description: "Review changes made in current session", Handler: cmdDiff})
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(cmd *Command) {
	r.commands[cmd.Name] = cmd
	for _, alias := range cmd.Aliases {
		r.aliases[alias] = cmd.Name
	}
}

// Get retrieves a command by name or alias
func (r *CommandRegistry) Get(name string) *Command {
	// Check direct command name
	if cmd, ok := r.commands[name]; ok {
		return cmd
	}
	// Check aliases
	if realName, ok := r.aliases[name]; ok {
		return r.commands[realName]
	}
	return nil
}

// List returns all commands sorted by name
func (r *CommandRegistry) List() []*Command {
	var cmds []*Command
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name < cmds[j].Name
	})
	return cmds
}

// Match returns commands that match the given prefix
func (r *CommandRegistry) Match(prefix string) []*Command {
	prefix = strings.ToLower(prefix)
	var matches []*Command
	for _, cmd := range r.commands {
		if strings.HasPrefix(strings.ToLower(cmd.Name), prefix) {
			matches = append(matches, cmd)
			continue
		}
		for _, alias := range cmd.Aliases {
			if strings.HasPrefix(strings.ToLower(alias), prefix) {
				matches = append(matches, cmd)
				break
			}
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})
	return matches
}

// Execute runs a command by parsing the input line
func (r *CommandRegistry) Execute(m *Model, line string) CommandResult {
	// Parse command and args
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return CommandResult{}
	}

	// Remove leading slash
	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	cmd := r.Get(cmdName)
	if cmd == nil {
		return CommandResult{
			Error: &UnknownCommandError{Name: cmdName},
		}
	}

	return cmd.Handler(m, args)
}

// UnknownCommandError is returned when a command is not found
type UnknownCommandError struct {
	Name string
}

func (e *UnknownCommandError) Error() string {
	return "unknown command: /" + e.Name
}

// Command handlers
func cmdHelp(m *Model, _ []string) CommandResult {
	var sb strings.Builder
	sb.WriteString("Commands:\n")
	for _, cmd := range m.commands.List() {
		sb.WriteString("  ")
		sb.WriteString(m.styles.CommandSlash.Render("/"))
		sb.WriteString(m.styles.CommandName.Render(cmd.Name))
		if len(cmd.Aliases) > 0 {
			sb.WriteString(m.styles.CommandDesc.Render(", /" + strings.Join(cmd.Aliases, ", /")))
		}
		sb.WriteString("  ")
		sb.WriteString(m.styles.CommandDesc.Render(cmd.Description))
		sb.WriteString("\n")
	}

	sb.WriteString("\nKey bindings:\n")
	for _, b := range KeyMapToSlice(m.keyMap) {
		keys := b.Keys()
		if len(keys) == 0 {
			continue
		}
		sb.WriteString("  ")
		sb.WriteString(m.styles.CommandName.Render(keys[0]))
		if help := b.Help(); help.Desc != "" {
			sb.WriteString("  ")
			sb.WriteString(m.styles.CommandDesc.Render(help.Desc))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(m.styles.HintText.Render("Type a question to chat with Copilot about Azure resources."))
	return CommandResult{Output: sb.String(), Raw: true}
}

func cmdExit(_ *Model, _ []string) CommandResult {
	return CommandResult{Exit: true}
}

func cmdClear(m *Model, _ []string) CommandResult {
	m.history = nil
	m.conversationTurns = 0
	return CommandResult{
		Output:    m.styles.FormatSuccess("Conversation cleared"),
		ClearChat: true,
		Raw:       true,
	}
}

func cmdModel(m *Model, args []string) CommandResult {
	if len(args) == 0 {
		var sb strings.Builder
		sb.WriteString("Current model: " + m.styles.StatusModel.Render(m.config.Model) + "\n")

		// Fetch available models from SDK
		if m.client != nil {
			models, err := m.client.ListModels(context.Background())
			if err == nil && len(models) > 0 {
				sb.WriteString("\nAvailable models:\n")
				for _, model := range models {
					if model.ID == m.config.Model {
						sb.WriteString(m.styles.SuccessText.Render("  ✓ "+model.ID) + "\n")
					} else {
						sb.WriteString(m.styles.HintText.Render("    "+model.ID) + "\n")
					}
				}
			} else if err != nil {
				sb.WriteString("\n" + m.styles.HintText.Render("(could not fetch models: "+err.Error()+")"))
			}
		}
		sb.WriteString("\n" + m.styles.HintText.Render("Usage: /model <name> to switch models"))
		return CommandResult{
			Output: sb.String(),
			Raw:    true,
		}
	}

	// Validate model name by fetching from SDK
	if m.client != nil {
		models, err := m.client.ListModels(context.Background())
		if err == nil && len(models) > 0 {
			valid := false
			for _, model := range models {
				if model.ID == args[0] {
					valid = true
					break
				}
			}
			if !valid {
				return CommandResult{
					Output: m.styles.FormatError("Unknown model: " + args[0] + "\nUse /model to see available models"),
					Raw:    true,
				}
			}
		}
	}

	m.config.Model = args[0]
	_ = m.config.Save()
	return CommandResult{
		Output: m.styles.FormatSuccess("Switched to model: " + args[0]),
		Raw:    true,
	}
}

func cmdContext(m *Model, _ []string) CommandResult {
	// Count messages by role
	userCount := 0
	assistantCount := 0
	for _, entry := range m.history {
		if entry.Role == "user" {
			userCount++
		} else {
			assistantCount++
		}
	}

	return CommandResult{
		Output: "Context window usage:\n" +
			m.styles.StatusModel.Render("  Turns: ") + fmt.Sprintf("%d", m.conversationTurns) + "\n" +
			m.styles.StatusModel.Render("  Messages: ") + fmt.Sprintf("%d (%d user, %d assistant)", len(m.history), userCount, assistantCount) + "\n" +
			m.styles.StatusModel.Render("  Mode: ") + m.styles.ModeIndicator(m.config.Mode) + "\n" +
			m.styles.HintText.Render("  Infinite sessions: enabled (auto-compaction at 80%% context usage)"),
		Raw: true,
	}
}

func cmdCompact(m *Model, _ []string) CommandResult {
	// With infinite sessions enabled in the SDK, compaction happens automatically.
	// This command manually clears local history while preserving session context.
	historyLen := len(m.history)
	if historyLen == 0 {
		return CommandResult{
			Output: m.styles.HintText.Render("No conversation history to compact"),
			Raw:    true,
		}
	}

	// Keep only the last few messages for context
	keepCount := 4
	if historyLen <= keepCount {
		return CommandResult{
			Output: m.styles.HintText.Render("History is already minimal"),
			Raw:    true,
		}
	}

	// Clear local history but preserve recent context
	m.history = m.history[historyLen-keepCount:]
	compactedCount := historyLen - keepCount

	return CommandResult{
		Output: m.styles.FormatSuccess(fmt.Sprintf("Compacted %d messages, kept last %d for context", compactedCount, keepCount)+"\n") +
			m.styles.HintText.Render("Note: Server-side context is managed automatically via infinite sessions"),
		Raw: true,
	}
}

func cmdScan(m *Model, args []string) CommandResult {
	// Signal that we want to run a scan through the copilot
	if len(args) == 0 {
		m.pendingPrompt = "Run an Azure resource compliance scan on all services"
	} else {
		m.pendingPrompt = "Run an Azure resource compliance scan on these services: " + strings.Join(args, ", ")
	}
	return CommandResult{}
}

func cmdRules(m *Model, args []string) CommandResult {
	if len(args) > 0 {
		m.pendingPrompt = "Show me the azqr recommendations catalog filtered by: " + strings.Join(args, " ")
	} else {
		m.pendingPrompt = "Show me the complete azqr recommendations catalog"
	}
	return CommandResult{}
}

func cmdServices(m *Model, _ []string) CommandResult {
	m.pendingPrompt = "List all Azure services supported by azqr"
	return CommandResult{}
}

func cmdMode(m *Model, args []string) CommandResult {
	if len(args) == 0 {
		return CommandResult{
			Output: "Current mode: " + m.styles.ModeIndicator(m.config.Mode) + "\n" +
				m.styles.HintText.Render("  ask   - Default Q&A mode\n") +
				m.styles.HintText.Render("  plan  - Multi-step planning mode\n") +
				m.styles.HintText.Render("  agent - Autonomous execution mode\n") +
				m.styles.HintText.Render("Usage: /mode <ask|plan|agent> or press shift+tab to cycle"),
			Raw: true,
		}
	}

	mode := strings.ToLower(args[0])
	switch mode {
	case "ask", "plan", "agent":
		m.config.Mode = mode
		return CommandResult{
			Output: m.styles.FormatSuccess("Switched to " + m.styles.ModeIndicator(mode) + " mode"),
			Raw:    true,
		}
	default:
		return CommandResult{
			Output: m.styles.FormatError("Invalid mode. Use: ask, plan, or agent"),
			Raw:    true,
		}
	}
}

func cmdSession(m *Model, args []string) CommandResult {
	if len(args) > 0 && strings.ToLower(args[0]) == "new" {
		oldID := ""
		if m.session != nil {
			oldID = m.session.SessionID
			_ = m.client.DeleteSession(context.Background(), oldID)
		}

		// Clear local history
		m.history = nil
		m.conversationTurns = 0

		return CommandResult{
			Output: m.styles.FormatSuccess(fmt.Sprintf("Session %s deleted. A new session will start on next launch.\n", oldID)) +
				m.styles.HintText.Render("Note: Current session remains active until you exit."),
			ClearChat: true,
			Raw:       true,
		}
	}

	// Show current session info
	var sb strings.Builder
	sb.WriteString("Session Information:\n")
	sb.WriteString(m.styles.StatusModel.Render("  Session ID: "))
	if m.session != nil && m.session.SessionID != "" {
		sb.WriteString(m.session.SessionID + "\n")
	} else {
		sb.WriteString(m.styles.HintText.Render("(none - new session)") + "\n")
	}
	sb.WriteString(m.styles.StatusModel.Render("  Turns: ") + fmt.Sprintf("%d\n", m.conversationTurns))
	sb.WriteString(m.styles.StatusModel.Render("  Messages: ") + fmt.Sprintf("%d\n", len(m.history)))
	sb.WriteString("\n" + m.styles.HintText.Render("Use /session new to delete current session (takes effect on next launch)"))
	return CommandResult{
		Output: sb.String(),
		Raw:    true,
	}
}

func cmdDiff(_ *Model, _ []string) CommandResult {
	// Placeholder for diff functionality
	return CommandResult{
		Output: "No changes made in current session.",
	}
}

func cmdEditor(m *Model, _ []string) CommandResult {
	// Delegate to the model's openEditor — the returned tea.Cmd is handled
	// by the caller via a special non-nil Cmd field.
	if cmd := m.openEditor(); cmd != nil {
		return CommandResult{EditorCmd: cmd}
	}
	return CommandResult{}
}
