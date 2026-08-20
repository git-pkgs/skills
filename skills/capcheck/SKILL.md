---
name: capcheck
description: Fail the build when Go code or a dependency gains a new privileged capability (spawn a process, open a socket, call cgo, write files). Use in Go projects when reviewing dependency updates for supply-chain risk, setting up capability baselines, or investigating why CI is reporting a capability drift.
allowed-tools: Bash(capcheck *)
---

# Go Capability Drift

`capcheck` records the set of privileged operations reachable from a Go module (via [google/capslock](https://github.com/google/capslock)) into a lock file, then fails when that set grows. `govulncheck` reports CVEs; `capcheck` reports when a dependency gains the ability to do something it couldn't before, whether or not a CVE exists yet.

Go only. Run from the module root.

## Commands

**First-time setup** (writes `capcheck.json` config and `capcheck.lock.json` baseline; commit both):
```bash
capcheck init ./...
```

**Check against the baseline** (default command, exit 1 on new capabilities):
```bash
capcheck ./...
```

**Accept the current state** after reviewing a reported change:
```bash
capcheck update ./...
```

**Print current capabilities** without a baseline:
```bash
capcheck list ./...
```

Flags (all commands):
- `-f text|json|github` - output format; `github` emits workflow annotations
- `--strict` - also fail on removed capabilities
- `--ignore CAP` - ignore a capability (repeatable, stacks with config)
- `--granularity package|function` - `function` is more precise but noisier
- `-C DIR` - run as if in DIR
- `--baseline PATH` - lock file path

## Configuration

`capcheck.json` (all keys optional):

```json
{
  "granularity": "package",
  "timeout": "5m",
  "goos": "linux",
  "goarch": "amd64",
  "ignore": ["FILES", "NETWORK", "REFLECT", "RUNTIME"]
}
```

Most projects ignore the noisy capabilities and watch for `EXEC`, `CGO`, `ARBITRARY_EXECUTION`, `UNSAFE_POINTER`, and `MODIFY_SYSTEM_STATE`. `ignore` matches hierarchically, so `MODIFY_SYSTEM_STATE` also covers `MODIFY_SYSTEM_STATE/ENV`.

Results depend on which stdlib files compile in, so pin `goos`/`goarch` to whatever CI runs on and keep it fixed; a lock file written on macOS will not match one written on Linux.

## GitHub Action

```yaml
- uses: git-pkgs/capcheck@v1
  with:
    packages: ./...
```

Runs `check --format github` and annotates the first line of user code in each new capability's call path.

## Reading a failure

```
github.com/you/app gained EXEC
  cmd/server/main.go:42:3   main.run
  handler/upload.go:88:5    handler.processImage
                            github.com/disintegration/imaging.Resize
                            os/exec.Command
```

The path shows how user code reaches the new capability. If the change is expected, `capcheck update ./...` and commit the lock file. If it came from a dependency bump you didn't write, treat it the same as a new advisory: read the dependency's diff before accepting.

## When to use

- Setting up supply-chain checks in a Go project's CI
- Reviewing why a dependency bump added exec/network/cgo access
- Auditing a Go module for what privileged operations it can reach
