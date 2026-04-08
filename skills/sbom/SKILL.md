---
name: sbom
description: Generate a Software Bill of Materials in CycloneDX or SPDX format. Use when producing SBOMs for compliance, supply chain transparency, or software composition analysis.
allowed-tools: Bash(git-pkgs sbom *)
---

# Generate SBOM

Produce a Software Bill of Materials listing all dependencies and their metadata.

## Commands

**Generate SBOM in CycloneDX format** (default):
```bash
git-pkgs sbom
```

**Generate SBOM in SPDX format:**
```bash
git-pkgs sbom --format=spdx
```

Options:
- `--format=FORMAT` - output format: `cyclonedx` (default), `spdx`
- `--output=FILE` - write to file instead of stdout
- `--json` - JSON output (default for both formats)

## When to use

- When compliance requires an SBOM (EO 14028, EU CRA)
- When sharing a software inventory with customers or auditors
- When feeding into vulnerability scanning or composition analysis tools
- When documenting what ships in a release
