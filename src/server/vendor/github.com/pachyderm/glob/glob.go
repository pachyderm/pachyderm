package glob

import (
	"regexp"

	"github.com/pachyderm/glob/compiler"
	"github.com/pachyderm/glob/syntax"
)

// Glob represents compiled glob pattern.
type Glob struct {
	r *regexp.Regexp
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
//                    capture one of pipe-separated subpatterns
//        `*(` { `|` pattern } `)`
//                    capture any number of of pipe-separated subpatterns
//        `+(` { `|` pattern } `)`
//                    capture one or more of of pipe-separated subpatterns
//        `?(` { `|` pattern } `)`
//                    capture zero or one of of pipe-separated subpatterns
//
func Compile(pattern string, separators ...rune) (*Glob, error) {
	ast, err := syntax.Parse(pattern)
	if err != nil {
		return nil, err
	}

	regex, err := compiler.Compile(ast, separators)
	if err != nil {
		return nil, err
	}

	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	return &Glob{r}, nil
}

// MustCompile is the same as Compile, except that if Compile returns error, this will panic
func MustCompile(pattern string, separators ...rune) *Glob {
	g, err := Compile(pattern, separators...)
	if err != nil {
		panic(err)
	}
	return g
}

func (g *Glob) Match(fixture string) bool {
	return g.r.MatchString(fixture)
}

func (g *Glob) Capture(fixture string) []string {
	return g.r.FindStringSubmatch(fixture)
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
