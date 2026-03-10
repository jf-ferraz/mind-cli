# Git Discipline

Standards for commit practices, branch management, and change tracking.

## Known-Good Increment

Every commit leaves the codebase in a working state:
- All tests pass
- Code compiles / lints / type-checks
- No partial implementations exposed (use feature flags if necessary)
- Could be deployed without immediate rollback

**Never commit a broken state.** If you're in the middle of a change and need to switch context, either finish the current task or stash your changes.

## Commit Practices

### When to Commit
- After completing a logical unit of work
- Before starting a refactor (clean rollback point)
- Before context switches (switching to a different task or agent)
- At natural stopping points during long implementations
- After every passing test cycle during TDD

### Commit Messages
```
{type}: {concise description}

{optional body — what and why, not how}
```

**Types**: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `wip`

Examples:
```
feat: add user authentication endpoint
fix: handle null response from payment gateway
refactor: extract validation logic into shared module
test: add regression tests for order calculation
docs: update API documentation for v2 endpoints
wip: planning complete — analyst + architect done
```

The `wip:` type is reserved for session splits. It signals an incomplete workflow that will resume in a new session.

### What NOT to Commit
- Debug logging or temporary test files (clean exit invariant)
- Commented-out code (delete it — git has history)
- Generated files that should be in .gitignore
- Secrets, credentials, or environment-specific configuration

### Commit Hashes in changes.md

The developer records commit short hashes in `docs/iterations/{descriptor}/changes.md` so the reviewer can trace each change to its commit:

```markdown
| File | Change | Reason | Commit |
|------|--------|--------|--------|
| src/auth.py | Add JWT validation | FR-3 | a1b2c3d |
```

## Branch Strategy

Branches align with iteration descriptors:

```
feature/{descriptor}     # NEW_PROJECT and ENHANCEMENT
bugfix/{descriptor}      # BUG_FIX
refactor/{descriptor}    # REFACTOR
```

Examples:
```
feature/user-auth
feature/barcode-scanning
bugfix/login-500-error
refactor/data-layer
```

### Branch Lifecycle

1. **Created** by orchestrator at iteration start (Step 4)
2. **Worked on** by developer with atomic commits
3. **Completed** when reviewer signs off
4. **Merged** via PR (orchestrator generates PR summary)

## During Agent Workflow

- **Orchestrator**: Creates the branch, commits iteration overview, commits at session splits (`wip:`), commits at completion
- **Developer**: Commits after completing each logical unit of implementation (stage + commit per unit)
- **Tester**: Commits after test suite additions
- **Reviewer**: Does not commit — only reviews

If the workflow is interrupted (context limit, error, user pause):
1. Commit current state with message: `wip: {what was in progress}`
2. Update `docs/state/workflow.md` with full structured handoff
3. Run `mind lock` to sync the lock file (if manifest exists)
4. The workflow can resume from the last commit via `docs/state/workflow.md`

## PR Summary

After the reviewer signs off, the orchestrator generates a PR summary:

```markdown
## Summary
{What was done — 2-3 sentences}

## Changes
{Key files modified/created — from changes.md}

## Test Results
{Pass/fail counts from deterministic gate}

## Reviewer Assessment
{APPROVED / APPROVED_WITH_NOTES / NEEDS_REVISION}
```
