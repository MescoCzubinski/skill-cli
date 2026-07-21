package e2e

import (
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

func TestAdd_LocalPath(t *testing.T) {
	env := newEnv(t)
	stdout, _, code := run(t, env, "add", fixture("first.md"))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "added: first-skill") {
		t.Errorf("unexpected stdout: %q", stdout)
	}

	metaPath := filepath.Join(env, "skill-cli", "meta", "first-skill.json")
	if _, err := os.Stat(metaPath); err != nil {
		t.Errorf("meta file not created: %v", err)
	}
	skillPath := filepath.Join(env, "skill-cli", "skills", "first-skill", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("skill file not created: %v", err)
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

func TestAdd_URL(t *testing.T) {
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

func TestAdd_URL_MultiFile(t *testing.T) {
	g := getGHFake(t)
	files := map[string]string{
		"SKILL.md": readFixture(t, "multi-skill/SKILL.md"),
		"extra.md": readFixture(t, "multi-skill/extra.md"),
	}
	userURL, _, _, _ := g.register(files)

	env := newEnv(t)
	claudeDir := filepath.Join(env, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := run(t, env, "add", userURL)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "added: multi-skill") {
		t.Errorf("unexpected stdout: %q", stdout)
	}

	for rel, want := range files {
		canonical := filepath.Join(env, "skill-cli", "skills", "multi-skill", rel)
		got, err := os.ReadFile(canonical)
		if err != nil {
			t.Errorf("canonical %s missing: %v", rel, err)
			continue
		}
		if string(got) != want {
			t.Errorf("canonical %s content mismatch", rel)
		}

		deployed := filepath.Join(claudeDir, "skills", "multi-skill", rel)
		if _, err := os.Stat(deployed); err != nil {
			t.Errorf("deployed %s missing: %v", rel, err)
		}
	}
}

// TestAdd_URL_NonMarkdownFiles verifies that a directory (tree) URL pulls every
// file in the skill dir, not just markdown — Python scripts and nested non-.md
// data files land in both the canonical and deploy trees byte-for-byte. This is
// the ".md + python scripts" case (e.g. Anthropic's xlsx skill).
func TestAdd_URL_NonMarkdownFiles(t *testing.T) {
	g := getGHFake(t)
	files := map[string]string{
		"SKILL.md":             "---\nname: scripted-skill\ndescription: skill shipping scripts\n---\n# Scripted skill\nRun scripts/run.py\n",
		"scripts/run.py":       "import sys\nprint(\"hello\")\n",
		"references/data.json": "{\"answer\": 42}\n",
	}
	userURL, _, _, _ := g.register(files)

	env := newEnv(t)
	claudeDir := filepath.Join(env, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := run(t, env, "add", userURL)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "added: scripted-skill") {
		t.Errorf("unexpected stdout: %q", stdout)
	}

	for rel, want := range files {
		canonical := filepath.Join(env, "skill-cli", "skills", "scripted-skill", rel)
		got, err := os.ReadFile(canonical)
		if err != nil {
			t.Errorf("canonical %s missing: %v", rel, err)
			continue
		}
		if string(got) != want {
			t.Errorf("canonical %s content mismatch: got %q want %q", rel, got, want)
		}

		deployed := filepath.Join(claudeDir, "skills", "scripted-skill", rel)
		got, err = os.ReadFile(deployed)
		if err != nil {
			t.Errorf("deployed %s missing: %v", rel, err)
			continue
		}
		if string(got) != want {
			t.Errorf("deployed %s content mismatch: got %q want %q", rel, got, want)
		}
	}
}

// TestAdd_LocalDir_NonMarkdownFiles is the local-directory analogue: adding a
// directory path pulls all files (scripts, nested data) while skipping dotfiles.
func TestAdd_LocalDir_NonMarkdownFiles(t *testing.T) {
	dir := t.TempDir()
	tree := map[string]string{
		"SKILL.md":             "---\nname: local-scripted\ndescription: local skill with scripts\n---\n# Local scripted skill\n",
		"scripts/run.py":       "print(\"local\")\n",
		"references/data.json": "{\"k\": 1}\n",
	}
	for rel, content := range tree {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// A dotfile must not be copied into the skill tree.
	if err := os.WriteFile(filepath.Join(dir, ".DS_Store"), []byte("junk"), 0644); err != nil {
		t.Fatal(err)
	}

	env := newEnv(t)
	stdout, stderr, code := run(t, env, "add", dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "added: local-scripted") {
		t.Errorf("unexpected stdout: %q", stdout)
	}

	for rel, want := range tree {
		canonical := filepath.Join(env, "skill-cli", "skills", "local-scripted", rel)
		got, err := os.ReadFile(canonical)
		if err != nil {
			t.Errorf("canonical %s missing: %v", rel, err)
			continue
		}
		if string(got) != want {
			t.Errorf("canonical %s content mismatch: got %q want %q", rel, got, want)
		}
	}

	dotfile := filepath.Join(env, "skill-cli", "skills", "local-scripted", ".DS_Store")
	if _, err := os.Stat(dotfile); !os.IsNotExist(err) {
		t.Errorf("expected dotfile to be skipped, stat returned err=%v", err)
	}
}

func TestAdd_Claude_LocalPath(t *testing.T) {
	env := newEnv(t)
	claudeDir := filepath.Join(env, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := run(t, env, "add", fixture("CLAUDE.md"))
	if code != 0 {
		t.Fatalf("expected exit 0, got %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "added: claude") {
		t.Errorf("unexpected stdout: %q", stdout)
	}

	canonical := filepath.Join(env, "skill-cli", "claude", "CLAUDE.md")
	if _, err := os.Stat(canonical); err != nil {
		t.Errorf("canonical CLAUDE.md not created: %v", err)
	}

	deployed := filepath.Join(claudeDir, "CLAUDE.md")
	data, err := os.ReadFile(deployed)
	if err != nil {
		t.Fatalf("deploy CLAUDE.md missing: %v", err)
	}
	want := readFixture(t, "CLAUDE.md")
	if string(data) != want {
		t.Errorf("deployed CLAUDE.md content mismatch")
	}
}
