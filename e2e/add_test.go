package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
