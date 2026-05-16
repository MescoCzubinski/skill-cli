package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MescoCzubinski/skill-cli/core"
)

const claudeDescription = "global CLAUDE.md file"

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

	basename, content, rawURL := loadSource(args[0])

	var name string
	switch basename {
	case "CLAUDE.md":
		addClaude(content, rawURL)
		name = "claude"
	default:
		name = addSkill(content, rawURL)
	}

	err := core.SyncSkillFiles()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.SyncClaude()
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

func loadSource(input string) (string, string, string) {
	if strings.HasPrefix(input, "http") {
		rawURL, err := core.GetRawURL(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		u, err := url.Parse(rawURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		content, err := core.Fetch(rawURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return path.Base(u.Path), content, rawURL
	}

	resolved, err := core.ResolveLocalPath(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	content, err := core.GetLocal(resolved)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return filepath.Base(resolved), content, ""
}

func addClaude(content, rawURL string) {
	_, err := core.GetClaudeMeta()
	if err == nil {
		fmt.Fprintln(os.Stderr, "CLAUDE.md already installed")
		fmt.Fprintln(os.Stderr, "  run `skill-cli update claude` to refresh it")
		fmt.Fprintln(os.Stderr, "  run `skill-cli remove claude` first to reinstall from a different source")
		os.Exit(1)
	}
	if !errors.Is(err, core.ErrSkillNotFound) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, err = core.SaveClaudeFile(content)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = core.SaveSkillMeta("claude", claudeDescription, rawURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addSkill(content, rawURL string) string {
	name, description, err := core.ParseFrontmatter(content)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = core.ValidateSkillName(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

	return name
}
