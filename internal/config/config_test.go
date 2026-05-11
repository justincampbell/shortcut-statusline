package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTokenEnvWins(t *testing.T) {
	t.Setenv("SHORTCUT_API_TOKEN", "env-token")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	got, err := Token()
	if err != nil {
		t.Fatal(err)
	}
	if got != "env-token" {
		t.Errorf("got %q", got)
	}
}

func TestTokenFromXDGConfigHome(t *testing.T) {
	t.Setenv("SHORTCUT_API_TOKEN", "")
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "shortcut-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"token":"xdg-token"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Token()
	if err != nil {
		t.Fatal(err)
	}
	if got != "xdg-token" {
		t.Errorf("got %q", got)
	}
}

func TestTokenFromHomeDotConfig(t *testing.T) {
	t.Setenv("SHORTCUT_API_TOKEN", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfgDir := filepath.Join(dir, ".config", "shortcut-cli")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(`{"token":"home-token"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Token()
	if err != nil {
		t.Fatal(err)
	}
	if got != "home-token" {
		t.Errorf("got %q", got)
	}
}

func TestTokenMissing(t *testing.T) {
	t.Setenv("SHORTCUT_API_TOKEN", "")
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	if _, err := Token(); err == nil {
		t.Errorf("expected error")
	}
}
