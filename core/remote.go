package core

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrNoRemote = errors.New("no remote registered")

func remotePath() string {
	return filepath.Join(ConfigDir(), "remote")
}

func GetRemote() (string, error) {
	data, err := os.ReadFile(remotePath())
	if os.IsNotExist(err) {
		return "", ErrNoRemote
	}
	if err != nil {
		return "", err
	}
	remoteURL := strings.TrimSpace(string(data))
	if remoteURL == "" {
		return "", ErrNoRemote
	}
	return remoteURL, nil
}

func SetRemote(remoteURL string) error {
	dir := ConfigDir()
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(remotePath(), []byte(remoteURL+"\n"), 0644)
}

func ClearRemote() error {
	err := os.Remove(remotePath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func ValidateRemoteURL(remoteURL string) error {
	err := exec.Command("git", "ls-remote", remoteURL).Run()
	if err != nil {
		return errors.New("repository not found or not accessible: " + remoteURL)
	}
	return nil
}
