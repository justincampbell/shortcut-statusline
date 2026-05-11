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
	"github.com/justincampbell/shortcut-statusline/internal/shortcut"
)

var version = "dev"

const formatHelp = "Format string. Tokens: {story.name|id|url|status}, {epic.name|id|url|status}, {objective.name|id|url}"

func main() {
	formatFlag := flag.String("format", "{story.name}", formatHelp)
	flag.StringVar(formatFlag, "f", "{story.name}", "Format string (shorthand)")
	noCacheFlag := flag.Bool("no-cache", false, "Bypass the on-disk cache")
	refreshFlag := flag.Bool("refresh", false, "Clear cache for the current branch and refetch")
	versionFlag := flag.Bool("version", false, "Show version")
	flag.BoolVar(versionFlag, "v", false, "Show version (shorthand)")
	flag.Parse()

	if *versionFlag {
		fmt.Println("shortcut-statusline", version)
		return
	}

	if err := run(*formatFlag, *noCacheFlag, *refreshFlag); err != nil {
		// Quiet failure: log to stderr but exit 0 so the statusline keeps moving.
		fmt.Fprintln(os.Stderr, "shortcut-statusline:", err)
	}
}

func run(formatStr string, noCache, refresh bool) error {
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
	wantStoryStatus := format.HasField(formatStr, format.NSStory, "status")
	wantEpicStatus := format.HasField(formatStr, format.NSEpic, "status")
	if wantEpicStatus {
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
			if hasNeededData(b, wantEpic, wantObj, wantStoryStatus, wantEpicStatus) {
				bundle = b
			}
		}
	}

	if bundle == nil {
		fetched, fetchErr := fetchBundle(c, storyID, wantEpic, wantObj, wantStoryStatus, wantEpicStatus, noCache, refresh)
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

	out, err := format.Render(formatStr, makeResolver(bundle))
	if err != nil {
		return err
	}
	out = format.CollapseWhitespace(out)
	if out != "" {
		fmt.Println(out)
	}
	return nil
}

func hasNeededData(b *cache.Bundle, wantEpic, wantObj, wantStoryStatus, wantEpicStatus bool) bool {
	if wantEpic && b.Story != nil && b.Story.EpicID != nil && b.Epic == nil {
		return false
	}
	if wantObj && b.Epic != nil && b.Epic.MilestoneID != nil && b.Objective == nil {
		return false
	}
	if wantStoryStatus && b.StoryStatus == "" {
		return false
	}
	if wantEpicStatus && b.Epic != nil && b.EpicStatus == "" {
		return false
	}
	return true
}

func fetchBundle(c *cache.Cache, storyID int, wantEpic, wantObj, wantStoryStatus, wantEpicStatus, noCache, refresh bool) (*cache.Bundle, error) {
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

	if wantStoryStatus || wantEpicStatus {
		states, err := loadWorkflowStates(ctx, c, client, noCache || refresh)
		if err != nil {
			return nil, err
		}
		if wantStoryStatus && b.Story != nil {
			b.StoryStatus = states.Story[b.Story.WorkflowStateID]
		}
		if wantEpicStatus && b.Epic != nil {
			b.EpicStatus = states.Epic[b.Epic.EpicStateID]
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

func makeResolver(b *cache.Bundle) format.Resolver {
	return func(ns, field string) (string, error) {
		switch ns {
		case format.NSStory:
			return storyField(b, field)
		case format.NSEpic:
			return epicField(b, field)
		case format.NSObjective:
			return objectiveField(b.Objective, field)
		}
		return "", errors.New("unknown namespace " + ns)
	}
}

func storyField(b *cache.Bundle, field string) (string, error) {
	if b.Story == nil {
		return "", nil
	}
	s := b.Story
	switch field {
	case "name":
		return s.Name, nil
	case "id":
		return strconv.Itoa(s.ID), nil
	case "url":
		return s.AppURL, nil
	case "status":
		return b.StoryStatus, nil
	}
	return "", fmt.Errorf("unknown field story.%s", field)
}

func epicField(b *cache.Bundle, field string) (string, error) {
	if b.Epic == nil {
		return "", nil
	}
	e := b.Epic
	switch field {
	case "name":
		return e.Name, nil
	case "id":
		return strconv.Itoa(e.ID), nil
	case "url":
		return e.AppURL, nil
	case "status":
		return b.EpicStatus, nil
	}
	return "", fmt.Errorf("unknown field epic.%s", field)
}

func objectiveField(o *shortcut.Objective, field string) (string, error) {
	if o == nil {
		return "", nil
	}
	switch field {
	case "name":
		return o.Name, nil
	case "id":
		return strconv.Itoa(o.ID), nil
	case "url":
		return o.AppURL, nil
	}
	return "", fmt.Errorf("unknown field objective.%s", field)
}
