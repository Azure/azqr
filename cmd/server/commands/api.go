// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Package commands provides the cobra command to start the azqr REST API server.
package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Azure/azqr/cmd/server/api"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(apiCmd)
}

// apiCmd represents the REST API server command.
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the azqr REST API server",
	Long:  "Start the azqr REST API server to expose supported commands as REST endpoints.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		port := os.Getenv("AZQR_API_PORT")
		if port == "" {
			port = "8080"
		}
		mux := http.NewServeMux()
		api.RegisterRoutes(mux)
		fmt.Printf("azqr REST API server started on :%s\n", port)
		http.ListenAndServe(":"+port, mux)
	},
}
