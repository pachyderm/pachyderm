package protofix

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FixAllPBGOFilesInDirectory fixes all *.pb.go files in rootPath.
func FixAllPBGOFilesInDirectory(rootPath string) {
	filepath.Walk(rootPath, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(f.Name(), ".pb.go") {
			fmt.Printf("Repairing %v\n", path)
			repairFile(path)
		}
		return nil
	})
}

// RevertAllPBGOFilesInDirectory reverts all *.pb.go files in rootPath.
func RevertAllPBGOFilesInDirectory(rootPath string) {
	filepath.Walk(rootPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return nil
		}
		if strings.HasSuffix(f.Name(), ".pb.go") {
			fmt.Printf("Reverting %v\n", path)
			args := []string{"checkout", path}
			_, err := exec.Command("git", args...).Output()
			if err != nil {
				fmt.Printf("error reverting %v : %v\n", path, err)
				os.Exit(1)
			}
		}
		return nil
	})
}

func repairedFileBytes(filename string) []byte {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	n := &walker{}

	ast.Walk(n, f)

	var buf bytes.Buffer

	config := &printer.Config{Mode: printer.UseSpaces + printer.TabIndent, Tabwidth: 8, Indent: 0}
	config.Fprint(&buf, fset, f)

	scanner := bufio.NewScanner(&buf)
	var out bytes.Buffer

	for scanner.Scan() {
		// this is retarded
		line := scanner.Text()
		if !strings.Contains(line, "grpc.SupportPackageIsVersion1") {
			out.WriteString(line + "\n")
		}
	}

	return out.Bytes()
}

func repairFile(filename string) {
	newFileContents := repairedFileBytes(filename)
	ioutil.WriteFile(filename, newFileContents, 0644)
}

func repairDeclaration(node ast.Node) {
	switch node := node.(type) {
	case *ast.Field:
		if len(node.Names) > 0 {
			declName := node.Names[0].Name
			if strings.HasSuffix(declName, "Id") {
				normalized := strings.TrimSuffix(declName, "Id")
				node.Names[0] = ast.NewIdent(fmt.Sprintf("%vID", normalized))
			}
		}
	}

}

type walker struct {
}

func (w *walker) Visit(node ast.Node) ast.Visitor {
	repairDeclaration(node)
	return w
}
