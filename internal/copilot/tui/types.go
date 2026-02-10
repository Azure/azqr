// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

// State represents the current UI state
type State int

const (
	StateReady State = iota
	StateThinking
	StateStreaming
)

// ToolInfo holds details about a tool execution
type ToolInfo struct {
	Name    string // Tool name
	Details string // Brief description of what it's doing (from args)
}

// ChunkType indicates the type of content chunk
type ChunkType int

const (
	ChunkReasoning ChunkType = iota
	ChunkAssistant
	ChunkTool
)

// Chunk represents a piece of content (reasoning, assistant message, or tool)
type Chunk struct {
	Type    ChunkType
	Content string   // Text content for reasoning/assistant
	Tool    ToolInfo // Tool info for tool chunks
}

// Activity tracks streaming content in order
type Activity struct {
	Chunks      []Chunk  // All content chunks in order
	CurrentTool ToolInfo // Tool currently running (not yet complete)
}

// AppendContent appends text to the last chunk of the given type, or creates a new chunk
func (a *Activity) AppendContent(chunkType ChunkType, content string) {
	if len(a.Chunks) > 0 && a.Chunks[len(a.Chunks)-1].Type == chunkType {
		a.Chunks[len(a.Chunks)-1].Content += content
	} else {
		a.Chunks = append(a.Chunks, Chunk{Type: chunkType, Content: content})
	}
}

// AddTool adds a completed tool chunk
func (a *Activity) AddTool(tool ToolInfo) {
	a.Chunks = append(a.Chunks, Chunk{Type: ChunkTool, Tool: tool})
}

// Message types for tea.Cmd
type activityMsg struct {
	reasoning   string // Append to reasoning
	toolStart   string // Tool name started
	toolDetails string // Tool details/args
	toolDone    bool   // Current tool completed
	assistant   string // Append to assistant message
}

type responseMsg struct{}

type errorMsg struct {
	err error
}

type ctrlCClearMsg struct{}

// statusLevel classifies the severity of a transient status bar message.
type statusLevel int

const (
	statusInfo statusLevel = iota
	statusWarn
	statusError
)

// statusMsg is a short-lived notice displayed in the status bar (not the viewport).
// After TTL the status bar clears automatically.
type statusMsg struct {
	text  string
	level statusLevel
}

// truncate shortens s to at most n bytes, appending "..." if truncated.
func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

// formatToolArgs extracts a brief description from tool arguments
func formatToolArgs(args interface{}) string {
	switch v := args.(type) {
	case map[string]interface{}:
		for _, key := range []string{"command", "cmd", "query", "path", "file", "subscription", "subscriptionId"} {
			if str, ok := v[key].(string); ok && str != "" {
				return truncate(str, 60)
			}
		}
	case string:
		return truncate(v, 60)
	}
	return ""
}

// formatToolDisplay formats a tool name and details for display
func formatToolDisplay(tool ToolInfo) string {
	if tool.Details != "" {
		return tool.Name + ": " + tool.Details
	}
	return tool.Name
}
