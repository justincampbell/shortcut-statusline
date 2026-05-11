package storyid

import (
	"context"
	"fmt"
	"os"

	"github.com/justincampbell/shortcut-statusline/internal/cache"
	"github.com/justincampbell/shortcut-statusline/internal/shortcut"
)

// BranchSearch resolves a story ID via Shortcut's /search/stories
// endpoint using the `branch:<name>` operator. Useful for branches that
// don't include `sc-NNNNN` (e.g. auto-created `chore/sc-new-story/*`
// branches), as long as Shortcut's VCS integration has linked the
// branch to its story.
//
// Results are cached in branches.json keyed by the branch name; a
// "no match" outcome is cached too so non-Shortcut branches like `main`
// don't trigger an API call on every prompt.
type BranchSearch struct {
	Cache *cache.Cache
	// Token returns a Shortcut API token. Deferred via a closure so we
	// don't read the config file just to bypass this resolver.
	Token func() (string, error)
	// NoCache forces a re-query (--no-cache / --refresh).
	NoCache bool
}

// Name implements Resolver.
func (BranchSearch) Name() string { return "branch-search" }

// Resolve implements Resolver. On API failure, falls back to a stale
// cached entry if one exists; otherwise returns the error.
func (s BranchSearch) Resolve(ctx context.Context, branchName string) (int, bool, error) {
	if !s.NoCache {
		if entry, fresh, _ := s.Cache.GetBranchStoryEntry(branchName); entry != nil && fresh {
			if entry.StoryID == 0 {
				return 0, false, nil
			}
			return entry.StoryID, true, nil
		}
	}

	token, err := s.Token()
	if err != nil {
		return 0, false, err
	}
	client := shortcut.New(token)

	stories, err := client.SearchStories(ctx, "branch:"+branchName)
	if err != nil {
		if entry, _, _ := s.Cache.GetBranchStoryEntry(branchName); entry != nil && entry.StoryID != 0 {
			return entry.StoryID, true, nil
		}
		return 0, false, err
	}

	var id int
	if len(stories) > 0 {
		id = stories[0].ID
	}
	if err := s.Cache.PutBranchStoryEntry(branchName, id); err != nil {
		fmt.Fprintln(os.Stderr, "shortcut-statusline: branch-story cache write:", err)
	}
	if id == 0 {
		return 0, false, nil
	}
	return id, true, nil
}
