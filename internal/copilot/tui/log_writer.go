// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"bytes"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// logLineMsg carries a single formatted log line from zerolog into the TUI.
type logLineMsg struct{ line string }

// tuiLogWriter is an io.Writer that buffers partial writes and forwards each
// complete line to a Bubbletea program as a logLineMsg.
type tuiLogWriter struct {
	program *tea.Program
	mu      sync.Mutex
	buf     bytes.Buffer
}

func (w *tuiLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf.Write(p)
	data := w.buf.Bytes()

	for {
		idx := bytes.IndexByte(data, '\n')
		if idx < 0 {
			break
		}
		line := bytes.TrimRight(data[:idx], "\r")
		if len(line) > 0 && w.program != nil {
			w.program.Send(logLineMsg{line: string(line)})
		}
		data = data[idx+1:]
	}

	w.buf.Reset()
	w.buf.Write(data)
	return len(p), nil
}
