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

	if args[0] == "--all" {
		updateAll()
	} else {
		updateSingle(args[0])
	}
}

func updateAll() {
	today := time.Now().Format("2006-01-02")

	skills, err := core.GetSkillsMeta()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for i := range skills {
		updateSkill(&skills[i], today)
	}
}

func updateSingle(name string) {
	today := time.Now().Format("2006-01-02")

	skill, err := core.FindSkillMeta(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	updateSkill(skill, today)
}

func updateSkill(skill *core.Skill, today string) {
	_, description, content, err := core.FetchSkill(skill.RawURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update %s: %v\n", skill.Name, err)
		os.Exit(1)
	}

	err = core.SaveSkillFile(skill.Name, content)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.UpdateSkillMeta(skill.Name, description, today)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("updated: %s\n", skill.Name)
}
