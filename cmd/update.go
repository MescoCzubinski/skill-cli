package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mieszko/skill-cli/skill"
)

func Update(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli update <name>|--all")
		os.Exit(1)
	}

	skills, err := skill.LoadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	today := time.Now().Format("2006-01-02")

	if args[0] == "--all" {
		for i := range skills {
			if err := skills[i].FetchSkill(today); err != nil {
				fmt.Fprintf(os.Stderr, "update %s: %v\n", skills[i].Name, err)
				os.Exit(1)
			}
			fmt.Printf("updated: %s\n", skills[i].Name)
		}
	} else {
		name := args[0]
		i, _ := skill.FindByName(skills, name)
		if i == -1 {
			fmt.Fprintf(os.Stderr, "skill %q not found\n", name)
			os.Exit(1)
		}
		if err := skills[i].FetchSkill(today); err != nil {
			fmt.Fprintf(os.Stderr, "update %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("updated: %s\n", name)
	}

	if err := skill.SaveAll(skills); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
