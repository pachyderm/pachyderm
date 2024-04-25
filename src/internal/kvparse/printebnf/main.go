package main

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/pachyderm/pachyderm/v2/src/internal/kvparse/kvgrammar"
)

func main() {
	lex := participle.Lexer(lexer.MustStateful(kvgrammar.Lexer))
	fmt.Println(participle.MustBuild[kvgrammar.KV](lex).String())
	fmt.Println()
	fmt.Println(participle.MustBuild[kvgrammar.KVs](lex).String())
}
