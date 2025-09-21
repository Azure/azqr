// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/Azure/azqr/internal/viewer"
	"github.com/spf13/cobra"
)

func init() {
	showCmd.Flags().StringP("file", "f", "", "Path to azqr report file (JSON or Excel format, required)")
	showCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	showCmd.Flags().Bool("open", true, "Open browser automatically")
	_ = showCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Launch local web dashboard for azqr reports",
	Long:  "Launches a local web server with an embedded dashboard to explore azqr reports. Supports both JSON reports (generated with --json flag) and Excel reports (default azqr output format).",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		port, _ := cmd.Flags().GetInt("port")
		open, _ := cmd.Flags().GetBool("open")

		if _, err := os.Stat(file); err != nil {
			return fmt.Errorf("cannot access file: %w", err)
		}
		ds, err := viewer.LoadDataStore(file)
		if err != nil {
			return err
		}
		if err := portAvailable(port); err != nil {
			return err
		}

		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Starting azqr dashboard on http://localhost%s (Ctrl+C to stop)\n", addr)
		if open {
			go func() { time.Sleep(300 * time.Millisecond); _ = openBrowser(fmt.Sprintf("http://localhost%s", addr)) }()
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return viewer.StartServer(ctx, addr, ds)
	},
}

func portAvailable(p int) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
	if err != nil {
		return fmt.Errorf("port %d not available: %w", p, err)
	}
	_ = l.Close()
	return nil
}
func openBrowser(url string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", url)
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		c = exec.Command("xdg-open", url)
	}
	if c == nil {
		return errors.New("unsupported platform for auto-open")
	}
	return c.Start()
}
