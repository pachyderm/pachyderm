package protofix

import(
        "fmt"
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

	for _, s := range f.Imports {
		fmt.Println(s.Path.Value)
	}
	
}
