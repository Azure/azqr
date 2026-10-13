// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azqr/internal/renderers/csv"
	"github.com/Azure/azqr/internal/renderers/excel"
	"github.com/Azure/azqr/internal/renderers/json"
	"github.com/Azure/azqr/internal/renderers/sarif"
	"github.com/Azure/azqr/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	sessionListCmd.Flags().StringP("status", "", "", "Filter by session status (running, completed, partial, failed, cancelled)")
	sessionListCmd.Flags().StringP("kind", "", "", "Filter by session kind (full, plugin_only)")
	sessionListCmd.Flags().IntP("limit", "", 20, "Maximum number of sessions to return")
	sessionListCmd.Flags().IntP("offset", "", 0, "Number of sessions to skip (for pagination)")
	sessionCmd.AddCommand(sessionListCmd)

	sessionCmd.AddCommand(sessionShowCmd)

	sessionExportCmd.Flags().StringP("output-name", "o", "", "Output file name without extension (default: the session id)")
	sessionExportCmd.Flags().BoolP("xlsx", "", true, "Export an Excel report (default)")
	sessionExportCmd.Flags().BoolP("csv", "", false, "Export CSV reports")
	sessionExportCmd.Flags().BoolP("json", "", false, "Export a JSON report")
	sessionExportCmd.Flags().BoolP("sarif", "", false, "Export a SARIF 2.1.0 report")
	sessionExportCmd.Flags().BoolP("mask", "m", true, "Mask the subscription id in the report (default)")
	sessionCmd.AddCommand(sessionExportCmd)

	sessionCmd.AddCommand(sessionDeleteCmd)

	sessionPruneCmd.Flags().DurationP("max-age", "", store.DefaultPruneMaxAge, "Remove sessions older than this duration (0 disables the age guard)")
	sessionPruneCmd.Flags().IntP("max-count", "", store.DefaultPruneMaxCount, "Keep at most this many of the newest sessions (0 disables the count guard)")
	sessionPruneCmd.Flags().BoolP("dry-run", "", false, "Preview sessions that would be pruned without deleting anything")
	sessionPruneCmd.Flags().StringP("export-dir", "", "", "Export each session to an Excel report in this directory before deleting it")
	sessionCmd.AddCommand(sessionPruneCmd)

	rootCmd.AddCommand(sessionCmd)
}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Inspect and manage recorded scan sessions",
	Long:  "List, show, export, and delete scan sessions persisted in the azqr SQLite session store.",
}

var sessionListCmd = &cobra.Command{
	Use:          "list",
	Short:        "List recorded scan sessions",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		statusFlag, _ := cmd.Flags().GetString("status")
		kindFlag, _ := cmd.Flags().GetString("kind")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		return withStore(cmd, func(st *store.Store) error {
			filter := store.SessionFilter{
				Status: store.SessionStatus(statusFlag),
				Kind:   store.SessionKind(kindFlag),
				Limit:  limit,
				Offset: offset,
			}
			sessions, err := st.ListSessions(filter)
			if err != nil {
				return err
			}
			if len(sessions) == 0 {
				fmt.Println("No sessions found.")
				return nil
			}

			fmt.Printf("%-28s %-12s %-8s %-10s %-20s %s\n", "ID", "KIND", "TRIGGER", "STATUS", "STARTED", "DURATION")
			for _, sess := range sessions {
				duration := "-"
				if sess.DurationMs != nil {
					duration = (time.Duration(*sess.DurationMs) * time.Millisecond).String()
				}
				fmt.Printf("%-28s %-12s %-8s %-10s %-20s %s\n",
					sess.ID, sess.Kind, sess.Trigger, sess.Status,
					sess.StartedAt.Format(time.RFC3339), duration)
			}
			return nil
		})
	},
}

var sessionShowCmd = &cobra.Command{
	Use:          "show <session-id>",
	Short:        "Show details for one recorded scan session",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := args[0]
		return withStore(cmd, func(st *store.Store) error {
			sess, err := st.GetSession(sessionID)
			if err != nil {
				return err
			}

			fmt.Printf("ID:           %s\n", sess.ID)
			fmt.Printf("Kind:         %s\n", sess.Kind)
			fmt.Printf("Trigger:      %s\n", sess.Trigger)
			fmt.Printf("Status:       %s\n", sess.Status)
			fmt.Printf("AZQR Version: %s\n", sess.AzqrVersion)
			fmt.Printf("Started:      %s\n", sess.StartedAt.Format(time.RFC3339))
			if sess.FinishedAt != nil {
				fmt.Printf("Finished:     %s\n", sess.FinishedAt.Format(time.RFC3339))
			}
			if sess.DurationMs != nil {
				fmt.Printf("Duration:     %s\n", (time.Duration(*sess.DurationMs) * time.Millisecond).String())
			}
			if sess.ErrorMessage != "" {
				fmt.Printf("Error:        %s\n", sess.ErrorMessage)
			}
			if len(sess.Params.Subscriptions) > 0 {
				fmt.Printf("Subscriptions: %s\n", strings.Join(sess.Params.Subscriptions, ", "))
			}
			if len(sess.Params.ManagementGroups) > 0 {
				fmt.Printf("Management Groups: %s\n", strings.Join(sess.Params.ManagementGroups, ", "))
			}
			if len(sess.Params.ResourceGroups) > 0 {
				fmt.Printf("Resource Groups: %s\n", strings.Join(sess.Params.ResourceGroups, ", "))
			}

			runs, err := st.ListStageRuns(sessionID)
			if err != nil {
				return err
			}
			if len(runs) > 0 {
				fmt.Printf("\n%-32s %-10s %-10s %s\n", "STAGE", "STATUS", "DURATION", "ERROR")
				for _, run := range runs {
					duration := "-"
					if run.DurationMs != nil {
						duration = (time.Duration(*run.DurationMs) * time.Millisecond).String()
					}
					fmt.Printf("%-32s %-10s %-10s %s\n", run.StageName, run.Status, duration, run.ErrorMessage)
				}
			}
			return nil
		})
	},
}

var sessionExportCmd = &cobra.Command{
	Use:          "export <session-id>",
	Short:        "Export a recorded scan session to report file(s)",
	Long:         "Reconstructs a session's stored results and renders them through the same exporters a live scan uses (Excel, CSV, JSON, SARIF).",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := args[0]
		outputName, _ := cmd.Flags().GetString("output-name")
		xlsx, _ := cmd.Flags().GetBool("xlsx")
		csvEnabled, _ := cmd.Flags().GetBool("csv")
		jsonEnabled, _ := cmd.Flags().GetBool("json")
		sarifEnabled, _ := cmd.Flags().GetBool("sarif")
		mask, _ := cmd.Flags().GetBool("mask")

		if outputName == "" {
			outputName = sessionID
		}

		return withStore(cmd, func(st *store.Store) error {
			sess, err := st.GetSession(sessionID)
			if err != nil {
				return err
			}

			rd, err := st.LoadReportData(sessionID, outputName, mask)
			if err != nil {
				return fmt.Errorf("load session %q report data: %w", sessionID, err)
			}

			if xlsx {
				excel.CreateExcelReport(rd)
			}
			if csvEnabled {
				csv.CreateCsvReport(rd)
			}
			if jsonEnabled {
				json.CreateJsonReport(rd)
			}
			if sarifEnabled {
				if err := sarif.CreateReport(rd, sess.AzqrVersion); err != nil {
					return fmt.Errorf("generate sarif report for session %q: %w", sessionID, err)
				}
			}

			fmt.Printf("Session %s exported to %s.*\n", sessionID, outputName)
			return nil
		})
	},
}

var sessionDeleteCmd = &cobra.Command{
	Use:          "delete <session-id>",
	Short:        "Delete a recorded scan session and all its persisted data",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := args[0]
		return withStore(cmd, func(st *store.Store) error {
			if err := st.DeleteSession(sessionID); err != nil {
				return err
			}
			fmt.Printf("Session %s deleted.\n", sessionID)
			return nil
		})
	},
}

var sessionPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove old scan sessions according to a retention policy",
	Long: fmt.Sprintf(`Remove scan sessions according to an age-based and/or count-based retention policy.
Pruning cascades through the same foreign-key relationships as "session delete", so a pruned
session's resources, recommendations, and plugin results are removed along with it.

Defaults: --max-age=%s, --max-count=%d. Set either flag to 0 to disable that guard; at least
one of --max-age/--max-count must remain positive.

Pass --export-dir to export each session to an Excel report in that directory immediately
before it is deleted, as a safety net against losing data pruning would otherwise discard.`, store.DefaultPruneMaxAge, store.DefaultPruneMaxCount),
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		maxAge, _ := cmd.Flags().GetDuration("max-age")
		maxCount, _ := cmd.Flags().GetInt("max-count")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		exportDir, _ := cmd.Flags().GetString("export-dir")

		return withStore(cmd, func(st *store.Store) error {
			policy := store.PrunePolicy{MaxAge: maxAge, MaxCount: maxCount, DryRun: dryRun}
			if exportDir != "" {
				if !dryRun {
					if err := os.MkdirAll(exportDir, 0750); err != nil {
						return fmt.Errorf("create export directory %q: %w", exportDir, err)
					}
				}
				policy.BeforeDelete = func(sess *store.Session) error {
					outputName := filepath.Join(exportDir, sess.ID)
					rd, err := st.LoadReportData(sess.ID, outputName, true)
					if err != nil {
						return fmt.Errorf("load report data for export: %w", err)
					}
					excel.CreateExcelReport(rd)
					return nil
				}
			}

			result, err := st.PruneSessions(policy)
			if err != nil {
				return err
			}

			if len(result.Removed) == 0 {
				fmt.Println("No sessions matched the prune policy.")
				return nil
			}

			verb := "Pruned"
			if result.DryRun {
				verb = "Would prune"
			}
			fmt.Printf("%s %d session(s):\n", verb, len(result.Removed))
			fmt.Printf("%-28s %-12s %-10s %s\n", "ID", "KIND", "STATUS", "STARTED")
			for _, sess := range result.Removed {
				fmt.Printf("%-28s %-12s %-10s %s\n", sess.ID, sess.Kind, sess.Status, sess.StartedAt.Format(time.RFC3339))
			}
			if exportDir != "" && !result.DryRun {
				fmt.Printf("Exported each pruned session to %s before deletion.\n", exportDir)
			}
			return nil
		})
	},
}

// withStore opens the session store configured via --store-path (falling back to
// store.DefaultPath() when unset), passes it to fn, and closes it afterward. Unlike
// store.OpenForScan (used by scan commands, where store unavailability must never block a
// scan), session subcommands exist purely to read/manage the store, so an open failure here is
// a real, user-facing error rather than something to silently fall back from.
func withStore(cmd *cobra.Command, fn func(*store.Store) error) error {
	storePath, _ := cmd.Flags().GetString("store-path")
	st, err := store.Open(storePath)
	if err != nil {
		return fmt.Errorf("open session store: %w", err)
	}
	defer func() { _ = st.Close() }()
	return fn(st)
}
