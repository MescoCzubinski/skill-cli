package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
