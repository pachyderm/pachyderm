package protofix

import(
        "fmt"
	"bytes"
	"go/printer"
        "go/parser"
        "go/token"
	"go/ast"
)

func FixAllPBGOFilesInDirectory() {
}
        
func gatherFiles() {
}

func repairFile(filename string) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, nil, parser.DeclarationErrors)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("file stuff: %v\n", f)

	for _, s := range f.Imports {
		fmt.Println(s.Path.Value)
	}

	n := &lukeNodeWalker{}

	ast.Walk(n, f)

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)
	fmt.Printf("new file:\n%v\n", buf.String())
	
}

func repairDeclaration(node ast.Node) {
	switch node := node.(type) {
	case *ast.DeclStmt:
		fmt.Printf("Found declaration: %v\n", node)
	case *ast.Field:
		fmt.Printf("Found field (%v)\n", node)
		fmt.Printf("names: %v\n", node.Names)
		node.Names[0] = ast.NewIdent(fmt.Sprintf("%vOMG", node.Names[0]))
	default:
		fmt.Printf("the type is (%T)\n", node)
	}
	
}

type lukeNodeWalker struct {	
}

func (w *lukeNodeWalker) Visit(node ast.Node) (ast.Visitor) {
	fmt.Printf("walking over node: %v\n", node)
	repairDeclaration(node)
	return w
}
