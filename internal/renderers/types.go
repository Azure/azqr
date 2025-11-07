package renderers

import (
	"fmt"
	"sort"

	"github.com/Azure/azqr/internal/models"
)

type SupportedTypes struct{}

func (t SupportedTypes) GetAll() string {
	output := fmt.Sprintln("Abbreviation  | Resource Type ")
	output += fmt.Sprintln("---|---")
	keys := make([]string, 0, len(models.ScannerFactoryList))
	for key := range models.ScannerFactoryList {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, factory := range models.ScannerFactoryList[key] {
			// Create temporary instance to get resource types
			scanner := factory()
			for _, rt := range scanner.ResourceTypes() {
				output += fmt.Sprintf("%s | %s", key, rt)
				output += fmt.Sprintln()
			}
		}
	}
	return output
}
