package glob

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dlclark/regexp2"

	"github.com/pachyderm/ohmyglob/compiler"
	"github.com/pachyderm/ohmyglob/syntax"
)

// Glob represents compiled glob pattern.
type Glob struct {
	r *regexp2.Regexp
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
	r, err := regexp2.Compile(regex, 0)
	r.MatchTimeout = time.Minute * 5 // if it takes more than 5minutes to match a glob, something is very wrong
	if err != nil {
		return nil, err
	}
	return &Glob{r: r}, nil
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
	m, err := g.r.MatchString(fixture)
	if err != nil {
		// this is taking longer than 5 minutes, so something is seriously wrong
		panic(err)
	}
	return m
}

// Capture returns the list of subexpressions captured while testing the fixture against the compiled pattern
func (g *Glob) Capture(fixture string) []string {
	m, err := g.r.FindStringMatch(fixture)
	if err != nil {
		// this is taking longer than 5 minutes, so something is seriously wrong
		panic(err)
	}
	if m == nil {
		return nil
	}
	groups := m.Groups()
	captures := make([]string, 0, len(groups))

	for _, gp := range groups {
		captures = append(captures, gp.Capture.String())
	}
	return captures
}

// The code for the extract function is based on the extract function from https://golang.org/src/regexp/regexp.go
// Additionally, the Replace function is based on the expand function from https://golang.org/src/regexp/regexp.go
func extract(str string) (int, string, bool) {
	var name, rest string
	var num int
	if len(str) < 2 || str[0] != '$' {
		return num, rest, false
	}
	brace := false
	if str[1] == '{' {
		brace = true
		str = str[2:]
	} else {
		str = str[1:]
	}
	i := 0
	for i < len(str) {
		rune, size := utf8.DecodeRuneInString(str[i:])
		if !unicode.IsDigit(rune) {
			break
		}
		i += size
	}
	if i == 0 {
		// empty name is not okay
		return num, rest, false
	}
	name = str[:i]
	if brace {
		if i >= len(str) || str[i] != '}' {
			// missing closing brace
			return num, rest, false
		}
		i++
	}

	// Parse number.
	num = 0
	for i := 0; i < len(name); i++ {
		if name[i] < '0' || '9' < name[i] || num >= 1e8 {
			num = -1
			break
		}
		num = num*10 + int(name[i]) - '0'
	}
	// Disallow leading zeros.
	if name[0] == '0' && len(name) > 1 {
		num = -1
	}

	rest = str[i:]
	return num, rest, true
}

// Replace runs the compiled pattern on the given fixture, and then replaces any instance of $n (or ${n}) in the template with the nth capture group
func (g *Glob) Replace(fixture, template string) string {
	result := ""
	match := g.Capture(fixture)
	for len(template) > 0 {
		i := strings.Index(template, "$")
		if i < 0 {
			break
		}
		result += template[:i]
		template = template[i:]
		if len(template) > 1 && template[1] == '$' {
			// Treat $$ as $.
			result += "$"
			template = template[2:]
			continue
		}
		num, rest, ok := extract(template)

		if !ok {
			// Malformed; treat $ as raw text.
			result += "$"
			template = template[1:]
			continue
		}
		template = rest
		if num < len(match) {
			result += match[num]
		}
	}
	result += template
	return result
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
