package report_templates

import (
	"embed"
)

//go:embed *.md
var embededFiles embed.FS

func GetTemplates(templateName string) string {
	data, err := embededFiles.ReadFile(templateName)
	if err != nil {
		return ""
	}
	return string(data)
}
