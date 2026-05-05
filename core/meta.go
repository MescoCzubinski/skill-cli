package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

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

func jsonPath() string {
	return filepath.Join(ConfigDir(), "skills.json")
}

func GetSkillsMeta() ([]Skill, error) {
	data, err := os.ReadFile(jsonPath())
	if os.IsNotExist(err) {
		return []Skill{}, nil
	}
	if err != nil {
		return nil, err
	}

	var skills []Skill
	err = json.Unmarshal(data, &skills)
	if err != nil {
		return nil, err
	}
	return skills, nil
}

func SaveSkillsMeta(skills []Skill) error {
	dir := ConfigDir()
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(skills, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(jsonPath(), data, 0644)
}

func findInSkills(skills []Skill, name string) int {
	for i := range skills {
		if skills[i].Name == name {
			return i
		}
	}
	return -1
}

func FindSkillMeta(name string) (*Skill, error) {
	skills, err := GetSkillsMeta()
	if err != nil {
		return nil, err
	}
	i := findInSkills(skills, name)
	if i == -1 {
		return nil, nil
	}
	return &skills[i], nil
}

func UpdateSkillMeta(name, description, today string) error {
	skills, err := GetSkillsMeta()
	if err != nil {
		return err
	}
	i := findInSkills(skills, name)
	if i == -1 {
		return fmt.Errorf("skill %q not found", name)
	}
	skills[i].Description = description
	skills[i].UpdatedAt = today
	return SaveSkillsMeta(skills)
}

func RemoveSkillMeta(name string) error {
	skills, err := GetSkillsMeta()
	if err != nil {
		return err
	}
	i := findInSkills(skills, name)
	if i == -1 {
		return nil
	}
	skills = append(skills[:i], skills[i+1:]...)
	return SaveSkillsMeta(skills)
}

func SaveSkillMeta(name, description, rawURL string) error {
	skills, err := GetSkillsMeta()
	if err != nil {
		return err
	}
	if findInSkills(skills, name) != -1 {
		return fmt.Errorf("skill %q already installed", name)
	}
	today := time.Now().Format("2006-01-02")
	skills = append(skills, Skill{
		Name:        name,
		Description: description,
		RawURL:      rawURL,
		InstalledAt: today,
		UpdatedAt:   today,
	})
	return SaveSkillsMeta(skills)
}
