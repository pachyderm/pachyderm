package main

// Parses the deploy cmds file to figure out what version of other software
// (e.g. dash) we're pinning to.
// TODO: This is pretty hackey, at some point we should build out a more
// robust way of dealing with compatibility matrices. See RFC #17
// (https://github.com/pachyderm/rfcs/pull/17) for one proposal, though it was
// initially rejected due to trade-offs in ops complexity.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
)

const deployCmdsFile = "./src/server/pkg/deploy/cmds/cmds.go"

func main() {
	if len(os.Args) <= 1 {
		panic("missing ident")
	}
	ident := os.Args[1]

	fset := token.NewFileSet()
	src, err := ioutil.ReadFile(deployCmdsFile)
	if err != nil {
		panic(err)
	}
	f, err := parser.ParseFile(fset, "", src, parser.DeclarationErrors)
	if err != nil {
		panic(err)
	}

	for _, d := range f.Decls {
		if decl, ok := d.(*ast.GenDecl); ok {
			if decl.Tok == token.CONST || decl.Tok == token.VAR {
				for _, rawSpec := range decl.Specs {
					if spec, ok := rawSpec.(*ast.ValueSpec); ok {
						if len(spec.Names) == 1 && spec.Names[0].Name == ident {
							rawVal := spec.Values[0]
							if val, ok := rawVal.(*ast.BasicLit); ok {
								fmt.Println(strings.Trim(val.Value, "\""))
								return
							}
						}
					}
				}
			}
		}
	}

	panic(fmt.Sprintf("expected ident %q is missing from %q", ident, deployCmdsFile))
}
