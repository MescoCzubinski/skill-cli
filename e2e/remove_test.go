package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestRemove_DeletesMetaAndSkillFile(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))

	stdout, _, code := run(t, env, "remove", "first-skill")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "removed: first-skill") {
		t.Errorf("unexpected stdout: %q", stdout)
	}

	metaPath := filepath.Join(env, "skill-cli", "meta", "first-skill.json")
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Errorf("meta file should be deleted after remove")
	}

	skillDir := filepath.Join(env, "skill-cli", "skills", "first-skill")
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Errorf("skill dir should be deleted after remove")
	}
}

func TestRemove_DeletesClaudeMetaAndFile(t *testing.T) {
	env := newEnv(t)
	claudeDir := filepath.Join(env, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(other, []byte(`{"keep": true}`), 0644); err != nil {
		t.Fatal(err)
	}

	run(t, env, "add", fixture("CLAUDE.md"))
	stdout, _, code := run(t, env, "remove", "claude")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "removed: claude") {
		t.Errorf("unexpected stdout: %q", stdout)
	}

	metaPath := filepath.Join(env, "skill-cli", "meta", "claude.json")
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Errorf("claude meta should be deleted after remove")
	}

	canonical := filepath.Join(env, "skill-cli", "claude", "CLAUDE.md")
	if _, err := os.Stat(canonical); !os.IsNotExist(err) {
		t.Errorf("canonical claude file should be deleted after remove")
	}

	deployed := filepath.Join(claudeDir, "CLAUDE.md")
	if _, err := os.Stat(deployed); !os.IsNotExist(err) {
		t.Errorf("deploy claude file should be deleted after remove")
	}

	if _, err := os.Stat(other); err != nil {
		t.Errorf("unrelated file in ~/.claude was deleted: %v", err)
	}
}
