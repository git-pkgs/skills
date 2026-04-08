---
name: history
description: Track dependency change history, find who added each dependency, and understand why packages were introduced. Use when investigating when a dependency was added or changed, who introduced it, or which commits modified dependencies.
allowed-tools: Bash(git-pkgs history *) Bash(git-pkgs log *) Bash(git-pkgs blame *) Bash(git-pkgs why *) Bash(git-pkgs bisect *) Bash(git-pkgs stats *)
---

# Dependency History

Understand how and why dependencies changed over time.

## Commands

**Show change history for a package** (or all packages):
```bash
git-pkgs history lodash
git-pkgs history
```

**List commits that modified dependencies:**
```bash
git-pkgs log
```

**Show who added each dependency:**
```bash
git-pkgs blame
```

**Show why a dependency was added** (commit, author, message):
```bash
git-pkgs why lodash
git-pkgs why pkg:npm/lodash
```

**Binary search for a dependency-related change** (like `git bisect` but only considers dependency commits):
```bash
git-pkgs bisect start
git-pkgs bisect good v1.0.0
git-pkgs bisect bad HEAD
```

**Show aggregate dependency statistics:**
```bash
git-pkgs stats
```

## When to use

- When investigating when or why a dependency was added
- When finding who to ask about a specific package
- When tracking down when a dependency change caused a problem
- When reviewing the overall dependency change velocity of a project
