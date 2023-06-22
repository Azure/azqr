// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azqr

import (
	"github.com/Azure/azqr/internal/renderers"
	"github.com/spf13/cobra"
)

func init() {
	pbiCmd.PersistentFlags().StringP("excel-report", "x", "", "Path to azqr Excel report file")
	rootCmd.AddCommand(pbiCmd)
}

var pbiCmd = &cobra.Command{
	Use:   "pbi",
	Short: "Creates PowerBI desktop dashboard template",
	Long:  "Creates PowerBI desktop dashboard template",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		xlsx, _ := cmd.Flags().GetString("excel-report")
		renderers.CreatePBIReport(xlsx)
	},
}
