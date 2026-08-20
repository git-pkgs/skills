---
name: brief
description: Detect a project's toolchain, languages, package managers, test runners, linters, and conventions. Use when starting work on a project, onboarding to a codebase, or checking what tools are configured. Also detects tooling gaps, threat categories, dangerous sink functions, and packs source trees for LLM context.
allowed-tools: Bash(brief *)
---

# Project Toolchain Detection

Use `brief` to understand a project's toolchain before doing any work. It detects languages, package managers, test runners, linters, formatters, build tools, CI, and more.

## Commands

**Scan a project:**
```bash
brief .
```

**See what changed on this branch:**
```bash
brief diff
```

**Find missing recommended tooling:**
```bash
brief missing .
```

**Enrich with external metadata** (download counts, dependents, OpenSSF Scorecard, runtime EOL):
```bash
brief enrich .
```

**Threat categories implied by the detected stack** (deterministic mapping from tool taxonomy tags to CWE/OWASP categories):
```bash
brief threat-model .
```

**Dangerous sink functions** for the detected stack (SQL exec, HTML rendering, shell spawn, path open) so you know what to grep for:
```bash
brief sinks .
```

**Pack a source tree into one document for LLM context** (directory tree plus each file with function bodies stripped, honouring `.gitignore`):
```bash
brief outline .
brief outline -full .              # keep full bodies
brief outline -ignore "docs/,*.md" .
```

**Inspect a native binary or package archive** (format, architecture, linked libraries, build metadata; accepts ELF, Mach-O, PE, wheels, gems, JARs, tar/zip):
```bash
brief inspect build/tool
brief inspect dist/package.whl
```

**Scan a remote repo or registry package:**
```bash
brief https://github.com/expressjs/express
brief npm:express
brief gem:rails
brief pypi:requests
```

**List everything brief knows about:**
```bash
brief list tools
brief list ecosystems
```

## Output

JSON when piped, human-readable on a TTY. Force with `-json`, `-human`, or `-markdown`. Filter to a single category with `-category test`. Use `-tracked` to ignore files not tracked by git.

Use `-verbose` to include homepage, docs, and repo links for each detected tool.

## When to use

- At the start of a conversation to understand the project
- Before suggesting tools or configurations
- When the user asks what tools are configured
- When reviewing a branch to understand which parts of the toolchain are affected
- When checking for tooling gaps before setting up CI or improving DX
- When starting a security review, to seed the threat model and sink list
