---
name: forge
description: Interact with GitHub, GitLab, Gitea, Forgejo, Codeberg, and Bitbucket repositories through one CLI. Use instead of `gh` or `glab` when the git remote is not github.com, when working against a self-hosted forge, or when the same workflow needs to run unchanged across forges. Covers pull/merge requests, issues, releases, CI pipelines, labels, and repo management.
allowed-tools: Bash(forge *)
---

# Multi-Forge CLI

`forge` is a single CLI that talks to GitHub, GitLab, Gitea, Forgejo, and Bitbucket Cloud. Commands are modelled on `gh`, so most `gh` invocations work by swapping the binary name. The forge backend is detected from the git remote, so the same command works regardless of where the repo is hosted.

Check `git remote -v` first. If `origin` is not `github.com`, use `forge` rather than `gh`.

## Repository selection

The current repo is inferred from the `origin` remote. Override with:

- `-R OWNER/REPO` or `-R HOST/OWNER/REPO` to target a specific repo
- `--remote NAME` to use a different remote
- `--host HOST` to force a host
- `--forge-type github|gitlab|gitea|forgejo|bitbucket` when auto-detection fails

Output is a table by default. Use `-o json` for machine-readable output or `-o plain` for scripting.

## Commands

**Authentication:**
```bash
forge auth login         # store credentials for a domain (interactive)
forge auth status        # list configured domains
```

Tokens are also read from `GITHUB_TOKEN`, `GITLAB_TOKEN`, `GITEA_TOKEN`, `FORGEJO_TOKEN`, `BITBUCKET_TOKEN`, or `FORGE_TOKEN`.

**Pull / merge requests** (`mr` is an alias for `pr`):
```bash
forge pr list
forge pr view 42
forge pr diff 42
forge pr checkout 42
forge pr create --title "..." --body "..."
forge pr edit 42 --title "..."
forge pr merge 42
forge pr review approve 42
forge pr comment 42 --body "..."
```

**Issues:**
```bash
forge issue list --state open
forge issue view 17
forge issue create --title "..." --body "..."
forge issue edit 17 --label bug
forge issue close 17
forge issue comment 17 --body "..."
```

**Repositories:**
```bash
forge repo view
forge repo clone OWNER/REPO
forge repo fork
forge repo list OWNER
forge repo create NAME --private
```

**CI / pipelines:**
```bash
forge ci list
forge ci view RUN_ID       # prints job IDs
forge ci log JOB_ID
forge ci retry RUN_ID
forge ci cancel RUN_ID
```

**Releases:**
```bash
forge release list
forge release view v1.2.0
forge release create --tag v1.2.0 --title "..." --body "..."
forge release upload v1.2.0 FILE
forge release download v1.2.0 ASSET
```

**Other resources:**
```bash
forge label list | create | edit | delete
forge milestone list | create | edit | close
forge branch list | delete
forge notification list | read
forge secret set --name NAME --value VALUE
forge deploy-key add | list | delete
forge collaborator add | list | remove
forge status list SHA
forge file cat PATH                # print file content (use --ref for non-default branch)
forge file ls [PATH]               # list directory entries
forge search repos QUERY
forge api PATH                     # raw authenticated API request
forge browse                       # open in browser
forge rate-limit
```

## Unsupported operations

Not every forge supports every operation. When a backend cannot do something (e.g. Bitbucket milestones), `forge` exits non-zero with an "operation not supported by this forge" error. Treat that as a signal to fall back to the forge's native CLI or web UI rather than retrying.

## When to use

- The git remote points at gitlab.com, codeberg.org, gitea.com, bitbucket.org, or a self-hosted instance
- The user mentions GitLab, Gitea, Forgejo, Codeberg, or Bitbucket
- A merge request rather than a pull request
- The same script or workflow needs to run against multiple forges
- `gh` fails because the repo is not on GitHub
