# skill-cli

A minimal CLI for managing AI skills.

Install skills from GitHub/GitLab URLs, add, update and remove them, and sync across devices — no UI, no external registries (sync via your own GitHub/GitLab repository).

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
  add     <url>           Install a skill from a URL
  list                    List installed skills
  remove  <name>          Remove a skill by name
  update  <name>|--all    Re-fetch and update skill(s)
  remote  <url>           Attach a git remote for sync
  sync                    Pull latest skills from remote
```

### Add a skill

```bash
skill-cli add https://github.com/JuliusBrussee/caveman/blob/main/skills/caveman/SKILL.md
```

### List installed skills

```bash
skill-cli list
```

```
NAME         DESCRIPTION                                UPDATED
caveman      Ultra-compressed communication mode...     2026-05-01
grill-me     Interview the user to surface intent...    2026-04-23
```

### Update skills

```bash
skill-cli update caveman
skill-cli update --all
```

### Remove a skill

```bash
skill-cli remove caveman
```

## Git

`skill-cli` works without a git remote. Attaching one enables syncing your skills across multiple devices.

Git must be installed and authentication (SSH or HTTPS) must be configured. `skill-cli` calls `git` directly and inherits your existing credentials.

### Attach a remote

Create an empty repository on GitHub or GitLab, then run:

**HTTPS:**

```bash
skill-cli remote https://github.com/{yourname}/{repository}
```

```bash
skill-cli remote https://gitlab.com/{yourname}/{repository}
```

**SSH:**

```bash
skill-cli remote git@github.com:{yourname}/{repository}.git
```

```bash
skill-cli remote git@gitlab.com:{yourname}/{repository}.git
```

### Sync across devices

```bash
skill-cli sync
```

On a new device, attach the same remote and run `sync` to pull your existing skills.

## Storage

Skills are stored locally in:

```
~/.config/skill-cli/
├── skills.json          # metadata: name, description, raw_url, installed_at, updated_at
└── skills/
    ├── grill-me.md
    └── caveman.md
```
