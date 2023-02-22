package embeded

import (
	"embed"
)

//go:embed *.png
var embededFiles embed.FS

// GetTemplates - Returns the template for the given name
func GetTemplates(templateName string) []byte {
	data, err := embededFiles.ReadFile(templateName)
	if err != nil {
		return nil
	}
	return data
}
