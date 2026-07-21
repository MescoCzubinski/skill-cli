package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const maxSkillBytes = 5 << 20 // 5 MiB
const maxTreeAPIBytes = 10 << 20

var httpClient = &http.Client{Timeout: 30 * time.Second}

var (
	githubAPIBase = envOr("SKILL_CLI_GITHUB_API_BASE", "https://api.github.com")
	githubRawBase = envOr("SKILL_CLI_GITHUB_RAW_BASE", "https://raw.githubusercontent.com")
	gitlabAPIBase = envOr("SKILL_CLI_GITLAB_API_BASE", "https://gitlab.com/api/v4")
	gitlabRawBase = envOr("SKILL_CLI_GITLAB_RAW_BASE", "https://gitlab.com")
)

func envOr(name, def string) string {
	v := os.Getenv(name)
	if v == "" {
		return def
	}
	return v
}

const (
	SourceKindGitHub = "github"
	SourceKindGitLab = "gitlab"
	SourceKindLocal  = "local"
)

const (
	ResourceTypeClaude = "claude"
	ResourceTypeSkill  = "skill"
)

type SourceRef struct {
	Type       string
	SourceKind string
	Owner      string
	Repo       string
	Branch     string
	Dir        string
	LocalPath  string
	URL        string
	Input      string
}

type treeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

func ValidateSkillName(name string) error {
	var skillNameRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
	matched := skillNameRegexp.MatchString(name)
	if !matched {
		return fmt.Errorf("invalid skill name %q: must match [a-z0-9][a-z0-9._-]{0,63}", name)
	}
	if name == "claude" {
		return fmt.Errorf("skill name %q is reserved", name)
	}

	return nil
}

func ResolveSource(candidates ...string) (SourceRef, error) {
	input := ""
	for _, c := range candidates {
		if c != "" {
			input = c
			break
		}
	}
	if input == "" {
		return SourceRef{}, fmt.Errorf("no source URL provided")
	}

	isHTTPS := strings.HasPrefix(input, "https://")
	isHTTP := strings.HasPrefix(input, "http://")
	if !isHTTPS && !isHTTP {
		info, err := os.Stat(input)
		if err != nil {
			return SourceRef{}, err
		}
		typ := ResourceTypeSkill
		if !info.IsDir() && filepath.Base(input) == "CLAUDE.md" {
			typ = ResourceTypeClaude
		}
		return SourceRef{Type: typ, SourceKind: SourceKindLocal, LocalPath: input, Input: input}, nil
	}

	u, err := url.Parse(input)
	if err != nil {
		return SourceRef{}, fmt.Errorf("invalid URL: %w", err)
	}

	switch u.Host {
	case "github.com":
		return resolveGitHub(input, u)
	case "gitlab.com":
		return resolveGitLab(input, u)
	case "raw.githubusercontent.com":
		return SourceRef{}, fmt.Errorf("raw URLs are not supported; pass a github.com blob or tree URL instead: %s", input)
	default:
		return SourceRef{}, fmt.Errorf("unsupported host %q: only github.com and gitlab.com are supported", u.Host)
	}
}

func resolveGitHub(input string, u *url.URL) (SourceRef, error) {
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) < 4 {
		return SourceRef{}, fmt.Errorf("unrecognized GitHub URL: %s", input)
	}
	owner, repo, kind, branch := parts[0], parts[1], parts[2], parts[3]
	rest := parts[4:]

	switch kind {
	case "blob":
		if len(rest) == 0 {
			return SourceRef{}, fmt.Errorf("GitHub blob URL missing path: %s", input)
		}
		last := rest[len(rest)-1]
		if last == "CLAUDE.md" {
			url := fmt.Sprintf("%s/%s/%s/%s/%s", githubRawBase, owner, repo, branch, strings.Join(rest, "/"))
			return SourceRef{Type: ResourceTypeClaude, SourceKind: SourceKindGitHub, URL: url, Input: input}, nil
		}
		dir := strings.Join(rest[:len(rest)-1], "/")
		return SourceRef{Type: ResourceTypeSkill, SourceKind: SourceKindGitHub, Owner: owner, Repo: repo, Branch: branch, Dir: dir, Input: input}, nil
	case "tree":
		dir := strings.Join(rest, "/")
		return SourceRef{Type: ResourceTypeSkill, SourceKind: SourceKindGitHub, Owner: owner, Repo: repo, Branch: branch, Dir: dir, Input: input}, nil
	default:
		return SourceRef{}, fmt.Errorf("unrecognized GitHub URL type %q: %s", kind, input)
	}
}

func resolveGitLab(input string, u *url.URL) (SourceRef, error) {
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) < 5 || parts[2] != "-" {
		return SourceRef{}, fmt.Errorf("unrecognized GitLab URL: %s", input)
	}
	owner, repo, kind, branch := parts[0], parts[1], parts[3], parts[4]
	rest := parts[5:]

	switch kind {
	case "raw":
		return SourceRef{}, fmt.Errorf("raw URLs are not supported; pass a gitlab.com blob or tree URL instead: %s", input)
	case "blob":
		if len(rest) == 0 {
			return SourceRef{}, fmt.Errorf("GitLab blob URL missing path: %s", input)
		}
		last := rest[len(rest)-1]
		if last == "CLAUDE.md" {
			url := fmt.Sprintf("%s/%s/%s/-/raw/%s/%s", gitlabRawBase, owner, repo, branch, strings.Join(rest, "/"))
			return SourceRef{Type: ResourceTypeClaude, SourceKind: SourceKindGitLab, URL: url, Input: input}, nil
		}
		dir := strings.Join(rest[:len(rest)-1], "/")
		return SourceRef{Type: ResourceTypeSkill, SourceKind: SourceKindGitLab, Owner: owner, Repo: repo, Branch: branch, Dir: dir, Input: input}, nil
	case "tree":
		dir := strings.Join(rest, "/")
		return SourceRef{Type: ResourceTypeSkill, SourceKind: SourceKindGitLab, Owner: owner, Repo: repo, Branch: branch, Dir: dir, Input: input}, nil
	default:
		return SourceRef{}, fmt.Errorf("unrecognized GitLab URL type %q: %s", kind, input)
	}
}

func GitHubListTree(owner, repo, branch, dir string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/git/trees/%s?recursive=1", githubAPIBase, owner, repo, branch)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub trees API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub trees API: status %d", resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxTreeAPIBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read trees body: %w", err)
	}
	if int64(len(body)) > maxTreeAPIBytes {
		return nil, fmt.Errorf("GitHub trees API: response exceeds %d bytes", maxTreeAPIBytes)
	}

	var payload struct {
		Tree      []treeEntry `json:"tree"`
		Truncated bool        `json:"truncated"`
	}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		return nil, fmt.Errorf("parse trees body: %w", err)
	}
	if payload.Truncated {
		return nil, fmt.Errorf("GitHub trees API: tree truncated; skill too large to list")
	}

	return filterTreePaths(payload.Tree, dir), nil
}

func GitLabListTree(owner, repo, branch, dir string) ([]string, error) {
	rel := []string{}
	page := 1
	for {
		params := url.Values{}
		params.Set("ref", branch)
		params.Set("recursive", "true")
		params.Set("per_page", "100")
		params.Set("page", fmt.Sprintf("%d", page))
		if dir != "" {
			params.Set("path", dir)
		}
		endpoint := fmt.Sprintf("%s/projects/%s%%2F%s/repository/tree?%s", gitlabAPIBase, owner, repo, params.Encode())

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		token := os.Getenv("GITLAB_TOKEN")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("GitLab tree API: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("GitLab tree API: status %d", resp.StatusCode)
		}

		limited := io.LimitReader(resp.Body, maxTreeAPIBytes+1)
		body, err := io.ReadAll(limited)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read trees body: %w", err)
		}
		if int64(len(body)) > maxTreeAPIBytes {
			return nil, fmt.Errorf("GitLab tree API: response exceeds %d bytes", maxTreeAPIBytes)
		}

		var entries []treeEntry
		err = json.Unmarshal(body, &entries)
		if err != nil {
			return nil, fmt.Errorf("parse trees body: %w", err)
		}

		rel = append(rel, filterTreePaths(entries, dir)...)

		next := resp.Header.Get("X-Next-Page")
		if next == "" {
			break
		}
		page++
	}
	return rel, nil
}

func filterTreePaths(entries []treeEntry, dir string) []string {
	prefix := ""
	if dir != "" {
		prefix = strings.TrimSuffix(dir, "/") + "/"
	}
	out := []string{}
	for _, entry := range entries {
		if entry.Type != "blob" {
			continue
		}
		if prefix != "" {
			hasPrefix := strings.HasPrefix(entry.Path, prefix)
			if !hasPrefix {
				continue
			}
		}
		out = append(out, strings.TrimPrefix(entry.Path, prefix))
	}
	return out
}

func GitHubRawURL(owner, repo, branch, rel string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", githubRawBase, owner, repo, branch, rel)
}

func GitLabRawURL(owner, repo, branch, rel string) string {
	return fmt.Sprintf("%s/%s/%s/-/raw/%s/%s", gitlabRawBase, owner, repo, branch, rel)
}

func FetchSource(ref SourceRef) (map[string][]byte, error) {
	switch ref.Type {
	case ResourceTypeClaude:
		data, err := fetchSingle(ref)
		if err != nil {
			return nil, err
		}
		return map[string][]byte{"CLAUDE.md": data}, nil
	}

	switch ref.SourceKind {
	case SourceKindGitHub:
		paths, err := GitHubListTree(ref.Owner, ref.Repo, ref.Branch, ref.Dir)
		if err != nil {
			return nil, err
		}
		return fetchBlobs(paths, func(rel string) string {
			return GitHubRawURL(ref.Owner, ref.Repo, ref.Branch, joinSkillPath(ref.Dir, rel))
		})
	case SourceKindGitLab:
		paths, err := GitLabListTree(ref.Owner, ref.Repo, ref.Branch, ref.Dir)
		if err != nil {
			return nil, err
		}
		return fetchBlobs(paths, func(rel string) string {
			return GitLabRawURL(ref.Owner, ref.Repo, ref.Branch, joinSkillPath(ref.Dir, rel))
		})
	case SourceKindLocal:
		return readLocalSkillTree(ref.LocalPath)
	default:
		return nil, fmt.Errorf("unsupported source kind: %s", ref.SourceKind)
	}
}

func fetchSingle(ref SourceRef) ([]byte, error) {
	switch ref.SourceKind {
	case SourceKindLocal:
		return os.ReadFile(ref.LocalPath)
	default:
		content, err := Fetch(ref.URL)
		if err != nil {
			return nil, err
		}
		return []byte(content), nil
	}
}

func fetchBlobs(paths []string, urlFn func(string) string) (map[string][]byte, error) {
	out := map[string][]byte{}
	for _, rel := range paths {
		content, err := Fetch(urlFn(rel))
		if err != nil {
			return nil, err
		}
		out[rel] = []byte(content)
	}
	return out, nil
}

func readLocalSkillTree(path string) (map[string][]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return map[string][]byte{"SKILL.md": data}, nil
	}
	paths, err := ListLocalDir(path)
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, rel := range paths {
		data, err := os.ReadFile(filepath.Join(path, filepath.FromSlash(rel)))
		if err != nil {
			return nil, err
		}
		out[rel] = data
	}
	return out, nil
}

func joinSkillPath(dir, rel string) string {
	if dir == "" {
		return rel
	}
	return dir + "/" + rel
}

func ListLocalDir(root string) ([]string, error) {
	rel := []string{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := d.Name()
		if d.IsDir() {
			if name == ".git" {
				return filepath.SkipDir
			}
			isHidden := p != root && strings.HasPrefix(name, ".")
			if isHidden {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(name, ".") {
			return nil
		}
		relPath, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = append(rel, filepath.ToSlash(relPath))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rel, nil
}

func Fetch(rawURL string) (string, error) {
	resp, err := httpClient.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: status %d", rawURL, resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxSkillBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if int64(len(body)) > maxSkillBytes {
		return "", fmt.Errorf("fetch %s: body exceeds %d bytes", rawURL, maxSkillBytes)
	}

	return string(body), nil
}

func ParseFrontmatter(content string) (string, string, error) {
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
