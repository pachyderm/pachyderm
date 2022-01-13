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
		return fmt.Sprintf("github.com/pachyderm/pachyderm/v2/src/%s", path.Dir(*proto.Name))
	},
	"hasAPI": func(proto *descriptor.FileDescriptorProto) bool {
		for _, service := range proto.Service {
			if *service.Name == "API" || *service.Name == "Debug" {
				return true
			}
		}
		return false
	},
	"isAPI": func(proto *descriptor.ServiceDescriptorProto) bool {
		return *proto.Name == "API" || *proto.Name == "Debug"
	},
	"isClientStreaming": func(proto *descriptor.MethodDescriptorProto) bool {
		return proto.ClientStreaming != nil && *proto.ClientStreaming
	},
	"isServerStreaming": func(proto *descriptor.MethodDescriptorProto) bool {
		return proto.ServerStreaming != nil && *proto.ServerStreaming
	},
	"pkgName": func(proto *descriptor.FileDescriptorProto) string {
		return path.Base(*proto.Options.GoPackage)
	},
	"title": strings.Title,
	"typeName": func(t *string) string {
		parts := strings.Split(*t, ".")
		if strings.HasPrefix(*t, ".google.protobuf") {
			return fmt.Sprintf("types.%s", parts[len(parts)-1])
		}
		for i := range parts {
			parts[i] = strings.Replace(parts[i], "_v2", "", -1)
		}
		return strings.Join(parts[1:], ".")
	},
	"clientName": func(t string) string {
		return fmt.Sprintf("unsupported%sBuilderClient", strings.Title(t))
	},
}

var outTemplate = template.Must(template.New("transaction-client").Funcs(funcs).Parse(`
package client

import (
	"context"
{{range .Protos}}{{if hasAPI .}}
  "{{importPath .}}"{{end}}{{end}}
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	types "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

func unsupportedError(name string) error {
	return errors.Errorf("the '%s' API call is not supported in transactions", name)
}
{{range .Protos}}{{$pkg := pkgName .}}{{$client := clientName $pkg}}{{range .Service}}{{$service := .}}{{if isAPI .}}
type {{$client}} struct {}
{{range .Method}}
func (c *{{$client}}) {{.Name}}(_ context.Context,{{if not (isClientStreaming .)}} _ *{{typeName .InputType}},{{end}} opts ...grpc.CallOption) ({{if or (isServerStreaming .) (isClientStreaming .)}}{{$pkg}}.{{$service.Name}}_{{.Name}}Client{{else}}*{{typeName .OutputType}}{{end}}, error) {
	return nil, unsupportedError("{{.Name}}")
}
{{end}}{{end}}{{end}}{{end}}
`))

func (gen *TransactionClientGenerator) AddProto(proto *descriptor.FileDescriptorProto) error {
	gen.Protos = append(gen.Protos, proto)
	return nil
}

func (gen *TransactionClientGenerator) Finish() (*plugin.CodeGeneratorResponse_File, error) {
	buf := &bytes.Buffer{}
	if err := outTemplate.Execute(buf, gen); err != nil {
		return nil, err
	}

	filename := "client/transaction.gen.go"
	content := buf.String()
	return &plugin.CodeGeneratorResponse_File{
		Name:    &filename,
		Content: &content,
	}, nil
}
