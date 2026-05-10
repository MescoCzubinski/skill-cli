package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const gitignoreContent = `remote

# macOS
.DS_Store
.AppleDouble
.LSOverride

# Linux
*~
.directory
.Trash-*

# Windows
Thumbs.db
ehthumbs.db
Desktop.ini
`

func HasRemote() bool {
	_, err := GetRemote()
	if err != nil {
		return false
	}
	return true
}

func IsGitRepo() bool {
	dir := ConfigDir()
	err := exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Run()
	return err == nil
}

func WriteGitignore() error {
	dir := ConfigDir()
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignoreContent), 0644)
}

func GitInit(remoteURL string) error {
    dir := ConfigDir()
    err := os.MkdirAll(dir, 0755)
    if err != nil {
        return err
    }

    out, err := exec.Command("git", "-C", dir, "init").CombinedOutput()
    if err != nil {
        return fmt.Errorf("git init failed: %w\n%s", err, out)
    }

    out, err = exec.Command("git", "-C", dir, "remote", "add", "origin", remoteURL).CombinedOutput()
    if err != nil {
        return fmt.Errorf("git remote add failed: %w\n%s", err, out)
    }

    out, err = exec.Command("git", "-C", dir, "ls-remote", "--heads", "origin").CombinedOutput()
    if err != nil {
		return fmt.Errorf("git ls-remote failed: %w\n%s", err, out)
	}

	hasCommits := len(out) > 0
	if hasCommits {
		out, err = exec.Command("git", "-C", dir, "pull", "origin", "main").CombinedOutput()
		if err != nil {
			return fmt.Errorf("git pull failed: %w\n%s", err, out)
		}

		out, err = exec.Command("git", "-C", dir, "branch", "--set-upstream-to=origin/main", "main").CombinedOutput()
		if err != nil {
			return fmt.Errorf("git branch failed: %w\n%s", err, out)
		}
	} else {
		out, err = exec.Command("git", "-C", dir, "commit", "-m", "init").CombinedOutput()
		if err != nil {
			return fmt.Errorf("git commit failed: %w\n%s", err, out)
		}

		out, err = exec.Command("git", "-C", dir, "push", "-u", "origin", "main").CombinedOutput()
		if err != nil {
			return fmt.Errorf("git push failed: %w\n%s", err, out)
		}
	}

    return nil
}

func GitSetOrigin(remoteURL string) error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "remote", "set-url", "origin", remoteURL).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git remote set-url failed: %w\n%s", err, out)
	}
	return nil
}

func GitPull() error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "pull", "--rebase").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w\n%s", err, out)
	}
	return nil
}

func GitPush(msg string) error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	isDirty := len(out) > 0
	if isDirty {
		out, err = exec.Command("git", "-C", dir, "add", ".").CombinedOutput()
		if err != nil {
			return fmt.Errorf("git add failed: %w\n%s", err, out)
		}

		out, err = exec.Command("git", "-C", dir, "commit", "-m", msg).CombinedOutput()
		if err != nil {
			return fmt.Errorf("git commit failed: %w\n%s", err, out)
		}
	}

	out, err = exec.Command("git", "-C", dir, "push", "-u", "origin", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w\n%s", err, out)
	}

	return nil
}
