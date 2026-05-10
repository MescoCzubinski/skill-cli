# skill-cli

A minimal CLI for managing AI skills.

Install skills from URLs, update from source, sync across devices via your own git repository - works with Claude Code, Codex CLI, Cursor, Windsurf, Gemini CLI.

## Install

### Download binary

Go to [Releases](https://github.com/MescoCzubinski/skill-cli/releases) and download the binary for your platform:

| OS      | File                            |
| ------- | ------------------------------- |
| Linux   | `skill-cli_linux_amd64.tar.gz`  |
| macOS   | `skill-cli_darwin_arm64.tar.gz` |
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
  add     <url>             Install a skill from a URL
  list                      List installed skills
  remove  <name>            Remove a skill by name
  update  <name>|--all      Re-fetch and update skill(s)
  remote  <url>             Attach a git remote for sync
```

### Add a skill

```bash
skill-cli add {url}
```

### List installed skills

```bash
skill-cli list
```

```
NAME           DESCRIPTION             UPDATED
{skill_name}   {skill_description}     {skill_updated_at}
{skill_name}   {skill_description}     {skill_updated_at}
```

### Update skills

```bash
skill-cli update {skill_name}
skill-cli update --all
```

### Remove a skill

```bash
skill-cli remove {skill_name}
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
├── skills.json          # metadata: name, description, raw_url, installed_at, updated_at
├── remote               # configured git remote URL (created by `skill-cli remote <url>`)
└── skills/
    ├── {skill_name}/SKILL.md
    ├── {skill_name}/SKILL.md
    └── {skill_name}/SKILL.md
```
