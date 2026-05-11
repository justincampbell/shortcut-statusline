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
	"github.com/justincampbell/shortcut-statusline/internal/config"
	"github.com/justincampbell/shortcut-statusline/internal/format"
	"github.com/justincampbell/shortcut-statusline/internal/osc8"
	"github.com/justincampbell/shortcut-statusline/internal/shortcut"
)

var version = "dev"

const formatHelp = "Format string. Tokens: {story.name|id|url|state}, {epic.name|id|url|state}, {objective.name|id|url|state}"

func main() {
	const defaultFormat = "{story.id}: {story.name} ({epic.name})"
	formatFlag := flag.String("format", defaultFormat, formatHelp)
	flag.StringVar(formatFlag, "f", defaultFormat, "Format string (shorthand)")
	noCacheFlag := flag.Bool("no-cache", false, "Bypass the on-disk cache")
	refreshFlag := flag.Bool("refresh", false, "Clear cache for the current branch and refetch")
	noLinksFlag := flag.Bool("no-links", false, "Disable OSC8 hyperlinks regardless of terminal detection")
	versionFlag := flag.Bool("version", false, "Show version")
	flag.BoolVar(versionFlag, "v", false, "Show version (shorthand)")
	flag.Parse()

	if *versionFlag {
		fmt.Println("shortcut-statusline", version)
		return
	}

	links := osc8.Enabled() && !*noLinksFlag
	if err := run(*formatFlag, *noCacheFlag, *refreshFlag, links); err != nil {
		// Quiet failure: log to stderr but exit 0 so the statusline keeps moving.
		fmt.Fprintln(os.Stderr, "shortcut-statusline:", err)
	}
}

func run(formatStr string, noCache, refresh, links bool) error {
	br, err := branch.Current()
	if err != nil {
		// Not in a repo, detached HEAD, etc. — print nothing, succeed.
		return nil
	}

	storyID, ok := branch.StoryID(br)
	if !ok {
		return nil
	}

	namespaces := format.Namespaces(formatStr)
	wantEpic := namespaces[format.NSEpic] || namespaces[format.NSObjective]
	wantObj := namespaces[format.NSObjective]
	wantStoryState := format.HasField(formatStr, format.NSStory, "state")
	wantEpicState := format.HasField(formatStr, format.NSEpic, "state")
	if wantEpicState {
		wantEpic = true
	}

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
			if hasNeededData(b, wantEpic, wantObj, wantStoryState, wantEpicState) {
				bundle = b
			}
		}
	}

	if bundle == nil {
		fetched, fetchErr := fetchBundle(c, storyID, wantEpic, wantObj, wantStoryState, wantEpicState, noCache, refresh)
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

	out, err := format.Render(formatStr, makeResolver(bundle, links))
	if err != nil {
		return err
	}
	out = format.CollapseWhitespace(out)
	if out != "" {
		fmt.Println(out)
	}
	return nil
}

func hasNeededData(b *cache.Bundle, wantEpic, wantObj, wantStoryState, wantEpicState bool) bool {
	if wantEpic && b.Story != nil && b.Story.EpicID != nil && b.Epic == nil {
		return false
	}
	if wantObj && b.Epic != nil && b.Epic.MilestoneID != nil && b.Objective == nil {
		return false
	}
	if wantStoryState && b.StoryState == "" {
		return false
	}
	if wantEpicState && b.Epic != nil && b.EpicState == "" {
		return false
	}
	return true
}

func fetchBundle(c *cache.Cache, storyID int, wantEpic, wantObj, wantStoryState, wantEpicState, noCache, refresh bool) (*cache.Bundle, error) {
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

	if wantEpic && story.EpicID != nil {
		epic, err := client.GetEpic(ctx, *story.EpicID)
		if err != nil {
			return nil, err
		}
		b.Epic = epic

		if wantObj && epic.MilestoneID != nil {
			obj, err := client.GetObjective(ctx, *epic.MilestoneID)
			if err != nil {
				return nil, err
			}
			b.Objective = obj
		}
	}

	if wantStoryState || wantEpicState {
		states, err := loadWorkflowStates(ctx, c, client, noCache || refresh)
		if err != nil {
			return nil, err
		}
		if wantStoryState && b.Story != nil {
			b.StoryState = states.Story[b.Story.WorkflowStateID]
		}
		if wantEpicState && b.Epic != nil {
			b.EpicState = states.Epic[b.Epic.EpicStateID]
		}
	}

	return b, nil
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
		Story: map[int]string{},
		Epic:  map[int]string{},
	}
	for _, w := range wfs {
		for _, st := range w.States {
			s.Story[st.ID] = st.Name
		}
	}
	for _, st := range ew.EpicStates {
		s.Epic[st.ID] = st.Name
	}
	if err := c.PutWorkflowStates(s); err != nil {
		fmt.Fprintln(os.Stderr, "shortcut-statusline: workflow cache write:", err)
	}
	return s, nil
}

func makeResolver(b *cache.Bundle, links bool) format.Resolver {
	return func(ns, field string) (string, error) {
		switch ns {
		case format.NSStory:
			return storyField(b, field, links)
		case format.NSEpic:
			return epicField(b, field, links)
		case format.NSObjective:
			return objectiveField(b.Objective, field, links)
		}
		return "", errors.New("unknown namespace " + ns)
	}
}

func storyField(b *cache.Bundle, field string, links bool) (string, error) {
	if b.Story == nil {
		return "", nil
	}
	s := b.Story
	switch field {
	case "name":
		return link(s.Name, s.AppURL, links), nil
	case "id":
		return link(strconv.Itoa(s.ID), s.AppURL, links), nil
	case "url":
		return s.AppURL, nil
	case "state":
		return b.StoryState, nil
	}
	return "", fmt.Errorf("unknown field story.%s", field)
}

func epicField(b *cache.Bundle, field string, links bool) (string, error) {
	if b.Epic == nil {
		return "", nil
	}
	e := b.Epic
	switch field {
	case "name":
		return link(e.Name, e.AppURL, links), nil
	case "id":
		return link(strconv.Itoa(e.ID), e.AppURL, links), nil
	case "url":
		return e.AppURL, nil
	case "state":
		return b.EpicState, nil
	}
	return "", fmt.Errorf("unknown field epic.%s", field)
}

func objectiveField(o *shortcut.Objective, field string, links bool) (string, error) {
	if o == nil {
		return "", nil
	}
	switch field {
	case "name":
		return link(o.Name, o.AppURL, links), nil
	case "id":
		return link(strconv.Itoa(o.ID), o.AppURL, links), nil
	case "url":
		return o.AppURL, nil
	case "state":
		return o.State, nil
	}
	return "", fmt.Errorf("unknown field objective.%s", field)
}

func link(text, url string, enabled bool) string {
	if !enabled {
		return text
	}
	return osc8.Wrap(text, url)
}
