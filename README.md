# git-pkgs skills

A [Claude Code plugin](https://code.claude.com/docs/en/plugins) providing dependency and repository management skills powered by [git-pkgs](https://github.com/git-pkgs/git-pkgs), [brief](https://github.com/git-pkgs/brief), and [forge](https://github.com/git-pkgs/forge).

## Install

Add the marketplace and install the plugin:

```
/plugin marketplace add git-pkgs/skills
/plugin install git-pkgs@git-pkgs
```

Requires `git-pkgs`, `brief`, and `forge` to be installed:

```bash
brew install git-pkgs
go install github.com/git-pkgs/brief/cmd/brief@latest
go install github.com/git-pkgs/forge/cmd/forge@latest
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
| [`forge`](skills/forge/SKILL.md) | Work with GitHub, GitLab, Gitea, Forgejo, and Bitbucket repos through one CLI |

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
