package ast

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pachyderm/ohmyglob/syntax/lexer"
)

var posix = map[string]string{
	":alnum:":  "0-9A-Za-z",
	":alpha:":  "A-Za-z",
	":ascii:":  "\x00-\x7F",
	":blank:":  "\t ",
	":cntrl:":  "\x00-\x1F\x7F",
	":digit:":  "0-9",
	":graph:":  "!-~]",
	":lower:":  "a-z",
	":print:":  " -~",
	":punct:":  "!-/:-@[-`{-~",
	":space:":  "\t\n\v\f\r ",
	":upper:":  "A-Z",
	":word:":   "0-9A-Za-z_",
	":xdigit:": "0-9A-Fa-f",
}

type Lexer interface {
	Next() lexer.Token
}

type parseFn func(*Node, Lexer) (parseFn, *Node, error)

func Parse(lexer Lexer) (*Node, error) {
	var parser parseFn

	root := NewNode(KindPattern, nil)

	var (
		tree *Node
		err  error
	)
	for parser, tree = parserMain, root; parser != nil; {
		parser, tree, err = parser(tree, lexer)
		if err != nil {
			return nil, err
		}
	}

	return root, nil
}

func parserMain(tree *Node, lex Lexer) (parseFn, *Node, error) {
	for {
		token := lex.Next()
		switch token.Type {
		case lexer.EOF:
			return nil, tree, nil

		case lexer.Error:
			return nil, tree, errors.New(token.Raw)

		case lexer.Text:
			Insert(tree, NewNode(KindText, Text{Text: token.Raw}))
			return parserMain, tree, nil

		case lexer.Any:
			Insert(tree, NewNode(KindAny, nil))
			return parserMain, tree, nil

		case lexer.Super:
			Insert(tree, NewNode(KindSuper, nil))
			return parserMain, tree, nil

		case lexer.Single:
			Insert(tree, NewNode(KindSingle, nil))
			return parserMain, tree, nil

		case lexer.RangeOpen:
			return parserRange, tree, nil

		case lexer.TermsOpen:
			a := NewNode(KindAnyOf, nil)
			Insert(tree, a)

			p := NewNode(KindPattern, nil)
			Insert(a, p)

			return parserMain, p, nil

		case lexer.CaptureOpen:
			a := NewNode(KindCapture, Capture{token.Raw[:1]})
			Insert(tree, a)

			p := NewNode(KindPattern, nil)
			Insert(a, p)

			return parserMain, p, nil

		case lexer.Separator:
			p := NewNode(KindPattern, nil)
			Insert(tree.Parent, p)

			return parserMain, p, nil

		case lexer.TermsClose:
			return parserMain, tree.Parent.Parent, nil

		case lexer.CaptureClose:
			return parserMain, tree.Parent.Parent, nil

		default:
			return nil, tree, fmt.Errorf("unexpected token: %s", token)
		}
	}
	return nil, tree, fmt.Errorf("unknown error")
}

func parserRange(tree *Node, lex Lexer) (parseFn, *Node, error) {
	var (
		not   bool
		chars string
	)
	for {
		token := lex.Next()
		switch token.Type {
		case lexer.EOF:
			return nil, tree, errors.New("unexpected end")

		case lexer.Error:
			return nil, tree, errors.New(token.Raw)

		case lexer.Not:
			not = true

		case lexer.Text:
			chars = token.Raw
			for k, v := range posix {
				chars = strings.ReplaceAll(chars, k, v)
			}

		case lexer.RangeClose:
			Insert(tree, NewNode(KindList, List{
				Chars: chars,
				Not:   not,
			}))

			return parserMain, tree, nil
		}
	}
}
