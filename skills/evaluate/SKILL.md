---
name: evaluate
description: Evaluate a package before adding it as a dependency. Checks registry metadata, adoption, maintenance, security posture, and license. Use when assessing whether a package is trustworthy, comparing alternatives, or reviewing a new dependency.
allowed-tools: Bash(git-pkgs urls *) Bash(git-pkgs licenses *) Bash(git-pkgs vulns *) Bash(git-pkgs health *) Bash(git-pkgs maintainers *) Bash(git-pkgs provenance *) Bash(git-pkgs changelog *) Bash(curl *) Bash(git-pkgs search *)
---

# Evaluate a Package

Assess a package's trustworthiness and suitability before adding it as a dependency.

## Before suggesting any package

**Verify it exists on the registry.** Check that the name, maintainer, and purpose match expectations. Do not hallucinate package names. If verification fails (registry unreachable, metadata incomplete), say so and explain why.

## Evaluation steps

### 1. Check registry metadata

```bash
curl -s "https://packages.ecosyste.ms/api/v1/registries/<registry>/packages/<package>" | jq '{
  dependent_repos: .dependent_repos_count,
  dependent_packages: .dependent_packages_count,
  latest: .latest_release_number,
  latest_release: .latest_release_published_at,
  created: .first_release_published_at,
  maintainers: (.maintainers | length),
  repo: .repository_url,
  license: .normalized_licenses,
  advisories: (.advisories | length),
  archived: .repo_metadata.archived
}'
```

Registries: `npmjs.org`, `pypi.org`, `rubygems.org`, `crates.io`, `proxy.golang.org`, `nuget.org`, `repo1.maven.org`

### 2. Check risk thresholds

| Field | Threshold | Risk |
|-------|-----------|------|
| `latest_release_published_at` | >2 years ago | Possibly abandoned |
| `dependent_repos_count` | <500 | Low adoption |
| `first_release_published_at` | <90 days ago | Too new |
| `maintainers` (length) | <2 | Bus factor risk |
| `versions_count` | 1 | Immature |
| `repo_metadata.archived` | true | Abandoned |
| `advisories` | non-empty | Known vulnerabilities |

### 3. Check OpenSSF Scorecard

```bash
curl -s "https://api.scorecard.dev/projects/github.com/<owner>/<repo>" | jq '{
  score: .score,
  maintained: (.checks[] | select(.name == "Maintained") | .score),
  dangerous_workflow: (.checks[] | select(.name == "Dangerous-Workflow") | .score),
  code_review: (.checks[] | select(.name == "Code-Review") | .score)
}'
```

Score below 5 warrants a closer look. If `Maintained` or `Dangerous-Workflow` scores are below 3, flag as higher risk.

### 4. Check registry links and advisories

```bash
git-pkgs urls <package>
```

Gives registry page, download URL, docs, source repo, and PURL. Follow the registry page for open advisories.

### 5. Review recent changes

```bash
git-pkgs changelog <package> -e <ecosystem>
```

### 6. Check license

Verify the license is compatible with the project. See the licenses skill for categories.

## Red flags

- Package less than 90 days old
- Maintainer account created recently
- Name similar to popular package (typosquatting)
- No license or non-OSI license
- Very few downloads for claimed purpose
- Runs code on install (postinstall scripts, setup.py)
- Single version published

## Typosquatting patterns

Watch for: character substitution (`djang0` vs `django`), character omission (`loadsh` vs `lodash`), homoglyphs (`pyp1` vs `pypi`), delimiter variation (`cross-env` vs `crossenv`), scope confusion (`@angular-devkit/core` vs `@angulardevkit/core`), combosquatting (`lodash-js`, `axios-api`).

## Good signals

- High dependent count (other packages trust it)
- Few direct dependencies
- Multiple maintainers
- OSI-approved license
- Provenance attestation

## Evaluating already-installed dependencies

When the package is already in the tree, git-pkgs can score the whole set without hand-rolled curl:

```bash
git-pkgs health --threshold 40   # maintenance health scores at or below 40
git-pkgs maintainers --single    # bus-factor risk
git-pkgs provenance --missing    # missing trusted-publishing attestation
```

## Less reliable signals

- GitHub stars (gameable)
- Download counts (gameable)
- Commit frequency (stable packages don't need commits)

## When to use

- Before adding any new dependency
- When comparing alternative packages
- When the user asks if a package is safe or trustworthy
- When reviewing dependency changes in a PR
