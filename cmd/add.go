package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mieszko/skill-cli/skill"
)

func Add(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli add <url>")
		os.Exit(1)
	}

	rawURL, err := skill.ToRawURL(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	name, description, err := skill.FetchMeta(rawURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	skills, err := skill.LoadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if i, _ := skill.FindByName(skills, name); i != -1 {
		fmt.Fprintf(os.Stderr, "skill %q already installed\n", name)
		os.Exit(1)
	}

	today := time.Now().Format("2006-01-02")
	skills = append(skills, skill.Skill{
		Name:        name,
		Description: description,
		RawURL:      rawURL,
		InstalledAt: today,
		UpdatedAt:   today,
	})

	if err := skill.SaveAll(skills); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("added: %s\n", name)
}
