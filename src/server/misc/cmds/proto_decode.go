package cmds

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

func allProtoMessages() map[string]protoreflect.MessageDescriptor {
	result := make(map[string]protoreflect.MessageDescriptor)
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		ms := fd.Messages()
		for i := 0; i < ms.Len(); i++ {
			md := ms.Get(i)
			result[string(md.FullName())] = md
		}
		return true
	})
	return result
}

func decodeBinaryProto(md protoreflect.MessageDescriptor, b []byte) (proto.Message, error) {
	m := dynamicpb.NewMessage(md)
	if err := proto.Unmarshal(b, m); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	return m, nil
}
