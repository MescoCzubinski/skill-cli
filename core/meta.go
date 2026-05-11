package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrSkillNotFound = errors.New("skill not found")

func ConfigDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.Getenv("HOME") + "/.config"
	}

	return filepath.Join(base, "skill-cli")
}

func SkillsDir() string {
	return filepath.Join(ConfigDir(), "skills")
}

func metaDir() string {
	return filepath.Join(ConfigDir(), "meta")
}

func metaPath(name string) string {
	return filepath.Join(metaDir(), name+".json")
}

func readSkillMeta(name string) (*Skill, error) {
	data, err := os.ReadFile(metaPath(name))
	if err != nil {
		return nil, err
	}

	var meta SkillMeta
	err = json.Unmarshal(data, &meta)
	if err != nil {
		return nil, err
	}

	return &Skill{
		Name:        name,
		RawURL:      meta.RawURL,
		InstalledAt: meta.InstalledAt,
		UpdatedAt:   meta.UpdatedAt,
	}, nil
}

func writeSkillMeta(name string, meta SkillMeta) error {
	err := os.MkdirAll(metaDir(), 0755)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metaPath(name), data, 0644)
}

func GetSkillsMeta() ([]Skill, error) {
	entries, err := os.ReadDir(metaDir())
	if os.IsNotExist(err) {
		return []Skill{}, nil
	}
	if err != nil {
		return nil, err
	}

	skills := []Skill{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		hasSuffix := strings.HasSuffix(filename, ".json")
		if !hasSuffix {
			continue
		}
		name := strings.TrimSuffix(filename, ".json")
		skill, err := readSkillMeta(name)
		if err != nil {
			return nil, err
		}
		skills = append(skills, *skill)
	}

	return skills, nil
}

func SaveSkillsMeta(skills []Skill) error {
	for _, s := range skills {
		err := writeSkillMeta(s.Name, SkillMeta{
			RawURL:      s.RawURL,
			InstalledAt: s.InstalledAt,
			UpdatedAt:   s.UpdatedAt,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func GetSkillMeta(name string) (*Skill, error) {
	skill, err := readSkillMeta(name)
	if os.IsNotExist(err) {
		return nil, ErrSkillNotFound
	}
	if err != nil {
		return nil, err
	}

	return skill, nil
}

func UpdateSkillMeta(name, description, today string) error {
	data, err := os.ReadFile(metaPath(name))
	if os.IsNotExist(err) {
		return ErrSkillNotFound
	}
	if err != nil {
		return err
	}

	var meta SkillMeta
	err = json.Unmarshal(data, &meta)
	if err != nil {
		return err
	}
	if meta.RawURL == "" {
		return fmt.Errorf("skill %q was installed from a local file and cannot be updated", name)
	}

	meta.UpdatedAt = today
	return writeSkillMeta(name, meta)
}

func RemoveSkillMeta(name string) error {
	err := os.Remove(metaPath(name))
	if os.IsNotExist(err) {
		return nil
	}

	return err
}

func SaveSkillMeta(name, description, rawURL string) error {
	_, err := os.Stat(metaPath(name))
	if err == nil {
		return fmt.Errorf("skill %q already installed", name)
	}
	if !os.IsNotExist(err) {
		return err
	}

	today := time.Now().Format("2006-01-02")
	return writeSkillMeta(name, SkillMeta{
		RawURL:      rawURL,
		InstalledAt: today,
		UpdatedAt:   today,
	})
}
