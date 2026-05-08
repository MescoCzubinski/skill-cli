package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var deployTargets = []string{
	".claude/skills",
	".cursor/skills",
	".gemini/skills",
	".gemini/antigravity/skills",
	".opencode/skills",
	".codex/skills",
}

func SyncSkillFiles() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	src := SkillsDir()
	for _, target := range deployTargets {
		assistantDir := filepath.Join(home, strings.SplitN(target, "/", 2)[0])
		_, err = os.Stat(assistantDir)
		if os.IsNotExist(err) {
			continue
		}
		dst := filepath.Join(home, target)
		err = copyDir(src, dst)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			err = os.WriteFile(dstPath, data, 0644)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func HasOrigin() (bool, error) {
	dir := ConfigDir()
	err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Run()
	if err == nil {
		return true, nil
	}
	_, isExitErr := err.(*exec.ExitError)
	if isExitErr {
		return false, nil
	}
	return false, err
}

func GitInit() error {
	dir := ConfigDir()
	remoteURL, err := GetRemote()
	if err != nil {
		return err
	}
	for _, args := range [][]string{
		{"git", "-C", dir, "init"},
		{"git", "-C", dir, "remote", "add", "origin", remoteURL},
		{"git", "-C", dir, "add", "."},
		{"git", "-C", dir, "commit", "-m", "init"},
		{"git", "-C", dir, "push", "-u", "origin", "main"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("%s failed: %w", args[2], err)
		}
	}
	return nil
}

func GitPush() error {
	dir := ConfigDir()
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	if err != nil || len(out) == 0 {
		return nil
	}

	for _, args := range [][]string{
		{"git", "-C", dir, "add", "."},
		{"git", "-C", dir, "commit", "-m", "sync"},
		{"git", "-C", dir, "push"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("%s failed: %w", args[2], err)
		}
	}
	return nil
}

func GitPull() error {
	dir := ConfigDir()
	cmd := exec.Command("git", "-C", dir, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}
	return nil
}
