// Package osc8 emits OSC8 hyperlink escape sequences so terminals that
// support them render text as clickable links.
//
// OSC8 spec: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
//
// We default to emitting OSC8 unconditionally. The escape sequence is
// well-behaved: terminals that don't support it consume and discard the
// framing, leaving the visible text intact. Auto-detection isn't reliable
// from our vantage point anyway — we're often invoked several layers deep
// (tmux → Claude Code statusline → ccstatusline → shortcut-statusline),
// and TERM_PROGRAM rarely survives that chain. Users on a hostile terminal
// can opt out with SHORTCUT_STATUSLINE_NO_LINKS=1 or --no-links.
package osc8

import "os"

// Wrap returns text wrapped in an OSC8 hyperlink pointing at url. If url is
// empty, text is returned unchanged.
func Wrap(text, url string) string {
	if url == "" {
		return text
	}
	return "\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\"
}

// Enabled reports whether OSC8 hyperlinks should be emitted. On by default.
// Forced off by either:
//   - NO_COLOR set to any non-empty value (https://no-color.org/). The spec is
//     scoped to ANSI color, but OSC8 is an ANSI escape too and most tools
//     interpret NO_COLOR as "suppress all decorative escapes."
//   - SHORTCUT_STATUSLINE_NO_LINKS=1/true/yes/on (tool-specific opt-out for
//     users who want colors but not hyperlinks).
func Enabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return !truthy(os.Getenv("SHORTCUT_STATUSLINE_NO_LINKS"))
}

func truthy(s string) bool {
	switch s {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
