---
name: review-areas
description: 'In-depth code review that fans out parallel subagents across review areas — CRITICAL after non-trivial development. USE FOR: "in-depth review of a pull request", "code quality check or bug hunt", "review correctness, tests, security, performance and product areas", "review after any non-trivial development phase". DO NOT USE FOR: "plan a new implementation", "explain how existing code works".'
---

# Skill: Review Areas

Fan out parallel read-only subagents, each assigned a different review area, then synthesize the highest-signal findings. This surfaces issues that a single-pass review misses because each subagent goes deep on its area instead of skimming everything.

## Review Areas

Pick 2–4 areas based on the nature of the change. Not every review needs all areas — match the areas to the risk profile.

| Area | When to include | Focus |
|------|----------------|-------|
| **Correctness** | Always | Logic errors, type safety, race conditions, null/undefined paths, unsafe casts, wrong behavior |
| **Tests** | When tests exist or should exist | Run tests, check failing/missing coverage, validate assertions match intent |
| **Security** | Auth, input handling, data flow changes | Input validation, auth checks, injection, data exposure |
| **Performance** | Hot paths, data structures, async changes | Algorithm complexity, unnecessary allocations, blocking async patterns |
| **Product** | UI, UX, or user-facing behavior changes | UX implications, feature completeness, accessibility gaps |

## Workflow

### 1 — Scope

- Prefer the active pull request, current git diff, or an explicit file list from the user.
- If there is nothing to review, ask the user to point at a branch, PR, diff, or file set.

### 2 — Build a Change Summary

Before fanning out, build a concise change summary. Do **not** paste raw diffs into the subagent prompts — each subagent has tools to read files and inspect changes itself.

The summary should include:
1. **Intent** — what the change is trying to accomplish (from PR description, commit messages, or conversation context).
2. **Changed files** — list each file path with a one-line description of what changed (e.g., "added input validation", "refactored into helper").
3. **Risk areas** — anything risky: new dependencies, auth changes, hot-path modifications, API surface changes.
4. **How to inspect** — branch/PR/commit info so subagents can use their tools.

Keep the summary under ~50 lines. Subagents get better results reading code in context than scanning a wall of diff.

### 3 — Fan Out

Launch 2–4 parallel subagents using the area prompts below. Each subagent works in isolation — do not share one area's findings with another before synthesis.

Use unnamed subagents (no custom agent needed). Each gets a self-contained prompt with its area, the change summary, and the return format.

### 4 — Synthesize

When all subagents return:

1. Deduplicate findings that overlap across areas (e.g., a correctness bug that also shows up as a missing test).
2. Order by severity: breaking > wrong behavior > security > missing coverage > performance > product.
3. Apply the signal filter — drop anything that wouldn't block a PR.
4. If no blocking issues survive, say so and mention any meaningful testing gaps.

### 5 — Save Findings

Always save the synthesized findings to session memory at `/memories/session/review.md`. This makes them available for follow-up turns, fix planning, and cross-referencing with future reviews.

### 6 — Fix or Report

- **Review-only**: if the user said "only review", "just review", or "read-only", stop after reporting.
- **Default (review-and-fix)**:
  1. For each fixable finding, launch one or more parallel `Explore` subagents to investigate the fix — this is faster and deeper than reading files inline. Give each Explore agent the finding, the relevant file paths, and ask it to return the specific change needed (what to replace, where).
  2. Apply fixes in severity order based on the Explore agents' recommendations.
  3. After applying, re-check the changed lines and diagnostics to validate.
  4. Report what was fixed and any remaining items that need the user's input.

## Signal Filter

Keep only findings a senior engineer would block a PR for:
- Will fail to compile, type-check, or produce wrong results
- Clear, citable violation of workspace coding standards
- Security vulnerability with a concrete exploit path
- Missing error handling that causes silent failures

Drop: style preferences, linter-catchable issues, pre-existing problems, speculative concerns.

## Area Prompt

Each subagent gets this prompt with `{AREA}` and `{FOCUS}` filled in from the Review Areas table.

```text
You are a focused code review subagent. Your area is: {AREA}

## What changed
<change summary from step 2: intent, changed files with one-line descriptions, risk areas>

## How to inspect
<branch, PR number, or commit range the subagent can use with its tools>

Focus on: {FOCUS}

Use your tools to read the changed files, check diagnostics, and run focused tests. Do not rely solely on this summary — dig into the code yourself. Read functions end-to-end. Trace inputs through branches and error paths. Check callers when contracts change.

Rules:
- Stay read-only. Do not edit files.
- Only flag issues that would block a PR — things that break, regress, or expose a concrete vulnerability.
- Do not report issues outside your area.
- Do not suggest code edits — describe the problem and why it matters.
- Check loaded workspace instructions and skills before flagging standard violations.
- Keep your response short. No preamble, no style commentary.

Return format:

**Area**: {AREA}

**Findings** (0–5 items, severity order):
- [file:line] One-sentence description. Why it matters.

If nothing blocks approval: "No blocking issues found in {AREA}."
```

## Output Shape

**Changes Summary** (50 words max):
What changed, why, and expected impact.

**What's Done Well** (1–3 items):
Acknowledge good patterns worth reinforcing.

**Critical Issues** (0–5 items, severity order):
Each with file references and the area that surfaced it.

**Improvements** (0–5 items):
High-value suggestions that didn't quite reach "blocking" but are worth addressing.

**Verdict**: Ready / Needs Revisions / Blocked — with a specific next step.
