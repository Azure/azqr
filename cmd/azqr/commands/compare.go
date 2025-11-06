// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azqr/internal/comparator"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	compareCmd.Flags().StringP("file1", "a", "", "Path to first report file (required)")
	compareCmd.Flags().StringP("file2", "b", "", "Path to second report file (required)")
	compareCmd.Flags().StringP("format", "f", "excel", "Report format to compare (excel)")
	compareCmd.Flags().StringP("output", "o", "", "Output file for comparison results (optional, prints to console if not specified)")
	_ = compareCmd.MarkFlagRequired("file1")
	_ = compareCmd.MarkFlagRequired("file2")
	rootCmd.AddCommand(compareCmd)
}

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare two azqr scan reports",
	Long: `Compare two azqr scan reports to identify differences in recommendations and resources.
Supports Excel (.xlsx).

For Excel format, the comparison provides:
  - Row count comparison for each sheet
  - Detection of duplicate rows after row 4 (data rows)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		file1, _ := cmd.Flags().GetString("file1")
		file2, _ := cmd.Flags().GetString("file2")
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")

		// Validate files exist
		if _, err := os.Stat(file1); err != nil {
			return fmt.Errorf("cannot access file1: %w", err)
		}
		if _, err := os.Stat(file2); err != nil {
			return fmt.Errorf("cannot access file2: %w", err)
		}

		// Normalize format
		format = strings.ToLower(strings.TrimSpace(format))

		// Perform comparison based on format
		var resultStr string
		var err error

		switch format {
		case "excel", "xlsx":
			resultStr, err = compareExcelFiles(file1, file2)
		default:
			return fmt.Errorf("unsupported format: %s (supported: excel)", format)
		}

		if err != nil {
			return err
		}

		// Output results
		if output != "" {
			// Write to file
			if err := os.WriteFile(output, []byte(resultStr), 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			log.Info().Msgf("Comparison results written to: %s", output)
		} else {
			// Print to console
			fmt.Println(resultStr)
		}

		return nil
	},
}

// compareExcelFiles compares two Excel files and returns formatted results
func compareExcelFiles(file1, file2 string) (string, error) {
	log.Info().Msgf("Comparing Excel files...")

	result, err := comparator.CompareExcelFiles(file1, file2)
	if err != nil {
		return "", fmt.Errorf("failed to compare Excel files: %w", err)
	}

	return comparator.FormatComparisonResult(result), nil
}
