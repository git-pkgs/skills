---
name: manage
description: Add, remove, update, install, replace, or vendor dependencies using the project's detected package manager. Use when the user wants to modify their dependencies.
disable-model-invocation: true
allowed-tools: Bash(git-pkgs add *) Bash(git-pkgs remove *) Bash(git-pkgs update *) Bash(git-pkgs install *) Bash(git-pkgs vendor *) Bash(git-pkgs replace *)
---

# Manage Dependencies

Add, remove, update, install, or vendor dependencies. git-pkgs detects the package manager from lockfiles and runs the appropriate command.

## Commands

**Add a dependency:**
```bash
git-pkgs add lodash
git-pkgs add lodash 4.17.21
git-pkgs add lodash --dev
```

**Remove a dependency:**
```bash
git-pkgs remove lodash
```

**Update dependencies:**
```bash
git-pkgs update           # update all
git-pkgs update lodash    # update one package
```

**Redirect a dependency to a local checkout, git ref, or specific version** for downstream testing:
```bash
git-pkgs replace github.com/acme/lib --path ../lib
git-pkgs replace lodash --git https://github.com/fork/lodash --ref fix-branch
git-pkgs replace lodash 4.17.21
git-pkgs replace github.com/acme/lib --drop   # remove the override
```

**Install from lockfile:**
```bash
git-pkgs install
```

**Vendor dependencies into the project:**
```bash
git-pkgs vendor
```

Common flags on all of the above:
- `-m MANAGER` - override detected package manager
- `-e ECO` - filter to one ecosystem in a polyglot project
- `-x ARG` - pass an extra argument through to the underlying manager
- `--dry-run` - show what would be run without executing

## Before adding a dependency

Check in order:

1. **Standard library** - does the language already provide this?
2. **Transitive cost** - how many dependencies does it bring? Check with `git-pkgs tree` after adding.
3. **Smaller alternative** - is there a focused package that does just what's needed?
4. **Inline it** - can you write 20-50 lines instead of adding a dependency?

Always clarify if a dependency is for development only (`--dev`) to minimize the production attack surface.

## After modifying dependencies

- Run `git-pkgs reindex` to update the database
- Run `git-pkgs vulns scan` to check for new vulnerabilities
- Run tests to verify nothing broke
- Commit the lockfile changes

## When to use

- When the user asks to add, remove, or update a package
- When testing a fix in a dependency by pointing at a local checkout or fork
- When installing dependencies in a freshly cloned repo
- When vendoring for airgapped or reproducible builds
