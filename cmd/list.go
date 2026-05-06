package cmd

import (
	"fmt"
	"os"

	"github.com/mieszko/skill-cli/core"
)

func List(_ []string) {
	skills, err := core.GetSkillsMeta()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(skills) == 0 {
		fmt.Println("no skills installed")
		return
	}

	fmt.Printf("%-16s %-45s %s\n", "NAME", "DESCRIPTION", "UPDATED")
	for _, s := range skills {
		desc := s.Description
		if len(desc) > 42 {
			desc = desc[:42] + "..."
		}
		fmt.Printf("%-16s %-45s %s\n", s.Name, desc, s.UpdatedAt)
	}
}
