# git-pkgs skills

A [Claude Code plugin](https://code.claude.com/docs/en/plugins) providing dependency management skills powered by [git-pkgs](https://github.com/git-pkgs/git-pkgs) and [brief](https://github.com/git-pkgs/brief).

## Install

```
/plugin install git-pkgs/skills
```

Requires `git-pkgs` and `brief` to be installed:

```bash
brew install git-pkgs
go install github.com/git-pkgs/brief/cmd/brief@latest
```

## Skills

| Skill | Description |
|-------|-------------|
| [`brief`](skills/brief/SKILL.md) | Detect project toolchain, languages, package managers, test runners, linters |
| [`init`](skills/init/SKILL.md) | Initialize, reindex, or manage the git-pkgs database |
| [`list`](skills/list/SKILL.md) | List, search, and locate dependencies |
| [`diff`](skills/diff/SKILL.md) | Compare dependencies between commits, branches, or working tree |
| [`history`](skills/history/SKILL.md) | Track dependency change history, blame, bisect |
| [`audit`](skills/audit/SKILL.md) | Scan for known vulnerabilities via OSV |
| [`outdated`](skills/outdated/SKILL.md) | Find outdated or stale dependencies |
| [`licenses`](skills/licenses/SKILL.md) | License compliance checking |
| [`sbom`](skills/sbom/SKILL.md) | Generate CycloneDX or SPDX SBOMs |
| [`manage`](skills/manage/SKILL.md) | Add, remove, update, install, or vendor dependencies |
| [`evaluate`](skills/evaluate/SKILL.md) | Assess a package's trustworthiness before adding it |

Most skills are triggered automatically by Claude when relevant. `manage` is manual-only since it modifies dependencies.

## Usage

Invoke skills directly:

```
/git-pkgs:brief
/git-pkgs:audit
/git-pkgs:evaluate lodash
```

Or just ask Claude naturally and it will use the right skill:

```
What dependencies does this project have?
Are there any known vulnerabilities?
Is the left-pad package safe to use?
```

## License

MIT
