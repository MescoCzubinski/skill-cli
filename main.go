package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: skill-cli <command> [args]")
		fmt.Println("commands: add, list, remove, update")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		fmt.Println("add: not implemented")
	case "list":
		fmt.Println("list: not implemented")
	case "remove":
		fmt.Println("remove: not implemented")
	case "update":
		fmt.Println("update: not implemented")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
