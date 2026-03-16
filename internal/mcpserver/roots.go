// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// fileURIToPath converts a file:// URI to a local filesystem path.
// It handles both Unix paths (file:///home/user) and Windows paths
// (file:///C:/Users, file:///c%3A/Users), decoding percent-encoded characters.
func fileURIToPath(uri string) string {
	if !strings.HasPrefix(uri, "file://") {
		return uri
	}

	u, err := url.Parse(uri)
	if err != nil {
		// Fallback: strip "file://" and hope for the best
		return strings.TrimPrefix(uri, "file://")
	}

	// url.Parse decodes percent-encoded characters (e.g. %3A → :) in u.Path.
	path := u.Path

	// On Windows, MCP clients emit file:///C:/... so the parsed path is /C:/...
	// Strip the leading "/" before a Windows drive letter to get a valid path.
	if len(path) >= 3 && path[0] == '/' && isASCIILetter(path[1]) && path[2] == ':' {
		path = path[1:]
	}

	return path
}

// isASCIILetter reports whether b is an ASCII letter (a–z or A–Z).
func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// getCurrentFolder returns the first file:// root reported by the MCP client,
// or falls back to the process working directory.
func getCurrentFolder(ctx context.Context) (string, error) {
	if result, err := s.RequestRoots(ctx, mcp.ListRootsRequest{}); err == nil {
		for _, root := range result.Roots {
			if strings.HasPrefix(root.URI, "file://") {
				return fileURIToPath(root.URI), nil
			}
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	return dir, nil
}
