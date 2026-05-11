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

// DefaultWorkflowTTL is how long cached workspace-wide lookups (workflows,
// members, groups) are considered fresh. These change rarely, so this is
// much longer than the bundle TTL.
const DefaultWorkflowTTL = 7 * 24 * time.Hour

// BundleSchemaVersion is bumped whenever Bundle gains a derived field that
// can't be populated from an older cached payload. A cached bundle with a
// lower SchemaVersion is treated as needing a refetch even when otherwise
// fresh — see hasNeededData in main.go.
const BundleSchemaVersion = 3

// Bundle is the cached set of resources for one branch. Owner / requestor
// / team fields keep the full MemberInfo / GroupInfo records so callers
// can render whichever variant ({owner}, {ownerMention}, {ownerName}) the
// format requests without re-fetching.
type Bundle struct {
	SchemaVersion  int                 `json:"schema_version,omitempty"`
	FetchedAt      int64               `json:"fetched_at"`
	Story          *shortcut.Story     `json:"story,omitempty"`
	Epic           *shortcut.Epic      `json:"epic,omitempty"`
	Objective      *shortcut.Objective `json:"objective,omitempty"`
	StoryState     string              `json:"story_state,omitempty"`
	StoryStateType string              `json:"story_state_type,omitempty"`
	EpicState      string              `json:"epic_state,omitempty"`
	EpicStateType  string              `json:"epic_state_type,omitempty"`
	StoryOwner     *MemberInfo         `json:"story_owner,omitempty"`
	StoryRequestor *MemberInfo         `json:"story_requestor,omitempty"`
	StoryTeam      *GroupInfo          `json:"story_team,omitempty"`
	EpicOwner      *MemberInfo         `json:"epic_owner,omitempty"`
	EpicTeam       *GroupInfo          `json:"epic_team,omitempty"`
}

// StateInfo is the cached name+type for a single workflow state.
type StateInfo struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// WorkflowStates is the cached id→{name,type} lookup for workflow + epic
// states.
type WorkflowStates struct {
	FetchedAt int64             `json:"fetched_at"`
	Story     map[int]StateInfo `json:"story,omitempty"`
	Epic      map[int]StateInfo `json:"epic,omitempty"`
}

// MemberInfo is the cached display name(s) for a single workspace member.
// MentionName is preferred for rendering; Name is a fallback when
// MentionName is unset.
type MemberInfo struct {
	MentionName string `json:"mention_name,omitempty"`
	Name        string `json:"name,omitempty"`
}

// Members is the cached id→MemberInfo lookup for the whole workspace.
type Members struct {
	FetchedAt int64                 `json:"fetched_at"`
	Members   map[string]MemberInfo `json:"members,omitempty"`
}

// GroupInfo is the cached display name(s) for a single workspace group/team.
type GroupInfo struct {
	MentionName string `json:"mention_name,omitempty"`
	Name        string `json:"name,omitempty"`
}

// Groups is the cached id→GroupInfo lookup for the whole workspace.
type Groups struct {
	FetchedAt int64                `json:"fetched_at"`
	Groups    map[string]GroupInfo `json:"groups,omitempty"`
}

// BranchStoryEntry records the story ID resolved for a given branch via
// the Shortcut search API. StoryID == 0 is a cached "no match" — we
// asked Shortcut and got nothing back, so the next invocation can skip
// the API call too.
type BranchStoryEntry struct {
	StoryID   int   `json:"story_id"`
	FetchedAt int64 `json:"fetched_at"`
}

// BranchStories is the on-disk shape of branches.json: a single map of
// hashed branch names to story IDs.
type BranchStories struct {
	Branches map[string]BranchStoryEntry `json:"branches,omitempty"`
}

// Cache is a per-user filesystem cache.
type Cache struct {
	Dir         string
	TTL         time.Duration
	WorkflowTTL time.Duration
}

const (
	workflowFile = "workflows.json"
	membersFile  = "members.json"
	groupsFile   = "groups.json"
	branchesFile = "branches.json"
)

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

// Put writes b to the cache for branch, stamping FetchedAt + SchemaVersion.
func (c *Cache) Put(branch string, b *Bundle) error {
	b.FetchedAt = time.Now().Unix()
	b.SchemaVersion = BundleSchemaVersion
	return c.writeAtomic(c.path(branch), b)
}

// writeAtomic marshals v to JSON and replaces path via a temp+rename in the
// same directory, so a reader never sees a half-written file.
func (c *Cache) writeAtomic(path string, v any) error {
	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
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
	if err := os.Rename(tmpName, path); err != nil {
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
	s.FetchedAt = time.Now().Unix()
	return c.writeAtomic(filepath.Join(c.Dir, workflowFile), s)
}

// GetMembers returns the cached id→MemberInfo map and whether it is fresh.
func (c *Cache) GetMembers() (m *Members, fresh bool, err error) {
	data, err := os.ReadFile(filepath.Join(c.Dir, membersFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	m = &Members{}
	if err := json.Unmarshal(data, m); err != nil {
		return nil, false, err
	}
	age := time.Since(time.Unix(m.FetchedAt, 0))
	return m, age < c.WorkflowTTL, nil
}

// PutMembers writes the id→MemberInfo map, atomic via temp+rename.
func (c *Cache) PutMembers(m *Members) error {
	m.FetchedAt = time.Now().Unix()
	return c.writeAtomic(filepath.Join(c.Dir, membersFile), m)
}

// GetGroups returns the cached id→GroupInfo map and whether it is fresh.
func (c *Cache) GetGroups() (g *Groups, fresh bool, err error) {
	data, err := os.ReadFile(filepath.Join(c.Dir, groupsFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	g = &Groups{}
	if err := json.Unmarshal(data, g); err != nil {
		return nil, false, err
	}
	age := time.Since(time.Unix(g.FetchedAt, 0))
	return g, age < c.WorkflowTTL, nil
}

// PutGroups writes the id→GroupInfo map, atomic via temp+rename.
func (c *Cache) PutGroups(g *Groups) error {
	g.FetchedAt = time.Now().Unix()
	return c.writeAtomic(filepath.Join(c.Dir, groupsFile), g)
}

// GetBranchStoryEntry returns the cached branch→story-id entry, if any,
// along with whether it is still fresh (per WorkflowTTL). A nil entry
// means the branch isn't cached at all; an entry with StoryID=0 is a
// cached "no match" — useful for branches like `main` that we don't
// want to re-query on every prompt.
func (c *Cache) GetBranchStoryEntry(branch string) (entry *BranchStoryEntry, fresh bool, err error) {
	bs, err := c.loadBranchStories()
	if err != nil || bs == nil {
		return nil, false, err
	}
	e, ok := bs.Branches[hashBranch(branch)]
	if !ok {
		return nil, false, nil
	}
	age := time.Since(time.Unix(e.FetchedAt, 0))
	return &e, age < c.WorkflowTTL, nil
}

// PutBranchStoryEntry merges (branch → storyID) into branches.json. A
// storyID of 0 records a cached "no match". Other entries in the file
// are preserved (read-modify-write).
func (c *Cache) PutBranchStoryEntry(branch string, storyID int) error {
	bs, err := c.loadBranchStories()
	if err != nil {
		return err
	}
	if bs == nil {
		bs = &BranchStories{}
	}
	if bs.Branches == nil {
		bs.Branches = map[string]BranchStoryEntry{}
	}
	bs.Branches[hashBranch(branch)] = BranchStoryEntry{
		StoryID:   storyID,
		FetchedAt: time.Now().Unix(),
	}
	return c.writeAtomic(filepath.Join(c.Dir, branchesFile), bs)
}

func (c *Cache) loadBranchStories() (*BranchStories, error) {
	data, err := os.ReadFile(filepath.Join(c.Dir, branchesFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	bs := &BranchStories{}
	if err := json.Unmarshal(data, bs); err != nil {
		return nil, err
	}
	return bs, nil
}

func hashBranch(branch string) string {
	sum := sha256.Sum256([]byte(branch))
	return hex.EncodeToString(sum[:])[:16]
}
