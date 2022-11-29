package report_templates

import (
	"embed"
	"log"
)

//go:embed *.md
var embededFiles embed.FS

func GetTemplates(templateName string) string {
	data, err := embededFiles.ReadFile(templateName)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}
