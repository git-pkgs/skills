---
name: outdated
description: Find dependencies with newer versions available or that haven't been updated in a long time. Use when checking for available updates, finding stale packages, or planning a dependency update cycle.
allowed-tools: Bash(git-pkgs outdated *) Bash(git-pkgs stale *)
---

# Outdated and Stale Dependencies

Find packages that need attention.

## Commands

**Find packages with newer versions available** (queries ecosyste.ms):
```bash
git-pkgs outdated
```

**Find packages that haven't been changed in the longest time:**
```bash
git-pkgs stale
```

Options:
- `--ecosystem=ECO` - filter by ecosystem
- `--json` - JSON output

## Interpreting results

Not every outdated package needs updating. Prioritize by:

1. **Security patches** - update immediately, check with `git-pkgs vulns scan`
2. **Major versions behind** - plan a migration, check changelogs for breaking changes
3. **Minor/patch behind** - update in batches, verify tests pass
4. **Stale but working** - stable packages don't need commits; only flag if they have known issues

## When to use

- During dependency maintenance cycles
- Before starting a new feature (to reduce risk of conflicts)
- When planning major upgrades
- When reviewing overall project health
