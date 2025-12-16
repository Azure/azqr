// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package embeded

import (
	"testing"
)

func TestGetTemplates_ValidFile(t *testing.T) {
	// Test reading the embedded azqr.png file
	data := GetTemplates("azqr.png")

	if data == nil {
		t.Error("GetTemplates() returned nil for valid file azqr.png")
	}

	if len(data) == 0 {
		t.Error("GetTemplates() returned empty data for azqr.png")
	}

	// PNG files start with specific magic bytes
	if len(data) >= 4 {
		// PNG magic bytes: 89 50 4E 47
		if data[0] != 0x89 || data[1] != 0x50 || data[2] != 0x4E || data[3] != 0x47 {
			t.Error("GetTemplates() did not return valid PNG data")
		}
	}
}

func TestGetTemplates_InvalidFile(t *testing.T) {
	// Test reading a non-existent file
	data := GetTemplates("nonexistent.png")

	if data != nil {
		t.Error("GetTemplates() should return nil for non-existent file")
	}
}

func TestGetTemplates_EmptyFilename(t *testing.T) {
	// Test with empty filename
	data := GetTemplates("")

	if data != nil {
		t.Error("GetTemplates() should return nil for empty filename")
	}
}
