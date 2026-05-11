package format

import (
	"fmt"
	"testing"
)

func TestNamespaces(t *testing.T) {
	got := Namespaces("{story.name} • {epic.name} - {story.id}")
	if !got["story"] || !got["epic"] {
		t.Errorf("got %v", got)
	}
	if got["objective"] {
		t.Errorf("unexpected objective")
	}
}

func TestRender(t *testing.T) {
	r := func(ns, field string) (string, error) {
		return ns + "." + field, nil
	}
	got, err := Render("hello {story.name} and {epic.id}", r)
	if err != nil {
		t.Fatal(err)
	}
	want := "hello story.name and epic.id"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestRenderUnknownReturnsEmpty(t *testing.T) {
	r := func(_, _ string) (string, error) { return "", nil }
	got, err := Render("a{story.name}b", r)
	if err != nil {
		t.Fatal(err)
	}
	if got != "ab" {
		t.Errorf("got %q", got)
	}
}

func TestRenderError(t *testing.T) {
	r := func(_, _ string) (string, error) { return "", fmt.Errorf("boom") }
	if _, err := Render("{story.name}", r); err == nil {
		t.Errorf("expected error")
	}
}

func TestCollapseWhitespace(t *testing.T) {
	if got := CollapseWhitespace("  a    b  "); got != "a b" {
		t.Errorf("got %q", got)
	}
}
