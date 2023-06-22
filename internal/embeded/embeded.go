// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package embeded

import (
	"embed"
)

//go:embed *.png *.pbit
var embededFiles embed.FS

// GetTemplates - Returns the template for the given name
func GetTemplates(templateName string) []byte {
	data, err := embededFiles.ReadFile(templateName)
	if err != nil {
		return nil
	}
	return data
}
