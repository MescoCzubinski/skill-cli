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
