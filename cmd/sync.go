package cmd

import (
	"fmt"
	"os"

	"github.com/mieszko/skill-cli/core"
)

func Sync(_ []string) {
	_, err := core.GetRemote()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.Sync()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.SyncSkillFiles()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("synced")
}
