package storyid

import (
	"context"
	"errors"
	"testing"
)

type stubResolver struct {
	name string
	id   int
	ok   bool
	err  error
}

func (s stubResolver) Name() string { return s.name }
func (s stubResolver) Resolve(_ context.Context, _ string) (int, bool, error) {
	return s.id, s.ok, s.err
}

func TestResolve_FirstMatchWins(t *testing.T) {
	rs := []Resolver{
		stubResolver{name: "a", id: 0, ok: false},
		stubResolver{name: "b", id: 99, ok: true},
		stubResolver{name: "c", id: 42, ok: true},
	}
	id, src, ok, errs := Resolve(context.Background(), "any", rs)
	if !ok || id != 99 || src != "b" || len(errs) != 0 {
		t.Errorf("id=%d src=%q ok=%v errs=%v", id, src, ok, errs)
	}
}

func TestResolve_AllMiss(t *testing.T) {
	rs := []Resolver{
		stubResolver{name: "a"},
		stubResolver{name: "b"},
	}
	id, src, ok, errs := Resolve(context.Background(), "any", rs)
	if ok || id != 0 || src != "" || len(errs) != 0 {
		t.Errorf("id=%d src=%q ok=%v errs=%v", id, src, ok, errs)
	}
}

func TestResolve_ErrorDoesNotStopIteration(t *testing.T) {
	rs := []Resolver{
		stubResolver{name: "a", err: errors.New("boom")},
		stubResolver{name: "b", id: 7, ok: true},
	}
	id, src, ok, errs := Resolve(context.Background(), "any", rs)
	if !ok || id != 7 || src != "b" {
		t.Errorf("id=%d src=%q ok=%v", id, src, ok)
	}
	if len(errs) != 1 || errs[0].Error() != "a: boom" {
		t.Errorf("errs=%v", errs)
	}
}

func TestBranchRegex(t *testing.T) {
	r := BranchRegex()
	if r.Name() != "branch-regex" {
		t.Errorf("name = %q", r.Name())
	}
	cases := map[string]struct {
		id int
		ok bool
	}{
		"feature/sc-12345-something":      {12345, true},
		"bug/SC-99-uppercase":             {99, true},
		"chore/sc-new-story/no-id":        {0, false},
		"main":                            {0, false},
	}
	for in, want := range cases {
		id, ok, err := r.Resolve(context.Background(), in)
		if err != nil {
			t.Errorf("%q: unexpected err %v", in, err)
		}
		if id != want.id || ok != want.ok {
			t.Errorf("%q: got id=%d ok=%v, want id=%d ok=%v", in, id, ok, want.id, want.ok)
		}
	}
}
