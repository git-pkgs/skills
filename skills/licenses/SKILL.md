---
name: licenses
description: Show license information for all project dependencies, enforce allow/deny policies, and detect license drift between installed and latest versions. Use when checking license compliance, reviewing license compatibility, or auditing for copyleft or missing licenses.
allowed-tools: Bash(git-pkgs licenses *)
---

# License Information

Check license compliance across all project dependencies. Licenses are normalized to SPDX identifiers.

## Commands

**Show all dependency licenses:**
```bash
git-pkgs licenses
git-pkgs licenses --group   # group output by license
```

**Enforce a policy** (exit non-zero on violation, suitable for CI):
```bash
git-pkgs licenses --allow MIT,Apache-2.0,BSD-3-Clause,ISC
git-pkgs licenses --deny GPL-3.0-only,AGPL-3.0-only
```

**Flag categories:**
```bash
git-pkgs licenses --copyleft     # flag GPL/AGPL
git-pkgs licenses --permissive   # flag anything not permissive
git-pkgs licenses --unknown      # flag packages with no detectable license
```

**Detect license drift** (packages whose license changed between the installed version and the latest release):
```bash
git-pkgs licenses --drift
```

Options:
- `-e ECO` - filter by ecosystem
- `-f FORMAT` - output format: text, json, csv
- `-c REF` - check licenses at a specific commit
- `--offline` - use cached metadata, no network

## License categories

**Permissive** (safe for most uses): MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, Unlicense

**Weak copyleft** (review needed, may affect linking): LGPL-2.1, LGPL-3.0, MPL-2.0

**Strong copyleft** (affects distribution of derivative works): GPL-2.0, GPL-3.0, AGPL-3.0

**No license** means the code is not open source. Do not use packages without a license.

## When to use

- Before releasing software to check license compatibility
- In CI to fail the build on policy violations
- During compliance audits
- When adding new dependencies to verify license terms
- When preparing license attribution notices
