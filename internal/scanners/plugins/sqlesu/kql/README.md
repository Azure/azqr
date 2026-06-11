# SQL Server ESU Scanner

> **Version: 0.2.0-beta**

Scans SQL Server instances (Azure VMs and Arc-enabled) for EOL/ESU lifecycle status, estimates the full current cost of staying on SQL Server IaaS, and calculates monthly/annual/3-year savings from migrating to Azure SQL Managed Instance. The query is implemented in [sql-esu.kql](sql-esu.kql).

> **ESU is no longer free on Azure VMs.** All production editions (Standard, Enterprise, Web) are billed at full ESU rates on Azure VMs.

## What it Does

1. Discovers `microsoft.sqlvirtualmachine/sqlvirtualmachines` and `microsoft.azurearcdata/sqlserverinstances` via Azure Resource Graph.
2. Resolves vCore count by joining the underlying Azure VM SKU for IaaS instances.
3. Assigns an `EOLStatus`: `Supported` → `Upcoming ESU` → `ESU Active` → `Expired`.
4. Generates a `MigrationRecommendation` and auto-selects the SQL MI target tier by edition (Enterprise → Business Critical, Standard/Web → General Purpose).
5. Calculates current total cost and estimated SQL MI cost, producing monthly/annual/3-year savings and a migration verdict.

## Assumptions

| Area | Value |
|------|-------|
| **ESU rates** | Standard/Web: $139/core/month · Enterprise: $540.50/core/month · Developer/Express/Free: $0. Blended 3-year planning estimate (Y1=75%, Y2=100%, Y3=125% of license cost). |
| **Minimum billable cores** | 4 cores per instance (Microsoft minimum) |
| **VM compute** | Blended PAYG by family: M-series $140, E-series $46, L-series $57, F-series $31, D/B-series $36 (per vCore/month). Windows multiplier ×1.8, West Europe ×1.13. Arc/on-prem = $0. |
| **Patch ops** | $160/month per instance (2 hrs × $80/hr operational overhead) |
| **SQL license cost (PAYG only)** | Enterprise: $274 · Standard: $73 · Web: $6 (per vCore/month). AHUB instances carry no hourly charge — SA is a sunk cost. |
| **SQL MI target tier** | Auto-selected: Enterprise → Business Critical, Standard/Web → General Purpose. Developer/Express/Free → N/A (not migration candidates). |
| **SQL MI cost (PAYG)** | GP: $123/vCore/month · BC: $367/vCore/month |
| **SQL MI cost (AHUB)** | GP: $49/vCore/month · BC: $147/vCore/month |
| **Savings formula** | `Current (VM compute + SQL license if PAYG + ESU + patch ops) − Est SQL MI cost` |

## Savings by Scenario

| Scenario | VM Compute | SQL License | ESU | Patch Ops | − SQL MI |
|---|---|---|---|---|---|
| Azure VM · PAYG · ESU Active/Upcoming | ✅ | ✅ | ✅ | $160 | − MI PAYG |
| Azure VM · PAYG · Supported/Expired | ✅ | ✅ | $0 | $160 | − MI PAYG |
| Azure VM · AHUB · ESU Active/Upcoming | ✅ | $0 | ✅ | $160 | − MI AHUB |
| Azure VM · AHUB · Supported/Expired | ✅ | $0 | $0 | $160 | − MI AHUB |
| Arc/On-Prem · PAYG · ESU Active/Upcoming | $0 | ✅ | ✅ | $160 | − MI PAYG |
| Arc/On-Prem · AHUB · ESU Active/Upcoming | $0 | $0 | ✅ | $160 | − MI AHUB |
| Arc/On-Prem · Supported/Expired | $0 | $0 | $0 | $160 | − MI AHUB |
| Developer / Express | — | — | — | — | $0 (N/A) |

> A negative saving means SQL MI costs more than the current setup — common for large AHUB instances not yet in ESU. Migration can still be justified by PaaS, security, and operational benefits.

## Output Columns

| Column | Description |
|--------|-------------|
| Name / ResourceGroup / Subscription / Location | Resource identity |
| CloudType | `Azure VM (SQL IaaS)` or `Arc-enabled (On-Prem)` |
| SQLVersion / Edition / vCores / BillableCores | Instance details |
| EOLStatus | `Supported` · `Upcoming ESU` · `ESU Active` · `Expired` |
| MigrationRecommendation | Actionable text guidance based on edition and EOL status |
| MigrationTargetTier | `General Purpose` · `Business Critical` · `N/A` |
| ESUStartDate / ESUEndDate | When ESU charges begin and when all support ends |
| ESUMonthlyCostPerCore | Per-core ESU rate applied in calculations |
| SQLLicenseType | `PAYG` · `AHUB` · `DR` |
| SQLLicenseMonthlyCostPerCore / SQLLicenseMonthlyCost / SQLLicenseAnnualCost | SQL license list rate (factored into savings for PAYG only) |
| VMCostPerCorePerMonth / EstVMComputeMonthlyCost / AnnualCost / ThreeYearCost | VM compute cost estimate ($0 for Arc/on-prem) |
| EstESUMonthlyCost / AnnualCost / ThreeYearCost | ESU cost projection ($0 if Supported or Expired) |
| PatchOpsMonthlyCost / AnnualCost / ThreeYearCost | Operational overhead |
| CurrentMonthlyCost / AnnualCost / ThreeYearCost | Total current spend (all components) |
| EstSQLMIMonthlyCost / AnnualCost / ThreeYearCost | Estimated SQL MI cost at recommended tier |
| EstSQLMIMonthlySaving / AnnualSaving / ThreeYearSaving | Saving vs current spend (negative = cost increase) |
| SQLMIMigrationVerdict | `Cost Savings` · `Break Even` · `Cost Increase (justified by PaaS/security benefits)` |

## Limitations

- Costs reflect public list pricing — EA/CSP discounts are not factored in.
- VM compute rates are blended family-level estimates (Linux PAYG baseline), not exact billing.
- SQL MI estimates do not account for reserved capacity pricing or workload-specific sizing.
- SQL Server inside containers or on non-Arc VMs is not discoverable via Azure Resource Graph.

## Usage

```bash
# Plugin-only mode (fast)
azqr sql-esu

# As part of a full scan
azqr scan --plugin sql-esu
```

## License

[MIT](../../../../LICENSE)
