---
name: kql-validator
description: Validate KQL (Kusto Query Language) files used in Azure Quick Review (azqr) against their recommendation definitions. Use when the user wants to validate KQL syntax, check semantic alignment with recommendations, verify property names against Azure REST API schemas, or audit KQL queries before a pull request. WHEN: "validate kql", "check kql files", "kql syntax error", "validate aks kql", "run kql validator", "validate azure resource graph queries", "check recommendations alignment", "validate kql for <service>".
---

# KQL Validator Skill

This skill validates KQL query files in Azure Quick Review by delegating to the `KQLValidator` subagent. The subagent handles the full validation pipeline — discovery, structural/semantic/schema checks, and report generation.

## What the subagent validates

- **Syntax**: Correct KQL grammar, type casts, projection clauses
- **Semantic alignment**: Query finds violations (not compliance), `recommendationId` matches `aprlGuid`
- **Schema validity**: Property references exist in Azure REST API specs (GitHub lookup, best-effort)
- **100% coverage**: Never samples — validates every file in scope

## How to invoke

### Step 1: Determine scope from the user's request

Parse what the user wants to validate:

| User says | Scope argument to pass |
|-----------|----------------------|
| "validate all KQL" / no specific target | *(empty — validate everything)* |
| "validate ContainerService" / "validate AKS" | Service name, e.g. `ContainerService` |
| "validate aks-004.kql" | Filename, e.g. `aks-004.kql` |
| "validate internal/graph/azqr/azure-resources/Sql/" | Directory path |

### Step 2: Launch the KQLValidator subagent

Spawn the subagent using the agent file at `.agents/skills/kql-validator/agents/kql-validator.md`. Pass the scope as the user argument.

**Subagent prompt template:**

```
Read and follow the instructions in .agents/skills/kql-validator/agents/kql-validator.md exactly.

User argument: <SCOPE_ARGUMENT_OR_EMPTY>

Workspace: /home/cmendibl3/_dev/azqr
```

The subagent is self-contained — it will discover files, validate them, and generate a full markdown report. Do not attempt to re-implement its logic inline.

### Step 3: Surface results to the user

Once the subagent completes, present its markdown report directly. If the subagent encountered issues, highlight the critical ones and suggest fixes.

## Important notes

- The subagent uses GitHub to look up Azure REST API schemas. If GitHub is unavailable, it continues with structural and semantic validation only and notes the limitation.
- For large scopes (all services), the run may take several minutes. Let the user know and launch the subagent in background mode.
- For focused scopes (single service or file), sync mode is fine.
