// Package color emits ANSI SGR escape sequences to colorize statusline
// output by Shortcut workflow-state semantics (unstarted/started/done).
//
// We default to emitting color unconditionally and let users opt out, for
// the same reasons as the osc8 package: terminals that don't render SGR
// either strip it cleanly or display garbage we can't predict from inside
// a deep tmux/Claude Code pipeline.
package color

import "os"

// ANSI SGR codes for the buckets we use. State buckets reflect Shortcut's
// web UI roughly: gray for unstarted/backlog, magenta for in-progress, green
// for done. Story-type buckets pick distinct hues so {story.type} pops next
// to the state-colored name.
const (
	Reset  = "\x1b[0m"
	Gray   = "\x1b[90m"
	Red    = "\x1b[31m"
	Green  = "\x1b[32m"
	Yellow = "\x1b[33m"
	Purple = "\x1b[35m"
	Cyan   = "\x1b[36m"
)

// Wrap returns text wrapped in code…Reset. Empty code returns text unchanged.
func Wrap(text, code string) string {
	if code == "" || text == "" {
		return text
	}
	return code + text + Reset
}

// ForStateType maps a Shortcut workflow-state `type` to an ANSI color.
// Types: "backlog", "unstarted", "started", "done". Anything else → "".
func ForStateType(t string) string {
	switch t {
	case "done":
		return Green
	case "started":
		return Purple
	case "backlog", "unstarted":
		return Gray
	}
	return ""
}

// ForStoryType maps the Shortcut `story_type` field ("feature", "bug",
// "chore") to an ANSI color. Anything else → "".
func ForStoryType(t string) string {
	switch t {
	case "bug":
		return Red
	case "chore":
		return Yellow
	case "feature":
		return Cyan
	}
	return ""
}

// ForObjectiveState maps the objective `state` string (Shortcut returns
// "to do", "in progress", "done" directly) to an ANSI color.
func ForObjectiveState(s string) string {
	switch s {
	case "done":
		return Green
	case "in progress":
		return Purple
	case "to do":
		return Gray
	}
	return ""
}

// Enabled reports whether ANSI color should be emitted. On by default.
// Forced off by either:
//   - NO_COLOR set to any non-empty value (https://no-color.org/).
//   - SHORTCUT_STATUSLINE_NO_COLOR=1/true/yes/on.
func Enabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return !truthy(os.Getenv("SHORTCUT_STATUSLINE_NO_COLOR"))
}

func truthy(s string) bool {
	switch s {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
