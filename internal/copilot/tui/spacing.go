// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package tui

import (
	"strings"
	"unicode/utf8"
)

// WrapText wraps text to the given width, preserving existing line breaks.
// It wraps on word boundaries when possible.
func WrapText(text string, width int) string {
	if width-4 < 20 {
		width = 20
	}
	var result strings.Builder
	writeLine := func(line string) {
		if result.Len() > 0 {
			result.WriteByte('\n')
		}
		result.WriteString(line)
	}
	for _, line := range strings.Split(text, "\n") {
		if utf8.RuneCountInString(line) <= width {
			writeLine(line)
			continue
		}
		// Wrap long line on word boundaries
		words := strings.Fields(line)
		currentLine := ""
		for _, word := range words {
			if currentLine == "" {
				currentLine = word
			} else if utf8.RuneCountInString(currentLine)+1+utf8.RuneCountInString(word) <= width {
				currentLine += " " + word
			} else {
				writeLine(currentLine)
				currentLine = word
			}
		}
		if currentLine != "" {
			writeLine(currentLine)
		}
	}
	return result.String()
}
