---
name: outdated
description: Find dependencies with newer versions available, that haven't been updated in a long time, or whose installed version has been deprecated or yanked. Use when checking for available updates, finding stale packages, or planning a dependency update cycle.
allowed-tools: Bash(git-pkgs outdated *) Bash(git-pkgs stale *) Bash(git-pkgs freshness *) Bash(git-pkgs deprecated *) Bash(git-pkgs changelog *)
---

# Outdated and Stale Dependencies

Find packages that need attention.

## Commands

**Find packages with newer versions available** (queries ecosyste.ms):
```bash
git-pkgs outdated
git-pkgs outdated --major   # only major bumps
git-pkgs outdated --minor   # skip patch-only updates
git-pkgs outdated --at 2026-01-01   # what was outdated on this date
```

**Show release-date lag** (how many days each dependency trails its latest release):
```bash
git-pkgs freshness
git-pkgs freshness -n 20    # top 20 lagging packages
```

**Find packages that haven't been changed in this repo for the longest time:**
```bash
git-pkgs stale
git-pkgs stale --days 365
```

**Find installed versions that are deprecated, yanked, or retracted:**
```bash
git-pkgs deprecated
```

**Show changelog entries for a package** between two versions (before deciding to upgrade):
```bash
git-pkgs changelog lodash --from 4.17.20 --to 4.17.21
git-pkgs changelog pkg:cargo/serde --from 1.0.0
```

Options:
- `-e ECO` - filter by ecosystem
- `-f json` - JSON output

## Interpreting results

Not every outdated package needs updating. Prioritize by:

1. **Security patches** - update immediately, check with `git-pkgs vulns scan`
2. **Deprecated or yanked** - the maintainer has withdrawn this version; move off it
3. **Major versions behind** - plan a migration, read `git-pkgs changelog` for breaking changes
4. **Minor/patch behind** - update in batches, verify tests pass
5. **Stale but working** - stable packages don't need commits; only flag if they have known issues

## When to use

- During dependency maintenance cycles
- Before starting a new feature (to reduce risk of conflicts)
- When planning major upgrades
- When reviewing overall project health
