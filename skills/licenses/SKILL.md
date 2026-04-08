---
name: licenses
description: Show license information for all project dependencies. Use when checking license compliance, reviewing license compatibility, or auditing for copyleft or missing licenses.
allowed-tools: Bash(git-pkgs licenses *)
---

# License Information

Check license compliance across all project dependencies. Licenses are normalized to SPDX identifiers.

## Commands

**Show all dependency licenses:**
```bash
git-pkgs licenses
```

Options:
- `--ecosystem=ECO` - filter by ecosystem
- `--json` - JSON output

## License categories

**Permissive** (safe for most uses): MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, Unlicense

**Weak copyleft** (review needed, may affect linking): LGPL-2.1, LGPL-3.0, MPL-2.0

**Strong copyleft** (affects distribution of derivative works): GPL-2.0, GPL-3.0, AGPL-3.0

**No license** means the code is not open source. Do not use packages without a license.

## When to use

- Before releasing software to check license compatibility
- During compliance audits
- When adding new dependencies to verify license terms
- When preparing license attribution notices
