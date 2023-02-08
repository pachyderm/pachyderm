package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

var outputPath = "/tmp/api-test/"
var packageLine = "package main_test"
var imports = []string{"\"testing\"", "\"github.com/pachyderm/pachyderm/v2/src/client\""}

// var declareFunction = "func Fuzz%s(f *testing.F) {" // TODO templatize
var newClient = []string{
	"pachClient, err := NewForTest()",
	"if err != nil {",
	"	return err",
	"}"}
var final = []string{"}"}

// TODO add add() for known cases https://go.dev/doc/tutorial/fuzz
func closeParamBlock(currentIndent int, currentBlock []string) []string {
	return append(currentBlock, withIndent(currentIndent, ")")...)
}
func closeFunction(currentIndent int, currentFunction []string) ([]string, int) {
	currentIndent -= 1
	return append(currentFunction, withIndent(currentIndent, "}")...), currentIndent
}
func declareFunction(currentIndent int, lines []string, funcLine string) ([]string, int) {
	lines = append(lines, withIndent(currentIndent, funcLine)...)
	return lines, currentIndent + 1
}
func startFuzzBlock(currentIndent int, lines []string) ([]string, int) {
	lines = append(lines, withIndent(currentIndent, "f.Fuzz(func(t *testing.T, %s) {")...)
	return lines, currentIndent + 1
}
func createApiCall(rpc *descriptorpb.MethodDescriptorProto, proto *descriptorpb.FileDescriptorProto, imports []string) ([]string, []string) {
	retLines := []string{}
	retLines = append(retLines, fmt.Sprintf("_, err := pachClient.PfsAPIClient.%s(", rpc.GetName()))
	// TODO make client generic
	pkg := strings.Split(proto.GetOptions().GetGoPackage(), "/")
	pkgName := pkg[len(pkg)-1]
	inputType := strings.Split(rpc.GetInputType(), ".")
	inputTypeName := inputType[len(inputType)-1]
	retLines = append(retLines, fmt.Sprintf("	&%s.%s{", pkgName, inputTypeName)) // using the go type - can we use reflection to calculate this
	retLines = append(retLines, "		c.Ctx(),")
	// FieldDescriptorProto_Type
	for _, mt := range proto.GetMessageType() {
		if inputTypeName == mt.GetName() {
			f2, _ := os.Create(fmt.Sprintf("%s%s", outputPath, "types.txt"))
			defer f2.Close()
			f2.WriteString(fmt.Sprintf("DNJ type: %v", mt))
			for _, field := range mt.GetField() {
				// // TODO how to return structure as well as primitives.
				// if field.GetType() == FieldDescriptorProto_Type.MESSAGE_TYPE {
				// 	// recurse and repeat type finding
				// } else {
				// 	// translate proto type to go primitive types
				// }
			}
		}
	}
	retLines = append(retLines, "		Repo: NewProjectRepo(projectName, repoName),") // TODO multiple inputs?
	retLines = append(retLines, "})")
	return retLines, append(imports, fmt.Sprintf("%s \"%s\"", pkgName, proto.GetOptions().GetGoPackage()))
}
func createImports(imports []string) []string {
	imports = append([]string{"import ("}, withIndent(1, imports...)...)
	return append(imports, ")", "")
}
func withIndent(indent int, strings ...string) []string {
	retLines := []string{}
	for _, line := range strings {
		indentStr := ""
		for i := 0; i < indent; i++ {
			indentStr += "	"
		}
		retLines = append(retLines, indentStr+line)
	}
	return retLines
}
func writeToFile(fileName string, lines []string) error {
	testFile, err := os.Create(fmt.Sprintf("%s%s", outputPath, fileName))
	// gofakeit
	if err != nil {
		return err
	}
	defer testFile.Close()
	for _, line := range lines {
		_, err = testFile.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// var filePath = "/home/docksonwedge/Projects/pachyderm/src/pfs/pfs.proto"
func run() error {
	// "/home/docksonwedge/Projects/pachyderm/src/pfs/pfs.proto"

	// protogen.Options{}.Run(func(plugin *protogen.Plugin) error {
	// 	for _, f := range plugin.Files {
	// 		fmt.Printf("%s",f.Desc.Path())
	// 	}
	// 	return nil
	// })
	req := &plugin.CodeGeneratorRequest{}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	} else if err := proto.Unmarshal(data, req); err != nil {
		return err
	}
	f, err := os.Create(fmt.Sprintf("%s%s", outputPath, "protos.txt"))
	defer f.Close()
	for _, proto := range req.ProtoFile {
		f.WriteString(fmt.Sprintf("DNJ proto: %s, ", *proto.Name))
		if *proto.Name == "pfs/pfs.proto" {

			// f2, err := os.Create(fmt.Sprintf("%s%s", outputPath, "types.txt"))
			// if err != nil {
			// 	return err
			// }
			// defer f2.Close()

			for _, service := range proto.Service {

				for _, rpc := range service.GetMethod() {

					currentIndent := 0
					if rpc.GetName() == "CreateRepo" {
						testLines := []string{}
						// primitives := []interface{}{}
						fileName := fmt.Sprintf("fuzz_%s_test.go", strings.ToLower(rpc.GetName()))

						funcLines, currentIndent := declareFunction(currentIndent, testLines, fmt.Sprintf("func Fuzz%s(f *testing.F) {", rpc.GetName()))
						funcLines = append(funcLines, withIndent(currentIndent, newClient...)...)

						funcLines, currentIndent = startFuzzBlock(currentIndent, funcLines)
						apiCallBlock, imports := createApiCall(rpc, proto, imports)
						funcLines = append(funcLines, withIndent(currentIndent, apiCallBlock...)...)
						funcLines, currentIndent = closeFunction(currentIndent, funcLines)
						funcLines = closeParamBlock(currentIndent, funcLines)
						funcLines, currentIndent = closeFunction(currentIndent, funcLines)
						testLines = append(testLines, funcLines...)

						testLines = append(createImports(imports), testLines...)
						writeToFile(fileName, testLines)
					}
				}
			}
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}
