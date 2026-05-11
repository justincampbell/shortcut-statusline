// Package cache stores the resolved story/epic/objective bundle on disk
// keyed by branch, so repeated statusline invocations don't hit the API.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/justincampbell/shortcut-statusline/internal/shortcut"
)

// DefaultTTL is how long a cached bundle is considered fresh.
const DefaultTTL = time.Hour

// DefaultWorkflowTTL is how long cached workflow state lookups are considered
// fresh. Workflows change rarely, so this is much longer than the bundle TTL.
const DefaultWorkflowTTL = 7 * 24 * time.Hour

// Bundle is the cached set of resources for one branch.
type Bundle struct {
	FetchedAt   int64               `json:"fetched_at"`
	Story       *shortcut.Story     `json:"story,omitempty"`
	Epic        *shortcut.Epic      `json:"epic,omitempty"`
	Objective   *shortcut.Objective `json:"objective,omitempty"`
	StoryStatus string              `json:"story_status,omitempty"`
	EpicStatus  string              `json:"epic_status,omitempty"`
}

// WorkflowStates is the cached id→name lookup for workflow + epic states.
type WorkflowStates struct {
	FetchedAt int64          `json:"fetched_at"`
	Story     map[int]string `json:"story,omitempty"`
	Epic      map[int]string `json:"epic,omitempty"`
}

// Cache is a per-user filesystem cache.
type Cache struct {
	Dir         string
	TTL         time.Duration
	WorkflowTTL time.Duration
}

const workflowFile = "workflows.json"

// New returns a Cache rooted at ~/.cache/shortcut-statusline (or
// $XDG_CACHE_HOME equivalent).
func New() (*Cache, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	ttl := DefaultTTL
	if v := os.Getenv("SHORTCUT_STATUSLINE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			ttl = d
		}
	}
	workflowTTL := DefaultWorkflowTTL
	if v := os.Getenv("SHORTCUT_STATUSLINE_WORKFLOW_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			workflowTTL = d
		}
	}
	return &Cache{
		Dir:         filepath.Join(base, "shortcut-statusline"),
		TTL:         ttl,
		WorkflowTTL: workflowTTL,
	}, nil
}

func (c *Cache) path(branch string) string {
	sum := sha256.Sum256([]byte(branch))
	return filepath.Join(c.Dir, hex.EncodeToString(sum[:])[:16]+".json")
}

// Get returns the bundle for branch and whether it is still fresh. Stale
// bundles are still returned (fresh=false) so callers can use them as a
// fallback on API failure.
func (c *Cache) Get(branch string) (b *Bundle, fresh bool, err error) {
	data, err := os.ReadFile(c.path(branch))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	b = &Bundle{}
	if err := json.Unmarshal(data, b); err != nil {
		return nil, false, err
	}
	age := time.Since(time.Unix(b.FetchedAt, 0))
	return b, age < c.TTL, nil
}

// Put writes b to the cache for branch, stamping FetchedAt to now.
func (c *Cache) Put(branch string, b *Bundle) error {
	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return err
	}
	b.FetchedAt = time.Now().Unix()
	data, err := json.Marshal(b)
	if err != nil {
		return err
	}
	final := c.path(branch)
	tmp, err := os.CreateTemp(c.Dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		// Best-effort cleanup if rename never happened.
		_ = os.Remove(tmpName)
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, final); err != nil {
		return fmt.Errorf("rename cache: %w", err)
	}
	return nil
}

// Delete removes the cache entry for branch (if any).
func (c *Cache) Delete(branch string) error {
	err := os.Remove(c.path(branch))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// GetWorkflowStates returns the cached id→name maps for workflow + epic
// states and whether they are fresh.
func (c *Cache) GetWorkflowStates() (s *WorkflowStates, fresh bool, err error) {
	data, err := os.ReadFile(filepath.Join(c.Dir, workflowFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	s = &WorkflowStates{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, false, err
	}
	age := time.Since(time.Unix(s.FetchedAt, 0))
	return s, age < c.WorkflowTTL, nil
}

// PutWorkflowStates writes the id→name maps, atomic via temp+rename.
func (c *Cache) PutWorkflowStates(s *WorkflowStates) error {
	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return err
	}
	s.FetchedAt = time.Now().Unix()
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	final := filepath.Join(c.Dir, workflowFile)
	tmp, err := os.CreateTemp(c.Dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, final); err != nil {
		return fmt.Errorf("rename workflow cache: %w", err)
	}
	return nil
}
