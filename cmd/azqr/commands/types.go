// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"fmt"

	"github.com/Azure/azqr/internal/renderers"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(typesCmd)
}

var typesCmd = &cobra.Command{
	Use:   "types",
	Short: "Print all supported azure resource types",
	Long:  "Print all supported azure resource types",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		st := renderers.SupportedTypes{}
		output := st.GetAll()
		fmt.Println(output)
	},
}
