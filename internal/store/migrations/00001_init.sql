-- +goose Up
CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  trigger TEXT NOT NULL,
  status TEXT NOT NULL,
  azqr_version TEXT,
  scan_params_json TEXT NOT NULL,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  duration_ms INTEGER,
  error_message TEXT,
  scope_id TEXT,
  tenant_id TEXT,
  summary_json TEXT
);

CREATE INDEX idx_sessions_started_at ON sessions (started_at);
CREATE INDEX idx_sessions_status ON sessions (status);

CREATE TABLE session_stage_runs (
  session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  stage_name TEXT NOT NULL,
  position INTEGER NOT NULL,
  status TEXT NOT NULL,
  duration_ms INTEGER,
  error_message TEXT,
  PRIMARY KEY (session_id, stage_name)
);

CREATE TABLE resources (
  session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  resource_id TEXT NOT NULL,
  subscription_id TEXT,
  resource_group TEXT,
  type TEXT,
  location TEXT,
  name TEXT,
  sku_name TEXT,
  sku_tier TEXT,
  sku_family TEXT,
  sku_capacity INTEGER,
  kind TEXT,
  sla TEXT,
  tags_json TEXT,
  excluded INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (session_id, resource_id)
);

CREATE INDEX idx_resources_session_type ON resources (session_id, type);

CREATE TABLE recommendation_catalog (
  recommendation_id TEXT PRIMARY KEY,
  source TEXT NOT NULL,
  category TEXT,
  impact TEXT,
  resource_type TEXT,
  recommendation TEXT,
  long_description TEXT,
  potential_benefits TEXT,
  learn_url TEXT,
  automation_available TEXT,
  content_hash TEXT NOT NULL,
  first_seen_azqr_version TEXT,
  last_seen_azqr_version TEXT
);

CREATE TABLE recommendation_results (
  session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  recommendation_id TEXT NOT NULL REFERENCES recommendation_catalog (recommendation_id),
  resource_id TEXT NOT NULL,
  subscription_id TEXT,
  subscription_name TEXT,
  resource_group TEXT,
  name TEXT,
  tags TEXT,
  param1 TEXT,
  param2 TEXT,
  param3 TEXT,
  param4 TEXT,
  param5 TEXT,
  PRIMARY KEY (session_id, recommendation_id, resource_id)
);

CREATE INDEX idx_recommendation_results_session ON recommendation_results (session_id);

CREATE TABLE advisor_results (
  session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  recommendation_id TEXT NOT NULL,
  resource_id TEXT NOT NULL,
  subscription_id TEXT,
  subscription_name TEXT,
  type TEXT,
  name TEXT,
  category TEXT,
  impact TEXT,
  description TEXT,
  PRIMARY KEY (session_id, recommendation_id, resource_id)
);

CREATE TABLE defender_results (
  session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  subscription_id TEXT NOT NULL,
  subscription_name TEXT,
  name TEXT NOT NULL,
  tier TEXT,
  PRIMARY KEY (session_id, subscription_id, name)
);

CREATE TABLE defender_recommendations (
  session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  resource_id TEXT NOT NULL,
  recommendation_name TEXT NOT NULL,
  subscription_id TEXT,
  subscription_name TEXT,
  resource_group_name TEXT,
  resource_type TEXT,
  resource_name TEXT,
  category TEXT,
  recommendation_severity TEXT,
  action_description TEXT,
  remediation_description TEXT,
  az_portal_link TEXT,
  PRIMARY KEY (session_id, resource_id, recommendation_name)
);

CREATE TABLE azure_policy_results (
  session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  resource_id TEXT NOT NULL,
  policy_assignment_id TEXT NOT NULL,
  subscription_id TEXT,
  subscription_name TEXT,
  policy_display_name TEXT,
  policy_description TEXT,
  compliance_state TEXT,
  type TEXT,
  name TEXT,
  resource_group_name TEXT,
  time_stamp TEXT,
  policy_definition_name TEXT,
  policy_definition_id TEXT,
  policy_assignment_name TEXT,
  PRIMARY KEY (session_id, resource_id, policy_assignment_id)
);

CREATE TABLE arc_sql_results (
  session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  azure_arc_server TEXT NOT NULL,
  sql_instance TEXT NOT NULL,
  subscription_id TEXT,
  subscription_name TEXT,
  status TEXT,
  resource_group TEXT,
  version TEXT,
  build TEXT,
  patch_level TEXT,
  edition TEXT,
  v_cores TEXT,
  license TEXT,
  dps_status TEXT,
  tel_status TEXT,
  defender_status TEXT,
  PRIMARY KEY (session_id, azure_arc_server, sql_instance)
);

CREATE TABLE cost_results (
  session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  subscription_id TEXT NOT NULL,
  service_name TEXT NOT NULL,
  subscription_name TEXT,
  value TEXT,
  currency TEXT,
  from_date TEXT,
  to_date TEXT,
  PRIMARY KEY (session_id, subscription_id, service_name)
);

CREATE TABLE plugin_executions (
  session_id TEXT NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
  plugin_name TEXT NOT NULL,
  plugin_version TEXT,
  status TEXT NOT NULL,
  started_at TEXT,
  finished_at TEXT,
  error_message TEXT,
  PRIMARY KEY (session_id, plugin_name)
);

-- data is always written via encoding/json.Marshal (see Recorder.writePluginResults), so
-- it is guaranteed valid JSON at the application layer. The column intentionally omits a
-- CHECK (json_valid(data)) constraint: SQLite's json_valid() has no portable Postgres
-- equivalent, and this migration tree is shared by both dialects (goose applies it
-- verbatim regardless of backend).
CREATE TABLE plugin_result_rows (
  session_id TEXT NOT NULL,
  plugin_name TEXT NOT NULL,
  sheet_name TEXT NOT NULL,
  row_ordinal INTEGER NOT NULL,
  data TEXT NOT NULL,
  FOREIGN KEY (session_id, plugin_name) REFERENCES plugin_executions (session_id, plugin_name) ON DELETE CASCADE,
  PRIMARY KEY (session_id, plugin_name, sheet_name, row_ordinal)
);

-- +goose Down
DROP TABLE plugin_result_rows;
DROP TABLE plugin_executions;
DROP TABLE cost_results;
DROP TABLE arc_sql_results;
DROP TABLE azure_policy_results;
DROP TABLE defender_recommendations;
DROP TABLE defender_results;
DROP TABLE advisor_results;
DROP TABLE recommendation_results;
DROP TABLE recommendation_catalog;
DROP TABLE resources;
DROP TABLE session_stage_runs;
DROP TABLE sessions;
