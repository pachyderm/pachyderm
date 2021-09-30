package main

import (
	"fmt"
	"os"

	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func init() {
	CodeGenerators = append(CodeGenerators, &MockClientGenerator{})
}

type MockClientGenerator struct {
}

func (gen *MockClientGenerator) AddProto(proto *descriptor.FileDescriptorProto) error {
	for _, service := range proto.Service {
		for _, method := range service.Method {
			fmt.Fprintf(os.Stderr, "%s.%s\n", *service.Name, *method.Name)
		}
	}
	return nil
}

func (gen *MockClientGenerator) Finish() (*plugin.CodeGeneratorResponse_File, error) {
	return nil, fmt.Errorf("unimplemented")
}
