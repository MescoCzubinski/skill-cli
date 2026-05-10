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

	_, err = core.GetSkillMeta(name)
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

	err = core.SyncSkillFiles()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if hasRemote {
		err = core.GitPush("remove: " + name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fmt.Printf("removed: %s\n", name)
}
