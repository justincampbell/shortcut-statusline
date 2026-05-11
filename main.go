// shortcut-statusline prints info about the current Shortcut story for use
// in a shell or Claude Code statusline.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/justincampbell/shortcut-statusline/internal/branch"
	"github.com/justincampbell/shortcut-statusline/internal/cache"
	"github.com/justincampbell/shortcut-statusline/internal/color"
	"github.com/justincampbell/shortcut-statusline/internal/config"
	"github.com/justincampbell/shortcut-statusline/internal/format"
	"github.com/justincampbell/shortcut-statusline/internal/osc8"
	"github.com/justincampbell/shortcut-statusline/internal/shortcut"
)

var version = "dev"

const formatHelp = "Format string. Tokens: {story.name|id|idName|url|state|type|owner|requestor|team}, {epic.name|id|idName|url|state|owner|team}, {objective.name|id|idName|url|state}"

func main() {
	const defaultFormat = "{story.idName} ({epic.idName})"
	formatFlag := flag.String("format", defaultFormat, formatHelp)
	flag.StringVar(formatFlag, "f", defaultFormat, "Format string (shorthand)")
	noCacheFlag := flag.Bool("no-cache", false, "Bypass the on-disk cache")
	refreshFlag := flag.Bool("refresh", false, "Clear cache for the current branch and refetch")
	noLinksFlag := flag.Bool("no-links", false, "Disable OSC8 hyperlinks regardless of terminal detection")
	noColorFlag := flag.Bool("no-color", false, "Disable ANSI color regardless of environment")
	versionFlag := flag.Bool("version", false, "Show version")
	flag.BoolVar(versionFlag, "v", false, "Show version (shorthand)")
	flag.Parse()

	if *versionFlag {
		fmt.Println("shortcut-statusline", version)
		return
	}

	links := osc8.Enabled() && !*noLinksFlag
	colors := color.Enabled() && !*noColorFlag
	if err := run(*formatFlag, *noCacheFlag, *refreshFlag, links, colors); err != nil {
		// Quiet failure: log to stderr but exit 0 so the statusline keeps moving.
		fmt.Fprintln(os.Stderr, "shortcut-statusline:", err)
	}
}

func run(formatStr string, noCache, refresh, links, colors bool) error {
	br, err := branch.Current()
	if err != nil {
		// Not in a repo, detached HEAD, etc. — print nothing, succeed.
		return nil
	}

	storyID, ok := branch.StoryID(br)
	if !ok {
		return nil
	}

	w := wantsFor(formatStr, colors)

	c, err := cache.New()
	if err != nil {
		return fmt.Errorf("cache init: %w", err)
	}

	if refresh {
		_ = c.Delete(br)
	}

	var bundle *cache.Bundle
	if !noCache && !refresh {
		b, fresh, err := c.Get(br)
		if err == nil && fresh && b != nil && b.Story != nil && b.Story.ID == storyID {
			if hasNeededData(b, w) {
				bundle = b
			}
		}
	}

	if bundle == nil {
		fetched, fetchErr := fetchBundle(c, storyID, w, noCache, refresh)
		if fetchErr != nil {
			// On API failure, fall back to stale cache.
			if stale, _, err := c.Get(br); err == nil && stale != nil && stale.Story != nil && stale.Story.ID == storyID {
				bundle = stale
			} else {
				return fetchErr
			}
		} else {
			bundle = fetched
			if !noCache {
				if err := c.Put(br, bundle); err != nil {
					fmt.Fprintln(os.Stderr, "shortcut-statusline: cache write:", err)
				}
			}
		}
	}

	out, err := format.Render(formatStr, makeResolver(bundle, links, colors))
	if err != nil {
		return err
	}
	out = format.CollapseWhitespace(out)
	if out != "" {
		fmt.Println(out)
	}
	return nil
}

// wants captures which optional resources / derived fields the format
// string and color setting require. It's the single argument threaded
// through cache-freshness and fetch logic so neither call site has to
// juggle six positional booleans.
type wants struct {
	Epic, Obj                       bool
	StoryState, EpicState           bool
	StoryType                       bool
	StoryOwner, EpicOwner           bool
	StoryRequestor                  bool
	StoryTeam, EpicTeam             bool
}

func wantsFor(formatStr string, colors bool) wants {
	namespaces := format.Namespaces(formatStr)
	w := wants{
		Epic: namespaces[format.NSEpic] || namespaces[format.NSObjective],
		Obj:  namespaces[format.NSObjective],
		// Color wraps the namespace's name/id by its workflow state, so any
		// referenced namespace needs its state fetched even when {.state}
		// isn't in the format.
		StoryState:     format.HasField(formatStr, format.NSStory, "state") || (colors && namespaces[format.NSStory]),
		EpicState:      format.HasField(formatStr, format.NSEpic, "state") || (colors && namespaces[format.NSEpic]),
		StoryType:      format.HasField(formatStr, format.NSStory, "type"),
		StoryOwner:     format.HasField(formatStr, format.NSStory, "owner"),
		StoryRequestor: format.HasField(formatStr, format.NSStory, "requestor"),
		StoryTeam:      format.HasField(formatStr, format.NSStory, "team"),
		EpicOwner:      format.HasField(formatStr, format.NSEpic, "owner"),
		EpicTeam:       format.HasField(formatStr, format.NSEpic, "team"),
	}
	if w.EpicState || w.EpicOwner || w.EpicTeam {
		w.Epic = true
	}
	return w
}

func (w wants) members() bool { return w.StoryOwner || w.EpicOwner || w.StoryRequestor }
func (w wants) groups() bool  { return w.StoryTeam || w.EpicTeam }

func hasNeededData(b *cache.Bundle, w wants) bool {
	// A cached bundle written before a derived field was introduced will
	// still pass the per-field checks below but render the new field
	// empty. The schema version forces a refetch on upgrade.
	if b.SchemaVersion < cache.BundleSchemaVersion {
		return false
	}
	if w.Epic && b.Story != nil && b.Story.EpicID != nil && b.Epic == nil {
		return false
	}
	if w.Obj && b.Epic != nil && b.Epic.MilestoneID != nil && b.Objective == nil {
		return false
	}
	if w.StoryState && (b.StoryState == "" || b.StoryStateType == "") {
		return false
	}
	if w.EpicState && b.Epic != nil && (b.EpicState == "" || b.EpicStateType == "") {
		return false
	}
	return true
}

func fetchBundle(c *cache.Cache, storyID int, w wants, noCache, refresh bool) (*cache.Bundle, error) {
	token, err := config.Token()
	if err != nil {
		return nil, err
	}
	client := shortcut.New(token)
	ctx := context.Background()

	story, err := client.GetStory(ctx, storyID)
	if err != nil {
		return nil, err
	}
	b := &cache.Bundle{Story: story}

	if w.Epic && story.EpicID != nil {
		epic, err := client.GetEpic(ctx, *story.EpicID)
		if err != nil {
			return nil, err
		}
		b.Epic = epic

		if w.Obj && epic.MilestoneID != nil {
			obj, err := client.GetObjective(ctx, *epic.MilestoneID)
			if err != nil {
				return nil, err
			}
			b.Objective = obj
		}
	}

	force := noCache || refresh

	if w.StoryState || w.EpicState {
		states, err := loadWorkflowStates(ctx, c, client, force)
		if err != nil {
			return nil, err
		}
		if w.StoryState && b.Story != nil {
			info := states.Story[b.Story.WorkflowStateID]
			b.StoryState = info.Name
			b.StoryStateType = info.Type
		}
		if w.EpicState && b.Epic != nil {
			info := states.Epic[b.Epic.EpicStateID]
			b.EpicState = info.Name
			b.EpicStateType = info.Type
		}
	}

	if w.members() {
		members, err := loadMembers(ctx, c, client, force)
		if err != nil {
			return nil, err
		}
		if w.StoryOwner {
			b.StoryOwner = firstOwnerName(b.Story.OwnerIDs, members)
		}
		if w.StoryRequestor && b.Story.RequestedByID != "" {
			b.StoryRequestor = memberLabel(members[b.Story.RequestedByID])
		}
		if w.EpicOwner && b.Epic != nil {
			b.EpicOwner = firstOwnerName(b.Epic.OwnerIDs, members)
		}
	}

	if w.groups() {
		groups, err := loadGroups(ctx, c, client, force)
		if err != nil {
			return nil, err
		}
		if w.StoryTeam && b.Story.GroupID != nil {
			b.StoryTeam = groupLabel(groups[*b.Story.GroupID])
		}
		if w.EpicTeam && b.Epic != nil && b.Epic.GroupID != nil {
			b.EpicTeam = groupLabel(groups[*b.Epic.GroupID])
		}
	}

	return b, nil
}

func firstOwnerName(ids []string, members map[string]cache.MemberInfo) string {
	if len(ids) == 0 {
		return ""
	}
	return memberLabel(members[ids[0]])
}

func memberLabel(m cache.MemberInfo) string {
	if m.MentionName != "" {
		return m.MentionName
	}
	return m.Name
}

func groupLabel(g cache.GroupInfo) string {
	if g.MentionName != "" {
		return g.MentionName
	}
	return g.Name
}

// loadWorkflowStates returns the workflow + epic state lookup maps, using a
// long-TTL on-disk cache. Force=true skips the cache.
func loadWorkflowStates(ctx context.Context, c *cache.Cache, client *shortcut.Client, force bool) (*cache.WorkflowStates, error) {
	if !force {
		if s, fresh, err := c.GetWorkflowStates(); err == nil && fresh && s != nil {
			return s, nil
		}
	}

	wfs, err := client.GetWorkflows(ctx)
	if err != nil {
		return nil, err
	}
	ew, err := client.GetEpicWorkflow(ctx)
	if err != nil {
		return nil, err
	}

	s := &cache.WorkflowStates{
		Story: map[int]cache.StateInfo{},
		Epic:  map[int]cache.StateInfo{},
	}
	for _, w := range wfs {
		for _, st := range w.States {
			s.Story[st.ID] = cache.StateInfo{Name: st.Name, Type: st.Type}
		}
	}
	for _, st := range ew.EpicStates {
		s.Epic[st.ID] = cache.StateInfo{Name: st.Name, Type: st.Type}
	}
	if err := c.PutWorkflowStates(s); err != nil {
		fmt.Fprintln(os.Stderr, "shortcut-statusline: workflow cache write:", err)
	}
	return s, nil
}

// loadMembers returns the id→member lookup, using a long-TTL on-disk cache.
// Force=true skips the cache. Resolved label is just the cached MemberInfo;
// callers pick mention_name vs. name via memberLabel.
func loadMembers(ctx context.Context, c *cache.Cache, client *shortcut.Client, force bool) (map[string]cache.MemberInfo, error) {
	if !force {
		if m, fresh, err := c.GetMembers(); err == nil && fresh && m != nil {
			return m.Members, nil
		}
	}
	members, err := client.GetMembers(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]cache.MemberInfo, len(members))
	for _, m := range members {
		out[m.ID] = cache.MemberInfo{MentionName: m.Profile.MentionName, Name: m.Profile.Name}
	}
	if err := c.PutMembers(&cache.Members{Members: out}); err != nil {
		fmt.Fprintln(os.Stderr, "shortcut-statusline: members cache write:", err)
	}
	return out, nil
}

// loadGroups returns the id→group lookup, using a long-TTL on-disk cache.
// Force=true skips the cache.
func loadGroups(ctx context.Context, c *cache.Cache, client *shortcut.Client, force bool) (map[string]cache.GroupInfo, error) {
	if !force {
		if g, fresh, err := c.GetGroups(); err == nil && fresh && g != nil {
			return g.Groups, nil
		}
	}
	groups, err := client.GetGroups(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]cache.GroupInfo, len(groups))
	for _, g := range groups {
		out[g.ID] = cache.GroupInfo{MentionName: g.MentionName, Name: g.Name}
	}
	if err := c.PutGroups(&cache.Groups{Groups: out}); err != nil {
		fmt.Fprintln(os.Stderr, "shortcut-statusline: groups cache write:", err)
	}
	return out, nil
}

func makeResolver(b *cache.Bundle, links, colors bool) format.Resolver {
	return func(ns, field string) (string, error) {
		switch ns {
		case format.NSStory:
			return storyField(b, field, links, colors)
		case format.NSEpic:
			return epicField(b, field, links, colors)
		case format.NSObjective:
			return objectiveField(b.Objective, field, links, colors)
		}
		return "", errors.New("unknown namespace " + ns)
	}
}

func storyField(b *cache.Bundle, field string, links, colors bool) (string, error) {
	if b.Story == nil {
		return "", nil
	}
	s := b.Story
	c := ""
	if colors {
		c = color.ForStateType(b.StoryStateType)
	}
	switch field {
	case "name":
		return decorate(s.Name, s.AppURL, c, links), nil
	case "id":
		return decorate(strconv.Itoa(s.ID), s.AppURL, c, links), nil
	case "idName":
		return decorate(strconv.Itoa(s.ID)+": "+s.Name, s.AppURL, c, links), nil
	case "url":
		return s.AppURL, nil
	case "state":
		return color.Wrap(b.StoryState, c), nil
	case "type":
		tc := ""
		if colors {
			tc = color.ForStoryType(s.Type)
		}
		// API returns lowercase ("bug" / "chore" / "feature"); render
		// capitalized to match Shortcut's web UI.
		return color.Wrap(capitalizeASCII(s.Type), tc), nil
	case "owner":
		return b.StoryOwner, nil
	case "requestor":
		return b.StoryRequestor, nil
	case "team":
		return b.StoryTeam, nil
	}
	return "", fmt.Errorf("unknown field story.%s", field)
}

func epicField(b *cache.Bundle, field string, links, colors bool) (string, error) {
	if b.Epic == nil {
		return "", nil
	}
	e := b.Epic
	c := ""
	if colors {
		c = color.ForStateType(b.EpicStateType)
	}
	switch field {
	case "name":
		return decorate(e.Name, e.AppURL, c, links), nil
	case "id":
		return decorate(strconv.Itoa(e.ID), e.AppURL, c, links), nil
	case "idName":
		return decorate(strconv.Itoa(e.ID)+": "+e.Name, e.AppURL, c, links), nil
	case "url":
		return e.AppURL, nil
	case "state":
		return color.Wrap(b.EpicState, c), nil
	case "owner":
		return b.EpicOwner, nil
	case "team":
		return b.EpicTeam, nil
	}
	return "", fmt.Errorf("unknown field epic.%s", field)
}

func objectiveField(o *shortcut.Objective, field string, links, colors bool) (string, error) {
	if o == nil {
		return "", nil
	}
	c := ""
	if colors {
		c = color.ForObjectiveState(o.State)
	}
	switch field {
	case "name":
		return decorate(o.Name, o.AppURL, c, links), nil
	case "id":
		return decorate(strconv.Itoa(o.ID), o.AppURL, c, links), nil
	case "idName":
		return decorate(strconv.Itoa(o.ID)+": "+o.Name, o.AppURL, c, links), nil
	case "url":
		return o.AppURL, nil
	case "state":
		return color.Wrap(o.State, c), nil
	}
	return "", fmt.Errorf("unknown field objective.%s", field)
}

// decorate wraps text in an OSC8 link (if links enabled and url present),
// then in an SGR color (if colors enabled and code present). Color sits
// outside the hyperlink so terminals render the visible text colored.
func decorate(text, url, colorCode string, links bool) string {
	if links {
		text = osc8.Wrap(text, url)
	}
	return color.Wrap(text, colorCode)
}

// capitalizeASCII upper-cases the first byte. Sufficient for the
// fixed set of story_type values, which are all-ASCII.
func capitalizeASCII(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	if b[0] >= 'a' && b[0] <= 'z' {
		b[0] -= 'a' - 'A'
	}
	return string(b)
}
