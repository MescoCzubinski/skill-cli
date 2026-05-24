package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestUpdate_LocalSkill_Refreshes(t *testing.T) {
	dir := t.TempDir()
	original := readFixture(t, "first.md")
	skillPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	env := newEnv(t)
	run(t, env, "add", dir)

	updated := strings.Replace(original, "A skill for testing purposes", "Locally refreshed", 1)
	if err := os.WriteFile(skillPath, []byte(updated), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := run(t, env, "update", "first-skill")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "updated: first-skill") {
		t.Errorf("expected 'updated' in stdout, got: %q", stdout)
	}

	canonical := filepath.Join(env, "skill-cli", "skills", "first-skill", "SKILL.md")
	got, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "Locally refreshed") {
		t.Errorf("canonical SKILL.md not refreshed: %s", got)
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

func TestUpdate_Skill_ReplaceURL(t *testing.T) {
	first := readFixture(t, "first.md")
	url1 := serveSkill(t, first)

	env := newEnv(t)
	run(t, env, "add", url1)

	updated := strings.Replace(first, "A skill for testing purposes", "Refreshed via switch URL", 1)
	url2 := serveSkill(t, updated)

	stdout, _, code := run(t, env, "update", "first-skill", url2)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "updated: first-skill") {
		t.Errorf("expected 'updated: first-skill', got: %q", stdout)
	}

	metaPath := filepath.Join(env, "skill-cli", "meta", "first-skill.json")
	metaData, _ := os.ReadFile(metaPath)
	if !strings.Contains(string(metaData), url2) {
		t.Errorf("expected new url in meta, got: %s", metaData)
	}
}

func TestUpdate_Claude_SwitchURL(t *testing.T) {
	original := readFixture(t, "CLAUDE.md")
	v2 := readFixture(t, "CLAUDE-v2.md")
	url1 := serveClaude(t, original)
	url2 := serveClaude(t, v2)

	env := newEnv(t)
	if err := os.MkdirAll(filepath.Join(env, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}
	run(t, env, "add", url1)

	stdout, _, code := run(t, env, "update", "claude", url2)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "updated: claude") {
		t.Errorf("expected 'updated: claude', got: %q", stdout)
	}

	deployed := filepath.Join(env, ".claude", "CLAUDE.md")
	data, _ := os.ReadFile(deployed)
	if string(data) != v2 {
		t.Errorf("deploy file should hold v2 content")
	}
}

func TestUpdate_MultiFile_RemovedFileDisappears(t *testing.T) {
	g := getGHFake(t)
	skillMD := readFixture(t, "multi-skill/SKILL.md")
	extraMD := readFixture(t, "multi-skill/extra.md")
	v1 := map[string]string{"SKILL.md": skillMD, "extra.md": extraMD}
	userURL, _, _, _ := g.register(v1)

	env := newEnv(t)
	if _, _, code := run(t, env, "add", userURL); code != 0 {
		t.Fatal("add failed")
	}

	g.mu.Lock()
	for key := range g.repos {
		g.repos[key] = map[string]string{"SKILL.md": skillMD}
	}
	g.mu.Unlock()

	stdout, stderr, code := run(t, env, "update", "multi-skill")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "updated: multi-skill") {
		t.Errorf("expected 'updated: multi-skill', got: %q", stdout)
	}

	gone := filepath.Join(env, "skill-cli", "skills", "multi-skill", "extra.md")
	if _, err := os.Stat(gone); !os.IsNotExist(err) {
		t.Errorf("expected extra.md to be removed, but stat returned err=%v", err)
	}
}
