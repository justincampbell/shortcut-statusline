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
		Story: map[int]string{1: "Backlog", 2: "Done"},
		Epic:  map[int]string{10: "Not Started"},
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
	if got.Story[1] != "Backlog" || got.Epic[10] != "Not Started" {
		t.Errorf("got %+v", got)
	}
}

func TestWorkflowStatesStale(t *testing.T) {
	c := &Cache{Dir: t.TempDir(), TTL: time.Hour, WorkflowTTL: time.Nanosecond}
	if err := c.PutWorkflowStates(&WorkflowStates{Story: map[int]string{1: "x"}}); err != nil {
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
