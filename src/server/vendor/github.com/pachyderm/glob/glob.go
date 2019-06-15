package glob

import (
	"fmt"

	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"

	"github.com/pachyderm/glob/compiler"
	"github.com/pachyderm/glob/syntax"
)

// Glob represents compiled glob pattern.
type Glob struct {
	p *pcre.Regexp
}

// Compile creates Glob for given pattern and strings (if any present after pattern) as separators.
// The pattern syntax is:
//
//    pattern:
//        { term }
//
//    term:
//        `*`         matches any sequence of non-separator characters
//        `**`        matches any sequence of characters
//        `?`         matches any single non-separator character
//        `[` [ `!` ] { character-range } `]`
//                    character class (must be non-empty)
//        `{` pattern-list `}`
//                    pattern alternatives
//        c           matches character c (c != `*`, `**`, `?`, `\`, `[`, `{`, `}`)
//        `\` c       matches character c
//
//    character-range:
//        c           matches character c (c != `\\`, `-`, `]`)
//        `\` c       matches character c
//        lo `-` hi   matches character c for lo <= c <= hi
//
//    pattern-list:
//        pattern { `,` pattern }
//                    comma-separated (without spaces) patterns
//
//    extended-glob:
//        `(` { `|` pattern } `)`
//        `@(` { `|` pattern } `)`
//                    match and capture one of pipe-separated subpatterns
//        `*(` { `|` pattern } `)`
//                    match and capture any number of the pipe-separated subpatterns
//        `+(` { `|` pattern } `)`
//                    match and capture one or more of the pipe-separated subpatterns
//        `?(` { `|` pattern } `)`
//                    match and capture zero or one of the pipe-separated subpatterns
//        `!(` { `|` pattern } `)`
//                    match and capture anything except one of the pipe-separated subpatterns
//
func Compile(pattern string, separators ...rune) (*Glob, error) {
	tree, err := syntax.Parse(pattern)
	if err != nil {
		return nil, err
	}

	regex, err := compiler.Compile(tree, separators)
	if err != nil {
		return nil, err
	}
	p, pcreErr := pcre.Compile(regex, 0)
	if pcreErr != nil {
		return nil, fmt.Errorf(pcreErr.String())
	}
	return &Glob{p: &p}, nil
}

// MustCompile is the same as Compile, except that if Compile returns error, this will panic
func MustCompile(pattern string, separators ...rune) *Glob {
	g, err := Compile(pattern, separators...)
	if err != nil {
		panic(err)
	}
	return g
}

// Match tests the fixture against the compiled pattern, and return true for a match
func (g *Glob) Match(fixture string) bool {
	m := g.p.MatcherString(fixture, 0)
	return m.MatchString(fixture, 0)
}

// Capture returns the list of subexpressions captured while testing the fixture against the compiled pattern
func (g *Glob) Capture(fixture string) []string {
	m := g.p.MatcherString(fixture, 0)
	num := m.Groups()
	groups := make([]string, 0, num)
	if m.MatchString(fixture, 0) {
		for i := 0; i <= num; i++ {
			groups = append(groups, m.GroupString(i))
		}
	}
	return groups
}

// QuoteMeta returns a string that quotes all glob pattern meta characters
// inside the argument text; For example, QuoteMeta(`*(foo*)`) returns `\*\(foo\*\)`.
func QuoteMeta(s string) string {
	b := make([]byte, 2*len(s))

	// a byte loop is correct because all meta characters are ASCII
	j := 0
	for i := 0; i < len(s); i++ {
		if syntax.Special(s[i]) {
			b[j] = '\\'
			j++
		}
		b[j] = s[i]
		j++
	}

	return string(b[0:j])
}
