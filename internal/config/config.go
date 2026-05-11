// Package config resolves the Shortcut API token from env or shortcut-cli config.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Token returns the Shortcut API token. SHORTCUT_API_TOKEN env var takes
// precedence over the shortcut-cli config file. Returns an error if neither
// source has a token.
func Token() (string, error) {
	if t := os.Getenv("SHORTCUT_API_TOKEN"); t != "" {
		return t, nil
	}

	path, err := configPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}

	var cfg struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.Token == "" {
		return "", fmt.Errorf("no token in %s and SHORTCUT_API_TOKEN unset", path)
	}
	return cfg.Token, nil
}

// configPath returns the location of the shortcut-cli config file. The
// `short` CLI uses XDG-style paths on all platforms (~/.config/shortcut-cli/),
// not the macOS-native ~/Library/Application Support/. Match that behavior so
// we share a single config.
func configPath() (string, error) {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "shortcut-cli", "config.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "shortcut-cli", "config.json"), nil
}
