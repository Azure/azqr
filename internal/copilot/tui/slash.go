// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import "strings"

// slashAction represents the result of parsing a slash command.
type slashAction int

const (
	slashNone    slashAction = iota // not a slash command
	slashHelp                       // /help
	slashClear                      // /clear
	slashExit                       // /exit, /quit
	slashModel                      // /model
	slashNew                        // /new
	slashUnknown                    // unrecognised slash command
)

// parseSlash returns the action and any display message for a given input line.
// If input does not start with '/' it returns (slashNone, "").
func parseSlash(input string) (slashAction, string) {
	if !strings.HasPrefix(input, "/") {
		return slashNone, ""
	}
	// Strip leading slash and any extra whitespace.
	parts := strings.Fields(strings.TrimPrefix(input, "/"))
	if len(parts) == 0 {
		return slashNone, ""
	}
	switch strings.ToLower(parts[0]) {
	case "help", "h", "?":
		return slashHelp, ""
	case "clear", "cls":
		return slashClear, ""
	case "exit", "quit", "q":
		return slashExit, ""
	case "model":
		return slashModel, ""
	case "new":
		return slashNew, ""
	default:
		return slashUnknown, "Unknown command: /" + parts[0] + ". Type /help for available commands."
	}
}
