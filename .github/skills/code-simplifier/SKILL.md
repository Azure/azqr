---
name: code-simplifier
description: Analyzes recently modified code and creates pull requests with simplifications that improve clarity, consistency, and maintainability while preserving functionality
---

# Code Simplifier Skill

Expert guidance for analyzing and simplifying recently modified code to improve clarity, consistency, and maintainability while preserving exact functionality.

## Overview

This skill enables agents to:
- Analyze code changes from the last 24 hours
- Apply targeted simplifications that improve code quality
- Create pull requests with improvements
- Validate changes through tests and linting

## When to Use This Skill

Use this skill when:
- User asks to simplify or refactor recent code changes
- User wants to improve code quality without changing functionality
- User requests code review with automatic improvements
- Scheduled code quality checks are needed

## Simplification Principles

### 1. Preserve Functionality
- **NEVER** change what the code does - only how it does it
- All original features, outputs, and behaviors must remain intact
- Run tests before and after to ensure no behavioral changes

### 2. Enhance Clarity
- Reduce unnecessary complexity and nesting
- Eliminate redundant code and abstractions
- Improve readability through clear variable and function names
- Consolidate related logic
- Remove unnecessary comments that describe obvious code
- **IMPORTANT**: Avoid nested ternary operators - prefer switch statements or if/else chains
- Choose clarity over brevity - explicit code is often better than compact code

### 3. Apply Project Standards
- Use project-specific conventions and patterns
- Follow established naming conventions
- Apply consistent formatting
- Use appropriate language features (modern syntax where beneficial)

### 4. Maintain Balance
Avoid over-simplification that could:
- Reduce code clarity or maintainability
- Create overly clever solutions that are hard to understand
- Combine too many concerns into single functions
- Remove helpful abstractions that improve code organization
- Prioritize "fewer lines" over readability
- Make the code harder to debug or extend

## Workflow

### Phase 1: Identify Recently Modified Code

#### 1.1 Find Recent Changes

Identify files modified in the last 24 hours from **all sources**:

**Option A: Local File Changes (filesystem-based)**
```bash
# Find files modified in the last 24 hours (includes uncommitted changes)
find . -type f -mtime -1 -not -path './.git/*' -not -path './node_modules/*' -not -path './vendor/*'

# Or using git to find modified files (staged and unstaged)
git status --porcelain

# Find recently modified tracked files
git diff --name-only HEAD~10  # Adjust based on commit frequency
```

**Option B: Git History (committed changes)**
```bash
# Get yesterday's date in ISO format
YESTERDAY=$(date -d '1 day ago' '+%Y-%m-%d' 2>/dev/null || date -v-1d '+%Y-%m-%d')

# List recent commits
git log --since="24 hours ago" --pretty=format:"%H %s" --no-merges

# List files changed in recent commits
git log --since="24 hours ago" --name-only --pretty=format:"" | sort -u
```

**Option C: GitHub PRs and Commits (remote changes)**

Use GitHub tools to:
- Search for pull requests merged in the last 24 hours
- Get details of merged PRs to understand what files were changed
- List commits from the last 24 hours to identify modified files

**Combine all sources** to get a comprehensive list of recently modified files.

#### 1.2 Extract Changed Files

From all identified sources (local changes, commits, PRs):
- Collect files modified on the filesystem in the last 24 hours
- List changed files from recent commits
- List changed files from merged PRs
- **Deduplicate** the combined file list
- Focus on source code files (`.go`, `.js`, `.ts`, `.tsx`, `.jsx`, `.py`, `.rb`, `.java`, `.cs`, `.php`, `.cpp`, `.c`, `.rs`, etc.)
- Exclude test files, lock files, generated files, and vendored dependencies

```bash
# Example: Combine filesystem and git changes, filter source files
{
  find . -type f -mtime -1 -not -path './.git/*' 2>/dev/null
  git log --since="24 hours ago" --name-only --pretty=format:"" 2>/dev/null
  git diff --name-only 2>/dev/null
} | sort -u | grep -E '\.(go|js|ts|tsx|jsx|py|rb|java|cs|php|cpp|c|rs)$'
```

#### 1.3 Determine Scope

If **no files were changed in the last 24 hours** (no local modifications, no commits, no merged PRs), exit gracefully:

```
✅ No code changes detected in the last 24 hours.
Code simplifier has nothing to process today.
```

If **files were changed** (from any source: local edits, commits, or PRs), proceed to Phase 2.

### Phase 2: Analyze and Simplify Code

#### 2.1 Review Project Standards

Before simplifying, review the project's coding standards:
- Check for style guides, coding conventions, or contribution guidelines
- Look for language-specific conventions (`STYLE.md`, `CONTRIBUTING.md`, `README.md`)
- Identify established patterns in the codebase

#### 2.2 Perform Code Analysis

For each changed file:

1. **Read the file contents**
2. **Identify refactoring opportunities**:
   - Long functions that could be split
   - Duplicate code patterns
   - Complex conditionals that could be simplified
   - Unclear variable names
   - Missing or excessive comments
   - Non-idiomatic patterns
3. **Design the simplification**:
   - What specific changes will improve clarity?
   - How can complexity be reduced?
   - What patterns should be applied?
   - Will this maintain all functionality?

#### 2.3 Apply Simplifications

Make surgical, focused changes that preserve all original behavior. Use targeted edits rather than full file rewrites.

### Phase 3: Validate Changes

#### 3.1 Run Tests

After making simplifications, run the project's test suite:

```bash
# Common test commands (adapt to the project)
make test          # If Makefile exists
npm test           # For Node.js projects
pytest             # For Python projects
./gradlew test     # For Gradle projects
mvn test           # For Maven projects
cargo test         # For Rust projects
go test ./...      # For Go projects
```

If tests fail:
- Review the failures carefully
- Revert changes that broke functionality
- Adjust simplifications to preserve behavior
- Re-run tests until they pass

#### 3.2 Run Linters

Ensure code style is consistent:

```bash
# Common lint commands (adapt to the project)
make lint          # If Makefile exists
npm run lint       # For Node.js projects
pylint . || flake8 . # For Python projects
cargo clippy       # For Rust projects
golangci-lint run  # For Go projects
```

Fix any linting issues introduced by the simplifications.

#### 3.3 Check Build

Verify the project still builds successfully:

```bash
# Common build commands (adapt to the project)
make build         # If Makefile exists
npm run build      # For Node.js projects
./gradlew build    # For Gradle projects
mvn package        # For Maven projects
cargo build        # For Rust projects
go build ./...     # For Go projects
```

### Phase 4: Create Pull Request

#### 4.1 Determine If PR Is Needed

Only create a PR if:
- ✅ Actual code simplifications were made
- ✅ All tests pass (or no tests exist)
- ✅ Linting is clean (or no linter configured)
- ✅ Build succeeds (or no build step exists)
- ✅ Changes improve code quality without breaking functionality

If no improvements were made or changes broke tests, exit gracefully:

```
✅ Code analyzed from last 24 hours.
No simplifications needed - code already meets quality standards.
```

#### 4.2 Generate PR Description

Use this structure for the PR:

```markdown
## Code Simplification - [Date]

This PR simplifies recently modified code to improve clarity, consistency, and maintainability while preserving all functionality.

### Files Simplified

- `path/to/file1.ext` - [Brief description of improvements]
- `path/to/file2.ext` - [Brief description of improvements]

### Improvements Made

1. **Reduced Complexity**
   - [Specific example]

2. **Enhanced Clarity**
   - [Specific example]

3. **Applied Project Standards**
   - [Specific example]

### Changes Based On

Recent changes from:
- #[PR_NUMBER] - [PR title]
- Commit [SHORT_SHA] - [Commit message]

### Testing

- ✅ All tests pass (or indicate if no tests exist)
- ✅ Linting passes (or indicate if no linter configured)
- ✅ Build succeeds (or indicate if no build step)
- ✅ No functional changes - behavior is identical

### Review Focus

Please verify:
- Functionality is preserved
- Simplifications improve code quality
- Changes align with project conventions
- No unintended side effects

---

*Automated by Code Simplifier Agent*
```

## Common Simplification Patterns

### Reduce Nesting
```go
// Before
func process(data []Item) error {
    if data != nil {
        if len(data) > 0 {
            for _, item := range data {
                if item.Valid {
                    // process
                }
            }
        }
    }
    return nil
}

// After
func process(data []Item) error {
    if len(data) == 0 {
        return nil
    }
    for _, item := range data {
        if !item.Valid {
            continue
        }
        // process
    }
    return nil
}
```

### Replace Nested Ternaries with Switch/If
```javascript
// Before
const status = isError ? 'error' : isLoading ? 'loading' : isSuccess ? 'success' : 'idle';

// After
let status;
if (isError) {
    status = 'error';
} else if (isLoading) {
    status = 'loading';
} else if (isSuccess) {
    status = 'success';
} else {
    status = 'idle';
}
```

### Consolidate Related Logic
```python
# Before
def validate_user(user):
    if user.name is None:
        return False
    if len(user.name) == 0:
        return False
    if user.email is None:
        return False
    if '@' not in user.email:
        return False
    return True

# After
def validate_user(user):
    if not user.name:
        return False
    if not user.email or '@' not in user.email:
        return False
    return True
```

### Improve Variable Names
```go
// Before
func calc(d []int) int {
    t := 0
    for _, v := range d {
        t += v
    }
    return t
}

// After
func calculateSum(values []int) int {
    total := 0
    for _, value := range values {
        total += value
    }
    return total
}
```

## Scope Control Guidelines

- **Focus on recent changes**: Only refine code modified in the last 24 hours (local files, commits, or merged PRs)
- **Include all change sources**: Consider filesystem modifications, staged changes, commits, and merged PRs
- **Don't over-refactor**: Avoid touching unrelated code
- **Preserve interfaces**: Don't change public APIs
- **Incremental improvements**: Make targeted, surgical changes

## Exit Conditions

Exit gracefully without creating a PR if:
- No code was changed in the last 24 hours
- No simplifications are beneficial
- Tests fail after changes
- Build fails after changes
- Changes are too risky or complex

## Quality Checklist

Before creating a PR, verify:

- [ ] Functionality is preserved (all tests pass)
- [ ] Code is clearer and more readable
- [ ] No nested ternary operators introduced
- [ ] Variable names are descriptive
- [ ] Unnecessary complexity removed
- [ ] Project conventions followed
- [ ] Build succeeds
- [ ] Linting passes
- [ ] Changes are focused and surgical

## Output Requirements

The agent MUST either:

1. **If no changes in last 24 hours**: Output a brief status message
2. **If no simplifications beneficial**: Output a brief status message
3. **If simplifications made**: Create a PR with the changes and detailed description
