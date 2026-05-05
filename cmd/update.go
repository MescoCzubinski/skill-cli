package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mieszko/skill-cli/core"
)

func Update(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli update <name>|--all")
		os.Exit(1)
	}

	today := time.Now().Format("2006-01-02")
	if args[0] == "--all" {
		updateAll(today)
	} else {
		updateSingle(args[0], today)
	}
}

func updateAll(today string) {
	skills, err := core.GetSkillsMeta()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for i := range skills {
		description, err := updateSkill(&skills[i], today)
		if err != nil {
			fmt.Fprintf(os.Stderr, "update %s: %v\n", skills[i].Name, err)
			os.Exit(1)
		}
		skills[i].Description = description
		skills[i].UpdatedAt = today
		fmt.Printf("updated: %s\n", skills[i].Name)
	}

	err = core.SaveSkillsMeta(skills)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func updateSingle(name, today string) {
	skill, err := core.FindSkillMeta(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if skill == nil {
		fmt.Fprintf(os.Stderr, "skill %q not found\n", name)
		os.Exit(1)
	}

	description, err := updateSkill(skill, today)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update %s: %v\n", name, err)
		os.Exit(1)
	}

	err = core.UpdateSkillMeta(name, description, today)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("updated: %s\n", name)
}

func updateSkill(s *core.Skill, today string) (description string, err error) {
	_, description, content, err := core.FetchSkill(s.RawURL)
	if err != nil {
		return
	}
	err = core.SaveSkillFile(s.Name, content)
	return
}
