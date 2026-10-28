# SQL Server ESU Scanner

> **Version: 0.7.0-beta**

Scans SQL Server instances (Azure VMs and Arc-enabled) for EOL/ESU lifecycle status, estimates the full current monthly cost of staying on SQL Server IaaS, and calculates monthly savings from migrating to Azure SQL Managed Instance. The query is implemented in [sql-eol.kql](sql-eol.kql).

> **ESU is no longer free on Azure VMs.** All production editions (Standard, Enterprise, Web) are billed at full ESU rates on Azure VMs.

## What it Does

1. Discovers `microsoft.sqlvirtualmachine/sqlvirtualmachines` and `microsoft.azurearcdata/sqlserverinstances` via Azure Resource Graph.
2. Resolves vCore count by joining the underlying Azure VM SKU for IaaS instances.
3. Assigns an `EOLStatus`: `Supported` → `Upcoming ESU` → `ESU Active` → `Expired`.
4. Generates a `MigrationRecommendation` and estimates the SQL MI target cost using a **conservative GP-only model** (all non-UVB cases use General Purpose pricing). Arc-enabled Enterprise AHUB instances apply the **Unlimited Virtualization Benefit (UVB)** → General Purpose at `max(4, vCores÷4)` vCores. Enterprise editions without UVB receive a note that a database assessment is required before committing to a tier.
5. Calculates current total monthly cost and estimated SQL MI cost, producing monthly savings and a migration verdict.
6. For Arc-enabled instances with a usable **Azure Arc migration assessment** (`properties.migration.assessment`), sizes the SQL MI from the assessment instead of the heuristic: assessed vCores, assessed tier (General Purpose or Business Critical) and predicted data+log storage. Pricing still uses the plugin's own list rates so every row shares one price basis; the assessment's own prices are shown only as a reference. All other rows keep the heuristic. The `EstimateBasis` column states which was used.

### Arc assessment rules

- A row uses the assessment only when it is enabled, uploaded within the last 14 days (`_assessmentFreshnessDays`), has a managed-instance result with status `Ready` or `ReadyWithConditions`, a compute size and a recognised tier.
- `NotReady` assessments produce the verdict `Blocked (assessment: NotReady)`; cost columns still use the heuristic.
- The headline cost, saving and verdict always use the conservative **2:1** consolidation, for assessed and heuristic rows alike.
- Assessed rows do not apply the UVB `vCores÷4` reduction; the assessed size is used at the plugin AHUB or PAYG rate.
- The assessment is Windows-only, SQL Server 2014+, excludes failover cluster instances, and uses a 30-day lookback at the 95th percentile by default. Under-sizing is possible if the lookback missed peak periods.
- Assessment currency, region and discount may differ from the plugin rates (`AssessmentPriceBasis` flags non-USD).

> **Breaking change (0.7.0-beta):** assessed rows now size the MI from the assessment (tier, vCores, storage), so headline cost and verdicts can change. New columns are appended after `SQLMIMigrationVerdict`; the verdict is no longer the last column.

## Assumptions

| Area | Value |
|------|-------|
| **ESU rates** | Standard/Web: $139/core/month · Enterprise: $540.50/core/month · Developer/Express/Free: $0. Blended 3-year planning estimate (Y1=75%, Y2=100%, Y3=125% of license cost). |
| **Minimum billable cores** | 4 cores per instance (Microsoft minimum) |
| **VM compute** | Blended PAYG by family: M-series $140, E-series $46, L-series $57, F-series $31, D/B-series $36 (per vCore/month). Windows multiplier ×1.8, West Europe ×1.13. Arc/on-prem = $0. |
| **Patch ops** | $160/month per instance (2 hrs × $80/hr operational overhead) |
| **SQL license cost (PAYG only)** | Enterprise: $274 · Standard: $73 · Web: $6 (per vCore/month). AHUB instances carry no hourly charge — SA is a sunk cost. |
| **SQL MI target tier** | All non-UVB cases → **General Purpose** (conservative estimate). Arc Enterprise AHUB (on-prem) → General Purpose via UVB. Enterprise editions without UVB → General Purpose with a note that a database assessment is required before committing to BC. Developer/Express/Free → N/A. |
| **Unlimited Virtualization Benefit (UVB)** | Arc-enabled Enterprise AHUB only. 1 on-prem core with SA → up to 4 SQL MI GP vCores. Sized at `max(4, vCores ÷ 4)` × $49/vCore AHUB. Azure VMs excluded — UVB applies to on-prem workloads only. Source: [microsoft.com/licensing/faqs/1#92](https://www.microsoft.com/licensing/faqs/1#92). |
| **SQL MI cost (PAYG)** | GP: $123/vCore/month |
| **SQL MI cost (AHUB)** | GP: $49/vCore/month |
| **SQL MI storage** | **Not included by default** (`_estimatedStorageGB = 0`). Set `_estimatedStorageGB` in `sql-eol.kql` to your environment's expected DB size to add storage to the MI cost estimate. GP rate: $0.115/GB/month (e.g. 512 GB → $58.88/month, 1,024 GB → $117.76/month). Free editions: $0. Backup storage (first 100% of provisioned size) is included free in SQL MI and is not modelled. |
| **Consolidation ratio** | 2:1 source-to-target. Two source VMs (or Arc instances) are assumed to consolidate onto a single SQL MI, so the estimated SQL MI cost is split equally between both. This conservative ratio reflects typical PaaS consolidation; actual consolidation opportunities should be validated per workload. |
| **Savings formula** | `Current (VM compute + SQL license if PAYG + ESU + patch ops) − Est SQL MI cost` |

## Savings by Scenario

| Scenario | VM Compute | SQL License | ESU | Patch Ops | − SQL MI |
|---|---|---|---|---|---|
| Azure VM · PAYG · ESU Active/Upcoming | ✅ | ✅ | ✅ | $160 | − MI PAYG |
| Azure VM · PAYG · Supported/Expired | ✅ | ✅ | $0 | $160 | − MI PAYG |
| Azure VM · AHUB · ESU Active/Upcoming | ✅ | $0 | ✅ | $160 | − MI AHUB |
| Azure VM · AHUB · Supported/Expired | ✅ | $0 | $0 | $160 | − MI AHUB |
| Arc/On-Prem · PAYG · ESU Active/Upcoming | $0 | ✅ | ✅ | $160 | − MI PAYG |
| Arc/On-Prem · Enterprise AHUB · UVB · ESU Active/Upcoming | $0 | $0 | ✅ | $160 | − MI GP UVB |
| Arc/On-Prem · Enterprise AHUB · UVB · Supported/Expired | $0 | $0 | $0 | $160 | − MI GP UVB |
| Arc/On-Prem · AHUB · ESU Active/Upcoming | $0 | $0 | ✅ | $160 | − MI AHUB |
| Arc/On-Prem · Supported/Expired | $0 | $0 | $0 | $160 | − MI AHUB |
| Developer / Express | — | — | — | — | $0 (N/A) |

> A negative saving means SQL MI costs more than the current setup — common for large AHUB instances not yet in ESU. Migration can still be justified by PaaS, security, and operational benefits.

## Output Columns

| Column | Description |
|--------|-------------|
| Subscription / ResourceGroup / Name / Location | Resource identity |
| CloudType | `Azure VM (SQL IaaS)` · `Arc-enabled (On-Prem)` · `Arc-enabled (AWS)` · `Arc-enabled (GCP)` · `Arc-enabled (<provider>)` for any other registered cloud provider |
| SQLVersion / Edition | Instance details |
| EOLStatus | `Supported` · `Upcoming ESU` · `ESU Active` · `Expired` |
| ESUStartDate / ESUEndDate | When ESU charges begin and when all support ends |
| MigrationTargetTier | `General Purpose` · `General Purpose (Database assessment required. Server had an Enterprise license)` · `N/A`. Arc Enterprise AHUB (on-prem) → `General Purpose` via UVB. All other non-free cases → `General Purpose` (conservative GP-only estimate). |
| MigrationRecommendation | Actionable text guidance based on edition and EOL status |
| vCores / BillableCores | Core count and Microsoft-minimum billable cores |
| ESUMonthlyCostPerCore | Per-core ESU rate applied in calculations |
| SQLLicenseType | `PAYG` · `AHUB` · `DR` |
| SQLLicenseMonthlyCostPerCore / SQLLicenseMonthlyCost | SQL license list rate (factored into savings for PAYG only) |
| VMCostPerCorePerMonth / EstVMComputeMonthlyCost | VM compute cost estimate ($0 for Arc/on-prem) |
| EstESUMonthlyCost | ESU cost projection ($0 if Supported or Expired) |
| PatchOpsMonthlyCost | Operational overhead |
| CurrentMonthlyCost | Total current monthly spend (all components) |
| ConsolidationRatio | Conservative 2:1 source-to-target ratio used in MI cost allocation |
| EstSQLMIMonthlyCost | Estimated SQL MI cost at recommended tier: 2:1 consolidated for all rows; assessed rows use assessed size and predicted storage, heuristic rows compute + license only by default. Set `_estimatedStorageGB` in the KQL to include storage (GP rate: $0.115/GB/month). |
| EstSQLMIMonthlySaving | Saving vs current monthly spend (negative = cost increase) |
| SQLMIMigrationVerdict | `Cost Savings` · `Break Even` · `Cost Increase (justified by PaaS/security benefits)` · `Blocked (assessment: NotReady)` |
| EstimateBasis | `Arc assessment (<percentile>th percentile, <lookback>d, uploaded <date>)` or `Heuristic (<reason>)` |
| AssessmentReadiness / AssessmentBlockerCount / AssessmentUploadTime | Arc assessment MI readiness status, blocker count and upload time (empty without an assessment) |
| AssessedMITier / AssessedMIvCores / PredictedDataLogGB | Assessed target tier, vCores and predicted data+log size (only when the assessment is used) |
| AssessmentPriceReference / AssessmentPriceBasis | The assessment's own monthly total and its region, currency and strategy; reference only |

## Limitations

- For heuristic rows SQL MI storage is **not included by default** (`_estimatedStorageGB = 0`); assessed rows include predicted data+log storage. Set `_estimatedStorageGB` in `sql-eol.kql` to include storage in the MI cost estimate (GP rate: $0.115/GB/month). Actual storage costs depend on database size and backup retention settings.
- Costs reflect public list pricing — EA/CSP discounts are not factored in.
- VM compute rates are blended family-level estimates (Linux PAYG baseline), not exact billing.
- SQL MI estimates do not account for reserved capacity pricing or workload-specific sizing.
- UVB assumes maximum 4:1 ratio ("up to 4 vCores per on-prem core") — actual eligibility and ratio should be verified with your Microsoft licensing team.
- Arc SQL cloud provider detection (`CloudType`) requires the underlying `microsoft.hybridcompute/machines` resource to be registered and connected. Disconnected Arc machines may show `Arc-enabled (On-Prem)` even if hosted on another cloud.
- SQL Server inside containers or on non-Arc VMs is not discoverable via Azure Resource Graph.

## Usage

```bash
# Plugin-only mode (fast)
azqr sql-eol

# As part of a full scan
azqr scan --plugin sql-eol
```

## License

[MIT](../../../../LICENSE)
