---
name: list
description: List, search, and locate dependencies in the current project. Use when exploring what packages are installed, finding where a dependency is declared, searching for packages by name pattern, or viewing the dependency tree.
allowed-tools: Bash(git-pkgs list *) Bash(git-pkgs tree *) Bash(git-pkgs search *) Bash(git-pkgs where *) Bash(git-pkgs browse *) Bash(git-pkgs urls *) Bash(git-pkgs notes *) Bash(git-pkgs integrity *)
---

# List and Search Dependencies

Query the current state of dependencies in a project.

## Commands

**List all dependencies at HEAD:**
```bash
git-pkgs list
```

Options:
- `--commit=REF` - list at a specific commit
- `--ecosystem=ECO` - filter by ecosystem
- `--manifest=PATH` - filter by manifest file
- `--json` - JSON output

**Show dependency tree** (grouped by manifest and type):
```bash
git-pkgs tree
```

**Search by name pattern:**
```bash
git-pkgs search lodash
git-pkgs search "react-*"
```

**Find where a package is declared** (file path, line number, content):
```bash
git-pkgs where lodash
```

**Open a package's source in your editor:**
```bash
git-pkgs browse lodash
```

**Show registry URLs** (registry page, download, docs, PURL):
```bash
git-pkgs urls lodash
git-pkgs urls pkg:npm/lodash@4.17.21
```

**Show lockfile integrity hashes** and detect drift:
```bash
git-pkgs integrity
```

**Manage notes on packages:**
```bash
git-pkgs notes add pkg:npm/lodash "reviewed for security"
git-pkgs notes list pkg:npm/lodash
```

## When to use

- When the user asks what dependencies are installed
- When looking for a specific package in the project
- When checking which manifest file declares a dependency
- When exploring the dependency structure
