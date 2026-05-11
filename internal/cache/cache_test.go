package cache

import (
	"testing"
	"time"

	"github.com/justincampbell/shortcut-statusline/internal/shortcut"
)

func TestPutGet(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour}
	b := &Bundle{Story: &shortcut.Story{ID: 1, Name: "hi"}}
	if err := c.Put("feature/sc-1", b); err != nil {
		t.Fatal(err)
	}
	got, fresh, err := c.Get("feature/sc-1")
	if err != nil {
		t.Fatal(err)
	}
	if !fresh {
		t.Errorf("expected fresh")
	}
	if got.Story.Name != "hi" {
		t.Errorf("name = %q", got.Story.Name)
	}
}

func TestStale(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Nanosecond}
	b := &Bundle{Story: &shortcut.Story{ID: 1}}
	if err := c.Put("b", b); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Nanosecond)
	got, fresh, err := c.Get("b")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("got nil")
	}
	if fresh {
		t.Errorf("expected stale")
	}
}

func TestMissing(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour}
	got, fresh, err := c.Get("nope")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil || fresh {
		t.Errorf("expected miss")
	}
}

func TestDelete(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour}
	if err := c.Put("b", &Bundle{Story: &shortcut.Story{ID: 1}}); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete("b"); err != nil {
		t.Fatal(err)
	}
	got, _, _ := c.Get("b")
	if got != nil {
		t.Errorf("expected gone")
	}
	// Delete on missing is fine.
	if err := c.Delete("never"); err != nil {
		t.Errorf("delete missing: %v", err)
	}
}

func TestWorkflowStates(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour, WorkflowTTL: time.Hour}
	s := &WorkflowStates{
		Story: map[int]StateInfo{
			1: {Name: "Backlog", Type: "backlog"},
			2: {Name: "Done", Type: "done"},
		},
		Epic: map[int]StateInfo{
			10: {Name: "Not Started", Type: "unstarted"},
		},
	}
	if err := c.PutWorkflowStates(s); err != nil {
		t.Fatal(err)
	}
	got, fresh, err := c.GetWorkflowStates()
	if err != nil {
		t.Fatal(err)
	}
	if !fresh {
		t.Errorf("expected fresh")
	}
	if got.Story[1].Name != "Backlog" || got.Story[1].Type != "backlog" {
		t.Errorf("story[1] = %+v", got.Story[1])
	}
	if got.Epic[10].Name != "Not Started" || got.Epic[10].Type != "unstarted" {
		t.Errorf("epic[10] = %+v", got.Epic[10])
	}
}

func TestWorkflowStatesStale(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour, WorkflowTTL: time.Nanosecond}
	s := &WorkflowStates{Story: map[int]StateInfo{1: {Name: "x"}}}
	if err := c.PutWorkflowStates(s); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Nanosecond)
	got, fresh, err := c.GetWorkflowStates()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || fresh {
		t.Errorf("expected stale present, got=%v fresh=%v", got, fresh)
	}
}

func TestPutStampsSchemaVersion(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour}
	b := &Bundle{Story: &shortcut.Story{ID: 1}}
	if b.SchemaVersion != 0 {
		t.Fatalf("precondition: unstamped bundle should be version 0")
	}
	if err := c.Put("b", b); err != nil {
		t.Fatal(err)
	}
	if b.SchemaVersion != BundleSchemaVersion {
		t.Errorf("Put should stamp SchemaVersion; got %d want %d", b.SchemaVersion, BundleSchemaVersion)
	}
	got, _, err := c.Get("b")
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != BundleSchemaVersion {
		t.Errorf("Get returned SchemaVersion=%d, want %d", got.SchemaVersion, BundleSchemaVersion)
	}
}

func TestMembers(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour, WorkflowTTL: time.Hour}
	m := &Members{Members: map[string]MemberInfo{
		"u1": {MentionName: "alice", Name: "Alice"},
	}}
	if err := c.PutMembers(m); err != nil {
		t.Fatal(err)
	}
	got, fresh, err := c.GetMembers()
	if err != nil {
		t.Fatal(err)
	}
	if !fresh || got == nil || got.Members["u1"].MentionName != "alice" {
		t.Errorf("got=%+v fresh=%v", got, fresh)
	}
}

func TestGroups(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour, WorkflowTTL: time.Hour}
	g := &Groups{Groups: map[string]GroupInfo{
		"g1": {MentionName: "platform", Name: "Platform"},
	}}
	if err := c.PutGroups(g); err != nil {
		t.Fatal(err)
	}
	got, fresh, err := c.GetGroups()
	if err != nil {
		t.Fatal(err)
	}
	if !fresh || got == nil || got.Groups["g1"].MentionName != "platform" {
		t.Errorf("got=%+v fresh=%v", got, fresh)
	}
}

func TestMembersStale(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour, WorkflowTTL: time.Nanosecond}
	if err := c.PutMembers(&Members{Members: map[string]MemberInfo{"u1": {Name: "X"}}}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Nanosecond)
	got, fresh, err := c.GetMembers()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || fresh {
		t.Errorf("expected stale present, got=%v fresh=%v", got, fresh)
	}
}

func TestSlashesInBranchName(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour}
	branch := "feature/sc-12345/with-extra-slashes"
	if err := c.Put(branch, &Bundle{Story: &shortcut.Story{ID: 12345}}); err != nil {
		t.Fatal(err)
	}
	got, fresh, err := c.Get(branch)
	if err != nil {
		t.Fatal(err)
	}
	if !fresh || got == nil {
		t.Fatalf("got=%v fresh=%v", got, fresh)
	}
}
