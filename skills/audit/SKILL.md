---
name: audit
description: Scan dependencies for known vulnerabilities using OSV. Use when checking for CVEs, reviewing security posture, or investigating who introduced a vulnerable dependency.
allowed-tools: Bash(git-pkgs vulns *)
---

# Vulnerability Scanning

Scan project dependencies for known vulnerabilities using the OSV database.

## Commands

**Scan for vulnerabilities:**
```bash
git-pkgs vulns scan
```

**Show who introduced each vulnerable dependency:**
```bash
git-pkgs vulns blame
```

Options:
- `--ecosystem=ECO` - filter by ecosystem
- `--severity=LEVEL` - filter by severity (critical, high, medium, low)
- `--json` - JSON output

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
- In CI pipelines to catch new CVEs
