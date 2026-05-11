package color

import "testing"

func TestForStateType(t *testing.T) {
	cases := map[string]string{
		"done":      Green,
		"started":   Purple,
		"unstarted": Gray,
		"backlog":   Gray,
		"":          "",
		"weird":     "",
	}
	for in, want := range cases {
		if got := ForStateType(in); got != want {
			t.Errorf("ForStateType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestForObjectiveState(t *testing.T) {
	cases := map[string]string{
		"done":        Green,
		"in progress": Purple,
		"to do":       Gray,
		"":            "",
	}
	for in, want := range cases {
		if got := ForObjectiveState(in); got != want {
			t.Errorf("ForObjectiveState(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWrap(t *testing.T) {
	if got := Wrap("hello", Green); got != "\x1b[32mhello\x1b[0m" {
		t.Errorf("got %q", got)
	}
	if got := Wrap("hello", ""); got != "hello" {
		t.Errorf("got %q (empty code should pass through)", got)
	}
	if got := Wrap("", Green); got != "" {
		t.Errorf("got %q (empty text should pass through)", got)
	}
}

func TestEnabled(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("SHORTCUT_STATUSLINE_NO_COLOR", "")
	if !Enabled() {
		t.Errorf("expected default-on")
	}

	t.Setenv("NO_COLOR", "1")
	if Enabled() {
		t.Errorf("NO_COLOR=1 should disable")
	}

	t.Setenv("NO_COLOR", "")
	t.Setenv("SHORTCUT_STATUSLINE_NO_COLOR", "true")
	if Enabled() {
		t.Errorf("SHORTCUT_STATUSLINE_NO_COLOR=true should disable")
	}
}
