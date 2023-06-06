//nolint:wrapcheck
package main

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"

	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func init() {
	CodeGenerators = append(CodeGenerators, &TransactionClientGenerator{})
}

type TransactionClientGenerator struct {
	Protos []*descriptor.FileDescriptorProto
}

var funcs = map[string]interface{}{
	"importPath": func(proto *descriptor.FileDescriptorProto) string {
		return fmt.Sprintf(`%s "%s"`, *proto.Package, *proto.Options.GoPackage)
	},
	"shouldImport": func(proto *descriptor.FileDescriptorProto) bool {
		for _, service := range proto.Service {
			if *service.Name == "API" || *service.Name == "Debug" {
				return true
			}
		}
		// special case because taskapi has no services defined
		return *proto.Package == "taskapi"
	},
	"isAPI": func(proto *descriptor.ServiceDescriptorProto) bool {
		return *proto.Name == "API" || *proto.Name == "Debug"
	},
	"isClientStreaming": func(proto *descriptor.MethodDescriptorProto) bool {
		return proto.GetClientStreaming()
	},
	"isServerStreaming": func(proto *descriptor.MethodDescriptorProto) bool {
		return proto.GetServerStreaming()
	},
	"protoPkgName": func(proto *descriptor.FileDescriptorProto) string {
		return *proto.Package
	},
	"goPkgName": func(proto *descriptor.FileDescriptorProto) string {
		return path.Base(*proto.Options.GoPackage)
	},
	"title": strings.Title, //nolint:staticcheck
	"typeName": func(t *string) string {
		parts := strings.Split(*t, ".")
		if len(parts) == 4 && parts[1] == "google" && parts[2] == "protobuf" {
			// example .google.protobuf.Empty
			switch parts[3] {
			case "Empty":
				return "emptypb.Empty"
			// If you run into a new type here, you'll need to import it down below.
			default:
				panic(fmt.Sprintf("unknown well-known type %v", parts))
			}
		}
		// example .pfs_v2.CreateRepoRequest
		return strings.Join(parts[1:], ".")
	},
	"clientName": func(t string) string {
		return fmt.Sprintf("unsupported%sBuilderClient", strings.Title(t)) //nolint:staticcheck
	},
}

var outTemplate = template.Must(template.New("transaction-client").Funcs(funcs).Parse(`
package client

import (
	"context"
	{{range .Protos}}{{if shouldImport .}}
	{{importPath .}}{{end}}{{end}}
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func unsupportedError(name string) error {
	return errors.Errorf("the '%s' API call is not supported in transactions", name)
}
{{range .Protos}}{{$protoPkg := protoPkgName .}}{{$pkg := goPkgName .}}{{$client := clientName $pkg}}{{range .Service}}{{$service := .}}{{if isAPI .}}
type {{$client}} struct{}
{{range .Method}}
func (c *{{$client}}) {{.Name}}(_ context.Context,{{if not (isClientStreaming .)}} _ *{{typeName .InputType}},{{end}} opts ...grpc.CallOption) ({{if or (isServerStreaming .) (isClientStreaming .)}}{{$protoPkg}}.{{$service.Name}}_{{.Name}}Client{{else}}*{{typeName .OutputType}}{{end}}, error) {
	return nil, unsupportedError("{{.Name}}")
}
{{end}}{{end}}{{end}}{{end}}
`))

func (gen *TransactionClientGenerator) AddProto(proto *descriptor.FileDescriptorProto) error {
	gen.Protos = append(gen.Protos, proto)
	return nil
}

func (gen *TransactionClientGenerator) Finish() ([]*plugin.CodeGeneratorResponse_File, error) {
	buf := &bytes.Buffer{}
	if err := outTemplate.Execute(buf, gen); err != nil {
		return nil, err
	}

	publicFilename := "client/transaction.gen.go"
	privateFilename := "internal/client/transaction.gen.go"
	content := buf.String()
	return []*plugin.CodeGeneratorResponse_File{
		{
			Name:    &publicFilename,
			Content: &content,
		},
		{
			Name:    &privateFilename,
			Content: &content,
		},
	}, nil
}
