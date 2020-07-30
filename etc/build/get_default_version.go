package main

// I'm sorry.

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
