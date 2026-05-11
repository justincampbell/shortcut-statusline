// Package format renders {namespace.field} tokens in a template string.
package format

import (
	"fmt"
	"regexp"
	"strings"
)

// Namespaces recognized today.
const (
	NSStory     = "story"
	NSEpic      = "epic"
	NSObjective = "objective"
)

// Resolver returns the value for a single {namespace.field} placeholder.
// Returning ("", nil) is treated as a deliberate empty value (e.g. epic not
// set). Returning a non-nil error fails the whole render.
type Resolver func(namespace, field string) (string, error)

var tokenRegex = regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\.([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// Token is a single {namespace.field} placeholder.
type Token struct {
	Namespace, Field string
}

// Tokens returns every {ns.field} token found in template, in order.
func Tokens(template string) []Token {
	matches := tokenRegex.FindAllStringSubmatch(template, -1)
	out := make([]Token, 0, len(matches))
	for _, m := range matches {
		out = append(out, Token{Namespace: m[1], Field: m[2]})
	}
	return out
}

// Namespaces returns the set of namespaces referenced in template.
func Namespaces(template string) map[string]bool {
	out := map[string]bool{}
	for _, t := range Tokens(template) {
		out[t.Namespace] = true
	}
	return out
}

// HasField reports whether template references {ns.field}.
func HasField(template, ns, field string) bool {
	for _, t := range Tokens(template) {
		if t.Namespace == ns && t.Field == field {
			return true
		}
	}
	return false
}

// Render replaces every {ns.field} token in template using resolver.
func Render(template string, resolver Resolver) (string, error) {
	var firstErr error
	result := tokenRegex.ReplaceAllStringFunc(template, func(match string) string {
		m := tokenRegex.FindStringSubmatch(match)
		ns, field := m[1], m[2]
		v, err := resolver(ns, field)
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("resolve {%s.%s}: %w", ns, field, err)
		}
		return v
	})
	if firstErr != nil {
		return "", firstErr
	}
	return result, nil
}

// CollapseWhitespace tidies up the rendered output: collapse runs of spaces
// and strip dangling separators left by empty values. Useful when an epic
// or objective is missing and the template had " • " between them.
func CollapseWhitespace(s string) string {
	s = strings.TrimSpace(s)
	// Multiple spaces → one.
	s = regexp.MustCompile(`[ \t]+`).ReplaceAllString(s, " ")
	return s
}
