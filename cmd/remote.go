package cmd

import (
	"fmt"
	"os"

	"github.com/MescoCzubinski/skill-cli/core"
)

func Remote(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli remote <url>")
		os.Exit(1)
	}
	remoteURL := args[0]

	var err error
	if !core.IsGitRepo() {
		err = core.GitInit(remoteURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		err = core.GitSetOrigin(remoteURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	err = core.GitPull()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.SyncSkillFiles()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("set remote: %s\n", remoteURL)
}
