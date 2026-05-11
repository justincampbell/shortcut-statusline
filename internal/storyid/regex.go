package storyid

import (
	"context"

	"github.com/justincampbell/shortcut-statusline/internal/branch"
)

// BranchRegex resolves a story ID by parsing `sc-NNNNN` out of the
// branch name. Pure string operation, no I/O.
func BranchRegex() Resolver { return branchRegex{} }

type branchRegex struct{}

func (branchRegex) Name() string { return "branch-regex" }

func (branchRegex) Resolve(_ context.Context, branchName string) (int, bool, error) {
	id, ok := branch.StoryID(branchName)
	return id, ok, nil
}
