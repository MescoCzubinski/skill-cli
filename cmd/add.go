package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MescoCzubinski/skill-cli/core"
)

func Add(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: skill-cli add <url|path>")
		os.Exit(1)
	}

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

	var name, description, content, rawURL string
	if strings.HasPrefix(args[0], "http") {
		rawURL, err = core.GetRawURL(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		name, description, content, err = core.FetchSkill(rawURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		name, description, content, err = core.GetLocalSkill(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	_, err = core.GetSkillMeta(name)
	if err == nil {
		fmt.Fprintf(os.Stderr, "skill %q already installed\n", name)
		fmt.Fprintf(os.Stderr, "  run `skill-cli update %s` to refresh it\n", name)
		fmt.Fprintf(os.Stderr, "  run `skill-cli remove %s` first to reinstall from a different source\n", name)
		os.Exit(1)
	}
	if !errors.Is(err, core.ErrSkillNotFound) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = core.SaveSkillFile(name, content)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.SaveSkillMeta(name, description, rawURL)
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
		err = core.GitPush("add: " + name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fmt.Printf("added: %s\n", name)
}
