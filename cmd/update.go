package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MescoCzubinski/skill-cli/core"
)

func Update(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli update <name>|--all [<url>]")
		os.Exit(1)
	}

	overrideURL := ""
	if len(args) > 1 {
		if args[0] == "--all" {
			fmt.Fprintln(os.Stderr, "cannot pass <url> with --all")
			os.Exit(1)
		}
		overrideURL = args[1]
	}

	hasRemote := core.HasRemote()
	if hasRemote {
		err := core.GitPull()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	skills, claude := loadSkills(args[0])
	today := time.Now().Format("2006-01-02")

	changed := []string{}
	hadError := false

	if claude != nil {
		didChange, err := updateClaude(claude, today, overrideURL)
		switch {
		case err != nil:
			fmt.Fprintf(os.Stderr, "update %s: %v\n", claude.Name, err)
			hadError = true
		case didChange:
			fmt.Printf("updated: %s\n", claude.Name)
			changed = append(changed, claude.Name)
		default:
			fmt.Printf("unchanged: %s\n", claude.Name)
		}
	}

	for i := range skills {
		skill := &skills[i]
		didChange, err := updateSkill(skill, today, overrideURL)
		switch {
		case err != nil:
			fmt.Fprintf(os.Stderr, "update %s: %v\n", skill.Name, err)
			hadError = true
		case didChange:
			fmt.Printf("updated: %s\n", skill.Name)
			changed = append(changed, skill.Name)
		default:
			fmt.Printf("unchanged: %s\n", skill.Name)
		}
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

	if hasRemote && len(changed) > 0 {
		msg := "update: " + strings.Join(changed, ", ")
		err = core.GitPush(msg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if hadError {
		os.Exit(1)
	}
}

func loadSkills(arg string) ([]core.Skill, *core.Skill) {
	switch arg {
	case "--all":
		skills, err := core.GetSkillsMeta()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		claude, err := core.GetClaudeMeta()
		if err != nil && !errors.Is(err, core.ErrSkillNotFound) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return skills, claude
	case "claude":
		claude, err := core.GetClaudeMeta()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return nil, claude
	default:
		skill, err := core.GetSkillMeta(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		return []core.Skill{*skill}, nil
	}
}

func updateClaude(skill *core.Skill, today, overrideURL string) (bool, error) {
	ref, err := core.ResolveSource(overrideURL, skill.URL)
	if err != nil {
		return false, err
	}
	if ref.Type != core.ResourceTypeClaude {
		return false, fmt.Errorf("source for claude must point at a CLAUDE.md file")
	}

	files, err := core.FetchSource(ref)
	if err != nil {
		return false, err
	}

	changed, err := core.SaveClaude(files)
	if err != nil {
		return false, err
	}

	err = core.UpdateSkillMeta(skill.Name, claudeDescription, today, overrideURL)
	if err != nil {
		return false, err
	}

	return changed, nil
}

func updateSkill(skill *core.Skill, today, overrideURL string) (bool, error) {
	ref, err := core.ResolveSource(overrideURL, skill.URL)
	if err != nil {
		return false, err
	}
	if ref.Type == core.ResourceTypeClaude {
		return false, fmt.Errorf("source for skill %q points at a CLAUDE.md file", skill.Name)
	}

	files, err := core.FetchSource(ref)
	if err != nil {
		return false, err
	}

	skillMD, ok := files["SKILL.md"]
	if !ok {
		return false, fmt.Errorf("no SKILL.md found in source")
	}

	parsedName, parsedDesc, err := core.ParseFrontmatter(string(skillMD))
	if err != nil {
		return false, err
	}
	if overrideURL != "" && parsedName != skill.Name {
		return false, fmt.Errorf("fetched skill name %q does not match %q", parsedName, skill.Name)
	}

	changed, err := core.SaveSkill(skill.Name, files)
	if err != nil {
		return false, err
	}

	err = core.UpdateSkillMeta(skill.Name, parsedDesc, today, overrideURL)
	if err != nil {
		return false, err
	}

	return changed, nil
}
