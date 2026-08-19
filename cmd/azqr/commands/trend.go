// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"fmt"

	"github.com/Azure/azqr/internal/history"
	"github.com/spf13/cobra"
)

func init() {
	trendCmd.Flags().String("history-file", "", "History JSONL path (defaults to the user config directory)")
	trendCmd.Flags().Int("last", 12, "Number of recent scans to show")
	trendCmd.Flags().String("scope", "", "Opaque scan scope ID (defaults to the latest scope)")
	trendCmd.Flags().String("format", "table", "Output format (table, json)")
	rootCmd.AddCommand(trendCmd)
}

var trendCmd = &cobra.Command{
	Use:   "trend",
	Short: "Show aggregate scan trends",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		override, _ := cmd.Flags().GetString("history-file")
		last, _ := cmd.Flags().GetInt("last")
		scopeID, _ := cmd.Flags().GetString("scope")
		format, _ := cmd.Flags().GetString("format")

		path, err := history.ResolvePath(override)
		if err != nil {
			return err
		}
		records, err := history.Read(path)
		if err != nil {
			return err
		}
		selected, selectedScope, err := history.Select(records, scopeID, last)
		if err != nil {
			return err
		}
		output, err := history.Render(selected, selectedScope, format)
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil
	},
}
