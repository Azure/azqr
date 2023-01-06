package templates

import (
	"embed"
)

//go:embed *.md
var embededFiles embed.FS

// GetTemplates - Returns the template for the given name
func GetTemplates(templateName string) string {
	data, err := embededFiles.ReadFile(templateName)
	if err != nil {
		return ""
	}
	return string(data)
}
