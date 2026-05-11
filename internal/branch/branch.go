// Package branch resolves the current git branch and extracts a Shortcut
// story ID from it.
package branch

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var storyIDRegex = regexp.MustCompile(`(?i)sc-(\d+)`)

// Current returns the current git branch name, walking up from the current
// working directory to find the .git directory. Returns an error if not in
// a git repository or if HEAD is detached.
func Current() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return currentFrom(wd)
}

func currentFrom(start string) (string, error) {
	gitDir, err := findGitDir(start)
	if err != nil {
		return "", err
	}

	headPath := filepath.Join(gitDir, "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return "", fmt.Errorf("read HEAD: %w", err)
	}
	head := strings.TrimSpace(string(data))

	const refPrefix = "ref: refs/heads/"
	if !strings.HasPrefix(head, refPrefix) {
		return "", fmt.Errorf("detached HEAD")
	}
	return strings.TrimPrefix(head, refPrefix), nil
}

// findGitDir walks up from start looking for a .git directory or file.
// Supports git worktrees, where .git is a file containing "gitdir: <path>".
func findGitDir(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, ".git")
		info, err := os.Stat(candidate)
		if err == nil {
			if info.IsDir() {
				return candidate, nil
			}
			// .git file (worktree): "gitdir: <path>"
			data, err := os.ReadFile(candidate)
			if err != nil {
				return "", err
			}
			line := strings.TrimSpace(string(data))
			const prefix = "gitdir: "
			if !strings.HasPrefix(line, prefix) {
				return "", fmt.Errorf("unexpected .git file contents")
			}
			gitDir := strings.TrimPrefix(line, prefix)
			if !filepath.IsAbs(gitDir) {
				gitDir = filepath.Join(dir, gitDir)
			}
			return gitDir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository")
		}
		dir = parent
	}
}

// StoryID returns the first Shortcut story ID found in the branch name, if
// any. The match is case-insensitive on the "sc-" prefix.
func StoryID(branch string) (int, bool) {
	m := storyIDRegex.FindStringSubmatch(branch)
	if m == nil {
		return 0, false
	}
	id, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return id, true
}
