---
name: init
description: Initialize, reindex, or manage the git-pkgs dependency database. Use when setting up git-pkgs in a repository for the first time, updating the database after pulling new commits, or checking database status.
allowed-tools: Bash(git-pkgs init *) Bash(git-pkgs reindex *) Bash(git-pkgs upgrade *) Bash(git-pkgs hooks *) Bash(git-pkgs info *) Bash(git-pkgs ecosystems *)
---

# Initialize git-pkgs

Set up and maintain the git-pkgs dependency database for a repository.

## Commands

**First-time setup** (walks git history, builds SQLite database in `.git/pkgs.sqlite3`):
```bash
git-pkgs init
```

Options:
- `--branch=NAME` - analyze a specific branch
- `--since=SHA` - start from a specific commit
- `--force` - rebuild from scratch
- `--no-hooks` - skip installing git hooks

**Update after pulling new commits:**
```bash
git-pkgs reindex
```

**Check database status:**
```bash
git-pkgs info
```

**Upgrade database schema** (rebuilds if schema changed):
```bash
git-pkgs upgrade
```

**Manage git hooks** for automatic reindexing:
```bash
git-pkgs hooks install
git-pkgs hooks remove
```

**List supported ecosystems:**
```bash
git-pkgs ecosystems
```

## When to use

- When a user wants to start tracking dependencies in a repo
- When `git-pkgs` commands fail because no database exists
- After pulling or rebasing to update the index
- When the database needs upgrading after a git-pkgs version update
