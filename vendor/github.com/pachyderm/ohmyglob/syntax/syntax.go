package syntax

import (
	"github.com/pachyderm/ohmyglob/syntax/ast"
	"github.com/pachyderm/ohmyglob/syntax/lexer"
)

func Parse(s string) (*ast.Node, error) {
	return ast.Parse(lexer.NewLexer(s))
}

func Special(b byte) bool {
	return lexer.Special(b)
}
