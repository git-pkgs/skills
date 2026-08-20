---
name: pin
description: Vendor browser assets (htmx, CSS kits, icon sets) from npm packages or GitHub releases without npm, node_modules, or install scripts. Use when a server-rendered project needs a handful of frontend files, when adding or updating a pinned asset, or when verifying vendored files against their lockfile.
allowed-tools: Bash(pin *)
---

# Vendor Browser Assets Without npm

`pin` fetches named files from published packages, verifies them against the registry tarball hash, writes them under a vendored directory, and records a `pin.lock` that is also a valid CycloneDX SBOM. No `node_modules`, no transitive dependencies, no lifecycle hooks executed on install.

## Manifest

`pin.yaml`:

```yaml
out: "internal/web/static/vendor"

assets:
  - name: "htmx.org"
    version: "^2.0"
    files: ["dist/htmx.min.js"]

  - name: "@tailwindcss/browser"
    version: "4.1.13"

  - name: "highlight.js"
    version: "11.11.1"
    source: "github:highlightjs/cdn-release"
    files: ["build/highlight.min.js", "build/styles/github.min.css"]
```

`version` accepts exact pins (`2.0.6`), semver ranges (`^2.0`, `~0.3.11`), or npm dist-tags. When `files` is omitted for an npm source, the entry point is picked from the package's `jsdelivr || unpkg || browser || module || main` field. `source:` can be `github:owner/repo` (resolved to a commit SHA) or `url:https://...` (hashed on first fetch, verified thereafter).

## Commands

**Write a starter manifest:**
```bash
pin init
```

**Resolve, fetch, verify, write lockfile and vendored files** (alias `pin install`):
```bash
pin sync
pin sync --frozen       # CI: fail before any network if manifest and lockfile disagree
pin sync --no-fetch     # frozen + re-hash on-disk files, no network at all
pin sync --dry-run
```

**Add or remove an asset** (edits `pin.yaml` and syncs):
```bash
pin add htmx.org@^2.0 dist/htmx.min.js
pin add lodash --exact
pin rm htmx.org
```

**Bump within the manifest range** (ignores the lock):
```bash
pin update            # all
pin update htmx.org   # one
```

**Check for newer versions, deprecations, yanks, license drift:**
```bash
pin outdated
```

**Re-hash vendored files against the lockfile** (exit 4 on drift):
```bash
pin verify
pin verify --strict   # also re-derive each hash from the registry tarball
```

**Inspect the lock:**
```bash
pin list
pin path htmx.org     # on-disk paths for one package
```

**Emit the lockfile as an SBOM:**
```bash
pin sbom -f cyclonedx
pin sbom -f spdx -o sbom.json
```

## Provenance

`pin sync` records npm sigstore attestations in the lockfile when the registry publishes them. Tighten with:

- `--strict-provenance` - fail if any npm entry has no attestation
- `--verify-provenance` - cryptographically verify each bundle against the live Sigstore TUF root
- `--require-publisher-matches-repository` - fail if the attestation's build workflow lives on a different repo than the package's declared `repository.url`
- `--signature-mode enforce` - fail on missing or invalid `dist.signatures`

## When to use

- A Rails, Django, Phoenix, or Go-rendered app needs a few JS/CSS files and nothing else from the npm ecosystem
- Adding, bumping, or removing a pinned frontend asset
- CI wants to assert vendored files match the lockfile without network access
- Producing an SBOM that covers vendored browser assets
