package cmd

import (
	"fmt"
	"os"

	"github.com/mieszko/skill-cli/skill"
)

func Remove(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli remove <name>")
		os.Exit(1)
	}

	name := args[0]
	skills, err := skill.LoadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	i, _ := skill.FindByName(skills, name)
	if i == -1 {
		fmt.Fprintf(os.Stderr, "skill %q not found\n", name)
		os.Exit(1)
	}

	skills = append(skills[:i], skills[i+1:]...)
	if err := skill.SaveAll(skills); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("removed: %s\n", name)
}
