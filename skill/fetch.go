package skill

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func ToRawURL(input string) (raw string, err error) {
	u, err := url.Parse(input)
	if err != nil {
		err = fmt.Errorf("invalid URL: %w", err)
		return
	}

	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")

	switch u.Host {
	case "raw.githubusercontent.com":
		return input, nil

	case "github.com":
		if len(parts) < 4 {
			err = fmt.Errorf("unrecognized GitHub URL: %s", input)
			return
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
			err = fmt.Errorf("unrecognized GitHub URL type %q: %s", kind, input)
			return
		}

	case "gitlab.com":
		if len(parts) < 5 || parts[2] != "-" {
			err = fmt.Errorf("unrecognized GitLab URL: %s", input)
			return
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
			err = fmt.Errorf("unrecognized GitLab URL type %q: %s", kind, input)
			return
		}

	default:
		err = fmt.Errorf("unsupported host %q: only github.com and gitlab.com supported", u.Host)
		return
	}
}

func (s *Skill) FetchSkill(today string) error {
	_, description, err := FetchMeta(s.RawURL)
	if err != nil {
		return err
	}
	s.Description = description
	s.UpdatedAt = today
	return nil
}

func FetchMeta(rawURL string) (name, description string, err error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		err = fmt.Errorf("fetch %s: %w", rawURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("fetch %s: status %d", rawURL, resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("read body: %w", err)
		return
	}

	name, description, err = parseFrontmatter(string(body))
	if err != nil {
		err = fmt.Errorf("parse frontmatter from %s: %w", rawURL, err)
	}
	return
}

func parseFrontmatter(content string) (name, description string, err error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		err = fmt.Errorf("no frontmatter found")
		return
	}

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
		err = fmt.Errorf("frontmatter missing 'name' field")
	}
	return
}
