package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/MescoCzubinski/skill-cli/core"
)

func Remove(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli remove <name>")
		os.Exit(1)
	}

	name := args[0]

	hasRemote := core.HasRemote()
	if hasRemote {
		err := core.GitPull()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	switch name {
	case "claude":
		removeClaude()
	default:
		removeSkill(name)
	}

	err := core.SyncSkillFiles()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.SyncClaude()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if hasRemote {
		err = core.GitPush("remove: " + name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fmt.Printf("removed: %s\n", name)
}

func removeClaude() {
	_, err := core.GetClaudeMeta()
	if errors.Is(err, core.ErrSkillNotFound) {
		fmt.Fprintln(os.Stderr, "CLAUDE.md is not installed")
		fmt.Fprintln(os.Stderr, "  run `skill-cli list` to see installed entries")
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.RemoveClaudeFile()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.RemoveSkillMeta("claude")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func removeSkill(name string) {
	_, err := core.GetSkillMeta(name)
	if errors.Is(err, core.ErrSkillNotFound) {
		fmt.Fprintf(os.Stderr, "skill %q is not installed\n", name)
		fmt.Fprintln(os.Stderr, "  run `skill-cli list` to see installed skills")
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.RemoveSkillFile(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.RemoveSkillMeta(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
