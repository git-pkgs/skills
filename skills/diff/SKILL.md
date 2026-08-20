---
name: diff
description: Compare dependencies between commits, branches, or the working tree. Use when reviewing what dependencies changed in a PR, between releases, or since a specific commit.
allowed-tools: Bash(git-pkgs diff *) Bash(git-pkgs diff-file *) Bash(git-pkgs diff-driver *) Bash(git-pkgs show *)
---

# Compare Dependencies

See what dependencies changed between two points in time.

## Commands

**Compare HEAD against working tree** (like `git diff`):
```bash
git-pkgs diff
```

**Compare between commits or branches:**
```bash
git-pkgs diff main..feature
git-pkgs diff --from=HEAD~10
git-pkgs diff --from=v1.0.0 --to=v2.0.0
```

**Show dependency changes in a specific commit:**
```bash
git-pkgs show
git-pkgs show abc1234
```

**Compare two manifest or lockfiles directly:**
```bash
git-pkgs diff-file Gemfile.lock.old Gemfile.lock
```

**Install a git textconv driver** so plain `git diff` on lockfiles shows sorted dependency lists instead of raw churn:
```bash
git-pkgs diff-driver --install
```

Options:
- `-e ECO` - filter by ecosystem
- `-t TYPE` - filter by dependency type (runtime, development)
- `--kind manifest|lockfile` - only compare declarations or only resolved versions
- `--stat` - aggregate change counts instead of full listing
- `-f json` - JSON output

## When to use

- Reviewing dependency changes in a pull request
- Comparing dependencies between releases
- Checking what changed after running an update
- Understanding the dependency impact of a specific commit
