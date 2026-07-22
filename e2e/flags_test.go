package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlag_ListShortcut(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))

	for _, arg := range []string{"--list", "-l"} {
		stdout, _, code := run(t, env, arg)
		if code != 0 {
			t.Fatalf("%s: expected exit 0, got %d", arg, code)
		}
		if !strings.Contains(stdout, "first-skill") {
			t.Errorf("%s: expected skill name in output, got: %q", arg, stdout)
		}
	}
}

func TestFlag_Help(t *testing.T) {
	env := newEnv(t)
	for _, arg := range []string{"--help", "-h", "help"} {
		stdout, _, code := run(t, env, arg)
		if code != 0 {
			t.Fatalf("%s: expected exit 0, got %d", arg, code)
		}
		if !strings.Contains(stdout, "Commands:") {
			t.Errorf("%s: expected 'Commands:' in stdout, got: %q", arg, stdout)
		}
		if !strings.Contains(stdout, "--check") {
			t.Errorf("%s: expected flags section in stdout, got: %q", arg, stdout)
		}
	}
}

func TestFlag_Unknown(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "add", "--bogus", fixture("first.md"))
	if code == 0 {
		t.Fatal("expected exit 1 for unknown flag")
	}
	if !strings.Contains(stderr, "unknown flag") {
		t.Errorf("expected 'unknown flag' in stderr, got: %q", stderr)
	}
}

func TestUpdate_Check_UpToDate(t *testing.T) {
	content := readFixture(t, "first.md")
	url := serveSkill(t, content)

	env := newEnv(t)
	run(t, env, "add", url)

	stdout, _, code := run(t, env, "update", "first-skill", "--check")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "up to date: first-skill") {
		t.Errorf("expected 'up to date', got: %q", stdout)
	}
}

func TestUpdate_Check_UpdateAvailable_WritesNothing(t *testing.T) {
	original := readFixture(t, "first.md")
	url := serveSkill(t, original)

	env := newEnv(t)
	run(t, env, "add", url)

	updated := strings.Replace(original, "A skill for testing purposes", "Updated description", 1)
	updatedURL := serveSkill(t, updated)

	metaPath := filepath.Join(env, "skill-cli", "meta", "first-skill.json")
	metaBefore, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	rewritten := strings.Replace(string(metaBefore), url, updatedURL, 1)
	if err := os.WriteFile(metaPath, []byte(rewritten), 0644); err != nil {
		t.Fatalf("rewrite meta: %v", err)
	}

	skillPath := filepath.Join(env, "skill-cli", "skills", "first-skill", "SKILL.md")
	fileBefore, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read skill file: %v", err)
	}

	stdout, _, code := run(t, env, "update", "first-skill", "--check")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "update available: first-skill") {
		t.Errorf("expected 'update available', got: %q", stdout)
	}

	fileAfter, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read skill file after check: %v", err)
	}
	if string(fileAfter) != string(fileBefore) {
		t.Errorf("--check modified the skill file")
	}
	metaAfter, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta after check: %v", err)
	}
	if string(metaAfter) != rewritten {
		t.Errorf("--check modified the meta file")
	}
}

func TestUpdate_Check_All(t *testing.T) {
	c1 := readFixture(t, "first.md")
	c2 := readFixture(t, "second.md")
	url1 := serveSkill(t, c1)
	url2 := serveSkill(t, c2)

	env := newEnv(t)
	run(t, env, "add", url1)
	run(t, env, "add", url2)

	stdout, _, code := run(t, env, "update", "--all", "--check")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "first-skill") || !strings.Contains(stdout, "second-skill") {
		t.Errorf("expected both skills in output, got: %q", stdout)
	}
}

func TestUpdate_Check_LocalSkillFails(t *testing.T) {
	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))

	_, stderr, code := run(t, env, "update", "first-skill", "--check")
	if code == 0 {
		t.Fatal("expected exit 1 for local-only skill under --check")
	}
	if !strings.Contains(stderr, "local file") {
		t.Errorf("expected 'local file' in stderr, got: %q", stderr)
	}
}

func TestAdd_NoCommit_SkipsPush(t *testing.T) {
	repo := localBareRepo(t)

	envA := newEnv(t)
	run(t, envA, "add", fixture("first.md"))
	_, stderr, code := run(t, envA, "remote", repo)
	if code != 0 {
		t.Fatalf("remote attach failed: %d %q", code, stderr)
	}

	stdout, stderr, code := run(t, envA, "add", fixture("second.md"), "--no-commit")
	if code != 0 {
		t.Fatalf("add --no-commit failed: %d %q %q", code, stdout, stderr)
	}
	localSecond := filepath.Join(envA, "skill-cli", "skills", "second-skill", "SKILL.md")
	if _, err := os.Stat(localSecond); err != nil {
		t.Errorf("second skill should exist locally: %v", err)
	}

	envB := newEnv(t)
	_, stderr, code = run(t, envB, "remote", repo)
	if code != 0 {
		t.Fatalf("fresh remote pull failed: %d %q", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(envB, "skill-cli", "skills", "first-skill", "SKILL.md")); err != nil {
		t.Errorf("first skill should have been pushed and pulled: %v", err)
	}
	if _, err := os.Stat(filepath.Join(envB, "skill-cli", "skills", "second-skill", "SKILL.md")); !os.IsNotExist(err) {
		t.Errorf("second skill should NOT have been pushed (--no-commit)")
	}
}

func TestAdd_NoUpdate_SkipsPull(t *testing.T) {
	repo := localBareRepo(t)

	env := newEnv(t)
	run(t, env, "add", fixture("first.md"))
	_, stderr, code := run(t, env, "remote", repo)
	if code != 0 {
		t.Fatalf("remote attach failed: %d %q", code, stderr)
	}

	if err := os.RemoveAll(strings.TrimPrefix(repo, "file://")); err != nil {
		t.Fatalf("remove bare remote: %v", err)
	}

	_, _, code = run(t, env, "add", fixture("second.md"), "--no-commit")
	if code == 0 {
		t.Fatal("expected exit 1 when pull hits a missing remote")
	}

	stdout, stderr, code := run(t, env, "add", fixture("second.md"), "--no-update", "--no-commit")
	if code != 0 {
		t.Fatalf("add --no-update --no-commit failed: %d %q %q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "added: second-skill") {
		t.Errorf("expected 'added: second-skill', got: %q", stdout)
	}
}

func TestNoCommit_DeferredThenDefaultSyncs(t *testing.T) {
	repo := localBareRepo(t)
	url := serveSkill(t, readFixture(t, "first.md"))

	env := newEnv(t)
	run(t, env, "add", url)
	_, stderr, code := run(t, env, "remote", repo)
	if code != 0 {
		t.Fatalf("remote attach failed: %d %q", code, stderr)
	}

	_, stderr, code = run(t, env, "remove", "first-skill", "--no-commit")
	if code != 0 {
		t.Fatalf("remove --no-commit failed: %d %q", code, stderr)
	}

	stdout, stderr, code := run(t, env, "add", fixture("second.md"))
	if code != 0 {
		t.Fatalf("default add after deferred change failed: %d %q %q", code, stdout, stderr)
	}

	envB := newEnv(t)
	_, stderr, code = run(t, envB, "remote", repo)
	if code != 0 {
		t.Fatalf("fresh remote pull failed: %d %q", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(envB, "skill-cli", "skills", "second-skill", "SKILL.md")); err != nil {
		t.Errorf("second skill should have been pushed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(envB, "skill-cli", "skills", "first-skill", "SKILL.md")); !os.IsNotExist(err) {
		t.Errorf("deferred removal of first skill should have been pushed")
	}
}

func TestUpdate_All_IncludesPulledSkill(t *testing.T) {
	repo := localBareRepo(t)
	url1 := serveSkill(t, readFixture(t, "first.md"))
	url2 := serveSkill(t, readFixture(t, "second.md"))

	envA := newEnv(t)
	run(t, envA, "add", url1)
	_, stderr, code := run(t, envA, "remote", repo)
	if code != 0 {
		t.Fatalf("envA remote attach failed: %d %q", code, stderr)
	}

	envB := newEnv(t)
	run(t, envB, "add", url2)
	_, stderr, code = run(t, envB, "remote", repo)
	if code != 0 {
		t.Fatalf("envB remote attach failed: %d %q", code, stderr)
	}

	stdout, stderr, code := run(t, envA, "update", "--all")
	if code != 0 {
		t.Fatalf("update --all failed: %d %q %q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "second-skill") {
		t.Errorf("update --all should include the pulled second-skill, got: %q", stdout)
	}
}

func TestUpdate_TooManyArgs(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "update", "first-skill", "http://example.com", "extra")
	if code == 0 {
		t.Fatal("expected exit 1 for too many arguments")
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage in stderr, got: %q", stderr)
	}
}
