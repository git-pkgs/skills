# skills

A Go library for finding, parsing and staging [Agent Skills](https://agentskills.io/specification): directories containing a `SKILL.md` file with YAML frontmatter. It works on paths on disk, and on any `fs.FS`, so the same code handles a checkout, an unpacked package, or a package archive opened with [archives](https://github.com/git-pkgs/archives).

## Installation

```bash
go get github.com/git-pkgs/skills
```

## Usage

### Parse one skill

```go
skill, err := skills.Parse("skills/lint")

fmt.Println(skill.Name, skill.Description)
fmt.Println(skill.Files, skill.HasScripts())
```

`Parse` accepts a `SKILL.md` path, the directory holding one, or a standalone markdown file such as `review.md`, and sets `skill.SourcePath` to the absolute directory. `ParseFS(fsys, name)` does the same against an `fs.FS`, with `skill.Dir` relative to it. Spec violations (name format, field lengths, a name that differs from its directory) are recorded in `skill.Warnings` and the skill still loads, as the spec's client guide recommends. Invalid YAML is an error.

Frontmatter keys outside the spec, such as the `library` and `requires` fields TanStack Intent adds or Claude Code's `model`, are kept in `skill.Extra` rather than dropped, and each one adds a warning. `skill.SourceHash` is the SHA-256 of `SKILL.md` followed by any sibling `schema.json`.

### Find skills in a directory tree

```go
found, err := skills.Discover(os.DirFS("."))
```

`Discover` searches up to six directories deep and skips `.git`, `node_modules`, `.venv`, `__pycache__` and `vendor`. A skill that fails to parse is left out and its error joined into `err`, so one bad file does not hide the rest. Use `WithMaxDepth` and `WithSkipDirs` to change the defaults, or `WalkFS` to handle each skill as it is found. `Walk(root, fn)` walks a directory on disk with the same defaults, sets `SourcePath`, and stops at the first error.

### Find skills in a package archive

```go
r, err := archives.OpenWithPrefix("toolkit-2.13.0.tgz", f, "package/")
defer r.Close()

fsys, err := archives.NewFS(r)
found, err := skills.Discover(fsys)
```

This module does not import archives; any `fs.FS` works. Skill paths are relative to the archive root after the prefix is stripped.

### Stage a skill for an agent

```go
if err := skills.ValidateName(skill.Name); err != nil {
    return err
}
err := skills.Stage(nil, skill, workspace, filepath.Join(workspace, ".claude", "skills", skill.Name))
```

`Stage` clears the destination, copies the skill's sibling files from `fsys` (or from `skill.SourcePath` when `fsys` is nil), and writes a rendered `SKILL.md` and `schema.json`. The destination must resolve inside the workspace after following symlinks, and symlinks in the source are refused. Check a name with `ValidateName` before building a destination from it, since `filepath.Join` cleans `../` away before `Stage` sees the path. Extra frontmatter keys are written by default; pass `skills.WithoutExtra()` to stage only spec fields, for example when the skill comes from a repository you are auditing.

### Other helpers

`Render` writes a `Skill` back to `SKILL.md` bytes. `Concat` joins skill bodies into one prompt. `ValidateNamespace` rejects unknown keys under a vendor metadata prefix such as `scrutineer.`. `Match`, `PathIncluded` and friends implement doublestar globs for skills that scope themselves to parts of a repository, and `ParseRepoSpec` parses `owner/repo@ref` skill sources.

## License

[MIT](LICENSE).
