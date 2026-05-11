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

func main() {
	formatFlag := flag.String("format", "{story.name}", "Format string. Tokens: {story.name|id|url}, {epic.name|id|url}, {objective.name|id|url}")
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
			if hasNeededData(b, wantEpic, wantObj) {
				bundle = b
			}
		}
	}

	if bundle == nil {
		fetched, fetchErr := fetchBundle(storyID, wantEpic, wantObj)
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

func hasNeededData(b *cache.Bundle, wantEpic, wantObj bool) bool {
	if wantEpic && b.Story != nil && b.Story.EpicID != nil && b.Epic == nil {
		return false
	}
	if wantObj && b.Epic != nil && b.Epic.MilestoneID != nil && b.Objective == nil {
		return false
	}
	return true
}

func fetchBundle(storyID int, wantEpic, wantObj bool) (*cache.Bundle, error) {
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

	return b, nil
}

func makeResolver(b *cache.Bundle) format.Resolver {
	return func(ns, field string) (string, error) {
		switch ns {
		case format.NSStory:
			return storyField(b.Story, field)
		case format.NSEpic:
			return epicField(b.Epic, field)
		case format.NSObjective:
			return objectiveField(b.Objective, field)
		}
		return "", errors.New("unknown namespace " + ns)
	}
}

func storyField(s *shortcut.Story, field string) (string, error) {
	if s == nil {
		return "", nil
	}
	switch field {
	case "name":
		return s.Name, nil
	case "id":
		return strconv.Itoa(s.ID), nil
	case "url":
		return s.AppURL, nil
	}
	return "", fmt.Errorf("unknown field story.%s", field)
}

func epicField(e *shortcut.Epic, field string) (string, error) {
	if e == nil {
		return "", nil
	}
	switch field {
	case "name":
		return e.Name, nil
	case "id":
		return strconv.Itoa(e.ID), nil
	case "url":
		return e.AppURL, nil
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
