package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "skill-cli-bin-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not create temp dir:", err)
		os.Exit(1)
	}

	name := "skill-cli"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binaryPath = filepath.Join(tmp, name)

	out, err := exec.Command("go", "build", "-o", binaryPath, "../").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s\n", err, out)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// run executes the binary with the given XDG config home and args.
// Returns stdout, stderr, and the exit code.
func run(t *testing.T, xdgHome string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+xdgHome, "HOME="+xdgHome)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected run error: %v", err)
		}
	}
	return stdout.String(), stderr.String(), code
}

// newEnv returns a fresh temp directory to use as XDG_CONFIG_HOME.
func newEnv(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// fixture returns the absolute path to a testdata file.
func fixture(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

// readFixture reads a testdata file as a string.
func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(fixture(name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// ghFake simulates the GitHub trees API and raw blob endpoints from a single
// httptest server. Tests register virtual repos, each keyed by owner/repo/branch.
// Env vars SKILL_CLI_GITHUB_API_BASE / SKILL_CLI_GITHUB_RAW_BASE point the
// subprocess at the fake.
type ghFake struct {
	mu     sync.Mutex
	server *httptest.Server
	repos  map[string]map[string]string
	next   int
}

var (
	fakeMu       sync.Mutex
	fakeRegistry = map[*testing.T]*ghFake{}
)

func getGHFake(t *testing.T) *ghFake {
	t.Helper()
	fakeMu.Lock()
	defer fakeMu.Unlock()
	g, ok := fakeRegistry[t]
	if ok {
		return g
	}
	g = newGHFake(t)
	fakeRegistry[t] = g
	t.Cleanup(func() {
		fakeMu.Lock()
		delete(fakeRegistry, t)
		fakeMu.Unlock()
	})
	return g
}

func newGHFake(t *testing.T) *ghFake {
	t.Helper()
	g := &ghFake{
		repos: map[string]map[string]string{},
	}
	g.server = httptest.NewServer(http.HandlerFunc(g.handle))
	t.Cleanup(g.server.Close)
	t.Setenv("SKILL_CLI_GITHUB_API_BASE", g.server.URL+"/api")
	t.Setenv("SKILL_CLI_GITHUB_RAW_BASE", g.server.URL+"/raw")
	return g
}

func (g *ghFake) register(files map[string]string) (userURL, owner, repo, branch string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	owner = "testorg"
	repo = fmt.Sprintf("repo-%d", g.next)
	branch = "main"
	key := owner + "/" + repo + "/" + branch
	g.repos[key] = files
	userURL = fmt.Sprintf("https://github.com/%s/%s/tree/%s", owner, repo, branch)
	return
}

func (g *ghFake) handle(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	switch {
	case strings.HasPrefix(p, "/api/repos/"):
		g.serveTreesAPI(w, r, strings.TrimPrefix(p, "/api/repos/"))
	case strings.HasPrefix(p, "/raw/"):
		g.serveRaw(w, r, strings.TrimPrefix(p, "/raw/"))
	default:
		http.NotFound(w, r)
	}
}

func (g *ghFake) serveTreesAPI(w http.ResponseWriter, r *http.Request, rest string) {
	parts := strings.Split(rest, "/")
	if len(parts) < 5 || parts[2] != "git" || parts[3] != "trees" {
		http.NotFound(w, r)
		return
	}
	owner, repo, branch := parts[0], parts[1], parts[4]
	key := owner + "/" + repo + "/" + branch

	g.mu.Lock()
	files, ok := g.repos[key]
	g.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	type entry struct {
		Path string `json:"path"`
		Type string `json:"type"`
	}
	body := struct {
		Tree      []entry `json:"tree"`
		Truncated bool    `json:"truncated"`
	}{}
	for path := range files {
		body.Tree = append(body.Tree, entry{Path: path, Type: "blob"})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(body)
}

func (g *ghFake) serveRaw(w http.ResponseWriter, r *http.Request, rest string) {
	parts := strings.SplitN(rest, "/", 4)
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	owner, repo, branch, path := parts[0], parts[1], parts[2], parts[3]
	key := owner + "/" + repo + "/" + branch

	g.mu.Lock()
	files, ok := g.repos[key]
	g.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	content, ok := files[path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, content)
}

// serveSkill registers a single-file skill (just SKILL.md) on the per-test fake
// and returns the user-facing GitHub URL pointing at the tree.
func serveSkill(t *testing.T, content string) string {
	t.Helper()
	g := getGHFake(t)
	userURL, _, _, _ := g.register(map[string]string{"SKILL.md": content})
	return userURL
}

// serveClaude registers a single-file CLAUDE.md on the per-test fake and
// returns a blob URL pointing at it (CLAUDE.md is special-cased single-file).
func serveClaude(t *testing.T, content string) string {
	t.Helper()
	g := getGHFake(t)
	_, owner, repo, branch := g.register(map[string]string{"CLAUDE.md": content})
	return fmt.Sprintf("https://github.com/%s/%s/blob/%s/CLAUDE.md", owner, repo, branch)
}

// localBareRepo creates a bare git repo and returns its file:/// URL.
func localBareRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	out, err := exec.Command("git", "init", "--bare", dir).CombinedOutput()
	if err != nil {
		t.Fatalf("git init --bare: %v\n%s", err, out)
	}
	return "file://" + dir
}

// ---- unknown command ----
func TestUnknownCommand(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "bogus")
	if code == 0 {
		t.Fatal("expected exit 1")
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Errorf("expected 'unknown command' in stderr, got: %q", stderr)
	}
}
