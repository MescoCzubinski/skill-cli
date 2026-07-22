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
	case "list", "--list", "-l":
		cmd.List(os.Args[2:])
	case "remove":
		cmd.Remove(os.Args[2:])
	case "update":
		cmd.Update(os.Args[2:])
	case "remote":
		cmd.Remote(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("usage: skill-cli <command> [args] [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add     <url|path>            Install a skill or CLAUDE.md from a URL or local file")
	fmt.Println("  list                          List installed skills and CLAUDE.md")
	fmt.Println("  remove  <name>|claude         Remove a skill by name, or the global CLAUDE.md")
	fmt.Println("  update  <name>|claude|--all   Re-fetch and update skill(s) or CLAUDE.md")
	fmt.Println("  remote  <url>                 Attach a git remote for sync")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --list, -l                    Shortcut for `list`")
	fmt.Println("  --check                       (update) Report available updates without writing")
	fmt.Println("  --no-update                   (add/remove/update) Skip the pre-op `git pull`")
	fmt.Println("  --no-commit                   (add/remove/update) Skip the post-op commit + push")
	fmt.Println("  --help, -h                    Show this help")
}
