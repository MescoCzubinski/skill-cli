# skill-cli

A minimal CLI for managing [AI skills](https://addyosmani.com/blog/agent-skills/) and a global `CLAUDE.md`.

Install skills and `CLAUDE.md` from URLs, update from source, sync across devices via your own git repository - works with Claude Code, Codex CLI, Cursor, Windsurf, Gemini CLI.

## Install

### Download binary

Go to [Releases](https://github.com/MescoCzubinski/skill-cli/releases) and download the binary for your platform:

| OS      | File                            |
| ------- | ------------------------------- |
| macOS   | `skill-cli_darwin_amd64.tar.gz` |
| macOS   | `skill-cli_darwin_arm64.tar.gz` |
| Linux   | `skill-cli_linux_amd64.tar.gz`  |
| Linux   | `skill-cli_linux_arm64.tar.gz`  |
| Windows | `skill-cli_windows_amd64.zip`   |

**Linux & macOS:**

```bash
tar -xzf skill-cli_*.tar.gz
```

```bash
sudo mv skill-cli /usr/local/bin/
```

**Windows:** Extract the `.zip`, then move `skill-cli.exe` to a folder in your `%PATH%` (e.g. `C:\Windows\System32`).

### With Go installed

```bash
go install github.com/MescoCzubinski/skill-cli@latest
```

Or:

```bash
git clone https://github.com/MescoCzubinski/skill-cli
cd skill-cli
go build -o skill-cli .
```

## Usage

```
skill-cli <command> [args]

Commands:
  add     <url|path>                     Install a skill or CLAUDE.md from a URL or local file
  list                                   List installed skills and CLAUDE.md
  remove  <name>|claude                  Remove a skill by name, or the global CLAUDE.md
  update  <name>|claude|--all            Re-fetch and update skill(s) or CLAUDE.md
  update  <name> [<url>]|claude [<url>]  Switch the source URL and re-fetch
  remote  <url>                          Attach a git remote for sync
```

### Add a skill

From a URL:

```bash
skill-cli add {url}
```

From a local file:

```bash
skill-cli add ./path/to/SKILL.md
```

Skills installed from a local file cannot be updated with `skill-cli update` (there is no remote source to fetch from). To replace one, remove it first and re-add it.

#### Multi-file skills

A skill can be more than a single `SKILL.md`. Point `add` at a **directory** — a local folder or a repository `tree` URL — and every file in it is installed, including subdirectories and non-markdown files such as Python scripts, JSON, or data assets (dotfiles and `.git/` are skipped). This is useful for skills that ship helper scripts alongside their instructions.

```bash
skill-cli add https://github.com/{user}/{repo}/tree/main/skills/{skill_name}
skill-cli add ./path/to/skill-dir
```

The directory must contain a `SKILL.md` at its root — its frontmatter `name` becomes the skill name.

### Add a global CLAUDE.md

Same `add` command — type is detected by basename. The path/URL must end in `/CLAUDE.md`:

```bash
skill-cli add https://raw.githubusercontent.com/{user}/{repo}/main/CLAUDE.md
skill-cli add ./path/to/CLAUDE.md
```

Only one CLAUDE.md may be installed at a time, under the reserved name `claude`. It is mirrored to `~/.claude/CLAUDE.md` (existing file there is overwritten). To replace the source, run `skill-cli update claude <new-url>` or `skill-cli remove claude` first.

### List installed skills

```bash
skill-cli list
```

```
Skills:
NAME           DESCRIPTION             UPDATED
{skill_name}   {skill_description}     {skill_updated_at}
{skill_name}   {skill_description}     {skill_updated_at}

CLAUDE.md:
NAME           DESCRIPTION             UPDATED
claude         global CLAUDE.md file   {claude_updated_at}
```

### Update skills

```bash
skill-cli update {skill_name}
skill-cli update claude
skill-cli update --all
```

Pass a URL to switch the source for a skill or CLAUDE.md and re-fetch from it:

```bash
skill-cli update {skill_name} {new_url}
skill-cli update claude {new_url}
```

For skills, the fetched frontmatter `name` must match `{skill_name}` (prevents accidental clobber).

### Remove a skill or CLAUDE.md

```bash
skill-cli remove {skill_name}
skill-cli remove claude
```

## Git

`skill-cli` works without a git remote. Attaching one enables syncing your skills across multiple devices automatically - each `add`, `remove`, and `update` pulls from and pushes to the remote.

Git must be installed and authentication (SSH or HTTPS) must be configured. `skill-cli` calls `git` directly and inherits your existing credentials.

### Attach a remote

Create an empty repository on GitHub or GitLab, then run:

**HTTPS:**

```bash
skill-cli remote https://github.com/{yourname}/{repository_name}
```

```bash
skill-cli remote https://gitlab.com/{yourname}/{repository_name}
```

**SSH:**

```bash
skill-cli remote git@github.com:{yourname}/{repository_name}.git
```

```bash
skill-cli remote git@gitlab.com:{yourname}/{repository_name}.git
```

On a new device, run the same `skill-cli remote <url>` to fetch your existing skills.

#### Git commands used

Under the hood, `skill-cli` runs plain `git` commands inside `~/.config/skill-cli/`.

`remote <url>` when the config directory is not yet a repo:

```
git init
git remote add origin <your-remote-url>
git add .
git commit -m "init"
git push -u origin main
```

`remote <url>` when the repo already exists:

```
git remote set-url origin <your-remote-url>
```

Before each `add`, `remove`, or `update` (when a remote is configured):

```
git pull --rebase
```

After each `add`, `remove`, or `update` (when a remote is configured):

```
git add .
git commit -m "<add|remove|update>: <name>"
git push -u origin HEAD
```

## Storage

Skills are stored locally in:

```
~/.config/skill-cli/
└── meta/               # metadata: name, description, raw_url, installed_at, updated_at
    ├── {skill_name}.json
    ├── {skill_name}.json
    ├── {skill_name}.json
    └── claude.json     # present only if a CLAUDE.md is installed
└── skills/
    ├── {skill_name}/SKILL.md         # single-file skill
    ├── {skill_name}                  # multi-file skill
    │   ├── SKILL.md
    │   └── {other_file}
    └── {skill_name}/SKILL.md
└── claude/             # present only if a CLAUDE.md is installed
    └── CLAUDE.md
```
