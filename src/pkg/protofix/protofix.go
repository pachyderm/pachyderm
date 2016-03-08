package protofix

import(
        "fmt"
	"bytes"
	"go/printer"
        "go/parser"
        "go/token"
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

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)

	fmt.Printf("new file:\n%v\n", buf.String())
	
}
