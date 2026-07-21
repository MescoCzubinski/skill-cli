package cmd

import (
	"fmt"
	"strings"
)

const (
	addUsage    = "usage: skill-cli add <url|path> [--no-update] [--no-commit]"
	removeUsage = "usage: skill-cli remove <name>|claude [--no-update] [--no-commit]"
	updateUsage = "usage: skill-cli update <name>|claude|--all [<url>] [--check] [--no-update] [--no-commit]"
)

// flags holds the optional toggles a subcommand recognizes. Each subcommand
// reads only the fields relevant to it; the rest stay at their zero value.
type flags struct {
	all      bool // update: operate on every installed skill and CLAUDE.md
	check    bool // update: dry-run - report what would change, write nothing
	noUpdate bool // add/remove/update: skip the pre-op `git pull`
	noCommit bool // add/remove/update: skip the post-op `git commit` + push
}

// parseFlags splits args into positional arguments and flags. Every token that
// starts with "-" must appear in allowed, otherwise it is reported as unknown.
// The returned flags struct has the recognized toggles set.
func parseFlags(args []string, allowed map[string]bool) ([]string, flags, error) {
	positional := []string{}
	var f flags
	for _, arg := range args {
		isFlag := strings.HasPrefix(arg, "-")
		if !isFlag {
			positional = append(positional, arg)
			continue
		}
		ok := allowed[arg]
		if !ok {
			return nil, f, fmt.Errorf("unknown flag: %s", arg)
		}
		switch arg {
		case "--all":
			f.all = true
		case "--check":
			f.check = true
		case "--no-update":
			f.noUpdate = true
		case "--no-commit":
			f.noCommit = true
		}
	}
	return positional, f, nil
}
