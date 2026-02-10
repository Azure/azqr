---
description: 'Project-specific instructions for Azure Quick Review (azqr) — architecture, conventions, recommendation system, KQL patterns, and contribution workflow'
applyTo: '**'
---

# Azure Quick Review (azqr) — Project Instructions

**azqr** is a CLI tool that scans Azure subscriptions and resource groups, evaluates resources against best-practice recommendations, and produces Excel/CSV/JSON assessment reports.

## Project Overview

azqr uses two complementary scan engines:

- **Azure Resource Graph (ARG) engine** — loads embedded YAML recommendation definitions paired with KQL query files, executes them against Azure Resource Graph in batches, and collects non-compliant resources.
- **Azure Resource Manager (ARM) scanners** — Go SDK-based scanners that collect resource state (subscriptions, diagnostics, advisor findings, Defender plans, costs, Azure Policy compliance, Arc SQL).

Default output is an Excel workbook. JSON and CSV are also supported. An optional local dashboard (`azqr show`) renders results in the terminal.

Key CLI commands:
- `azqr scan` — run a full assessment
- `azqr scan <abbr>` — scan a single service (e.g., `azqr scan aks`)
- `azqr rules` — print all loaded recommendations
- `azqr types` — print supported Azure resource types
- `azqr show` — serve a local dashboard from a previous scan output
- `azqr compare` — diff two scan reports
- `azqr plugins list|info` — inspect loaded plugins

---

## Architecture Overview

```
cmd/azqr/
  main.go                    Entry point; registers internal plugins via blank imports
  commands/                  Cobra CLI commands (scan, show, rules, types, compare, plugins, copilot, mcp)

internal/
  models/                    Shared types, interfaces, scanner registry, filter logic
  pipeline/                  Composable scan pipeline (one file per stage)
  graph/                     Embedded YAML+KQL recommendation engine
    aprl/                    APRL upstream recommendations (git submodule — do not edit)
    azure-orphan-resources/  Orphaned resource checks (git submodule — do not edit)
    azqr/                    azqr-owned recommendations (add custom recommendations here)
      azure-resources/
        <Provider>/
          <resourceType>/
            recommendations.yaml
            kql/
              <aprlGuid>.kql
  scanners/                  ARM-based service data fetchers
    registry/                Single-import registration of all service scanners
    plugins/                 Internal Go plugins (carbon, openai, zone)
  plugins/                   YAML plugin loader and registry
  renderers/                 Excel/CSV/JSON report generation
  viewer/                    Local dashboard server (bubbletea TUI)
  mcpserver/                 MCP server implementation
  az/                        Azure SDK helpers and auth utilities
  to/                        Pointer helper utilities

data/
  recommendations.json       Generated snapshot — regenerate with `make json` after any recommendation change

examples/
  plugins/yaml-example/      Reference YAML plugin demonstrating inline and file-based KQL queries
```

---

## Recommendation System

### YAML Recommendation Files

Every recommendation is defined in a `recommendations.yaml` file:

```
internal/graph/azqr/azure-resources/<Provider>/<resourceType>/recommendations.yaml
```

Each entry in the file follows this structure:

```yaml
- description: AKS Cluster should be private
  aprlGuid: aks-004
  recommendationTypeId: null
  recommendationControl: Security          # High-level category
  recommendationImpact: High               # High | Medium | Low
  recommendationResourceType: Microsoft.ContainerService/managedClusters
  recommendationMetadataState: Active      # Active | Disabled
  longDescription: AKS Cluster should be private
  potentialBenefits: See recommendation details
  pgVerified: true                         # true if verified by Azure product group
  automationAvailable: true                # true when a matching .kql file exists
  tags: []
  learnMoreLink:
  - name: Learn more
    url: 'https://learn.microsoft.com/azure/aks/private-clusters'
```

Valid values for `recommendationControl`:
`SLA` | `Scalability` | `HighAvailability` | `BusinessContinuity` | `DisasterRecovery` | `Security` | `Governance` | `MonitoringAndAlerting` | `OtherBestPractices`

Rules:
- `aprlGuid` must be **globally unique** across all three recommendation sources (aprl, azure-orphan-resources, azqr).
- Use the service abbreviation as prefix, e.g. `aks-004`, `vm-001`, `st-003`.
- Set `automationAvailable: true` only when a corresponding `.kql` file exists.
- Set `recommendationMetadataState: Disabled` to suppress a recommendation without deleting it.
- Run `make json` after any change and commit the updated `data/recommendations.json`.

### KQL Query Files

Every recommendation with `automationAvailable: true` requires a KQL file:

```
internal/graph/azqr/azure-resources/<Provider>/<resourceType>/kql/<aprlGuid>.kql
```

The filename must match the `aprlGuid` exactly (e.g., `aks-004.kql` for `aprlGuid: aks-004`).

**Required output columns** (in order):
```kql
| project recommendationId, name, id, tags, param1 [, param2, param3, param4, param5]
```

- `recommendationId` — must equal the `aprlGuid` string literal
- `name` — resource name
- `id` — full resource ID
- `tags` — resource tags
- `param1`–`param5` — optional context fields shown in the report

**Critical rule — return non-compliant resources only.**  
KQL queries must filter to resources that **fail** the check, not those that pass:

```kql
// AKS Cluster should be private
resources
| where type =~ 'Microsoft.ContainerService/managedClusters'
| extend enablePrivateCluster = properties.apiServerAccessProfile.enablePrivateCluster
| where isnull(enablePrivateCluster) or enablePrivateCluster == false
| extend recommendationId = 'aks-004'
| extend param1 = 'Public cluster'
| project recommendationId, name, id, tags, param1
```

Use `extend recommendationId = '<aprlGuid>'` before the final `project` statement.

Common KQL patterns:
- Use `type =~` (case-insensitive) for resource type filtering
- Use `isnull(x) or x == false` to catch unset boolean properties
- Use `tostring()` / `todynamic()` for nested property access
- `param1`–`param5` are free-form strings for additional context in the report

### Recommendation Sources

| Directory | Owner | Notes |
|---|---|---|
| `internal/graph/aprl/` | Upstream APRL project | Git submodule — **never edit directly** |
| `internal/graph/azure-orphan-resources/` | Upstream orphan resources project | Git submodule — **never edit directly** |
| `internal/graph/azqr/` | azqr project | Add new custom recommendations here |

---

## Scanner / Service Registry

Service scanners map Azure resource types to a short abbreviation used for CLI subcommands and filtering.

Scanners are registered in `internal/scanners/registry/scanners.go` using the `init()` pattern:

```go
func init() {
    models.ScannerList["aks"] = []models.IAzureScanner{
        models.NewBaseScanner("Azure Kubernetes Service", "Microsoft.ContainerService/managedClusters"),
    }
}
```

- The key (`"aks"`) becomes the CLI subcommand: `azqr scan aks`
- `NewBaseScanner(serviceName, resourceTypes...)` takes one or more resource types
- Multiple resource types under one abbreviation are all scanned together
- `models.ScannerList` is the authoritative registry — all entries are reflected in `azqr types` output

The `internal/scanners/registry/` package is imported as a blank import so all `init()` functions run without requiring individual imports everywhere.

---

## How to Add Support for a New Azure Resource

1. **Create the directory structure**:
   ```
   internal/graph/azqr/azure-resources/<Provider>/<resourceType>/kql/
   ```

2. **Create `recommendations.yaml`** with one or more recommendation entries following the schema above.

3. **Create KQL files** for each recommendation with `automationAvailable: true`:
   ```
   internal/graph/azqr/azure-resources/<Provider>/<resourceType>/kql/<aprlGuid>.kql
   ```

4. **Register the scanner** in `internal/scanners/registry/scanners.go`:
   ```go
   func init() {
       models.ScannerList["<abbr>"] = []models.IAzureScanner{
           models.NewBaseScanner("<Service Display Name>", "Microsoft.<Provider>/<resourceType>"),
       }
   }
   ```

5. **Validate YAML**:
   ```sh
   make validate-yaml
   ```

6. **Regenerate recommendations snapshot** and verify no unexpected diff:
   ```sh
   make json
   ```

7. **Run full test suite**:
   ```sh
   make test
   ```

---

## Plugin System

### Internal Go Plugins

Internal plugins add optional scan stages that produce additional report sheets (e.g., carbon emissions, OpenAI throttling, zone mapping).

- Located in `internal/scanners/plugins/<name>/`
- Registered via blank imports in `cmd/azqr/main.go`:
  ```go
  import (
      _ "github.com/Azure/azqr/internal/scanners/plugins/carbon"
      _ "github.com/Azure/azqr/internal/scanners/plugins/openai"
      _ "github.com/Azure/azqr/internal/scanners/plugins/zone"
  )
  ```
- Each plugin implements the `plugins.Plugin` interface and calls `plugins.Register()` in its `init()` function.
- Internal plugins surface as top-level CLI commands (e.g., `azqr carbon-emissions`).

### YAML Plugins

YAML plugins allow defining custom recommendations without writing Go code.

- Loaded from disk at runtime via `internal/plugins/yaml.go`
- Passed to `azqr scan` with the `--plugins` flag
- Reference example: `examples/plugins/yaml-example/custom-checks.yaml`
- Queries can be:
  - **Inline**: `query: |` field in the YAML
  - **File-based**: `queryFile: kql/something.kql` referencing a path relative to the YAML file

---

## Pipeline Stages

The scan pipeline is assembled in `internal/pipeline/` with one file per stage:

| Stage file | Description | Default |
|---|---|---|
| `stage_initialization.go` | Auth, client setup | Always |
| `stage_subscription_discovery.go` | Discover subscriptions | Always |
| `stage_resource_discovery.go` | Inventory all resources | Always |
| `stage_graph_scan.go` | Run ARG/YAML+KQL recommendations | Always |
| `stage_diagnostics_scan.go` | Check diagnostic settings | Always |
| `stage_advisor.go` | Azure Advisor recommendations | Enabled by default |
| `stage_defender_status.go` | Defender for Cloud plans | Enabled by default |
| `stage_defender_recommendations.go` | Defender recommendations | Disabled by default |
| `stage_azure_policy.go` | Azure Policy compliance | Disabled by default |
| `stage_arc_sql.go` | Azure Arc SQL instances | Disabled by default |
| `stage_cost.go` | Cost data (last month) | Disabled by default |
| `stage_plugin_execution.go` | Internal and YAML plugins | Enabled if plugins loaded |
| `stage_report_rendering.go` | Generate output files | Always |

Control stages with the `--stages` flag:
```sh
azqr scan --stages +cost          # enable cost stage
azqr scan --stages -advisor       # disable advisor stage
azqr scan --stages +policy,+arc   # enable multiple optional stages
```

`ScanParams` and `StageConfigs` in `internal/models/` hold the runtime stage configuration.

---

## Development Workflow

| Command | When to use |
|---|---|
| `make build` | Compile the `azqr` binary |
| `make test` | Full quality gate: lint → vet → tidy → json → validate-yaml → validate-scanners → unit tests |
| `make json` | Regenerate `data/recommendations.json` — **run and commit after any recommendation change** |
| `make validate-yaml` | Validate all `recommendations.yaml` files against the schema |
| `make validate-scanners` | Verify APRL recommendation coverage |
| `make lint` | Run golangci-lint |
| `make vet` | Run go vet |
| `make tidy` | Tidy go modules |
| `make test-integration` | Integration tests against real Azure resources (requires `AZURE_SUBSCRIPTION_ID`, `AZURE_TENANT_ID`, Terraform) |
| `make clean` | Remove built binaries |

The `make test` target is the CI gate. All of the following must pass:
- `golangci-lint run`
- `go vet ./...`
- `go mod tidy` (no diff)
- `make json` (no diff in `data/recommendations.json`)
- `make validate-yaml`
- `make validate-scanners`
- `go test -race ./...`

---

## Key Conventions

### Copyright Header
All Go source files must start with:
```go
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
```

### Recommendation IDs (`aprlGuid`)
- Must be globally unique across all sources (aprl, azure-orphan-resources, azqr).
- Use the service abbreviation as a prefix followed by a zero-padded number: `<abbr>-<NNN>` (e.g., `aks-004`, `st-003`).
- Never reuse or reassign an `aprlGuid`, even if a recommendation is removed.

### KQL non-compliance logic
- Always return resources that **violate** the recommendation (not compliant ones).
- Use `isnull(x) or x == false` for boolean properties that may be absent.
- Use `type =~` for case-insensitive resource type matching.
- The `recommendationId` column value must be the `aprlGuid` string literal.

### Generated files
- `data/recommendations.json` is generated — never edit it manually. Always run `make json` and commit the result.
- `go.sum` is generated — run `go mod tidy` to keep it current.

### Scanner abbreviations
- Use lowercase 2–6 character abbreviations matching Azure resource naming conventions (e.g., `aks`, `vm`, `vnet`, `st`, `kv`).
- Abbreviations become CLI subcommands and are used in filter flags.

### Logging
- Use `github.com/rs/zerolog/log` for all logging.
- Use structured fields: `log.Error().Err(err).Msg("message")`.
- Do not use `fmt.Println` or `log.Printf` in non-CLI code.

---

## Testing Patterns

### Unit Tests
- Use table-driven tests with a `tests` slice of structs.
- Place test files next to the code they test (`_test.go` suffix).
- Mark helper functions with `t.Helper()`.
- Clean up resources with `t.Cleanup()`.

### Integration Tests
- Located in `test/integration/`.
- Terraform fixtures in `test/fixtures/terraform/`.
- Require `AZURE_SUBSCRIPTION_ID` and `AZURE_TENANT_ID` environment variables.
- Run with: `make test-integration`
- Tests use the [terratest](https://github.com/gruntwork-io/terratest) library.

### Running Tests
```sh
make test                  # unit tests (includes lint, vet, validate)
make test-integration      # integration tests (provisions real Azure resources)
go test -race ./...        # run unit tests directly with race detector
```
