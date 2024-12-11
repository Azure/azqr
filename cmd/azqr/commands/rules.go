// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"fmt"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().BoolP("json", "j", false, "Output rules list in JSON format")
	rootCmd.AddCommand(rulesCmd)
}

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Print all recommendations",
	Long:  "Print all recommendations as markdown table",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		oj, _ := cmd.Flags().GetBool("json")
		output := renderers.GetAllRecommendations(!oj)
		fmt.Println(output)
	},
}
