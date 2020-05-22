package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	src, err := ioutil.ReadFile("./src/server/pkg/deploy/cmds/cmds.go")
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
						if len(spec.Names) == 1 && spec.Names[0].Name == "defaultDashImage" {
							rawVal := spec.Values[0]
							if val, ok := rawVal.(*ast.BasicLit); ok {
								// We only want the version string
								dockerImage := strings.Trim(val.Value, "\"")
								tokens := strings.Split(dockerImage, ":")
								if len(tokens) != 2 {
									panic(fmt.Errorf("invalid dash docker image %v", dockerImage))
								}
								fmt.Println(tokens[1])
							}
						}
					}
				}
			}
		}
	}

}
