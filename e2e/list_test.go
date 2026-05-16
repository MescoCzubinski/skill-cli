package e2e

import (
	"strings"
	"testing"
)

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

func TestList_WithSkillsAndClaude(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))
	run(t, env, "add", fixture("CLAUDE.md"))

	stdout, _, code := run(t, env, "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "Skills:") {
		t.Errorf("expected Skills: section, got: %q", stdout)
	}
	if !strings.Contains(stdout, "CLAUDE.md:") {
		t.Errorf("expected CLAUDE.md: section, got: %q", stdout)
	}
	if !strings.Contains(stdout, "first-skill") {
		t.Errorf("expected first-skill row, got: %q", stdout)
	}
	skillsIdx := strings.Index(stdout, "Skills:")
	claudeIdx := strings.Index(stdout, "CLAUDE.md:")
	if skillsIdx > claudeIdx {
		t.Errorf("Skills section should appear before CLAUDE.md section")
	}
}
