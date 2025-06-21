// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package mcpserver

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func currentWorkspace(ctx context.Context) string {
	result, err := s.RequestRoots(ctx, mcp.ListRootsRequest{})
	if err == nil {
		for _, root := range result.Roots {
			uri := root.URI
			if strings.HasPrefix(uri, "file://") {
				return strings.TrimPrefix(uri, "file://")
			}
		}
	}

	return ""
}

func getCurrentFolder(ctx context.Context) (string, error) {
	currentDir := currentWorkspace(ctx)

	if currentDir != "" {
		return currentDir, nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	return currentDir, nil
}
