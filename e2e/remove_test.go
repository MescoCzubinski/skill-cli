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
