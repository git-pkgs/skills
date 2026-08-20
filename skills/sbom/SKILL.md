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
git-pkgs sbom --type=spdx
```

Options:
- `--type=TYPE` - SBOM type: `cyclonedx` (default), `spdx`
- `-f FORMAT` - encoding: `json` (default), `xml`
- `--name=NAME` - project name (defaults to git directory name)
- `--version=VER` - project version
- `-c REF` - generate for a specific commit
- `--skip-enrichment` - skip fetching license data from ecosyste.ms

Output goes to stdout; redirect to a file with `> sbom.json`.

## When to use

- When compliance requires an SBOM (EO 14028, EU CRA)
- When sharing a software inventory with customers or auditors
- When feeding into vulnerability scanning or composition analysis tools
- When documenting what ships in a release
