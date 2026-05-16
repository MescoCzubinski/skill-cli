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

	skills := loadSkillsToChange(args[0])
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
		parsedName, parsedDesc, err := core.ParseFrontmatter(content)
		if err != nil {
			return false, err
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
