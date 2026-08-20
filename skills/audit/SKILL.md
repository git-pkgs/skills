---
name: audit
description: Scan dependencies for known vulnerabilities using OSV, check provenance attestations, and find deprecated or yanked versions. Use when checking for CVEs, reviewing security posture, or investigating who introduced or fixed a vulnerable dependency.
allowed-tools: Bash(git-pkgs vulns *) Bash(git-pkgs provenance *) Bash(git-pkgs deprecated *)
---

# Vulnerability Scanning

Scan project dependencies for known vulnerabilities using the OSV database.

## Commands

**Scan for vulnerabilities:**
```bash
git-pkgs vulns scan
```

Options:
- `-e ECO` - filter by ecosystem
- `-s LEVEL` - minimum severity (critical, high, medium, low)
- `-f FORMAT` - output format: text, json, sarif
- `--live` - query OSV directly instead of the cached data
- `--no-sync` - skip auto-sync, use only cached data
- `-c REF` - scan dependencies at a specific commit

**Show details for a specific vulnerability ID:**
```bash
git-pkgs vulns show GHSA-xxxx-xxxx-xxxx
```

**Show who introduced each vulnerable dependency:**
```bash
git-pkgs vulns blame
```

**Show who fixed vulnerabilities** (opposite of blame):
```bash
git-pkgs vulns praise
```

**Compare vulnerabilities between commits** (added vs fixed):
```bash
git-pkgs vulns diff main HEAD
```

**How long has each vulnerability been present:**
```bash
git-pkgs vulns exposure
git-pkgs vulns exposure --summary   # aggregate metrics only
```

**Timeline of commits that introduced or fixed vulnerabilities:**
```bash
git-pkgs vulns log
git-pkgs vulns log --introduced
git-pkgs vulns log --fixed
```

**Vulnerability history for one package across all commits:**
```bash
git-pkgs vulns history lodash
```

**Refresh cached OSV data:**
```bash
git-pkgs vulns sync --force
```

## Provenance

**Check for trusted-publishing attestations** on installed dependencies:
```bash
git-pkgs provenance
git-pkgs provenance --missing   # only packages without provenance
```

Reports verified trusted-publishing signals where the registry exposes them, weaker signature attestations otherwise, and marks unsupported ecosystems explicitly rather than treating missing metadata as verified.

## Deprecated and yanked versions

**Find installed versions that have been deprecated, yanked, or retracted** by the registry:
```bash
git-pkgs deprecated
```

A yanked version is one the maintainer has withdrawn, usually because of a bug or security problem. Treat these as needing an update even if no CVE is filed.

## Responding to vulnerabilities

When vulnerabilities are found, check in order:

1. **Reachability** - is the vulnerable code path actually called? Many CVEs affect features you don't use.
2. **Update** - is a patched version available? Use `git-pkgs outdated <package>` to check.
3. **Override transitive** - force a newer version of the vulnerable transitive dependency.
4. **Fork and patch** - apply the security fix to a fork.
5. **Remove** - find an alternative or inline the functionality.
6. **Accept risk** - document why it's not exploitable in your context.

## When to use

- Before releases or deployments
- During security reviews
- When the user asks about vulnerabilities
- In CI pipelines to catch new CVEs (use `-f sarif` for code-scanning integration)
