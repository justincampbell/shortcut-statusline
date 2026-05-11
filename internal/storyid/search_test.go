package storyid

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/justincampbell/shortcut-statusline/internal/cache"
)

func newCache(t *testing.T) *cache.Cache {
	t.Helper()
	return &cache.Cache{Dir: t.TempDir(), TTL: time.Hour, WorkflowTTL: time.Hour}
}

func TestBranchSearch_CacheHitSkipsTokenAndAPI(t *testing.T) {
	c := newCache(t)
	const branch = "chore/sc-new-story/foo"
	if err := c.PutBranchStoryEntry(branch, 203310); err != nil {
		t.Fatal(err)
	}
	r := BranchSearch{Cache: c, Token: func() (string, error) {
		t.Fatalf("Token should not be called when cache is fresh")
		return "", nil
	}}
	id, ok, err := r.Resolve(context.Background(), branch)
	if err != nil || !ok || id != 203310 {
		t.Errorf("id=%d ok=%v err=%v", id, ok, err)
	}
}

func TestBranchSearch_CachedNoMatchSkipsAPI(t *testing.T) {
	c := newCache(t)
	const branch = "main"
	if err := c.PutBranchStoryEntry(branch, 0); err != nil {
		t.Fatal(err)
	}
	r := BranchSearch{Cache: c, Token: func() (string, error) {
		t.Fatalf("Token should not be called when no-match is cached")
		return "", nil
	}}
	id, ok, err := r.Resolve(context.Background(), branch)
	if err != nil || ok || id != 0 {
		t.Errorf("expected cached no-match: id=%d ok=%v err=%v", id, ok, err)
	}
}

func TestBranchSearch_TokenError(t *testing.T) {
	c := newCache(t)
	want := errors.New("no token")
	r := BranchSearch{Cache: c, Token: func() (string, error) {
		return "", want
	}}
	_, ok, err := r.Resolve(context.Background(), "any-uncached-branch")
	if ok {
		t.Errorf("expected miss on token error")
	}
	if !errors.Is(err, want) {
		t.Errorf("err = %v, want %v", err, want)
	}
}

func TestBranchSearch_Name(t *testing.T) {
	if (BranchSearch{}).Name() != "branch-search" {
		t.Errorf("unexpected name")
	}
}
