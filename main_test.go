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
		Story:      &shortcut.Story{ID: 1, Name: "story", EpicID: &epicID, AppURL: "u1"},
		Epic:       &shortcut.Epic{ID: epicID, Name: "epic", MilestoneID: &milestoneID, AppURL: "u2"},
		Objective:  &shortcut.Objective{ID: milestoneID, Name: "obj", AppURL: "u3", State: "in progress"},
		StoryState: "In Development",
		EpicState:  "In Progress",
	}
	r := makeResolver(b, false, false)

	out, err := format.Render("{story.name} • {story.id} • {story.state} • {epic.name} • {epic.id} • {epic.state} • {objective.name} • {objective.state}", r)
	if err != nil {
		t.Fatal(err)
	}
	want := "story • 1 • In Development • epic • " + strconv.Itoa(epicID) + " • In Progress • obj • in progress"
	if out != want {
		t.Errorf("got %q want %q", out, want)
	}
}

func TestResolverIdName(t *testing.T) {
	epicID := 42
	milestoneID := 99
	b := &cache.Bundle{
		Story:     &shortcut.Story{ID: 1, Name: "story", EpicID: &epicID},
		Epic:      &shortcut.Epic{ID: epicID, Name: "epic", MilestoneID: &milestoneID},
		Objective: &shortcut.Objective{ID: milestoneID, Name: "obj", State: "done"},
	}
	r := makeResolver(b, false, false)
	out, err := format.Render("{story.idName} • {epic.idName} • {objective.idName}", r)
	if err != nil {
		t.Fatal(err)
	}
	want := "1: story • 42: epic • 99: obj"
	if out != want {
		t.Errorf("got %q want %q", out, want)
	}
}

func TestResolverMissingNamespaces(t *testing.T) {
	b := &cache.Bundle{Story: &shortcut.Story{Name: "story"}}
	r := makeResolver(b, false, false)
	out, err := format.Render("{story.name} {epic.name}", r)
	if err != nil {
		t.Fatal(err)
	}
	if out != "story " {
		t.Errorf("got %q", out)
	}
}

func TestResolverWithLinks(t *testing.T) {
	epicID := 42
	milestoneID := 99
	b := &cache.Bundle{
		Story:     &shortcut.Story{ID: 1, Name: "story", EpicID: &epicID, AppURL: "https://app.shortcut.com/x/story/1"},
		Epic:      &shortcut.Epic{ID: epicID, Name: "epic", MilestoneID: &milestoneID, AppURL: "https://app.shortcut.com/x/epic/42"},
		Objective: &shortcut.Objective{ID: milestoneID, Name: "obj", AppURL: "https://app.shortcut.com/x/objective/99", State: "in progress"},
	}
	r := makeResolver(b, true, false)

	out, err := format.Render("{story.name} {story.id} {epic.name} {objective.name}", r)
	if err != nil {
		t.Fatal(err)
	}
	want := "\x1b]8;;https://app.shortcut.com/x/story/1\x1b\\story\x1b]8;;\x1b\\ " +
		"\x1b]8;;https://app.shortcut.com/x/story/1\x1b\\1\x1b]8;;\x1b\\ " +
		"\x1b]8;;https://app.shortcut.com/x/epic/42\x1b\\epic\x1b]8;;\x1b\\ " +
		"\x1b]8;;https://app.shortcut.com/x/objective/99\x1b\\obj\x1b]8;;\x1b\\"
	if out != want {
		t.Errorf("\n got %q\nwant %q", out, want)
	}
}

func TestResolverLinksLeaveUrlAndStateUnwrapped(t *testing.T) {
	b := &cache.Bundle{
		Story:      &shortcut.Story{ID: 1, Name: "s", AppURL: "https://example.com/story/1"},
		StoryState: "In Development",
	}
	r := makeResolver(b, true, false)
	out, err := format.Render("{story.url}|{story.state}", r)
	if err != nil {
		t.Fatal(err)
	}
	if out != "https://example.com/story/1|In Development" {
		t.Errorf("got %q", out)
	}
}

func TestResolverWithColors(t *testing.T) {
	b := &cache.Bundle{
		Story:          &shortcut.Story{ID: 1, Name: "s", AppURL: ""},
		StoryState:     "In Development",
		StoryStateType: "started",
	}
	r := makeResolver(b, false, true)
	out, err := format.Render("{story.idName} {story.state}", r)
	if err != nil {
		t.Fatal(err)
	}
	want := "\x1b[35m1: s\x1b[0m \x1b[35mIn Development\x1b[0m"
	if out != want {
		t.Errorf("\n got %q\nwant %q", out, want)
	}
}

func TestResolverStoryType(t *testing.T) {
	b := &cache.Bundle{Story: &shortcut.Story{ID: 1, Name: "s", Type: "bug"}}
	out, err := format.Render("{story.type}", makeResolver(b, false, false))
	if err != nil {
		t.Fatal(err)
	}
	if out != "bug" {
		t.Errorf("got %q want %q", out, "bug")
	}
}

func TestResolverStoryTypeColors(t *testing.T) {
	cases := map[string]string{
		"bug":     "\x1b[31mbug\x1b[0m",
		"chore":   "\x1b[33mchore\x1b[0m",
		"feature": "\x1b[36mfeature\x1b[0m",
	}
	for typ, want := range cases {
		b := &cache.Bundle{Story: &shortcut.Story{ID: 1, Name: "s", Type: typ}}
		out, err := format.Render("{story.type}", makeResolver(b, false, true))
		if err != nil {
			t.Fatal(err)
		}
		if out != want {
			t.Errorf("type=%q: got %q want %q", typ, out, want)
		}
	}
}

func TestResolverObjectiveColor(t *testing.T) {
	b := &cache.Bundle{
		Objective: &shortcut.Objective{ID: 99, Name: "obj", State: "done"},
	}
	r := makeResolver(b, false, true)
	out, err := format.Render("{objective.name}", r)
	if err != nil {
		t.Fatal(err)
	}
	want := "\x1b[32mobj\x1b[0m"
	if out != want {
		t.Errorf("got %q want %q", out, want)
	}
}

func TestResolverColorWithLinks(t *testing.T) {
	b := &cache.Bundle{
		Story:          &shortcut.Story{ID: 1, Name: "s", AppURL: "https://x/1"},
		StoryStateType: "done",
	}
	r := makeResolver(b, true, true)
	out, err := format.Render("{story.name}", r)
	if err != nil {
		t.Fatal(err)
	}
	// Color wraps the OSC8-wrapped text — SGR outside, hyperlink inside.
	want := "\x1b[32m\x1b]8;;https://x/1\x1b\\s\x1b]8;;\x1b\\\x1b[0m"
	if out != want {
		t.Errorf("\n got %q\nwant %q", out, want)
	}
}

func TestHasNeededData(t *testing.T) {
	epicID := 1
	milestoneID := 2
	storyWithEpic := &shortcut.Story{EpicID: &epicID}
	epicWithMilestone := &shortcut.Epic{ID: epicID, MilestoneID: &milestoneID}
	cur := cache.BundleSchemaVersion

	cases := []struct {
		name string
		b    *cache.Bundle
		w    wants
		ok   bool
	}{
		{"story only, want story", &cache.Bundle{SchemaVersion: cur, Story: &shortcut.Story{}}, wants{}, true},
		{"want epic, missing", &cache.Bundle{SchemaVersion: cur, Story: storyWithEpic}, wants{Epic: true}, false},
		{"want epic, present", &cache.Bundle{SchemaVersion: cur, Story: storyWithEpic, Epic: epicWithMilestone}, wants{Epic: true}, true},
		{"want obj, missing", &cache.Bundle{SchemaVersion: cur, Story: storyWithEpic, Epic: epicWithMilestone}, wants{Epic: true, Obj: true}, false},
		{"want obj, present", &cache.Bundle{SchemaVersion: cur, Story: storyWithEpic, Epic: epicWithMilestone, Objective: &shortcut.Objective{}}, wants{Epic: true, Obj: true}, true},
		{"story has no epic, want epic — fine", &cache.Bundle{SchemaVersion: cur, Story: &shortcut.Story{}}, wants{Epic: true, Obj: true}, true},
		{"want story state, missing", &cache.Bundle{SchemaVersion: cur, Story: &shortcut.Story{}}, wants{StoryState: true}, false},
		{"want story state, name but no type (old field)", &cache.Bundle{SchemaVersion: cur, Story: &shortcut.Story{}, StoryState: "X"}, wants{StoryState: true}, false},
		{"want story state, present", &cache.Bundle{SchemaVersion: cur, Story: &shortcut.Story{}, StoryState: "X", StoryStateType: "started"}, wants{StoryState: true}, true},
		{"want epic state, missing", &cache.Bundle{SchemaVersion: cur, Story: storyWithEpic, Epic: epicWithMilestone}, wants{Epic: true, EpicState: true}, false},
		{"want epic state, name but no type (old field)", &cache.Bundle{SchemaVersion: cur, Story: storyWithEpic, Epic: epicWithMilestone, EpicState: "X"}, wants{Epic: true, EpicState: true}, false},
		{"want epic state, present", &cache.Bundle{SchemaVersion: cur, Story: storyWithEpic, Epic: epicWithMilestone, EpicState: "X", EpicStateType: "started"}, wants{Epic: true, EpicState: true}, true},
		{"pre-schema-v2 bundle is stale even when fully populated", &cache.Bundle{Story: &shortcut.Story{}}, wants{}, false},
	}
	for _, c := range cases {
		if got := hasNeededData(c.b, c.w); got != c.ok {
			t.Errorf("%s: got %v want %v", c.name, got, c.ok)
		}
	}
}

func TestWantsFor(t *testing.T) {
	cases := []struct {
		name   string
		fmt    string
		colors bool
		check  func(t *testing.T, w wants)
	}{
		{"story only, no colors", "{story.name}", false, func(t *testing.T, w wants) {
			if w.Epic || w.Obj || w.StoryState || w.StoryOwner || w.members() || w.groups() {
				t.Errorf("expected story-name only; got %+v", w)
			}
		}},
		{"colors infer story state", "{story.name}", true, func(t *testing.T, w wants) {
			if !w.StoryState {
				t.Errorf("colors should force StoryState; got %+v", w)
			}
		}},
		{"epic field promotes wantEpic", "{epic.owner}", false, func(t *testing.T, w wants) {
			if !w.Epic || !w.EpicOwner || !w.members() {
				t.Errorf("expected Epic+EpicOwner+members; got %+v", w)
			}
		}},
		{"story team needs groups", "{story.team}", false, func(t *testing.T, w wants) {
			if !w.StoryTeam || !w.groups() || w.members() {
				t.Errorf("expected groups only; got %+v", w)
			}
		}},
		{"requestor needs members", "{story.requestor}", false, func(t *testing.T, w wants) {
			if !w.StoryRequestor || !w.members() {
				t.Errorf("expected members; got %+v", w)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.check(t, wantsFor(c.fmt, c.colors))
		})
	}
}

func TestResolverOwnerTeamRequestor(t *testing.T) {
	b := &cache.Bundle{
		Story:          &shortcut.Story{ID: 1, Name: "s"},
		Epic:           &shortcut.Epic{ID: 2, Name: "e"},
		StoryOwner:     "alice",
		StoryRequestor: "bob",
		StoryTeam:      "platform",
		EpicOwner:      "carol",
		EpicTeam:       "growth",
	}
	out, err := format.Render(
		"{story.owner}|{story.requestor}|{story.team}|{epic.owner}|{epic.team}",
		makeResolver(b, false, false),
	)
	if err != nil {
		t.Fatal(err)
	}
	want := "alice|bob|platform|carol|growth"
	if out != want {
		t.Errorf("got %q want %q", out, want)
	}
}

func TestFirstOwnerNameAndLabelFallback(t *testing.T) {
	members := map[string]cache.MemberInfo{
		"id1": {MentionName: "alice", Name: "Alice"},
		"id2": {Name: "No Handle"}, // mention_name empty → fallback to name
	}
	if got := firstOwnerName(nil, members); got != "" {
		t.Errorf("nil ids should yield empty; got %q", got)
	}
	if got := firstOwnerName([]string{}, members); got != "" {
		t.Errorf("empty ids should yield empty; got %q", got)
	}
	if got := firstOwnerName([]string{"id1", "id2"}, members); got != "alice" {
		t.Errorf("want first owner alice; got %q", got)
	}
	if got := firstOwnerName([]string{"id2"}, members); got != "No Handle" {
		t.Errorf("want name fallback; got %q", got)
	}
	if got := firstOwnerName([]string{"missing"}, members); got != "" {
		t.Errorf("unknown id should yield empty; got %q", got)
	}
}

func TestGroupLabelFallback(t *testing.T) {
	if got := groupLabel(cache.GroupInfo{MentionName: "platform", Name: "Platform"}); got != "platform" {
		t.Errorf("mention_name preferred; got %q", got)
	}
	if got := groupLabel(cache.GroupInfo{Name: "Just Name"}); got != "Just Name" {
		t.Errorf("name fallback; got %q", got)
	}
	if got := groupLabel(cache.GroupInfo{}); got != "" {
		t.Errorf("empty info → empty; got %q", got)
	}
}
