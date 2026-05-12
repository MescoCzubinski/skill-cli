package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func GetRawURL(input string) (string, error) {
	u, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	switch u.Host {
	case "raw.githubusercontent.com":
		return input, nil
	case "github.com":
		return getRawURLGitHub(input, parts)
	case "gitlab.com":
		return getRawURLGitLab(input, parts)
	default:
		return input, nil
	}
}

func getRawURLGitHub(input string, parts []string) (string, error) {
	if len(parts) < 4 {
		return "", fmt.Errorf("unrecognized GitHub URL: %s", input)
	}
	user, repo, kind, branch := parts[0], parts[1], parts[2], parts[3]
	rest := strings.Join(parts[4:], "/")
	switch kind {
	case "blob":
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", user, repo, branch, rest), nil
	case "tree":
		if rest != "" {
			return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s/SKILL.md", user, repo, branch, rest), nil
		}
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/SKILL.md", user, repo, branch), nil
	default:
		return "", fmt.Errorf("unrecognized GitHub URL type %q: %s", kind, input)
	}
}

func getRawURLGitLab(input string, parts []string) (string, error) {
	if len(parts) < 5 || parts[2] != "-" {
		return "", fmt.Errorf("unrecognized GitLab URL: %s", input)
	}
	kind := parts[3]
	switch kind {
	case "raw":
		return input, nil
	case "blob":
		parts[3] = "raw"
		return fmt.Sprintf("https://gitlab.com/%s", strings.Join(parts, "/")), nil
	case "tree":
		parts[3] = "raw"
		return fmt.Sprintf("https://gitlab.com/%s/SKILL.md", strings.Join(parts, "/")), nil
	default:
		return "", fmt.Errorf("unrecognized GitLab URL type %q: %s", kind, input)
	}
}

func GetLocalSkill(path string) (string, string, string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", "", "", err
	}
	if info.IsDir() {
		path = path + "/SKILL.md"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", "", err
	}

	content := string(data)
	name, description, err := parseFrontmatter(content)
	if err != nil {
		return "", "", "", fmt.Errorf("parse frontmatter from %s: %w", path, err)
	}

	return name, description, content, nil
}

func FetchSkill(rawURL string) (string, string, string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", "", "", fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("fetch %s: status %d", rawURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("read body: %w", err)
	}

	content := string(body)
	name, description, err := parseFrontmatter(content)
	if err != nil {
		return "", "", "", fmt.Errorf("parse frontmatter from %s: %w", rawURL, err)
	}

	return name, description, content, nil
}

func parseFrontmatter(content string) (string, string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return "", "", fmt.Errorf("no frontmatter found")
	}

	var name, description string
	var blockKey string
	var blockLines []string

	flushBlock := func() {
		if blockKey == "description" && len(blockLines) > 0 {
			description = strings.Join(blockLines, " ")
		}
		blockKey = ""
		blockLines = nil
	}

	for _, line := range lines[1:] {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "---" {
			break
		}
		if blockKey != "" {
			if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
				blockLines = append(blockLines, strings.TrimSpace(line))
				continue
			}
			flushBlock()
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if v == ">" || v == "|" {
			blockKey = k
			blockLines = nil
			continue
		}
		switch k {
		case "name":
			name = v
		case "description":
			description = v
		}
	}
	flushBlock()

	if name == "" {
		return "", "", fmt.Errorf("frontmatter missing 'name' field")
	}

	return name, description, nil
}
