package skill

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func configDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.Getenv("HOME") + "/.config"
	}
	return filepath.Join(base, "skill-cli")
}

func jsonPath() string {
	return filepath.Join(configDir(), "skills.json")
}

func LoadAll() ([]Skill, error) {
	data, err := os.ReadFile(jsonPath())
	if os.IsNotExist(err) {
		return []Skill{}, nil
	}
	if err != nil {
		return nil, err
	}

	var skills []Skill
	if err := json.Unmarshal(data, &skills); err != nil {
		return nil, err
	}
	return skills, nil
}

func SaveAll(skills []Skill) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(skills, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(jsonPath(), data, 0644)
}

func FindByName(skills []Skill, name string) (int, *Skill) {
	for i := range skills {
		if skills[i].Name == name {
			return i, &skills[i]
		}
	}
	return -1, nil
}
