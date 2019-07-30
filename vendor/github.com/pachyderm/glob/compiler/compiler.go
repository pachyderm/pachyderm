package compiler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pachyderm/glob/syntax/ast"
)

func compileChildren(tree *ast.Node, sep []rune, concatWith string) (string, error) {
	childRegex := make([]string, 0)
	for _, desc := range tree.Children {
		cr, err := compile(desc, sep)
		if err != nil {
			return "", err
		}
		childRegex = append(childRegex, cr)
	}
	return strings.Join(childRegex, concatWith), nil
}

func compile(tree *ast.Node, sep []rune) (string, error) {
	var err error
	regex := ""
	meta := regexp.QuoteMeta
	switch tree.Kind {
	// stuff between braces becomes a non-capturing group OR'd together
	case ast.KindAnyOf:
		if len(tree.Children) == 0 {
			return "", nil
		}
		anyOfRegex, err := compileChildren(tree, sep, "|")
		if err != nil {
			return "", err
		}
		return "(?:" + anyOfRegex + ")", nil

	// subexpresions are simply concatenated
	case ast.KindPattern:
		if len(tree.Children) == 0 {
			return "", nil
		}
		regex, err = compileChildren(tree, sep, "")
		if err != nil {
			return "", err
		}

	// capture groups become capture groups, with the stuff in them OR'd together
	case ast.KindCapture:
		if len(tree.Children) == 0 {
			return "", nil
		}
		c := tree.Value.(ast.Capture)
		captureRegex, err := compileChildren(tree, sep, "|")
		if err != nil {
			return "", err
		}
		captureRegex = "((?:" + captureRegex + ")"
		switch c.Quantifier {
		case "*":
			return captureRegex + "*)", nil
		case "?":
			return captureRegex + "?)", nil
		case "+":
			return captureRegex + "+)", nil
		case "@":
			return captureRegex + ")", nil
		case "!":
			// not implemented -- would require a non-regular expression
			// a future implementation might switch to using PCRE in place of Go regexp here
		}

		return "", fmt.Errorf("unimplemented quatifier %v", c.Quantifier)

	// glob `*` essentially becomes `.*`, but excluding any separators
	case ast.KindAny:
		if len(sep) == 0 {
			regex = ".*"
		} else {
			regex = fmt.Sprintf("[^%v]*", meta(string(sep)))
		}
	// glob `**` is just `.*`
	case ast.KindSuper:
		regex = ".*"

	// glob `?` essentially becomes `.`, but excluding any separators
	case ast.KindSingle:
		if len(sep) == 0 {
			regex = "."
		} else {
			regex = fmt.Sprintf("[^%v]", meta(string(sep)))
		}

	case ast.KindNothing:
		regex = ""

	// stuff in a list e.g. `[abcd]` is handled the same way by regexp
	case ast.KindList:
		l := tree.Value.(ast.List)
		sign := ""
		if l.Not {
			sign = "^"
		}
		regex = fmt.Sprintf("[%v%v]", sign, meta(string(l.Chars)))

	// POSIX classes like `[[:alpha:]]` are handled the same way by regexp
	case ast.KindPOSIX:
		p := tree.Value.(ast.POSIX)
		sign := ""
		if p.Not {
			sign = "^"
		}
		regex = fmt.Sprintf("[%v[:%v:]]", sign, meta(string(p.Class)))

	// stuff in a range e.g. `[a-d]` is handled the same way by regexp
	case ast.KindRange:
		r := tree.Value.(ast.Range)
		sign := ""
		if r.Not {
			sign = "^"
		}
		regex = fmt.Sprintf("[%v%v-%v]", sign, meta(string(r.Lo)), meta(string(r.Hi)))

	// text just matches text, after we escape any special regexp chars
	case ast.KindText:
		t := tree.Value.(ast.Text)
		regex = meta(t.Text)

	default:
		return "", fmt.Errorf("could not compile tree: unknown node type")
	}

	return regex, nil
}

// Compile takes a glob AST, and converts it into a regular expression
// Any separator characters (typically the path directory char: `/` or `\`)
// are passed in to allow the compiler to handle them correctly
func Compile(tree *ast.Node, sep []rune) (string, error) {
	regex, err := compile(tree, sep)
	if err != nil {
		return "", err
	}
	// globs are expected to match against the whole thing
	return "\\A" + regex + "\\z", nil
}
