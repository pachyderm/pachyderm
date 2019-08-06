package compiler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pachyderm/ohmyglob/syntax/ast"
)

// these dummy strings are not valid UTF-8, so we can use them without worrying about them also matching legitmate input
const (
	closeNegDummy = string(0xffff)
	boundaryDummy = string(0xfffe)
)

func dot(sep []rune) string {
	meta := regexp.QuoteMeta
	if len(sep) == 0 {
		return "."
	}
	return fmt.Sprintf("[^%v]", meta(string(sep)))
}

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
		var captureRegex string
		if c.Quantifier == "!" {
			// each subexpression in a negation needs a closeNegDummy marker
			captureRegex, err = compileChildren(tree, sep, closeNegDummy+"|")
		} else {
			captureRegex, err = compileChildren(tree, sep, "|")
		}
		if err != nil {
			return "", err
		}
		switch c.Quantifier {
		case "*":
			return "((?:" + captureRegex + ")*)", nil
		case "?":
			return "((?:" + captureRegex + ")?)", nil
		case "+":
			return "((?:" + captureRegex + ")+)", nil
		case "@":
			return "(" + captureRegex + ")", nil
		case "!":
			// not only does a negation capture require PCRE
			// it also requires a complicated function to determine the scope of what it should match
			// this requires global information, so we cannot do it here, instead we insert a `closeNegDummy`
			// to mark the spot where we might potentially need this
			return "((?:(?!(?:" + captureRegex + fmt.Sprintf("%v))%v*))", closeNegDummy, dot(sep)), nil
		}

		return "", fmt.Errorf("unimplemented quatifier %v", c.Quantifier)

	// glob `*` essentially becomes `.*`, but excluding any separators
	case ast.KindAny:
		// `*` is more aggressive than `!(...)`, so it bounds the scope of negation
		regex = dot(sep) + "*" + boundaryDummy

	// glob `**` is just `.*`
	case ast.KindSuper:
		// `**` is more aggressive than `!(...)`, so it bounds the scope of negation
		regex = ".*" + boundaryDummy

	// glob `?` essentially becomes `.`, but excluding any separators
	case ast.KindSingle:
		regex = dot(sep)

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
		// text is more aggressive than `!(...)`, so it bounds the scope of negation
		regex = meta(t.Text) + boundaryDummy

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

	// Negations require some global processing, which is done here
	// This process is derived from the one in https://github.com/micromatch/extglob/blob/master/lib/compilers.js

	// look for the end of a negative capture scope
	index := strings.Index(regex, closeNegDummy+")")
	if index > 0 {
		// negations should try to match until they reach a boundary
		if strings.Contains(regex[index:], boundaryDummy) {
			regex = strings.Replace(regex, closeNegDummy, "", -1)
		} else {
			// if no boundaries are imposed, match to the end of the line
			regex = strings.Replace(regex, closeNegDummy, "$", -1)
		}
	}
	// remove all the dummy markers
	regex = strings.Replace(regex, boundaryDummy, "", -1)

	// globs are expected to match against the whole thing
	return "\\A" + regex + "\\z", nil
}
