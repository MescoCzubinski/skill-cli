package core

import (
	"bytes"
	"os"
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

func SaveSkillFile(name, content string) (bool, error) {
	dir := filepath.Join(SkillsDir(), name)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return false, err
	}

	path := filepath.Join(dir, "SKILL.md")
	data := []byte(content)

	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, data) {
		return false, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}

func SkillFileChanged(name, content string) (bool, error) {
	path := filepath.Join(SkillsDir(), name, "SKILL.md")
	existing, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return !bytes.Equal(existing, []byte(content)), nil
}

func RemoveSkillFile(name string) error {
	err := os.RemoveAll(filepath.Join(SkillsDir(), name))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func claudeDir() string {
	return filepath.Join(ConfigDir(), "claude")
}

func claudeFilePath() string {
	return filepath.Join(claudeDir(), "CLAUDE.md")
}

func SaveClaudeFile(content string) (bool, error) {
	err := os.MkdirAll(claudeDir(), 0755)
	if err != nil {
		return false, err
	}

	path := claudeFilePath()
	data := []byte(content)

	existing, err := os.ReadFile(path)
	if err == nil && bytes.Equal(existing, data) {
		return false, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}

func ClaudeFileChanged(content string) (bool, error) {
	existing, err := os.ReadFile(claudeFilePath())
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return !bytes.Equal(existing, []byte(content)), nil
}

func RemoveClaudeFile() error {
	err := os.RemoveAll(claudeDir())
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func SyncClaude() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	assistantDir := filepath.Join(home, ".claude")
	_, err = os.Stat(assistantDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	dst := filepath.Join(assistantDir, "CLAUDE.md")
	srcData, err := os.ReadFile(claudeFilePath())
	if os.IsNotExist(err) {
		err = os.Remove(dst)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}

	return os.WriteFile(dst, srcData, 0644)
}

func SyncSkillFiles() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	src := SkillsDir()
	canonical, err := listSkillNames(src)
	if err != nil {
		return err
	}

	for _, target := range deployTargets {
		assistantDir := filepath.Join(home, strings.SplitN(target, "/", 2)[0])
		_, err = os.Stat(assistantDir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}

		dst := filepath.Join(home, target)
		err = os.MkdirAll(dst, 0755)
		if err != nil {
			return err
		}

		deployed, err := listSkillNames(dst)
		if err != nil {
			return err
		}

		for name := range deployed {
			_, keep := canonical[name]
			if keep {
				continue
			}
			err = os.RemoveAll(filepath.Join(dst, name))
			if err != nil {
				return err
			}
		}

		for name := range canonical {
			err = copyDir(filepath.Join(src, name), filepath.Join(dst, name))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func listSkillNames(dir string) (map[string]struct{}, error) {
	names := map[string]struct{}{}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return names, nil
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		names[entry.Name()] = struct{}{}
	}

	return names, nil
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
			if err != nil {
				return err
			}
			continue
		}
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		err = os.WriteFile(dstPath, data, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
