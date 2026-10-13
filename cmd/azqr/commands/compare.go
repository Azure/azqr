// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package commands

import (
	"fmt"
	"os"

	"github.com/Azure/azqr/internal/store"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	compareCmd.Flags().StringP("output", "o", "", "Output file for comparison results (optional, prints to console if not specified)")
	compareCmd.Flags().StringP("source", "", "", "Source session ID to compare")
	compareCmd.Flags().StringP("target", "", "", "Target session ID to compare")
	rootCmd.AddCommand(compareCmd)
}

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare two azqr scan sessions",
	Long: `Compare two azqr scan sessions to identify differences in recommendations and resources.
Sessions are persisted in the azqr SQLite session store.

The comparison provides:
  - Added/removed resources (by resource ID)
  - Added/removed/changed recommendation hit counts (by recommendation ID)

The comparison is refused (with no override flag) when both sessions have a scan scope
(scope_id) or tenant (tenant_id) recorded and those values differ, since comparing two sessions
scanned against different scopes or Azure AD tenants would produce a misleading diff. A session
predating this feature, or one where tenant capture failed, simply skips that specific check.`,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		source, _ := cmd.Flags().GetString("source")
		target, _ := cmd.Flags().GetString("target")

		if source == "" || target == "" {
			return fmt.Errorf("both --source and --target are required")
		}

		resultStr, err := compareSessions(cmd, source, target)
		if err != nil {
			return err
		}

		// Output results
		if output != "" {
			// Write to file
			if err := os.WriteFile(output, []byte(resultStr), 0600); err != nil {
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

// compareSessions compares two persisted sessions by ID and returns formatted results.
func compareSessions(cmd *cobra.Command, sourceID, targetID string) (string, error) {
	log.Info().Msgf("Comparing sessions %s and %s...", sourceID, targetID)

	var resultStr string
	err := withStore(cmd, func(st *store.Store) error {
		result, err := st.CompareSessions(sourceID, targetID)
		if err != nil {
			return fmt.Errorf("failed to compare sessions: %w", err)
		}
		resultStr = formatSessionComparisonResult(result)
		return nil
	})
	return resultStr, err
}

// formatSessionComparisonResult renders a store.SessionComparisonResult as a human-readable report.
func formatSessionComparisonResult(result *store.SessionComparisonResult) string {
	output := "Session Comparison Results:\n"
	output += fmt.Sprintf("Source Session: %s\n", result.SourceSessionID)
	output += fmt.Sprintf("Target Session: %s\n\n", result.TargetSessionID)

	var added, removed []store.ResourceChange
	for _, rc := range result.ResourceChanges {
		if rc.ChangeType == "added" {
			added = append(added, rc)
		} else {
			removed = append(removed, rc)
		}
	}

	output += fmt.Sprintf("Resources: %d added, %d removed\n", len(added), len(removed))
	if len(added) > 0 {
		output += "\nAdded Resources:\n"
		for _, rc := range added {
			output += fmt.Sprintf("  + %-60s %-30s %s\n", rc.ResourceID, rc.Type, rc.Name)
		}
	}
	if len(removed) > 0 {
		output += "\nRemoved Resources:\n"
		for _, rc := range removed {
			output += fmt.Sprintf("  - %-60s %-30s %s\n", rc.ResourceID, rc.Type, rc.Name)
		}
	}

	if len(result.RecommendationChanges) > 0 {
		output += "\nRecommendation Changes:\n"
		output += fmt.Sprintf("%-20s | %-8s | %-50s | %10s | %10s | %10s\n",
			"Category", "Change", "Recommendation", "Old", "New", "Diff")
		output += "---------------------------------------------------------------------------------------------------------\n"
		for _, change := range result.RecommendationChanges {
			diffStr := fmt.Sprintf("%+d", change.ImpactedResourcesDiff)
			output += fmt.Sprintf("%-20s | %-8s | %-50s | %10d | %10d | %10s\n",
				change.Category, change.ChangeType, change.Recommendation,
				change.OldImpactedResources, change.NewImpactedResources, diffStr)
		}
	}

	return output
}
