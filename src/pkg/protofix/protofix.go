package protofix

import(
        "fmt"
	"bytes"
	"strings"
	"go/printer"
        "go/parser"
        "go/token"
	"go/ast"
	"io/ioutil"
)

func FixAllPBGOFilesInDirectory() {
}
        
func gatherFiles() {
}

func repairedFileBytes(filename string) []byte {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, nil, parser.DeclarationErrors)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	n := &lukeNodeWalker{}

	ast.Walk(n, f)

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)

	return buf.Bytes()
}

func repairFile(filename string) {
	newFileContents := repairedFileBytes(filename)
	ioutil.WriteFile(filename, newFileContents, 0644)
}

func repairDeclaration(node ast.Node) {
	switch node := node.(type) {
	case *ast.Field:
		declName := node.Names[0].Name
		if strings.HasSuffix(declName, "Id") {
			normalized := strings.TrimSuffix(declName, "Id")
			node.Names[0] = ast.NewIdent(fmt.Sprintf("%vID", normalized))
		}
//	default:
//		fmt.Printf("the type is (%T)\n", node)
	}
	
}

type lukeNodeWalker struct {	
}

func (w *lukeNodeWalker) Visit(node ast.Node) (ast.Visitor) {
	repairDeclaration(node)
	return w
}
