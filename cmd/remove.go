package cmd

import (
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
	_, err := core.FindSkillMeta(name)
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

	fmt.Printf("removed: %s\n", name)
}
