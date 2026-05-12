package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

// serveSkill starts an httptest server serving content at /SKILL.md.
// Returns the URL to the raw file.
func serveSkill(t *testing.T, content string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, content)
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/SKILL.md"
}

// serve404 starts an httptest server that always returns 404.
func serve404(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/SKILL.md"
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

// readFixture reads a testdata file as a string.
func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(fixture(name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
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

// ---- add ----

func TestAdd_NoArgs(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "add")
	if code == 0 {
		t.Fatal("expected exit 1")
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage in stderr, got: %q", stderr)
	}
}

func TestAdd_LocalPath_First(t *testing.T) {
	env := newEnv(t)
	stdout, _, code := run(t, env, "add", fixture("first.md"))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "added: first-skill") {
		t.Errorf("unexpected stdout: %q", stdout)
	}
}

func TestAdd_LocalPath_AlreadyInstalled(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))
	_, stderr, code := run(t, env, "add", fixture("first.md"))
	if code == 0 {
		t.Fatal("expected exit 1 on duplicate add")
	}
	if !strings.Contains(stderr, "already installed") {
		t.Errorf("expected 'already installed' in stderr, got: %q", stderr)
	}
}

func TestAdd_LocalPath_FileNotFound(t *testing.T) {
	env := newEnv(t)
	_, _, code := run(t, env, "add", "/nonexistent/path/SKILL.md")
	if code == 0 {
		t.Fatal("expected exit 1")
	}
}

func TestAdd_LocalPath_NoFrontmatter(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "add", fixture("no-frontmatter.md"))
	if code == 0 {
		t.Fatal("expected exit 1")
	}
	if !strings.Contains(stderr, "frontmatter") {
		t.Errorf("expected frontmatter error in stderr, got: %q", stderr)
	}
}

func TestAdd_LocalDir(t *testing.T) {
	// Create a temp dir with a SKILL.md inside
	dir := t.TempDir()
	content := readFixture(t, "first.md")
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	env := newEnv(t)
	stdout, _, code := run(t, env, "add", dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "added: first-skill") {
		t.Errorf("unexpected stdout: %q", stdout)
	}
}

func TestAdd_URL_First(t *testing.T) {
	content := readFixture(t, "first.md")
	url := serveSkill(t, content)
	env := newEnv(t)
	stdout, _, code := run(t, env, "add", url)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "added: first-skill") {
		t.Errorf("unexpected stdout: %q", stdout)
	}
}

func TestAdd_URL_404(t *testing.T) {
	url := serve404(t)
	env := newEnv(t)
	_, stderr, code := run(t, env, "add", url)
	if code == 0 {
		t.Fatal("expected exit 1")
	}
	if !strings.Contains(stderr, "404") && !strings.Contains(stderr, "status") {
		t.Errorf("expected 404/status error in stderr, got: %q", stderr)
	}
}

func TestAdd_WritesMetaAndSkillFile(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))

	metaPath := filepath.Join(env, "skill-cli", "meta", "first-skill.json")
	if _, err := os.Stat(metaPath); err != nil {
		t.Errorf("meta file not created: %v", err)
	}

	skillPath := filepath.Join(env, "skill-cli", "skills", "first-skill", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("skill file not created: %v", err)
	}
}

func TestAdd_SyncsToDeploy(t *testing.T) {
	env := newEnv(t)
	claudeDir := filepath.Join(env, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	run(t, env, "add", fixture("first.md"))

	synced := filepath.Join(claudeDir, "skills", "first-skill", "SKILL.md")
	if _, err := os.Stat(synced); err != nil {
		t.Errorf("skill not synced to .claude/skills: %v", err)
	}
}

func TestAdd_RejectsMaliciousName(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "add", fixture("malicious-name.md"))
	if code == 0 {
		t.Fatal("expected exit 1 for malicious name")
	}
	if !strings.Contains(stderr, "invalid skill name") {
		t.Errorf("expected 'invalid skill name' in stderr, got: %q", stderr)
	}

	skillsRoot := filepath.Join(env, "skill-cli", "skills")
	entries, _ := os.ReadDir(skillsRoot)
	for _, e := range entries {
		if strings.Contains(e.Name(), "..") || strings.Contains(e.Name(), "escape") {
			t.Errorf("malicious dir leaked into skills/: %q", e.Name())
		}
	}

	configParent := filepath.Dir(env)
	if _, err := os.Stat(filepath.Join(configParent, "escape")); err == nil {
		t.Error("malicious path wrote outside config dir")
	}
}

func TestAdd_RejectsEmptyName(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "add", fixture("empty-name.md"))
	if code == 0 {
		t.Fatal("expected exit 1 for empty name")
	}
	if !strings.Contains(stderr, "invalid skill name") && !strings.Contains(stderr, "missing 'name'") {
		t.Errorf("expected name-related error in stderr, got: %q", stderr)
	}
}

func TestAdd_RejectsSlashName(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "add", fixture("slash-name.md"))
	if code == 0 {
		t.Fatal("expected exit 1 for slashed name")
	}
	if !strings.Contains(stderr, "invalid skill name") {
		t.Errorf("expected 'invalid skill name' in stderr, got: %q", stderr)
	}
}

func TestAdd_RejectsOversizeBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		chunk := strings.Repeat("a", 64*1024)
		for i := 0; i < 100; i++ {
			fmt.Fprint(w, chunk)
		}
	}))
	t.Cleanup(srv.Close)

	env := newEnv(t)
	_, stderr, code := run(t, env, "add", srv.URL+"/SKILL.md")
	if code == 0 {
		t.Fatal("expected exit 1 for oversize body")
	}
	if !strings.Contains(stderr, "exceeds") {
		t.Errorf("expected 'exceeds' in stderr, got: %q", stderr)
	}
}

// ---- list ----

func TestList_NoSkills(t *testing.T) {
	env := newEnv(t)
	stdout, _, code := run(t, env, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "no skills installed") {
		t.Errorf("expected 'no skills installed', got: %q", stdout)
	}
}

func TestList_OneSkill(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))

	stdout, _, code := run(t, env, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "first-skill") {
		t.Errorf("expected skill name in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "A skill for testing purposes") {
		t.Errorf("expected description in output, got: %q", stdout)
	}
}

func TestList_MultipleSkills(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))
	run(t, env, "add", fixture("second.md"))

	stdout, _, code := run(t, env, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "first-skill") {
		t.Errorf("expected first-skill in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "second-skill") {
		t.Errorf("expected second-skill in output, got: %q", stdout)
	}
}

// ---- remove ----

func TestRemove_NoArgs(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "remove")
	if code == 0 {
		t.Fatal("expected exit 1")
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage in stderr, got: %q", stderr)
	}
}

func TestRemove_ExistingSkill(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))

	stdout, _, code := run(t, env, "remove", "first-skill")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "removed: first-skill") {
		t.Errorf("unexpected stdout: %q", stdout)
	}
}

func TestRemove_NonExistent(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "remove", "no-such-skill")
	if code == 0 {
		t.Fatal("expected exit 1")
	}
	if !strings.Contains(stderr, "not installed") {
		t.Errorf("expected 'not installed' in stderr, got: %q", stderr)
	}
}

func TestRemove_DeletesMetaAndSkillFile(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))
	run(t, env, "remove", "first-skill")

	metaPath := filepath.Join(env, "skill-cli", "meta", "first-skill.json")
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Errorf("meta file should be deleted after remove")
	}

	skillDir := filepath.Join(env, "skill-cli", "skills", "first-skill")
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Errorf("skill dir should be deleted after remove")
	}
}

// ---- update ----

func TestUpdate_NoArgs(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "update")
	if code == 0 {
		t.Fatal("expected exit 1")
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage in stderr, got: %q", stderr)
	}
}

func TestUpdate_ContentChanged(t *testing.T) {
	originalContent := readFixture(t, "first.md")
	url := serveSkill(t, originalContent)

	env := newEnv(t)
	run(t, env, "add", url)

	updatedContent := strings.Replace(originalContent, "A skill for testing purposes", "Updated description", 1)
	updatedURL := serveSkill(t, updatedContent)

	metaPath := filepath.Join(env, "skill-cli", "meta", "first-skill.json")
	data, _ := os.ReadFile(metaPath)
	newData := strings.Replace(string(data), url, updatedURL, 1)
	os.WriteFile(metaPath, []byte(newData), 0644)

	stdout, _, code := run(t, env, "update", "first-skill")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "updated: first-skill") {
		t.Errorf("expected 'updated' in stdout, got: %q", stdout)
	}
}

func TestUpdate_ContentUnchanged(t *testing.T) {
	content := readFixture(t, "first.md")
	url := serveSkill(t, content)

	env := newEnv(t)
	run(t, env, "add", url)

	stdout, _, code := run(t, env, "update", "first-skill")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "unchanged: first-skill") {
		t.Errorf("expected 'unchanged' in stdout, got: %q", stdout)
	}
}

func TestUpdate_NonExistentSkill(t *testing.T) {
	env := newEnv(t)
	_, _, code := run(t, env, "update", "no-such-skill")
	if code == 0 {
		t.Fatal("expected exit 1")
	}
}

func TestUpdate_LocalSkillCannotUpdate(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))
	_, stderr, code := run(t, env, "update", "first-skill")
	if code == 0 {
		t.Fatal("expected exit 1 for local-only skill")
	}
	if !strings.Contains(stderr, "local file") {
		t.Errorf("expected 'local file' error in stderr, got: %q", stderr)
	}
}

func TestUpdate_AllNoSkills(t *testing.T) {
	env := newEnv(t)
	_, _, code := run(t, env, "update", "--all")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestUpdate_AllMultipleSkills(t *testing.T) {
	c1 := readFixture(t, "first.md")
	c2 := readFixture(t, "second.md")
	url1 := serveSkill(t, c1)
	url2 := serveSkill(t, c2)

	env := newEnv(t)
	run(t, env, "add", url1)
	run(t, env, "add", url2)

	stdout, _, code := run(t, env, "update", "--all")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "first-skill") {
		t.Errorf("expected first-skill in output, got: %q", stdout)
	}
	if !strings.Contains(stdout, "second-skill") {
		t.Errorf("expected second-skill in output, got: %q", stdout)
	}
}

func TestUpdateAll_ContinuesPastFailure(t *testing.T) {
	good := readFixture(t, "first.md")
	goodURL := serveSkill(t, good)
	badURL := serve404(t)

	env := newEnv(t)

	stdout, _, code := run(t, env, "add", goodURL)
	if code != 0 {
		t.Fatalf("setup add good failed: %d %q", code, stdout)
	}

	second := readFixture(t, "second.md")
	stagingURL := serveSkill(t, second)
	stdout, _, code = run(t, env, "add", stagingURL)
	if code != 0 {
		t.Fatalf("setup add second failed: %d %q", code, stdout)
	}

	metaPath := filepath.Join(env, "skill-cli", "meta", "second-skill.json")
	data, _ := os.ReadFile(metaPath)
	rewritten := strings.Replace(string(data), stagingURL, badURL, 1)
	os.WriteFile(metaPath, []byte(rewritten), 0644)

	updatedGood := strings.Replace(good, "A skill for testing purposes", "Updated good description", 1)
	updatedGoodURL := serveSkill(t, updatedGood)
	goodMeta := filepath.Join(env, "skill-cli", "meta", "first-skill.json")
	goodData, _ := os.ReadFile(goodMeta)
	goodRewritten := strings.Replace(string(goodData), goodURL, updatedGoodURL, 1)
	os.WriteFile(goodMeta, []byte(goodRewritten), 0644)

	stdout, stderr, code := run(t, env, "update", "--all")
	if code == 0 {
		t.Fatalf("expected non-zero exit when one upstream fails; stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stdout, "first-skill") {
		t.Errorf("expected first-skill to still update; stdout=%q", stdout)
	}
	if !strings.Contains(stderr, "second-skill") {
		t.Errorf("expected second-skill failure in stderr; stderr=%q", stderr)
	}
}

// ---- remote ----

func TestRemote_NoArgs(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "remote")
	if code == 0 {
		t.Fatal("expected exit 1")
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage in stderr, got: %q", stderr)
	}
}

func TestRemote_FreshConfigDir(t *testing.T) {
	env := newEnv(t)
	repoURL := localBareRepo(t)
	stdout, stderr, code := run(t, env, "remote", repoURL)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %q\nstderr: %q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "set remote:") {
		t.Errorf("expected 'set remote:' in stdout, got: %q", stdout)
	}

	gitDir := filepath.Join(env, "skill-cli", ".git")
	if _, err := os.Stat(gitDir); err != nil {
		t.Errorf(".git dir not found in config dir: %v", err)
	}
}

func TestRemote_PopulatedRemoteFreshLocal(t *testing.T) {
	repo := localBareRepo(t)

	seedEnv := newEnv(t)
	run(t, seedEnv, "add", fixture("first.md"))
	run(t, seedEnv, "remote", repo)

	env := newEnv(t)
	stdout, stderr, code := run(t, env, "remote", repo)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %q\nstderr: %q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "set remote:") {
		t.Errorf("expected 'set remote:' in stdout, got: %q", stdout)
	}

	skillPath := filepath.Join(env, "skill-cli", "skills", "first-skill", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("skill not pulled from remote: %v", err)
	}
}

func TestRemote_UpdateExistingRepo(t *testing.T) {
	env := newEnv(t)
	repo1 := localBareRepo(t)
	repo2 := localBareRepo(t)

	run(t, env, "remote", repo1)

	stdout, stderr, code := run(t, env, "remote", repo2)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %q\nstderr: %q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "set remote:") {
		t.Errorf("expected 'set remote:' in stdout, got: %q", stdout)
	}

	configDir := filepath.Join(env, "skill-cli")
	out, err := exec.Command("git", "-C", configDir, "remote", "get-url", "origin").Output()
	if err != nil {
		t.Fatalf("git remote get-url: %v", err)
	}
	if !strings.Contains(strings.TrimSpace(string(out)), strings.TrimPrefix(repo2, "file://")) {
		t.Errorf("remote URL not updated: got %q, want %q", strings.TrimSpace(string(out)), repo2)
	}
}

func TestRemote_RejectsBadScheme(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "remote", "ext::sh -c id")
	if code == 0 {
		t.Fatal("expected exit 1 for ext:: scheme")
	}
	if !strings.Contains(stderr, "invalid remote URL") {
		t.Errorf("expected 'invalid remote URL' in stderr, got: %q", stderr)
	}

	gitDir := filepath.Join(env, "skill-cli", ".git")
	if _, err := os.Stat(gitDir); err == nil {
		t.Error(".git should not have been created for bad scheme")
	}
}
