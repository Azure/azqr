// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"sort"

	"github.com/Azure/azqr/internal/models"
	"github.com/spf13/cobra"

	// Import registry to trigger scanner registration
	_ "github.com/Azure/azqr/internal/scanners/registry"
)

func init() {
	abbreviations := make([]string, 0, len(models.ScannerList))
	for abbr := range models.ScannerList {
		abbreviations = append(abbreviations, abbr)
	}
	sort.Strings(abbreviations)

	// Dynamically create a command for each registered scanner
	for _, abbr := range abbreviations {
		scanners := models.ScannerList[abbr]
		if len(scanners) == 0 {
			continue
		}

		// Get service name from the first scanner in the list
		serviceName := scanners[0].ServiceName()

		// Create command for this scanner
		cmd := &cobra.Command{
			Use:   abbr,
			Short: fmt.Sprintf("Scan %s", serviceName),
			Long:  fmt.Sprintf("Scan %s", serviceName),
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				// Capture the abbreviation in the closure
				scannerAbbr := abbr
				scan(cmd, []string{scannerAbbr})
			},
		}

		scanCmd.AddCommand(cmd)
	}
}
