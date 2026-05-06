package cmd

import (
	"fmt"
	"os"

	"github.com/mieszko/skill-cli/core"
)

func Remote(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli remote <add|clear> [args]")
		os.Exit(1)
	}
	switch args[0] {
	case "add":
		remoteAdd(args[1:])
	case "clear":
		remoteClear(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown remote command: %s\n", args[0])
		os.Exit(1)
	}
}

func remoteAdd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli remote add <url>")
		os.Exit(1)
	}
	remoteURL := args[0]

	err := core.ValidateRemoteURL(remoteURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.SetRemote(remoteURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("added remote: %s\n", remoteURL)
}

func remoteClear(args []string) {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli remote clear")
		os.Exit(1)
	}
	err := core.ClearRemote()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("remote cleared")
}
