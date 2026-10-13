// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/pipeline"
	"github.com/Azure/azqr/internal/renderers"
	"github.com/rs/zerolog/log"
)

// Recorder implements pipeline.StageObserver, writing each stage's slice of
// ScanContext.ReportData into the store's typed core tables, in one
// transaction per stage. Since every pipeline stage replaces its owned
// ReportData field(s) wholesale rather than appending incrementally, each
// domain write here deletes that session's prior rows for the domain before
// re-inserting the current (complete) set, rather than tracking a per-domain
// append cursor.
type Recorder struct {
	store       *Store
	sessionID   string
	azqrVersion string
}

// NewRecorder returns a Recorder that persists stage output under sessionID.
func NewRecorder(s *Store, sessionID string, azqrVersion string) *Recorder {
	return &Recorder{store: s, sessionID: sessionID, azqrVersion: azqrVersion}
}

var _ pipeline.StageObserver = (*Recorder)(nil)

// StageCompleted implements pipeline.StageObserver. It records the stage's run outcome in
// session_stage_runs and, for a successfully executed stage, writes that stage's current
// ReportData domain slice(s) into the matching core tables. All writes for one stage happen
// in a single transaction; a write failure is logged and does not stop the scan.
func (r *Recorder) StageCompleted(ctx *pipeline.ScanContext, stageName string, position int, status pipeline.StageStatus, duration time.Duration, stageErr error) {
	tx, err := r.store.db.Begin()
	if err != nil {
		log.Error().Err(err).Str("stage", stageName).Msg("store: begin transaction failed, skipping persistence for stage")
		return
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	errMsg := sql.NullString{}
	if stageErr != nil {
		errMsg = sql.NullString{String: stageErr.Error(), Valid: true}
	}

	_, err = tx.Exec(
		r.store.rebind(`INSERT INTO session_stage_runs (session_id, stage_name, position, status, duration_ms, error_message)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT (session_id, stage_name) DO UPDATE SET
		   position = excluded.position,
		   status = excluded.status,
		   duration_ms = excluded.duration_ms,
		   error_message = excluded.error_message`),
		r.sessionID, stageName, position, string(status), duration.Milliseconds(), errMsg,
	)
	if err != nil {
		log.Error().Err(err).Str("stage", stageName).Msg("store: record stage run failed")
		return
	}

	if status == pipeline.StageStatusExecuted && ctx != nil && ctx.ReportData != nil {
		if err := r.writeStageData(tx, stageName, ctx.ReportData); err != nil {
			log.Error().Err(err).Str("stage", stageName).Msg("store: persist stage data failed")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Str("stage", stageName).Msg("store: commit stage transaction failed")
		return
	}
	committed = true
}

// writeStageData dispatches to the domain-specific writer(s) for stageName. Stages with no
// persistable domain data (for example Initialization, Subscription Discovery) fall through
// with no action.
func (r *Recorder) writeStageData(tx *sql.Tx, stageName string, rd *renderers.ReportData) error {
	switch stageName {
	case "Initialization":
		return r.writeTenantID(tx, rd.TenantID)
	case "Resource Discovery":
		return r.writeResources(tx, rd.Resources)
	case "Graph Scan":
		return r.writeGraphResults(tx, rd.Graph)
	case "Advisor Scan":
		return r.writeAdvisorResults(tx, rd.Advisor)
	case "Defender Status Scan":
		return r.writeDefenderResults(tx, rd.Defender)
	case "Defender Recommendations Scan":
		return r.writeDefenderRecommendations(tx, rd.DefenderRecommendations)
	case "Azure Policy Scan":
		return r.writeAzurePolicyResults(tx, rd.AzurePolicy)
	case "Arc-enabled SQL Server Scan":
		return r.writeArcSQLResults(tx, rd.ArcSQL)
	case "Cost Analysis Scan":
		return r.writeCostResults(tx, rd.Cost)
	case "Plugin Execution":
		return r.writePluginResults(tx, rd.PluginResults)
	default:
		return nil
	}
}

// writeTenantID updates the session's tenant_id column when a tenant ID was captured during
// the Initialization stage. An empty tenantID (capture failed or was skipped) leaves the
// column unchanged (NULL from CreateSession).
func (r *Recorder) writeTenantID(tx *sql.Tx, tenantID string) error {
	if tenantID == "" {
		return nil
	}
	if _, err := tx.Exec(r.store.rebind(`UPDATE sessions SET tenant_id = ? WHERE id = ?`), tenantID, r.sessionID); err != nil {
		return fmt.Errorf("update tenant_id for session %q: %w", r.sessionID, err)
	}
	return nil
}

func (r *Recorder) writeResources(tx *sql.Tx, resources []*models.Resource) error {
	if _, err := tx.Exec(r.store.rebind(`DELETE FROM resources WHERE session_id = ?`), r.sessionID); err != nil {
		return fmt.Errorf("delete prior resources: %w", err)
	}
	stmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO resources (
			session_id, resource_id, subscription_id, resource_group, type, location, name,
			sku_name, sku_tier, sku_family, sku_capacity, kind, sla, tags_json, excluded
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`),
	)
	if err != nil {
		return fmt.Errorf("prepare resources insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, res := range resources {
		tagsJSON, err := json.Marshal(res.Tags)
		if err != nil {
			return fmt.Errorf("marshal tags for resource %q: %w", res.ID, err)
		}
		_, err = stmt.Exec(
			r.sessionID, res.ID, res.SubscriptionID, res.ResourceGroup, res.Type, res.Location, res.Name,
			res.SkuName, res.SkuTier, res.SkuFamily, res.SkuCapacity, res.Kind, res.SLA, string(tagsJSON),
		)
		if err != nil {
			return fmt.Errorf("insert resource %q: %w", res.ID, err)
		}
	}
	return nil
}

func (r *Recorder) writeGraphResults(tx *sql.Tx, results []*models.GraphResult) error {
	if _, err := tx.Exec(r.store.rebind(`DELETE FROM recommendation_results WHERE session_id = ?`), r.sessionID); err != nil {
		return fmt.Errorf("delete prior recommendation_results: %w", err)
	}

	catalogStmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO recommendation_catalog (
			recommendation_id, source, category, impact, resource_type, recommendation,
			long_description, potential_benefits, learn_url, automation_available, content_hash,
			first_seen_azqr_version, last_seen_azqr_version
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (recommendation_id) DO UPDATE SET
		   source = excluded.source,
		   category = excluded.category,
		   impact = excluded.impact,
		   resource_type = excluded.resource_type,
		   recommendation = excluded.recommendation,
		   long_description = excluded.long_description,
		   potential_benefits = excluded.potential_benefits,
		   learn_url = excluded.learn_url,
		   automation_available = excluded.automation_available,
		   content_hash = excluded.content_hash,
		   last_seen_azqr_version = excluded.last_seen_azqr_version`),
	)
	if err != nil {
		return fmt.Errorf("prepare recommendation_catalog upsert: %w", err)
	}
	defer func() { _ = catalogStmt.Close() }()

	resultStmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO recommendation_results (
			session_id, recommendation_id, resource_id, subscription_id, subscription_name,
			resource_group, name, tags, param1, param2, param3, param4, param5
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare recommendation_results insert: %w", err)
	}
	defer func() { _ = resultStmt.Close() }()

	seenCatalog := make(map[string]bool, len(results))
	for _, g := range results {
		if !seenCatalog[g.RecommendationID] {
			seenCatalog[g.RecommendationID] = true
			hash := recommendationContentHash(g)
			_, err = catalogStmt.Exec(
				g.RecommendationID, g.Source, string(g.Category), string(g.Impact), g.ResourceType,
				g.Recommendation, g.LongDescription, g.PotentialBenefits, g.Learn, g.AutomationAvailable,
				hash, r.azqrVersion, r.azqrVersion,
			)
			if err != nil {
				return fmt.Errorf("upsert recommendation_catalog %q: %w", g.RecommendationID, err)
			}
		}

		_, err = resultStmt.Exec(
			r.sessionID, g.RecommendationID, g.ResourceID, g.SubscriptionID, g.SubscriptionName,
			g.ResourceGroup, g.Name, g.Tags, g.Param1, g.Param2, g.Param3, g.Param4, g.Param5,
		)
		if err != nil {
			return fmt.Errorf("insert recommendation_result %q/%q: %w", g.RecommendationID, g.ResourceID, err)
		}
	}
	return nil
}

// recommendationContentHash derives a stable hash of a recommendation's definitional content,
// used to detect when a catalog entry's wording has changed across azqr versions.
func recommendationContentHash(g *models.GraphResult) string {
	h := sha256.Sum256([]byte(g.Recommendation + "\x00" + g.LongDescription + "\x00" + g.PotentialBenefits + "\x00" + g.Learn))
	return hex.EncodeToString(h[:])
}

func (r *Recorder) writeAdvisorResults(tx *sql.Tx, results []*models.AdvisorResult) error {
	if _, err := tx.Exec(r.store.rebind(`DELETE FROM advisor_results WHERE session_id = ?`), r.sessionID); err != nil {
		return fmt.Errorf("delete prior advisor_results: %w", err)
	}
	stmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO advisor_results (
			session_id, recommendation_id, resource_id, subscription_id, subscription_name,
			type, name, category, impact, description
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare advisor_results insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, a := range results {
		_, err = stmt.Exec(
			r.sessionID, a.RecommendationID, a.ResourceID, a.SubscriptionID, a.SubscriptionName,
			a.Type, a.Name, a.Category, a.Impact, a.Description,
		)
		if err != nil {
			return fmt.Errorf("insert advisor_result %q/%q: %w", a.RecommendationID, a.ResourceID, err)
		}
	}
	return nil
}

func (r *Recorder) writeDefenderResults(tx *sql.Tx, results []*models.DefenderResult) error {
	if _, err := tx.Exec(r.store.rebind(`DELETE FROM defender_results WHERE session_id = ?`), r.sessionID); err != nil {
		return fmt.Errorf("delete prior defender_results: %w", err)
	}
	stmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO defender_results (session_id, subscription_id, subscription_name, name, tier)
		 VALUES (?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare defender_results insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, d := range results {
		if _, err := stmt.Exec(r.sessionID, d.SubscriptionID, d.SubscriptionName, d.Name, d.Tier); err != nil {
			return fmt.Errorf("insert defender_result %q/%q: %w", d.SubscriptionID, d.Name, err)
		}
	}
	return nil
}

func (r *Recorder) writeDefenderRecommendations(tx *sql.Tx, results []*models.DefenderRecommendation) error {
	if _, err := tx.Exec(r.store.rebind(`DELETE FROM defender_recommendations WHERE session_id = ?`), r.sessionID); err != nil {
		return fmt.Errorf("delete prior defender_recommendations: %w", err)
	}
	stmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO defender_recommendations (
			session_id, resource_id, recommendation_name, subscription_id, subscription_name,
			resource_group_name, resource_type, resource_name, category, recommendation_severity,
			action_description, remediation_description, az_portal_link
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare defender_recommendations insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, d := range results {
		_, err = stmt.Exec(
			r.sessionID, d.ResourceId, d.RecommendationName, d.SubscriptionId, d.SubscriptionName,
			d.ResourceGroupName, d.ResourceType, d.ResourceName, d.Category, d.RecommendationSeverity,
			d.ActionDescription, d.RemediationDescription, d.AzPortalLink,
		)
		if err != nil {
			return fmt.Errorf("insert defender_recommendation %q/%q: %w", d.ResourceId, d.RecommendationName, err)
		}
	}
	return nil
}

func (r *Recorder) writeAzurePolicyResults(tx *sql.Tx, results []*models.AzurePolicyResult) error {
	if _, err := tx.Exec(r.store.rebind(`DELETE FROM azure_policy_results WHERE session_id = ?`), r.sessionID); err != nil {
		return fmt.Errorf("delete prior azure_policy_results: %w", err)
	}
	stmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO azure_policy_results (
			session_id, resource_id, policy_assignment_id, subscription_id, subscription_name,
			policy_display_name, policy_description, compliance_state, type, name,
			resource_group_name, time_stamp, policy_definition_name, policy_definition_id,
			policy_assignment_name
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare azure_policy_results insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, p := range results {
		_, err = stmt.Exec(
			r.sessionID, p.ResourceID, p.PolicyAssignmentID, p.SubscriptionID, p.SubscriptionName,
			p.PolicyDisplayName, p.PolicyDescription, p.ComplianceState, p.Type, p.Name,
			p.ResourceGroupName, p.TimeStamp, p.PolicyDefinitionName, p.PolicyDefinitionID,
			p.PolicyAssignmentName,
		)
		if err != nil {
			return fmt.Errorf("insert azure_policy_result %q/%q: %w", p.ResourceID, p.PolicyAssignmentID, err)
		}
	}
	return nil
}

func (r *Recorder) writeArcSQLResults(tx *sql.Tx, results []*models.ArcSQLResult) error {
	if _, err := tx.Exec(r.store.rebind(`DELETE FROM arc_sql_results WHERE session_id = ?`), r.sessionID); err != nil {
		return fmt.Errorf("delete prior arc_sql_results: %w", err)
	}
	stmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO arc_sql_results (
			session_id, azure_arc_server, sql_instance, subscription_id, subscription_name,
			status, resource_group, version, build, patch_level, edition, v_cores, license,
			dps_status, tel_status, defender_status
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare arc_sql_results insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, a := range results {
		_, err = stmt.Exec(
			r.sessionID, a.AzureArcServer, a.SQLInstance, a.SubscriptionID, a.SubscriptionName,
			a.Status, a.ResourceGroup, a.Version, a.Build, a.PatchLevel, a.Edition, a.VCores,
			a.License, a.DPSStatus, a.TELStatus, a.DefenderStatus,
		)
		if err != nil {
			return fmt.Errorf("insert arc_sql_result %q/%q: %w", a.AzureArcServer, a.SQLInstance, err)
		}
	}
	return nil
}

func (r *Recorder) writeCostResults(tx *sql.Tx, results []*models.CostResult) error {
	if _, err := tx.Exec(r.store.rebind(`DELETE FROM cost_results WHERE session_id = ?`), r.sessionID); err != nil {
		return fmt.Errorf("delete prior cost_results: %w", err)
	}
	stmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO cost_results (
			session_id, subscription_id, service_name, subscription_name, value, currency,
			from_date, to_date
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare cost_results insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, c := range results {
		_, err = stmt.Exec(
			r.sessionID, c.SubscriptionID, c.ServiceName, c.SubscriptionName, c.Value, c.Currency,
			c.From.Format(time.RFC3339Nano), c.To.Format(time.RFC3339Nano),
		)
		if err != nil {
			return fmt.Errorf("insert cost_result %q/%q: %w", c.SubscriptionID, c.ServiceName, err)
		}
	}
	return nil
}

func (r *Recorder) writePluginResults(tx *sql.Tx, results []*renderers.PluginResult) error {
	if _, err := tx.Exec(r.store.rebind(`DELETE FROM plugin_result_rows WHERE session_id = ?`), r.sessionID); err != nil {
		return fmt.Errorf("delete prior plugin_result_rows: %w", err)
	}
	if _, err := tx.Exec(r.store.rebind(`DELETE FROM plugin_executions WHERE session_id = ?`), r.sessionID); err != nil {
		return fmt.Errorf("delete prior plugin_executions: %w", err)
	}

	execStmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO plugin_executions (session_id, plugin_name, status, finished_at)
		 VALUES (?, ?, 'completed', ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare plugin_executions insert: %w", err)
	}
	defer func() { _ = execStmt.Close() }()

	rowStmt, err := tx.Prepare(
		r.store.rebind(`INSERT INTO plugin_result_rows (session_id, plugin_name, sheet_name, row_ordinal, data)
		 VALUES (?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare plugin_result_rows insert: %w", err)
	}
	defer func() { _ = rowStmt.Close() }()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	seenPlugins := make(map[string]bool, len(results))

	for _, pr := range results {
		if !seenPlugins[pr.PluginName] {
			seenPlugins[pr.PluginName] = true
			if _, err := execStmt.Exec(r.sessionID, pr.PluginName, now); err != nil {
				return fmt.Errorf("insert plugin_execution %q: %w", pr.PluginName, err)
			}
		}

		if len(pr.Table) == 0 {
			continue
		}
		headers := pr.Table[0]
		for rowIdx, row := range pr.Table[1:] {
			doc := make(map[string]string, len(headers))
			for i, h := range headers {
				if i < len(row) {
					doc[h] = row[i]
				}
			}
			data, err := json.Marshal(doc)
			if err != nil {
				return fmt.Errorf("marshal plugin row %q/%q#%d: %w", pr.PluginName, pr.SheetName, rowIdx, err)
			}
			if _, err := rowStmt.Exec(r.sessionID, pr.PluginName, pr.SheetName, rowIdx, string(data)); err != nil {
				return fmt.Errorf("insert plugin_result_row %q/%q#%d: %w", pr.PluginName, pr.SheetName, rowIdx, err)
			}
		}
	}
	return nil
}
