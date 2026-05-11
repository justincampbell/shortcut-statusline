package osc8

import "testing"

func TestWrap(t *testing.T) {
	got := Wrap("hello", "https://example.com")
	want := "\x1b]8;;https://example.com\x1b\\hello\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestWrapEmptyURL(t *testing.T) {
	if got := Wrap("hello", ""); got != "hello" {
		t.Errorf("got %q want %q", got, "hello")
	}
}

func TestEnabled(t *testing.T) {
	cases := []struct {
		name     string
		noLinks  string
		noColor  string
		want     bool
	}{
		{"unset defaults on", "", "", true},
		{"no-links=0 stays on", "0", "", true},
		{"no-links=false stays on", "false", "", true},
		{"no-links=1 forces off", "1", "", false},
		{"no-links=true forces off", "true", "", false},
		{"no-links=yes forces off", "yes", "", false},
		{"no-links=on forces off", "on", "", false},
		{"NO_COLOR=1 forces off", "", "1", false},
		{"NO_COLOR=anything forces off", "", "potato", false},
		{"NO_COLOR=0 forces off (per spec, presence matters)", "", "0", false},
		{"NO_COLOR empty stays on", "", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("SHORTCUT_STATUSLINE_NO_LINKS", c.noLinks)
			t.Setenv("NO_COLOR", c.noColor)
			if got := Enabled(); got != c.want {
				t.Errorf("got %v want %v", got, c.want)
			}
		})
	}
}
