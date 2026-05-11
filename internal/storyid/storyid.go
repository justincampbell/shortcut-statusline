// Package storyid resolves a Shortcut story ID for a given git branch.
//
// A Resolver returns (id, true, nil) for a match, (0, false, nil) for
// "no match", or an error for a lookup failure (e.g. API timeout). The
// Resolve orchestrator tries an ordered list of resolvers — first match
// wins. An error from one resolver doesn't stop iteration: later
// resolvers may still produce a match.
package storyid

import (
	"context"
	"fmt"
)

// Resolver looks up a Shortcut story ID from a branch name.
type Resolver interface {
	// Name identifies the resolver in diagnostics.
	Name() string
	// Resolve returns the matched story ID. ok=false with err=nil means
	// "no match — try the next resolver"; a non-nil error means the
	// lookup itself failed.
	Resolve(ctx context.Context, branch string) (id int, ok bool, err error)
}

// Resolve walks resolvers in order, returning the first match along
// with the matching resolver's Name. Errors are accumulated in errs so
// the caller can log them; iteration continues past errors.
func Resolve(ctx context.Context, branch string, resolvers []Resolver) (id int, source string, ok bool, errs []error) {
	for _, r := range resolvers {
		id, ok, err := r.Resolve(ctx, branch)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", r.Name(), err))
			continue
		}
		if ok {
			return id, r.Name(), true, errs
		}
	}
	return 0, "", false, errs
}
