// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/Azure/azqr/internal/findings"
	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/renderers"
)

// StageRun is one recorded session_stage_runs row, describing a single pipeline stage's
// outcome within a session.
type StageRun struct {
	StageName    string
	Position     int
	Status       string
	DurationMs   *int64
	ErrorMessage string
}

// SessionFilter narrows ListSessions results. A zero value matches every session, newest first.
type SessionFilter struct {
	Status SessionStatus // empty matches any status
	Kind   SessionKind   // empty matches any kind
	Since  *time.Time    // sessions started at or after this time
	Until  *time.Time    // sessions started at or before this time
	Limit  int           // 0 means no limit
	Offset int
}

// ListSessions returns sessions matching filter, ordered by started_at descending (newest
// first).
func (s *Store) ListSessions(filter SessionFilter) ([]*Session, error) {
	query := `SELECT id, kind, trigger, status, azqr_version, scan_params_json, started_at,
		finished_at, duration_ms, error_message, scope_id, tenant_id FROM sessions WHERE 1 = 1`
	args := []any{}

	if filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, string(filter.Status))
	}
	if filter.Kind != "" {
		query += ` AND kind = ?`
		args = append(args, string(filter.Kind))
	}
	if filter.Since != nil {
		query += ` AND started_at >= ?`
		args = append(args, filter.Since.UTC().Format(time.RFC3339Nano))
	}
	if filter.Until != nil {
		query += ` AND started_at <= ?`
		args = append(args, filter.Until.UTC().Format(time.RFC3339Nano))
	}

	query += ` ORDER BY started_at DESC`
	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += ` OFFSET ?`
			args = append(args, filter.Offset)
		}
	}

	rows, err := s.db.Query(s.rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	sessions := []*Session{}
	for rows.Next() {
		sess, err := scanSessionRow(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return sessions, nil
}

// GetSession returns the session with the given id, or an error if no such session exists.
func (s *Store) GetSession(id string) (*Session, error) {
	row := s.db.QueryRow(
		s.rebind(`SELECT id, kind, trigger, status, azqr_version, scan_params_json, started_at,
		 finished_at, duration_ms, error_message, scope_id, tenant_id FROM sessions WHERE id = ?`), id,
	)
	sess, err := scanSessionRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("session %q: not found", id)
		}
		return nil, err
	}
	return sess, nil
}

// sessionRowScanner is satisfied by both *sql.Row and *sql.Rows, letting scanSessionRow
// back both GetSession (single row) and ListSessions (row set).
type sessionRowScanner interface {
	Scan(dest ...any) error
}

func scanSessionRow(row sessionRowScanner) (*Session, error) {
	var (
		sess                              Session
		kind, trigger, status             string
		azqrVersion, paramsJSON, startRaw string
		finishedAtRaw, errMsg             sql.NullString
		scopeID, tenantID                 sql.NullString
		durationMs                        sql.NullInt64
	)

	if err := row.Scan(&sess.ID, &kind, &trigger, &status, &azqrVersion, &paramsJSON,
		&startRaw, &finishedAtRaw, &durationMs, &errMsg, &scopeID, &tenantID); err != nil {
		return nil, fmt.Errorf("scan session row: %w", err)
	}

	sess.Kind = SessionKind(kind)
	sess.Trigger = SessionTrigger(trigger)
	sess.Status = SessionStatus(status)
	sess.AzqrVersion = azqrVersion
	sess.ErrorMessage = errMsg.String
	sess.ScopeID = scopeID.String
	sess.TenantID = tenantID.String

	if err := json.Unmarshal([]byte(paramsJSON), &sess.Params); err != nil {
		return nil, fmt.Errorf("session %q: unmarshal scan_params_json: %w", sess.ID, err)
	}

	startedAt, err := time.Parse(time.RFC3339Nano, startRaw)
	if err != nil {
		return nil, fmt.Errorf("session %q: parse started_at %q: %w", sess.ID, startRaw, err)
	}
	sess.StartedAt = startedAt

	if finishedAtRaw.Valid {
		finishedAt, err := time.Parse(time.RFC3339Nano, finishedAtRaw.String)
		if err != nil {
			return nil, fmt.Errorf("session %q: parse finished_at %q: %w", sess.ID, finishedAtRaw.String, err)
		}
		sess.FinishedAt = &finishedAt
	}
	if durationMs.Valid {
		sess.DurationMs = &durationMs.Int64
	}

	return &sess, nil
}

// ListStageRuns returns a session's recorded stage runs, ordered by pipeline position.
func (s *Store) ListStageRuns(sessionID string) ([]StageRun, error) {
	rows, err := s.db.Query(
		s.rebind(`SELECT stage_name, position, status, duration_ms, error_message
		 FROM session_stage_runs WHERE session_id = ? ORDER BY position`), sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list stage runs for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	runs := []StageRun{}
	for rows.Next() {
		var run StageRun
		var durationMs sql.NullInt64
		var errMsg sql.NullString
		if err := rows.Scan(&run.StageName, &run.Position, &run.Status, &durationMs, &errMsg); err != nil {
			return nil, fmt.Errorf("scan stage run row: %w", err)
		}
		if durationMs.Valid {
			run.DurationMs = &durationMs.Int64
		}
		run.ErrorMessage = errMsg.String
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list stage runs for session %q: %w", sessionID, err)
	}
	return runs, nil
}

// DeleteSession removes a session and, via ON DELETE CASCADE foreign keys declared in
// migrations/00001_init.sql, every row recorded under it (resources, recommendation_results,
// advisor_results, defender_results, defender_recommendations, azure_policy_results,
// arc_sql_results, cost_results, plugin_executions, plugin_result_rows, session_stage_runs).
// recommendation_catalog rows are shared across sessions by recommendation_id and are never
// deleted here.
func (s *Store) DeleteSession(id string) error {
	res, err := s.db.Exec(s.rebind(`DELETE FROM sessions WHERE id = ?`), id)
	if err != nil {
		return fmt.Errorf("delete session %q: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete session %q: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("session %q: not found", id)
	}
	return nil
}

// LoadReportData reconstructs a *renderers.ReportData from a session's stored rows, so the
// existing xlsx/csv/json/sarif exporters can render it exactly as they would a live scan's
// in-memory result.
//
// Two known, accepted gaps versus a live scan's ReportData:
//   - The Recommendations catalog used to build rd.Summary (and therefore the "Recommendations"
//     export sheet) is derived only from recommendation_catalog rows that have at least one
//     recommendation_results row for this session. Recommendation rules that matched zero
//     resources during this particular session are not recreated, so they will not appear as
//     "Implemented: true, Impacted: 0" rows, unlike a live scan's full YAML-rule-derived view.
//   - Plugin result tables are reconstructed with columns in alphabetical order (the order
//     encoding/json uses when marshaling plugin_result_rows.data), which may differ from the
//     original column order produced by the live scan.
func (s *Store) LoadReportData(sessionID, outputFileName string, mask bool) (*renderers.ReportData, error) {
	sess, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Exporters (e.g. excel.renderSheet) gate each sheet on rd.Stages.IsStageEnabled and
	// dereference rd.Stages unconditionally, so it must never be left nil here.
	stages := models.NewStageConfigs()
	for _, stageName := range sess.Params.EnabledStages {
		_ = stages.EnableStage(stageName)
	}

	rd := renderers.NewReportData(outputFileName, mask, stages)
	rd.ScopeID = sessionID

	loaders := []func(*Store, string, *renderers.ReportData) error{
		(*Store).loadResources,
		(*Store).loadGraphResults,
		(*Store).loadAdvisorResults,
		(*Store).loadDefenderResults,
		(*Store).loadDefenderRecommendations,
		(*Store).loadAzurePolicyResults,
		(*Store).loadArcSQLResults,
		(*Store).loadCostResults,
		(*Store).loadPluginResults,
	}
	for _, load := range loaders {
		if err := load(s, sessionID, &rd); err != nil {
			return nil, err
		}
	}

	rd.ResourceTypeCount = computeResourceTypeCount(rd.Resources)
	rd.Summary = findings.Build(rd.Recommendations, rd.Graph, len(rd.Resources))

	return &rd, nil
}

func computeResourceTypeCount(resources []*models.Resource) []*models.ResourceTypeCount {
	type key struct{ subscription, resourceType string }
	counts := map[key]float64{}
	order := []key{}
	for _, r := range resources {
		k := key{subscription: r.SubscriptionID, resourceType: r.Type}
		if _, exists := counts[k]; !exists {
			order = append(order, k)
		}
		counts[k]++
	}

	result := make([]*models.ResourceTypeCount, 0, len(order))
	for _, k := range order {
		result = append(result, &models.ResourceTypeCount{
			Subscription: k.subscription,
			ResourceType: k.resourceType,
			Count:        counts[k],
		})
	}
	return result
}

const (
	resourceSelectColumns = `SELECT resource_id, subscription_id, resource_group, type, location, name,
		 sku_name, sku_tier, sku_family, sku_capacity, kind, sla, tags_json
		 FROM resources WHERE session_id = ? AND excluded = 0`

	graphResultSelectColumns = `SELECT rr.recommendation_id, rr.resource_id, rr.subscription_id, rr.subscription_name,
		 rr.resource_group, rr.name, rr.tags, rr.param1, rr.param2, rr.param3, rr.param4, rr.param5,
		 rc.source, rc.category, rc.impact, rc.resource_type, rc.recommendation,
		 rc.long_description, rc.potential_benefits, rc.learn_url, rc.automation_available
		 FROM recommendation_results rr
		 JOIN recommendation_catalog rc ON rc.recommendation_id = rr.recommendation_id
		 WHERE rr.session_id = ?`
)

// scanResourceRow scans one row produced by resourceSelectColumns into a *models.Resource,
// shared by loadResources (unbounded) and LoadResourcesPage (bounded).
func scanResourceRow(rows *sql.Rows) (*models.Resource, error) {
	var res models.Resource
	var tagsJSON string
	if err := rows.Scan(&res.ID, &res.SubscriptionID, &res.ResourceGroup, &res.Type,
		&res.Location, &res.Name, &res.SkuName, &res.SkuTier, &res.SkuFamily,
		&res.SkuCapacity, &res.Kind, &res.SLA, &tagsJSON); err != nil {
		return nil, fmt.Errorf("scan resource row: %w", err)
	}
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &res.Tags); err != nil {
			return nil, fmt.Errorf("resource %q: unmarshal tags_json: %w", res.ID, err)
		}
	}
	return &res, nil
}

// scanGraphResultRow scans one row produced by graphResultSelectColumns into a
// *models.GraphResult, shared by loadGraphResults (unbounded) and LoadGraphResultsPage
// (bounded).
func scanGraphResultRow(rows *sql.Rows) (*models.GraphResult, error) {
	var g models.GraphResult
	var category, impact string
	if err := rows.Scan(&g.RecommendationID, &g.ResourceID, &g.SubscriptionID, &g.SubscriptionName,
		&g.ResourceGroup, &g.Name, &g.Tags, &g.Param1, &g.Param2, &g.Param3, &g.Param4, &g.Param5,
		&g.Source, &category, &impact, &g.ResourceType, &g.Recommendation,
		&g.LongDescription, &g.PotentialBenefits, &g.Learn, &g.AutomationAvailable); err != nil {
		return nil, fmt.Errorf("scan recommendation result row: %w", err)
	}
	g.Category = models.RecommendationCategory(category)
	g.Impact = models.RecommendationImpact(impact)
	return &g, nil
}

func (s *Store) loadResources(sessionID string, rd *renderers.ReportData) error {
	rows, err := s.db.Query(s.rebind(resourceSelectColumns), sessionID)
	if err != nil {
		return fmt.Errorf("load resources for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		res, err := scanResourceRow(rows)
		if err != nil {
			return err
		}
		rd.Resources = append(rd.Resources, res)
	}
	return rows.Err()
}

func (s *Store) loadGraphResults(sessionID string, rd *renderers.ReportData) error {
	rows, err := s.db.Query(s.rebind(graphResultSelectColumns), sessionID)
	if err != nil {
		return fmt.Errorf("load recommendation results for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	catalog := map[string]map[string]*models.GraphRecommendation{}
	for rows.Next() {
		g, err := scanGraphResultRow(rows)
		if err != nil {
			return err
		}
		rd.Graph = append(rd.Graph, g)

		lowerType := g.ResourceType
		if catalog[lowerType] == nil {
			catalog[lowerType] = map[string]*models.GraphRecommendation{}
		}
		if _, exists := catalog[lowerType][g.RecommendationID]; !exists {
			catalog[lowerType][g.RecommendationID] = &models.GraphRecommendation{
				RecommendationID:    g.RecommendationID,
				Recommendation:      g.Recommendation,
				Category:            string(g.Category),
				Impact:              string(g.Impact),
				ResourceType:        g.ResourceType,
				LongDescription:     g.LongDescription,
				PotentialBenefits:   g.PotentialBenefits,
				AutomationAvailable: g.AutomationAvailable,
				Source:              g.Source,
			}
		}
	}
	rd.Recommendations = catalog
	return rows.Err()
}

func (s *Store) loadAdvisorResults(sessionID string, rd *renderers.ReportData) error {
	rows, err := s.db.Query(
		s.rebind(`SELECT recommendation_id, resource_id, subscription_id, subscription_name, type, name,
		 category, impact, description FROM advisor_results WHERE session_id = ?`), sessionID,
	)
	if err != nil {
		return fmt.Errorf("load advisor results for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var a models.AdvisorResult
		if err := rows.Scan(&a.RecommendationID, &a.ResourceID, &a.SubscriptionID, &a.SubscriptionName,
			&a.Type, &a.Name, &a.Category, &a.Impact, &a.Description); err != nil {
			return fmt.Errorf("scan advisor result row: %w", err)
		}
		rd.Advisor = append(rd.Advisor, &a)
	}
	return rows.Err()
}

func (s *Store) loadDefenderResults(sessionID string, rd *renderers.ReportData) error {
	rows, err := s.db.Query(
		s.rebind(`SELECT subscription_id, subscription_name, name, tier FROM defender_results WHERE session_id = ?`), sessionID,
	)
	if err != nil {
		return fmt.Errorf("load defender results for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var d models.DefenderResult
		if err := rows.Scan(&d.SubscriptionID, &d.SubscriptionName, &d.Name, &d.Tier); err != nil {
			return fmt.Errorf("scan defender result row: %w", err)
		}
		rd.Defender = append(rd.Defender, &d)
	}
	return rows.Err()
}

func (s *Store) loadDefenderRecommendations(sessionID string, rd *renderers.ReportData) error {
	rows, err := s.db.Query(
		s.rebind(`SELECT resource_id, recommendation_name, subscription_id, subscription_name,
		 resource_group_name, resource_type, resource_name, category, recommendation_severity,
		 action_description, remediation_description, az_portal_link
		 FROM defender_recommendations WHERE session_id = ?`), sessionID,
	)
	if err != nil {
		return fmt.Errorf("load defender recommendations for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var d models.DefenderRecommendation
		if err := rows.Scan(&d.ResourceId, &d.RecommendationName, &d.SubscriptionId, &d.SubscriptionName,
			&d.ResourceGroupName, &d.ResourceType, &d.ResourceName, &d.Category, &d.RecommendationSeverity,
			&d.ActionDescription, &d.RemediationDescription, &d.AzPortalLink); err != nil {
			return fmt.Errorf("scan defender recommendation row: %w", err)
		}
		rd.DefenderRecommendations = append(rd.DefenderRecommendations, &d)
	}
	return rows.Err()
}

func (s *Store) loadAzurePolicyResults(sessionID string, rd *renderers.ReportData) error {
	rows, err := s.db.Query(
		s.rebind(`SELECT resource_id, policy_assignment_id, subscription_id, subscription_name,
		 policy_display_name, policy_description, compliance_state, type, name,
		 resource_group_name, time_stamp, policy_definition_name, policy_definition_id,
		 policy_assignment_name FROM azure_policy_results WHERE session_id = ?`), sessionID,
	)
	if err != nil {
		return fmt.Errorf("load azure policy results for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var p models.AzurePolicyResult
		if err := rows.Scan(&p.ResourceID, &p.PolicyAssignmentID, &p.SubscriptionID, &p.SubscriptionName,
			&p.PolicyDisplayName, &p.PolicyDescription, &p.ComplianceState, &p.Type, &p.Name,
			&p.ResourceGroupName, &p.TimeStamp, &p.PolicyDefinitionName, &p.PolicyDefinitionID,
			&p.PolicyAssignmentName); err != nil {
			return fmt.Errorf("scan azure policy result row: %w", err)
		}
		rd.AzurePolicy = append(rd.AzurePolicy, &p)
	}
	return rows.Err()
}

func (s *Store) loadArcSQLResults(sessionID string, rd *renderers.ReportData) error {
	rows, err := s.db.Query(
		s.rebind(`SELECT azure_arc_server, sql_instance, subscription_id, subscription_name, status,
		 resource_group, version, build, patch_level, edition, v_cores, license, dps_status,
		 tel_status, defender_status FROM arc_sql_results WHERE session_id = ?`), sessionID,
	)
	if err != nil {
		return fmt.Errorf("load arc sql results for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var a models.ArcSQLResult
		if err := rows.Scan(&a.AzureArcServer, &a.SQLInstance, &a.SubscriptionID, &a.SubscriptionName,
			&a.Status, &a.ResourceGroup, &a.Version, &a.Build, &a.PatchLevel, &a.Edition, &a.VCores,
			&a.License, &a.DPSStatus, &a.TELStatus, &a.DefenderStatus); err != nil {
			return fmt.Errorf("scan arc sql result row: %w", err)
		}
		rd.ArcSQL = append(rd.ArcSQL, &a)
	}
	return rows.Err()
}

func (s *Store) loadCostResults(sessionID string, rd *renderers.ReportData) error {
	rows, err := s.db.Query(
		s.rebind(`SELECT subscription_id, service_name, subscription_name, value, currency, from_date, to_date
		 FROM cost_results WHERE session_id = ?`), sessionID,
	)
	if err != nil {
		return fmt.Errorf("load cost results for session %q: %w", sessionID, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var c models.CostResult
		var fromRaw, toRaw string
		if err := rows.Scan(&c.SubscriptionID, &c.ServiceName, &c.SubscriptionName, &c.Value,
			&c.Currency, &fromRaw, &toRaw); err != nil {
			return fmt.Errorf("scan cost result row: %w", err)
		}
		if fromRaw != "" {
			from, err := time.Parse(time.RFC3339Nano, fromRaw)
			if err != nil {
				return fmt.Errorf("cost result %q/%q: parse from_date %q: %w", c.SubscriptionID, c.ServiceName, fromRaw, err)
			}
			c.From = from
		}
		if toRaw != "" {
			to, err := time.Parse(time.RFC3339Nano, toRaw)
			if err != nil {
				return fmt.Errorf("cost result %q/%q: parse to_date %q: %w", c.SubscriptionID, c.ServiceName, toRaw, err)
			}
			c.To = to
		}
		rd.Cost = append(rd.Cost, &c)
	}
	return rows.Err()
}

func (s *Store) loadPluginResults(sessionID string, rd *renderers.ReportData) error {
	execRows, err := s.db.Query(
		s.rebind(`SELECT plugin_name FROM plugin_executions WHERE session_id = ? ORDER BY plugin_name`), sessionID,
	)
	if err != nil {
		return fmt.Errorf("load plugin executions for session %q: %w", sessionID, err)
	}
	pluginNames := []string{}
	for execRows.Next() {
		var name string
		if err := execRows.Scan(&name); err != nil {
			_ = execRows.Close()
			return fmt.Errorf("scan plugin execution row: %w", err)
		}
		pluginNames = append(pluginNames, name)
	}
	if err := execRows.Err(); err != nil {
		_ = execRows.Close()
		return fmt.Errorf("load plugin executions for session %q: %w", sessionID, err)
	}
	_ = execRows.Close()

	for _, pluginName := range pluginNames {
		sheetRows, err := s.db.Query(
			s.rebind(`SELECT DISTINCT sheet_name FROM plugin_result_rows WHERE session_id = ? AND plugin_name = ? ORDER BY sheet_name`),
			sessionID, pluginName,
		)
		if err != nil {
			return fmt.Errorf("load plugin sheets for plugin %q: %w", pluginName, err)
		}
		sheetNames := []string{}
		for sheetRows.Next() {
			var name string
			if err := sheetRows.Scan(&name); err != nil {
				_ = sheetRows.Close()
				return fmt.Errorf("scan plugin sheet row: %w", err)
			}
			sheetNames = append(sheetNames, name)
		}
		if err := sheetRows.Err(); err != nil {
			_ = sheetRows.Close()
			return fmt.Errorf("load plugin sheets for plugin %q: %w", pluginName, err)
		}
		_ = sheetRows.Close()

		for _, sheetName := range sheetNames {
			table, err := s.loadPluginSheetTable(sessionID, pluginName, sheetName)
			if err != nil {
				return err
			}
			rd.PluginResults = append(rd.PluginResults, &renderers.PluginResult{
				PluginName: pluginName,
				SheetName:  sheetName,
				Table:      table,
			})
		}
	}
	return nil
}

func (s *Store) loadPluginSheetTable(sessionID, pluginName, sheetName string) ([][]string, error) {
	rows, err := s.db.Query(
		s.rebind(`SELECT data FROM plugin_result_rows WHERE session_id = ? AND plugin_name = ? AND sheet_name = ?
		 ORDER BY row_ordinal`), sessionID, pluginName, sheetName,
	)
	if err != nil {
		return nil, fmt.Errorf("load plugin rows for %q/%q: %w", pluginName, sheetName, err)
	}
	defer func() { _ = rows.Close() }()

	docs := []map[string]string{}
	headerSet := map[string]bool{}
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("scan plugin row for %q/%q: %w", pluginName, sheetName, err)
		}
		doc := map[string]string{}
		if err := json.Unmarshal([]byte(data), &doc); err != nil {
			return nil, fmt.Errorf("plugin row %q/%q: unmarshal data: %w", pluginName, sheetName, err)
		}
		for h := range doc {
			headerSet[h] = true
		}
		docs = append(docs, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load plugin rows for %q/%q: %w", pluginName, sheetName, err)
	}

	headers := make([]string, 0, len(headerSet))
	for h := range headerSet {
		headers = append(headers, h)
	}
	sort.Strings(headers)

	table := make([][]string, 0, len(docs)+1)
	table = append(table, headers)
	for _, doc := range docs {
		row := make([]string, len(headers))
		for i, h := range headers {
			row[i] = doc[h]
		}
		table = append(table, row)
	}
	return table, nil
}
