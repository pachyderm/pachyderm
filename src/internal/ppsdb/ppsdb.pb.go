// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.4
// source: internal/ppsdb/ppsdb.proto

package ppsdb

import (
	_ "github.com/pachyderm/pachyderm/v2/src/pps"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ClusterDefaultsWrapper struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Json string `protobuf:"bytes,3,opt,name=json,proto3" json:"json,omitempty"`
}

func (x *ClusterDefaultsWrapper) Reset() {
	*x = ClusterDefaultsWrapper{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_ppsdb_ppsdb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterDefaultsWrapper) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterDefaultsWrapper) ProtoMessage() {}

func (x *ClusterDefaultsWrapper) ProtoReflect() protoreflect.Message {
	mi := &file_internal_ppsdb_ppsdb_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterDefaultsWrapper.ProtoReflect.Descriptor instead.
func (*ClusterDefaultsWrapper) Descriptor() ([]byte, []int) {
	return file_internal_ppsdb_ppsdb_proto_rawDescGZIP(), []int{0}
}

func (x *ClusterDefaultsWrapper) GetJson() string {
	if x != nil {
		return x.Json
	}
	return ""
}

type ProjectDefaultsWrapper struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Json string `protobuf:"bytes,3,opt,name=json,proto3" json:"json,omitempty"`
}

func (x *ProjectDefaultsWrapper) Reset() {
	*x = ProjectDefaultsWrapper{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_ppsdb_ppsdb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProjectDefaultsWrapper) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProjectDefaultsWrapper) ProtoMessage() {}

func (x *ProjectDefaultsWrapper) ProtoReflect() protoreflect.Message {
	mi := &file_internal_ppsdb_ppsdb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProjectDefaultsWrapper.ProtoReflect.Descriptor instead.
func (*ProjectDefaultsWrapper) Descriptor() ([]byte, []int) {
	return file_internal_ppsdb_ppsdb_proto_rawDescGZIP(), []int{1}
}

func (x *ProjectDefaultsWrapper) GetJson() string {
	if x != nil {
		return x.Json
	}
	return ""
}

var File_internal_ppsdb_ppsdb_proto protoreflect.FileDescriptor

var file_internal_ppsdb_ppsdb_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x70, 0x73, 0x64, 0x62,
	0x2f, 0x70, 0x70, 0x73, 0x64, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x70, 0x70,
	0x73, 0x5f, 0x76, 0x32, 0x1a, 0x0d, 0x70, 0x70, 0x73, 0x2f, 0x70, 0x70, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x5e, 0x0a, 0x16, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x44, 0x65,
	0x66, 0x61, 0x75, 0x6c, 0x74, 0x73, 0x57, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x12, 0x12, 0x0a,
	0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6a, 0x73, 0x6f,
	0x6e, 0x4a, 0x04, 0x08, 0x01, 0x10, 0x02, 0x4a, 0x04, 0x08, 0x02, 0x10, 0x03, 0x52, 0x0c, 0x64,
	0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x5f, 0x6a, 0x73, 0x6f, 0x6e, 0x52, 0x16, 0x65, 0x66, 0x66,
	0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x5f, 0x6a,
	0x73, 0x6f, 0x6e, 0x22, 0x2c, 0x0a, 0x16, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x44, 0x65,
	0x66, 0x61, 0x75, 0x6c, 0x74, 0x73, 0x57, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x12, 0x12, 0x0a,
	0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6a, 0x73, 0x6f,
	0x6e, 0x42, 0x36, 0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x70, 0x61, 0x63, 0x68, 0x79, 0x64, 0x65, 0x72, 0x6d, 0x2f, 0x70, 0x61, 0x63, 0x68, 0x79, 0x64,
	0x65, 0x72, 0x6d, 0x2f, 0x76, 0x32, 0x2f, 0x73, 0x72, 0x63, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x70, 0x73, 0x64, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_internal_ppsdb_ppsdb_proto_rawDescOnce sync.Once
	file_internal_ppsdb_ppsdb_proto_rawDescData = file_internal_ppsdb_ppsdb_proto_rawDesc
)

func file_internal_ppsdb_ppsdb_proto_rawDescGZIP() []byte {
	file_internal_ppsdb_ppsdb_proto_rawDescOnce.Do(func() {
		file_internal_ppsdb_ppsdb_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_ppsdb_ppsdb_proto_rawDescData)
	})
	return file_internal_ppsdb_ppsdb_proto_rawDescData
}

var file_internal_ppsdb_ppsdb_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_internal_ppsdb_ppsdb_proto_goTypes = []interface{}{
	(*ClusterDefaultsWrapper)(nil), // 0: pps_v2.ClusterDefaultsWrapper
	(*ProjectDefaultsWrapper)(nil), // 1: pps_v2.ProjectDefaultsWrapper
}
var file_internal_ppsdb_ppsdb_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_internal_ppsdb_ppsdb_proto_init() }
func file_internal_ppsdb_ppsdb_proto_init() {
	if File_internal_ppsdb_ppsdb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_ppsdb_ppsdb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterDefaultsWrapper); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_ppsdb_ppsdb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProjectDefaultsWrapper); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_ppsdb_ppsdb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_ppsdb_ppsdb_proto_goTypes,
		DependencyIndexes: file_internal_ppsdb_ppsdb_proto_depIdxs,
		MessageInfos:      file_internal_ppsdb_ppsdb_proto_msgTypes,
	}.Build()
	File_internal_ppsdb_ppsdb_proto = out.File
	file_internal_ppsdb_ppsdb_proto_rawDesc = nil
	file_internal_ppsdb_ppsdb_proto_goTypes = nil
	file_internal_ppsdb_ppsdb_proto_depIdxs = nil
}
