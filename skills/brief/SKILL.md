---
name: brief
description: Detect a project's toolchain, languages, package managers, test runners, linters, and conventions. Use when starting work on a project, onboarding to a codebase, or checking what tools are configured. Also detects tooling gaps and fetches external metadata.
allowed-tools: Bash(brief *)
---

# Project Toolchain Detection

Use `brief` to understand a project's toolchain before doing any work. It detects languages, package managers, test runners, linters, formatters, build tools, CI, and more.

## Commands

**Scan a project:**
```bash
brief .
```

**See what changed on this branch:**
```bash
brief diff
```

**Find missing recommended tooling:**
```bash
brief missing .
```

**Enrich with external metadata** (download counts, dependents, OpenSSF Scorecard, runtime EOL):
```bash
brief enrich .
```

**Scan a remote repo or registry package:**
```bash
brief https://github.com/expressjs/express
brief npm:express
brief gem:rails
brief pypi:requests
```

## Output

JSON when piped, human-readable on a TTY. Force either with `--json` or `--human`. Filter to a single category with `--category test`.

Use `--verbose` to include homepage, docs, and repo links for each detected tool.

## When to use

- At the start of a conversation to understand the project
- Before suggesting tools or configurations
- When the user asks what tools are configured
- When reviewing a branch to understand which parts of the toolchain are affected
- When checking for tooling gaps before setting up CI or improving DX
