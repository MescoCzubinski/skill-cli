package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/MescoCzubinski/skill-cli/core"
)

const claudeDescription = "global CLAUDE.md file"

func Add(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli add <url|path>")
		os.Exit(1)
	}

	hasRemote := core.HasRemote()
	if hasRemote {
		err := core.GitPull()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	ref, err := core.ResolveSource(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var name string
	switch ref.Type {
	case core.ResourceTypeClaude:
		addClaude(ref)
		name = "claude"
	default:
		name = addSkill(ref)
	}

	err = core.SyncSkillFiles()
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
		err = core.GitPush("add: " + name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fmt.Printf("added: %s\n", name)
}

func addClaude(ref core.SourceRef) {
	_, err := core.GetClaudeMeta()
	if err == nil {
		fmt.Fprintln(os.Stderr, "CLAUDE.md already installed")
		fmt.Fprintln(os.Stderr, "  run `skill-cli update claude` 		to refresh it")
		fmt.Fprintln(os.Stderr, "  run `skill-cli update claude <url>` 	to install from a different source")
		fmt.Fprintln(os.Stderr, "  run `skill-cli remove claude` 		to remove it")
		os.Exit(1)
	}
	if !errors.Is(err, core.ErrSkillNotFound) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	files, err := core.FetchSource(ref)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = core.SaveClaude(files)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.SaveSkillMeta("claude", claudeDescription, ref.Input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addSkill(ref core.SourceRef) string {
	files, err := core.FetchSource(ref)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	skillMD, ok := files["SKILL.md"]
	if !ok {
		fmt.Fprintln(os.Stderr, "no SKILL.md found in source")
		os.Exit(1)
	}

	name, description, err := core.ParseFrontmatter(string(skillMD))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = core.ValidateSkillName(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = core.GetSkillMeta(name)
	if err == nil {
		fmt.Fprintf(os.Stderr, "skill %q already installed\n", name)
		fmt.Fprintf(os.Stderr, "  run `skill-cli update %s` to refresh it\n", name)
		fmt.Fprintf(os.Stderr, "  run `skill-cli remove %s` first to reinstall from a different source\n", name)
		os.Exit(1)
	}
	if !errors.Is(err, core.ErrSkillNotFound) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = core.SaveSkill(name, files)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.SaveSkillMeta(name, description, ref.Input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return name
}
