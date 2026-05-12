# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build / test / lint

```bash
CGO_ENABLED=0 go build ./...        # build (CGO is disabled in release too)
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go fmt ./...
CGO_ENABLED=0 go test ./e2e/...     # full e2e suite — the only test package
CGO_ENABLED=0 go test -v -run TestName ./e2e/...   # one test
```

`CGO_ENABLED=0` is required in this environment because gcc isn't installed; it also matches the goreleaser build env, so use it everywhere.

The e2e suite (`e2e/e2e_test.go`) is the project's primary correctness signal — there are no unit tests. `TestMain` compiles the binary to a temp dir once, then each test runs it with a fresh `XDG_CONFIG_HOME` and `HOME` pointing at `t.TempDir()`. New behavior should get an e2e test, not a unit test.

Helpers in `e2e/e2e_test.go` worth reusing:
- `run(t, xdgHome, args...)` — invokes the binary, returns stdout/stderr/exit code.
- `serveSkill(t, content)` / `serve404(t)` — `httptest` servers for the HTTP path.
- `localBareRepo(t)` — `git init --bare` in a temp dir, returns `file://…` URL for the git path.
- `fixture(name)` / `readFixture(t, name)` — load from `e2e/testdata/`.

## Architecture

Three packages, one rule:

- `main.go` — `switch` on `os.Args[1]` to dispatch to `cmd.*`.
- `cmd/` — one file per subcommand (`add`, `list`, `remove`, `update`, `remote`). Owns orchestration: argument parsing, branching, user-facing error messages, `os.Exit`. Calls into `core/` and never invokes `exec.Command` or touches the filesystem directly.
- `core/` — single-purpose primitives. Each function does one thing (one shell call, one file op, one parse step) and returns an error. **No "if A then do B else do C" workflow logic lives here.** When a primitive starts to grow branches, lift them into `cmd/`.

This split is load-bearing — past PRs reverted attempts to push workflow logic into `core`. If you find yourself adding `GitInitAndPushOrPull(…)`, split it instead.

### Where data lives

Skills are canonical under `~/.config/skill-cli/` (via `os.UserConfigDir()` — falls back to `$HOME/.config`):

```
skill-cli/
├── meta/<name>.json     # one JSON per skill: description, raw_url, installed_at, updated_at
├── skills/<name>/SKILL.md
└── .git/                # only if `remote <url>` was run
```

`core.SyncSkillFiles()` mirrors `skills/` into per-assistant deploy targets under `$HOME`: `.claude/skills`, `.cursor/skills`, `.gemini/skills`, `.gemini/antigravity/skills`, `.opencode/skills`, `.codex/skills` — but **only for assistant dirs that already exist** (the top-level `.claude`/`.cursor`/etc. is the existence check). The deploy dirs are treated as derived: anything in them that isn't in `skills/` is deleted. Don't write user state there.

The `name` field in a skill's YAML frontmatter is used as a path segment in three places (`skills/<name>/`, `meta/<name>.json`, deploy mirrors), so it goes through `validateSkillName` (allowlist regex) immediately after `parseFrontmatter` in both `FetchSkill` and `GetLocalSkill`. Keep that check tight if you touch fetch logic.

### Subcommand shape

Every `cmd/*.go` follows the same flow:

1. Validate args, print usage and exit 1 if wrong.
2. `hasRemote := core.HasRemote()`; if true, `core.GitPull()` first.
3. Do the work (fetch, save, remove, …).
4. `core.SyncSkillFiles()` to push canonical → deploy mirrors.
5. If `hasRemote`, `core.GitPush("<verb>: <name>")`.

The pre-work `SyncSkillFiles` was deliberately removed — deploy dirs are recomputed post-write only.

### Git layer

`cmd/remote.go` orchestrates the init / set-origin / merge dance using primitives in `core/git.go` (`GitInitRepo`, `GitAddOrigin`, `IsRemoteEmpty`, `HasLocalChanges`, `GitFetchOrigin`, `GitCommitAllowEmpty`, `GitAddAll`, `GitCommit`, `GitBranchMain`, `GitMergeTheirsMain`, `GitSetUpstreamMain`, `GitCheckoutTrackMain`, `GitPushMain`, `GitSetOrigin`). Don't reach for `exec.Command` from `cmd/` — add a primitive.

`HasRemote()` deliberately swallows errors and returns `false` if git isn't installed or `.git` isn't there. That's the "feature off" path; don't treat its false as fatal.

User-supplied URLs are passed to git as argv with a `--` separator (`git remote add origin -- <url>`) to block `-`-prefixed URLs being read as flags. `ValidateRemoteURL` also enforces an allowlist of schemes (`https://`, `http://`, `ssh://`, `git://`, `file://`, `git@`).

### URL → raw URL translation

`core.GetRawURL` rewrites GitHub/GitLab blob/tree URLs to their `raw.githubusercontent.com` / `gitlab.com/.../-/raw/` equivalents. For `tree/` URLs it appends `/SKILL.md`. Anything else is passed through unchanged. Add new hosts by extending the switch in `getRawURL*`.

### Frontmatter parser

`parseFrontmatter` is a tiny hand-rolled YAML reader for `name` and `description` only. It supports inline scalars and `|`/`>` block scalars but is **not** a full YAML parser — it'll mis-handle nested keys. If a SKILL.md needs more, prefer fixing the SKILL.md over expanding the parser.

## Conventions (enforced in past PRs)

- **Dispatch on a subcommand or argument value uses `switch/case/default`**, never `if/else if`. This applies to `cmd/` and `main.go` where the CLI branches on a string arg. Ordinary `if/else if` chains inside `core/` (e.g. branching on parsed state) are fine.
- **Don't call functions inside `if` conditions** — assign the return value to a variable first, then branch on it. Applies to `err`, `bool`, anything.
- HTTP body reads in `core/fetch.go` go through `io.LimitReader(resp.Body, maxSkillBytes+1)`; the package-level `httpClient` has a 30s timeout. Use them — don't reach for `http.Get` directly.

## Release

CI is `.github/workflows/release.yml` driven by goreleaser. Tag a commit to publish; goreleaser builds `linux/darwin/windows × amd64/arm64` (no windows-arm64) with `CGO_ENABLED=0` and `-s -w` ldflags, and creates a draft GitHub release under `MescoCzubinski/skill-cli`.
