package main

import (
	"strconv"
	"testing"

	"github.com/justincampbell/shortcut-statusline/internal/cache"
	"github.com/justincampbell/shortcut-statusline/internal/format"
	"github.com/justincampbell/shortcut-statusline/internal/shortcut"
)

func TestResolver(t *testing.T) {
	epicID := 42
	milestoneID := 99
	b := &cache.Bundle{
		Story:     &shortcut.Story{ID: 1, Name: "story", EpicID: &epicID, AppURL: "u1"},
		Epic:      &shortcut.Epic{ID: epicID, Name: "epic", MilestoneID: &milestoneID, AppURL: "u2"},
		Objective: &shortcut.Objective{ID: milestoneID, Name: "obj", AppURL: "u3"},
	}
	r := makeResolver(b)

	out, err := format.Render("{story.name} • {story.id} • {epic.name} • {epic.id} • {objective.name} • {objective.url}", r)
	if err != nil {
		t.Fatal(err)
	}
	want := "story • 1 • epic • " + strconv.Itoa(epicID) + " • obj • u3"
	if out != want {
		t.Errorf("got %q want %q", out, want)
	}
}

func TestResolverMissingNamespaces(t *testing.T) {
	b := &cache.Bundle{Story: &shortcut.Story{Name: "story"}}
	r := makeResolver(b)
	out, err := format.Render("{story.name} {epic.name}", r)
	if err != nil {
		t.Fatal(err)
	}
	if out != "story " {
		t.Errorf("got %q", out)
	}
}

func TestHasNeededData(t *testing.T) {
	epicID := 1
	milestoneID := 2
	storyWithEpic := &shortcut.Story{EpicID: &epicID}
	epicWithMilestone := &shortcut.Epic{ID: epicID, MilestoneID: &milestoneID}

	cases := []struct {
		name           string
		b              *cache.Bundle
		wantEpic, wantObj bool
		ok             bool
	}{
		{"story only, want story", &cache.Bundle{Story: &shortcut.Story{}}, false, false, true},
		{"want epic, missing", &cache.Bundle{Story: storyWithEpic}, true, false, false},
		{"want epic, present", &cache.Bundle{Story: storyWithEpic, Epic: epicWithMilestone}, true, false, true},
		{"want obj, missing", &cache.Bundle{Story: storyWithEpic, Epic: epicWithMilestone}, true, true, false},
		{"want obj, present", &cache.Bundle{Story: storyWithEpic, Epic: epicWithMilestone, Objective: &shortcut.Objective{}}, true, true, true},
		{"story has no epic, want epic — fine", &cache.Bundle{Story: &shortcut.Story{}}, true, true, true},
	}
	for _, c := range cases {
		if got := hasNeededData(c.b, c.wantEpic, c.wantObj); got != c.ok {
			t.Errorf("%s: got %v want %v", c.name, got, c.ok)
		}
	}
}
