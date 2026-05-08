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
  add <url>                Install a skill from a URL
  list                     List installed skills
  remove <name>            Remove a skill by name
  update <name>|--all      Re-fetch and update skill(s)
  remote <url>             Attach a git remote for sync
  sync push|pull           Push or pull latest skills from remote
```

### Add a skill

```bash
skill-cli add https://github.com/{username}/{repository_name}/{skill_path}/{skill_name}/SKILL.md
```

### List installed skills

```bash
skill-cli list
```

```
NAME           DESCRIPTION                                UPDATED
{skill_name}   Ultra-compressed communication mode...     2026-05-01
grill-me       Interview the user to surface intent...    2026-04-23
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

`skill-cli` works without a git remote. Attaching one enables syncing your skills across multiple devices.

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

### Sync across devices

Push local skills to the remote:

```bash
skill-cli sync push
```

Pull skills from the remote:

```bash
skill-cli sync pull
```

On a new device, attach the same remote and run `skill-cli sync pull` to fetch your existing skills.

#### Git commands used

Under the hood, `skill-cli` runs plain `git` commands inside `~/.config/skill-cli/`

`sync push` (when there are local changes):

```
git add .
git commit -m "sync"
git push
```

`sync pull`:

```
git pull
```

First `sync push|pull` (when the config directory is not yet a repo):

```
git init
git remote add origin <your-remote-url>
git add .
git commit -m "init"
git push -u origin main
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
