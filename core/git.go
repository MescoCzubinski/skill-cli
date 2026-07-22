package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var allowedRemoteSchemes = []string{"https://", "http://", "ssh://", "git://", "file://", "git@"}

func ValidateRemoteURL(remoteURL string) error {
	for _, prefix := range allowedRemoteSchemes {
		if strings.HasPrefix(remoteURL, prefix) {
			return nil
		}
	}
	return fmt.Errorf("invalid remote URL %q: must start with one of %v", remoteURL, allowedRemoteSchemes)
}

func gitAvailable() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed - please install git to use remote sync")
	}

	return nil
}

func gitAddAll(dir string) error {
	out, err := exec.Command("git", "-C", dir, "add", ".").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, out)
	}

	return nil
}

func gitCommit(dir, msg string) error {
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}
	if len(out) == 0 {
		return nil
	}

	out, err = exec.Command("git", "-C", dir, "commit", "-m", msg).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %w\n%s", err, out)
	}

	return nil
}

func HasRemote() bool {
	err := gitAvailable()
	if err != nil {
		return false
	}
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "remote").Output()
	if err != nil {
		return false
	}

	return len(strings.TrimSpace(string(out))) > 0
}

func IsGitRepo() bool {
	dir := ConfigDir()
	err := exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Run()

	return err == nil
}

func GitInitRepo() error {
	dir := ConfigDir()
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	out, err := exec.Command("git", "-C", dir, "init").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git init failed: %w\n%s", err, out)
	}

	return nil
}

func GitAddOrigin(remoteURL string) error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "remote", "add", "origin", "--", remoteURL).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git remote add failed: %w\n%s", err, out)
	}

	return nil
}

func HasLocalChanges() (bool, error) {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}

	return len(out) > 0, nil
}

func GitFetchOrigin() error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "fetch", "origin").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %w\n%s", err, out)
	}

	return nil
}

func GitCommitAllowEmpty(msg string) error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", msg).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %w\n%s", err, out)
	}

	return nil
}

func GitAddAll() error {
	return gitAddAll(ConfigDir())
}

func GitCommit(msg string) error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "commit", "-m", msg).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %w\n%s", err, out)
	}

	return nil
}

func GitBranchMain() error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "branch", "-M", "main").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch rename failed: %w\n%s", err, out)
	}

	return nil
}

func GitMergeTheirsMain() error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "merge", "--allow-unrelated-histories", "-X", "theirs", "origin/main").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git merge failed: %w\n%s", err, out)
	}

	return nil
}

func GitSetUpstreamMain() error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "branch", "--set-upstream-to=origin/main", "main").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git set-upstream failed: %w\n%s", err, out)
	}

	return nil
}

func GitCheckoutTrackMain() error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "checkout", "-B", "main", "--track", "origin/main").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %w\n%s", err, out)
	}

	return nil
}

func GitAvailable() error {
	return gitAvailable()
}

func IsRemoteEmpty() (bool, error) {
	err := gitAvailable()
	if err != nil {
		return false, err
	}
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "ls-remote", "--heads", "origin").CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git ls-remote failed: %w\n%s", err, out)
	}

	return len(out) == 0, nil
}

func GitPushMain() error {
	err := gitAvailable()
	if err != nil {
		return err
	}
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "push", "-u", "origin", "main").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w\n%s", err, out)
	}

	return nil
}

func GitSetOrigin(remoteURL string) error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "remote", "set-url", "origin", "--", remoteURL).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git remote set-url failed: %w\n%s", err, out)
	}

	return nil
}

func GitPull() error {
	err := gitAvailable()
	if err != nil {
		return err
	}

	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "pull", "--rebase", "--autostash", "origin", "main").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w\n%s", err, out)
	}

	return nil
}

func GitPush(msg string) error {
	err := gitAvailable()
	if err != nil {
		return err
	}

	dir := ConfigDir()
	err = gitAddAll(dir)
	if err != nil {
		return err
	}
	err = gitCommit(dir, msg)
	if err != nil {
		return err
	}

	out, err := exec.Command("git", "-C", dir, "push", "-u", "origin", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w\n%s", err, out)
	}

	return nil
}
