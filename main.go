package main

import (
	"fmt"
	"os"

	"github.com/MescoCzubinski/skill-cli/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		cmd.Add(os.Args[2:])
	case "list":
		cmd.List(os.Args[2:])
	case "remove":
		cmd.Remove(os.Args[2:])
	case "update":
		cmd.Update(os.Args[2:])
	case "remote":
		cmd.Remote(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("usage: skill-cli <command> [args]")
	fmt.Println("commands: add <url|path>, list, remove <name>, update <name>|--all, update <name> <url>, remote <url>")
}
