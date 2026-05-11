package branch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoryID(t *testing.T) {
	tests := []struct {
		branch string
		want   int
		ok     bool
	}{
		{"feature/sc-12345-add-feature", 12345, true},
		{"bug/sc-23456-fix-bug", 23456, true},
		{"chore/sc-34567-cleanup", 34567, true},
		{"SC-45678-uppercase", 45678, true},
		{"sc-1-tiny", 1, true},
		{"main", 0, false},
		{"feature/no-story-id", 0, false},
		{"", 0, false},
		{"feature/sc-abc", 0, false},
	}
	for _, tt := range tests {
		got, ok := StoryID(tt.branch)
		if got != tt.want || ok != tt.ok {
			t.Errorf("StoryID(%q) = (%d, %v); want (%d, %v)", tt.branch, got, ok, tt.want, tt.ok)
		}
	}
}

func TestCurrentFrom(t *testing.T) {
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/feature/sc-12345-test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	sub := filepath.Join(tmp, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := currentFrom(sub)
	if err != nil {
		t.Fatal(err)
	}
	if got != "feature/sc-12345-test" {
		t.Errorf("got %q", got)
	}
}

func TestCurrentFromDetached(t *testing.T) {
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("abc123def\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := currentFrom(tmp); err == nil {
		t.Errorf("expected error for detached HEAD")
	}
}

func TestCurrentFromNotARepo(t *testing.T) {
	tmp := t.TempDir()
	if _, err := currentFrom(tmp); err == nil {
		t.Errorf("expected error outside repo")
	}
}

func TestCurrentFromWorktree(t *testing.T) {
	tmp := t.TempDir()
	mainGit := filepath.Join(tmp, "main", ".git")
	if err := os.MkdirAll(filepath.Join(mainGit, "worktrees", "wt"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mainGit, "worktrees", "wt", "HEAD"), []byte("ref: refs/heads/feature/sc-42-wt\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	wt := filepath.Join(tmp, "wt")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	gitFile := filepath.Join(wt, ".git")
	contents := "gitdir: " + filepath.Join(mainGit, "worktrees", "wt") + "\n"
	if err := os.WriteFile(gitFile, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := currentFrom(wt)
	if err != nil {
		t.Fatal(err)
	}
	if got != "feature/sc-42-wt" {
		t.Errorf("got %q", got)
	}
}
