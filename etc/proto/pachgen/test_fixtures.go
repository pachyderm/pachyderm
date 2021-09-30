package main

import (
	"fmt"
	"os"

	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func init() {
	CodeGenerators = append(CodeGenerators, &TestFixturesGenerator{})
}

type TestFixturesGenerator struct {
}

func (gen *TestFixturesGenerator) AddProto(proto *descriptor.FileDescriptorProto) error {
	for _, service := range proto.Service {
		for _, method := range service.Method {
			fmt.Fprintf(os.Stderr, "%s.%s\n", *service.Name, *method.Name)
		}
	}
	return nil
}

func (gen *TestFixturesGenerator) Finish() (*plugin.CodeGeneratorResponse_File, error) {
	return nil, fmt.Errorf("unimplemented")
}
