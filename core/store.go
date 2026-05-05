package core

import (
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

func SaveSkillFile(name, content string) error {
	dir := filepath.Join(SkillsDir(), name)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	data := []byte(content)
	err = os.WriteFile(filepath.Join(dir, "SKILL.md"), data, 0644)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	for _, target := range deployTargets {
		assistantDir := filepath.Join(home, strings.SplitN(target, "/", 2)[0])
		_, err = os.Stat(assistantDir)
		if os.IsNotExist(err) {
			continue
		}

		skillDir := filepath.Join(home, target, name)
		err = os.MkdirAll(skillDir, 0755)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveSkillFile(name string) error {
	err := os.RemoveAll(filepath.Join(SkillsDir(), name))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	for _, target := range deployTargets {
		os.RemoveAll(filepath.Join(home, target, name))
	}
	return nil
}
