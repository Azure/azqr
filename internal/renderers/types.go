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
	keys := make([]string, 0, len(models.ScannerList))
	for key := range models.ScannerList {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, t := range models.ScannerList[key] {
			for _, rt := range t.ResourceTypes() {
				output += fmt.Sprintf("%s | %s", key, rt)
				output += fmt.Sprintln()
			}
		}
	}
	return output
}
