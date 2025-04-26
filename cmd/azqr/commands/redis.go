// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	scanCmd.AddCommand(redisCmd)
}

var redisCmd = &cobra.Command{
	Use:   "redis",
	Short: "Scan Azure Cache for Redis",
	Long:  "Scan Azure Cache for Redis",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		scan(cmd, []string{"redis"})
	},
}
