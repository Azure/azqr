// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/Azure/azqr/internal/renderers/pbi"
	"github.com/spf13/cobra"
)

func init() {
	pbiCmd.PersistentFlags().StringP("template-path", "p", "", "Path were the PowerBI template will be created")
	rootCmd.AddCommand(pbiCmd)
}

var pbiCmd = &cobra.Command{
	Use:   "pbi",
	Short: "Creates Power BI Desktop dashboard template",
	Long:  "Creates Power BI Desktop dashboard template",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("template-path")
		pbi.CreatePBIReport(path)
	},
}
