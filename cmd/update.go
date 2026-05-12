package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MescoCzubinski/skill-cli/core"
)

func Update(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli update <name>|--all")
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

	err := core.SyncSkillFiles()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	skills := loadSkillsToChange(args[0])
	today := time.Now().Format("2006-01-02")

	changed := []string{}
	for i := range skills {
		didChange := updateSkill(&skills[i], today)
		if didChange {
			changed = append(changed, skills[i].Name)
		}
	}

	err = core.SyncSkillFiles()
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
}

func loadSkillsToChange(arg string) []core.Skill {
	if arg == "--all" {
		skills, err := core.GetSkillsMeta()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return skills
	}

	skill, err := core.GetSkillMeta(arg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return []core.Skill{*skill}
}

func updateSkill(skill *core.Skill, today string) bool {
	if skill.RawURL == "" {
		fmt.Fprintf(os.Stderr, "update %s: installed from a local file (cannot update)\n", skill.Name)
		os.Exit(1)
	}

	_, description, content, err := core.FetchSkill(skill.RawURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update %s: %v\n", skill.Name, err)
		os.Exit(1)
	}

	changed, err := core.SaveSkillFile(skill.Name, content)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.UpdateSkillMeta(skill.Name, description, today)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if changed {
		fmt.Printf("updated: %s\n", skill.Name)
	} else {
		fmt.Printf("unchanged: %s\n", skill.Name)
	}

	return changed
}
