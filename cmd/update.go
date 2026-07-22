package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MescoCzubinski/skill-cli/core"
)

const updateUsage = "usage: skill-cli update <name>|claude|--all [<url>] [--check] [--no-update] [--no-commit]"

func Update(args []string) {
	positional, flags, err := core.ParseFlags(args, map[string]bool{
		"--all":       true,
		"--check":     true,
		"--no-update": true,
		"--no-commit": true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, updateUsage)
		os.Exit(1)
	}

	all := flags["--all"]
	check := flags["--check"]
	noUpdate := flags["--no-update"]
	noCommit := flags["--no-commit"]

	target, overrideURL, err := updateTarget(positional, all)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, updateUsage)
		os.Exit(1)
	}

	if check {
		skills := loadSkillsToChange(target)
		checkSkills(skills, overrideURL)
		return
	}

	hasRemote := core.HasRemote()
	if hasRemote && !noUpdate {
		err = core.GitPull()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	skills := loadSkillsToChange(target)
	today := time.Now().Format("2006-01-02")

	changed := []string{}
	hadError := false
	for i := range skills {
		didChange, err := updateSkill(&skills[i], today, overrideURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "update %s: %v\n", skills[i].Name, err)
			hadError = true
			continue
		}
		if didChange {
			changed = append(changed, skills[i].Name)
		}
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

	if hasRemote && !noCommit && len(changed) > 0 {
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

func updateTarget(positional []string, all bool) (string, string, error) {
	switch {
	case all:
		if len(positional) > 0 {
			return "", "", errors.New("cannot pass a name or <url> with --all")
		}
		return "--all", "", nil
	case len(positional) == 0:
		return "", "", errors.New("a skill name, claude, or --all is required")
	case len(positional) == 1:
		return positional[0], "", nil
	case len(positional) == 2:
		return positional[0], positional[1], nil
	default:
		return "", "", errors.New("too many arguments")
	}
}

func loadSkillsToChange(arg string) []core.Skill {
	switch arg {
	case "--all":
		skills, err := core.GetSkillsMeta()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		claude, err := core.GetClaudeMeta()
		if err == nil {
			skills = append(skills, *claude)
		} else if !errors.Is(err, core.ErrSkillNotFound) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return skills
	case "claude":
		claude, err := core.GetClaudeMeta()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return []core.Skill{*claude}
	default:
		skill, err := core.GetSkillMeta(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return []core.Skill{*skill}
	}
}

func updateSkill(skill *core.Skill, today, overrideURL string) (bool, error) {
	fetchURL := skill.RawURL
	if overrideURL != "" {
		fetchURL = overrideURL
	}
	if fetchURL == "" {
		return false, fmt.Errorf("installed from a local file (cannot update)")
	}

	content, err := core.Fetch(fetchURL)
	if err != nil {
		return false, err
	}

	var description string
	var changed bool
	if skill.Name == "claude" {
		description = claudeDescription
		changed, err = core.SaveClaudeFile(content)
		if err != nil {
			return false, err
		}
	} else {
		parsedName, parsedDesc, parseErr := core.ParseFrontmatter(content)
		if parseErr != nil {
			return false, parseErr
		}
		if overrideURL != "" && parsedName != skill.Name {
			return false, fmt.Errorf("fetched skill name %q does not match %q", parsedName, skill.Name)
		}
		description = parsedDesc
		changed, err = core.SaveSkillFile(skill.Name, content)
		if err != nil {
			return false, err
		}
	}

	err = core.UpdateSkillMeta(skill.Name, description, today, overrideURL)
	if err != nil {
		return false, err
	}

	if changed {
		fmt.Printf("updated: %s\n", skill.Name)
	} else {
		fmt.Printf("unchanged: %s\n", skill.Name)
	}

	return changed, nil
}

func checkSkills(skills []core.Skill, overrideURL string) {
	hadError := false
	for i := range skills {
		err := checkSkill(&skills[i], overrideURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "check %s: %v\n", skills[i].Name, err)
			hadError = true
		}
	}
	if hadError {
		os.Exit(1)
	}
}

func checkSkill(skill *core.Skill, overrideURL string) error {
	fetchURL := skill.RawURL
	if overrideURL != "" {
		fetchURL = overrideURL
	}
	if fetchURL == "" {
		return fmt.Errorf("installed from a local file (cannot check)")
	}

	content, err := core.Fetch(fetchURL)
	if err != nil {
		return err
	}

	var changed bool
	if skill.Name == "claude" {
		changed, err = core.ClaudeFileChanged(content)
		if err != nil {
			return err
		}
	} else {
		parsedName, _, parseErr := core.ParseFrontmatter(content)
		if parseErr != nil {
			return parseErr
		}
		if overrideURL != "" && parsedName != skill.Name {
			return fmt.Errorf("fetched skill name %q does not match %q", parsedName, skill.Name)
		}
		changed, err = core.SkillFileChanged(skill.Name, content)
		if err != nil {
			return err
		}
	}

	if changed {
		fmt.Printf("update available: %s\n", skill.Name)
	} else {
		fmt.Printf("up to date: %s\n", skill.Name)
	}

	return nil
}
