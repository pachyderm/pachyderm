package graphqlproto

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

func GetGraphQLFile(file *descriptor.FileDescriptorProto) bool {
	if file.Options != nil {
		v, err := proto.GetExtension(file.Options, E_Graphql)
		if err == nil && v.(*bool) != nil {
			return (*v.(*bool))
		}
	}
	return false
}

func IsBitflags(file *descriptor.FileDescriptorProto, message *descriptor.DescriptorProto) bool {
	return proto.GetBoolExtension(message.Options, E_Bitflags, false)
}

func GetRequiredField(message *descriptor.DescriptorProto) bool {
	if message.Options != nil {
		v, err := proto.GetExtension(message.Options, E_Required)
		if err == nil && v.(*bool) != nil {
			return (*v.(*bool))
		}
	}
	return false
}
