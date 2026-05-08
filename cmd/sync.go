package cmd

import (
	"fmt"
	"os"

	"github.com/MescoCzubinski/skill-cli/core"
)

func Sync(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli sync <push|pull>")
		os.Exit(1)
	}

	_, err := core.GetRemote()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch args[0] {
	case "push":
		err = syncPush()
	case "pull":
		err = syncPull()
	default:
		fmt.Fprintf(os.Stderr, "unknown sync command: %s\n", args[0])
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("synced (%s)\n", args[0])
}

func syncPush() error {
	hasOrigin, err := core.HasOrigin()
	if err != nil {
		return err
	}

	if !hasOrigin {
		err := core.GitInit()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return nil
	}

	err = core.GitPush()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return nil
}

func syncPull() error {
	hasOrigin, err := core.HasOrigin()
	if err != nil {
		return err
	}

	if !hasOrigin {
		err := core.GitInit()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		err := core.GitPull()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	err = core.SyncSkillFiles()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return nil
}
