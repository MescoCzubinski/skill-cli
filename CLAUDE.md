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

The e2e suite is the project's primary correctness signal — there are no unit tests. `TestMain` (in `e2e/e2e_test.go`) compiles the binary to a temp dir once, then each test runs it with a fresh `XDG_CONFIG_HOME` and `HOME` pointing at `t.TempDir()`. The suite covers every command and its main options (local file, local dir, URL, multi-file, CLAUDE.md, `--all`, URL override, …) but **deliberately does not exercise every exception path** (oversize bodies, raw-URL rejection, 404s, malformed frontmatter, reserved names, …). Add a test when you add a new option; don't add one for every new error case.

Helpers in `e2e/e2e_test.go` worth reusing:

- `run(t, xdgHome, args...)` — invokes the binary, returns stdout/stderr/exit code.
- `getGHFake(t)` — per-test singleton fake. Backs a single `httptest` server that serves both the GitHub trees API (`/api/repos/...`) and raw blobs (`/raw/...`). Sets `SKILL_CLI_GITHUB_{API,RAW}_BASE` env vars so the subprocess routes there. `g.register(map[string]string{...})` adds a virtual repo and returns a user-facing `https://github.com/.../tree/main` URL.
- `serveSkill(t, content)` — convenience wrapper around `getGHFake`: registers a single-file SKILL.md repo, returns the tree URL.
- `serveClaude(t, content)` — registers a CLAUDE.md in a virtual repo, returns a `blob/.../CLAUDE.md` URL.
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
├── skills/<name>/       # multi-file skill: SKILL.md + any .md files / subdirs
│   ├── SKILL.md
│   └── <extra>.md
├── claude/              # only if a CLAUDE.md is installed
│   └── CLAUDE.md
└── .git/                # only if `remote <url>` was run
```

A skill is always a directory tree. `core.SaveSkill(name, files)` diffs the incoming tree against the on-disk tree (`readSkillTree` + `treesEqual`); if equal it returns `(false, nil)` to signal "unchanged", otherwise it removes the existing dir and writes the full map. That means **removed files in an update genuinely disappear**, and `update` can print `unchanged: <name>` without further bookkeeping. `core.SaveClaude(files)` is the single-file analogue keyed at `CLAUDE.md`, with the same change-detection contract.

The `raw_url` field in meta holds the **original user-supplied URL** (e.g. `https://github.com/.../tree/main/skills/foo`, a `blob/.../SKILL.md` URL, or a local path) — never a `raw.githubusercontent.com` URL. `update` re-resolves this through `core.ResolveSkillSource(skill, overrideURL)`, which prefers the override and falls back to the stored URL; an empty stored URL means "installed from a local file with no source" and is an error.

`core.SyncSkillFiles()` mirrors `skills/` into per-assistant deploy targets under `$HOME`: `.claude/skills`, `.cursor/skills`, `.gemini/skills`, `.gemini/antigravity/skills`, `.opencode/skills`, `.codex/skills` — but **only for assistant dirs that already exist** (the top-level `.claude`/`.cursor`/etc. is the existence check). The deploy dirs are treated as derived: anything in them that isn't in `skills/` is deleted. Don't write user state there. `core.SyncClaude()` is the one-file analogue that mirrors `claude/CLAUDE.md` to `~/.claude/CLAUDE.md` (only if `~/.claude/` exists; absence of the canonical file means delete the deploy copy).

The `name` field in a skill's YAML frontmatter is used as a path segment in three places (`skills/<name>/`, `meta/<name>.json`, deploy mirrors), so `cmd/add.go` runs `core.ValidateSkillName` (allowlist regex `^[a-z0-9][a-z0-9._-]{0,63}$`, plus a hard reject of the reserved `"claude"`) immediately after `ParseFrontmatter`. `cmd/update.go` does **not** re-validate — the name comes from on-disk meta (already validated at install time), not the fetched frontmatter; when an override URL is given, update only checks `parsedName == skill.Name` to prevent accidental clobber. `SaveSkill` also validates every relative path (`filepath.Clean`, no `..`, no absolute, must stay inside `skills/<name>/`) as defense against a malformed tree response.

### Subcommand shape

Every `cmd/*.go` follows the same flow:

1. Validate args, print usage and exit 1 if wrong.
2. `hasRemote := core.HasRemote()`; if true, `core.GitPull()` first.
3. Do the work (fetch, save, remove, …).
4. `core.SyncSkillFiles()` then `core.SyncClaude()` to push canonical → deploy mirrors.
5. If `hasRemote`, `core.GitPush("<verb>: <name>")`.

The pre-work sync was deliberately removed — deploy dirs are recomputed post-write only.

`cmd/update.go` is the one variant: it accumulates `changed`/`hadError` across all targets (CLAUDE.md first, then skills), prints per-target `updated:` / `unchanged:` / error lines, and exits non-zero at the end if any single update failed — `--all` continues past individual failures rather than aborting on the first.

### Git layer

`cmd/remote.go` orchestrates the init / set-origin / merge dance using primitives in `core/git.go` (`GitInitRepo`, `GitAddOrigin`, `IsRemoteEmpty`, `HasLocalChanges`, `GitFetchOrigin`, `GitCommitAllowEmpty`, `GitAddAll`, `GitCommit`, `GitBranchMain`, `GitMergeTheirsMain`, `GitSetUpstreamMain`, `GitCheckoutTrackMain`, `GitPushMain`, `GitSetOrigin`, `GitPull`, `GitPush`, `IsGitRepo`, `GitAvailable`). Don't reach for `exec.Command` from `cmd/` — add a primitive.

`HasRemote()` deliberately swallows errors and returns `false` if git isn't installed or `.git` isn't there. That's the "feature off" path; don't treat its false as fatal. `GitAvailable()` is the explicit check used by `remote` to surface a hard error when git is missing.

User-supplied URLs are passed to git as argv with a `--` separator (`git remote add origin -- <url>`, `git remote set-url origin -- <url>`) to block `-`-prefixed URLs being read as flags. `ValidateRemoteURL` also enforces an allowlist of schemes (`https://`, `http://`, `ssh://`, `git://`, `file://`, `git@`).

### Source resolution & fetch

`SourceRef` has two orthogonal fields: `Type` (`ResourceTypeClaude` | `ResourceTypeSkill`) — _what_ the resource is, used by `cmd/` to pick the right flow — and `SourceKind` (`SourceKindGitHub` | `SourceKindGitLab` | `SourceKindLocal`) — _where_ it comes from, used internally by `core.FetchSource`. `cmd/` only ever switches on `Type`. **Don't reintroduce a `switch ref.SourceKind` in `cmd/`** — when adding a host or changing fetch strategy, extend `FetchSource` (and its helpers `fetchSingle` / `fetchBlobs` / `readLocalSkillTree`) instead.

`core.FetchSource(ref)` is the single fetch entry point. It returns `map[string][]byte` keyed by path relative to the resource root for every kind of source. For `Type == ResourceTypeClaude` it short-circuits to one HTTP/file read and returns `{"CLAUDE.md": content}`; for skills it dispatches on `SourceKind` to `*ListTree` + per-blob `Fetch` (or walks the local dir). `core.SaveSkill(name, files)` and `core.SaveClaude(files)` mirror this — both take the same `map[string][]byte` shape and return `(changed bool, err error)`.

`core.ResolveSource(input)` parses a user-supplied URL or local path into a `SourceRef`:

- `github.com/<o>/<r>/blob/<branch>/<path>` — last segment is stripped; the parent dir becomes the skill dir. Exception: if the last segment is `CLAUDE.md`, the ref is `Type: ResourceTypeClaude` (single-file).
- `github.com/<o>/<r>/tree/<branch>/<path>` — `<path>` becomes the skill dir.
- `gitlab.com/<o>/<r>/-/{blob,tree}/<branch>/<path>` — analogous.
- `raw.githubusercontent.com` and GitLab `/-/raw/` URLs are **rejected** with a clear error. Other hosts are also rejected.
- Local file: treated as a single-file skill (the file's content becomes `SKILL.md`). A local file named `CLAUDE.md` is `Type: ResourceTypeClaude`.
- Local directory: walked recursively by `ListLocalDir` (skipping dotfiles and `.git/`).

`core.ResolveSkillSource(skill, overrideURL)` is the update-path wrapper: it picks `overrideURL` when set, otherwise `skill.URL`, and feeds the result to `ResolveSource`. An empty stored URL with no override is an error (the skill was installed from a local file with no source recorded).

`core.GitHubListTree` / `core.GitLabListTree` call the trees REST API and return paths **relative to the skill dir**. They send `Authorization: Bearer $GITHUB_TOKEN` / `$GITLAB_TOKEN` when set. GitHub `truncated: true` responses are an error; tree-API and per-blob bodies are both bounded by `io.LimitReader` (`maxTreeAPIBytes = 10 MiB`, `maxSkillBytes = 5 MiB`). The API/raw bases default to `https://api.github.com` / `https://raw.githubusercontent.com` (and analogous for GitLab); each is overridable via `SKILL_CLI_GITHUB_API_BASE` / `SKILL_CLI_GITHUB_RAW_BASE` / `SKILL_CLI_GITLAB_API_BASE` / `SKILL_CLI_GITLAB_RAW_BASE` — used by the e2e fake server, not a documented user knob.

`core.GitHubRawURL(o, r, b, relPath)` / `core.GitLabRawURL(...)` build the per-blob URL passed to `Fetch`. Add new hosts by extending the switch in `ResolveSource` and adding matching `*ListTree` / `*RawURL` primitives, then wire them into `FetchSource`.

### Frontmatter parser

`core.ParseFrontmatter` is a tiny hand-rolled YAML reader for `name` and `description` only. It supports inline scalars and `|`/`>` block scalars but is **not** a full YAML parser — it'll mis-handle nested keys. If a SKILL.md needs more, prefer fixing the SKILL.md over expanding the parser.

## Conventions (enforced in past PRs)

- **Dispatch on a subcommand or argument value uses `switch/case/default`**, never `if/else if`. This applies to `cmd/` and `main.go` where the CLI branches on a string arg. Ordinary `if/else if` chains inside `core/` (e.g. branching on parsed state) are fine.
- **Don't call functions inside `if` conditions** — assign the return value to a variable first, then branch on it. Applies to `err`, `bool`, anything.
- HTTP body reads in `core/fetch.go` go through `io.LimitReader(resp.Body, maxSkillBytes+1)` (or `maxTreeAPIBytes+1` for trees); the package-level `httpClient` has a 30s timeout. Use them — don't reach for `http.Get` directly.

## Release

CI is `.github/workflows/release.yml` driven by goreleaser. Tag a commit to publish; goreleaser builds `linux/darwin/windows × amd64/arm64` (no windows-arm64) with `CGO_ENABLED=0` and `-s -w` ldflags, and creates a draft GitHub release under `MescoCzubinski/skill-cli`.
