// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.1
// source: internal/storage/fileset/index/index.proto

package index

import (
	chunk "github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
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

// Index stores an index to and metadata about a range of files or a file.
type Index struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// NOTE: range and file are mutually exclusive.
	Range *Range `protobuf:"bytes,2,opt,name=range,proto3" json:"range,omitempty"`
	File  *File  `protobuf:"bytes,3,opt,name=file,proto3" json:"file,omitempty"`
	// NOTE: num_files and size_bytes did not exist in older versions of 2.x, so
	// they will not be set.
	NumFiles  int64 `protobuf:"varint,4,opt,name=num_files,json=numFiles,proto3" json:"num_files,omitempty"`
	SizeBytes int64 `protobuf:"varint,5,opt,name=size_bytes,json=sizeBytes,proto3" json:"size_bytes,omitempty"`
}

func (x *Index) Reset() {
	*x = Index{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_storage_fileset_index_index_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Index) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Index) ProtoMessage() {}

func (x *Index) ProtoReflect() protoreflect.Message {
	mi := &file_internal_storage_fileset_index_index_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Index.ProtoReflect.Descriptor instead.
func (*Index) Descriptor() ([]byte, []int) {
	return file_internal_storage_fileset_index_index_proto_rawDescGZIP(), []int{0}
}

func (x *Index) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Index) GetRange() *Range {
	if x != nil {
		return x.Range
	}
	return nil
}

func (x *Index) GetFile() *File {
	if x != nil {
		return x.File
	}
	return nil
}

func (x *Index) GetNumFiles() int64 {
	if x != nil {
		return x.NumFiles
	}
	return 0
}

func (x *Index) GetSizeBytes() int64 {
	if x != nil {
		return x.SizeBytes
	}
	return 0
}

type Range struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Offset   int64          `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	LastPath string         `protobuf:"bytes,2,opt,name=last_path,json=lastPath,proto3" json:"last_path,omitempty"`
	ChunkRef *chunk.DataRef `protobuf:"bytes,3,opt,name=chunk_ref,json=chunkRef,proto3" json:"chunk_ref,omitempty"`
}

func (x *Range) Reset() {
	*x = Range{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_storage_fileset_index_index_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Range) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Range) ProtoMessage() {}

func (x *Range) ProtoReflect() protoreflect.Message {
	mi := &file_internal_storage_fileset_index_index_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Range.ProtoReflect.Descriptor instead.
func (*Range) Descriptor() ([]byte, []int) {
	return file_internal_storage_fileset_index_index_proto_rawDescGZIP(), []int{1}
}

func (x *Range) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *Range) GetLastPath() string {
	if x != nil {
		return x.LastPath
	}
	return ""
}

func (x *Range) GetChunkRef() *chunk.DataRef {
	if x != nil {
		return x.ChunkRef
	}
	return nil
}

type File struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Datum    string           `protobuf:"bytes,1,opt,name=datum,proto3" json:"datum,omitempty"`
	DataRefs []*chunk.DataRef `protobuf:"bytes,2,rep,name=data_refs,json=dataRefs,proto3" json:"data_refs,omitempty"`
}

func (x *File) Reset() {
	*x = File{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_storage_fileset_index_index_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *File) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*File) ProtoMessage() {}

func (x *File) ProtoReflect() protoreflect.Message {
	mi := &file_internal_storage_fileset_index_index_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use File.ProtoReflect.Descriptor instead.
func (*File) Descriptor() ([]byte, []int) {
	return file_internal_storage_fileset_index_index_proto_rawDescGZIP(), []int{2}
}

func (x *File) GetDatum() string {
	if x != nil {
		return x.Datum
	}
	return ""
}

func (x *File) GetDataRefs() []*chunk.DataRef {
	if x != nil {
		return x.DataRefs
	}
	return nil
}

var File_internal_storage_fileset_index_index_proto protoreflect.FileDescriptor

var file_internal_storage_fileset_index_index_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x65, 0x74, 0x2f, 0x69, 0x6e, 0x64, 0x65, 0x78,
	0x2f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x69, 0x6e,
	0x64, 0x65, 0x78, 0x1a, 0x22, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x73, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x2f, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x2f, 0x63, 0x68, 0x75, 0x6e,
	0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9c, 0x01, 0x0a, 0x05, 0x49, 0x6e, 0x64, 0x65,
	0x78, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x22, 0x0a, 0x05, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x52, 0x61, 0x6e,
	0x67, 0x65, 0x52, 0x05, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x1f, 0x0a, 0x04, 0x66, 0x69, 0x6c,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e,
	0x46, 0x69, 0x6c, 0x65, 0x52, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x6e, 0x75,
	0x6d, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x6e,
	0x75, 0x6d, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x69, 0x7a, 0x65, 0x5f,
	0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73, 0x69, 0x7a,
	0x65, 0x42, 0x79, 0x74, 0x65, 0x73, 0x22, 0x69, 0x0a, 0x05, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x6c, 0x61, 0x73, 0x74, 0x5f,
	0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x61, 0x73, 0x74,
	0x50, 0x61, 0x74, 0x68, 0x12, 0x2b, 0x0a, 0x09, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x5f, 0x72, 0x65,
	0x66, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x2e,
	0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x66, 0x52, 0x08, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x52, 0x65,
	0x66, 0x22, 0x49, 0x0a, 0x04, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x64, 0x61, 0x74,
	0x75, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x64, 0x61, 0x74, 0x75, 0x6d, 0x12,
	0x2b, 0x0a, 0x09, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x72, 0x65, 0x66, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x52,
	0x65, 0x66, 0x52, 0x08, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x66, 0x73, 0x42, 0x46, 0x5a, 0x44,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x61, 0x63, 0x68, 0x79,
	0x64, 0x65, 0x72, 0x6d, 0x2f, 0x70, 0x61, 0x63, 0x68, 0x79, 0x64, 0x65, 0x72, 0x6d, 0x2f, 0x76,
	0x32, 0x2f, 0x73, 0x72, 0x63, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x73,
	0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x65, 0x74, 0x2f, 0x69,
	0x6e, 0x64, 0x65, 0x78, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_storage_fileset_index_index_proto_rawDescOnce sync.Once
	file_internal_storage_fileset_index_index_proto_rawDescData = file_internal_storage_fileset_index_index_proto_rawDesc
)

func file_internal_storage_fileset_index_index_proto_rawDescGZIP() []byte {
	file_internal_storage_fileset_index_index_proto_rawDescOnce.Do(func() {
		file_internal_storage_fileset_index_index_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_storage_fileset_index_index_proto_rawDescData)
	})
	return file_internal_storage_fileset_index_index_proto_rawDescData
}

var file_internal_storage_fileset_index_index_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_internal_storage_fileset_index_index_proto_goTypes = []interface{}{
	(*Index)(nil),         // 0: index.Index
	(*Range)(nil),         // 1: index.Range
	(*File)(nil),          // 2: index.File
	(*chunk.DataRef)(nil), // 3: chunk.DataRef
}
var file_internal_storage_fileset_index_index_proto_depIdxs = []int32{
	1, // 0: index.Index.range:type_name -> index.Range
	2, // 1: index.Index.file:type_name -> index.File
	3, // 2: index.Range.chunk_ref:type_name -> chunk.DataRef
	3, // 3: index.File.data_refs:type_name -> chunk.DataRef
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_internal_storage_fileset_index_index_proto_init() }
func file_internal_storage_fileset_index_index_proto_init() {
	if File_internal_storage_fileset_index_index_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_storage_fileset_index_index_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Index); i {
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
		file_internal_storage_fileset_index_index_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Range); i {
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
		file_internal_storage_fileset_index_index_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*File); i {
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
			RawDescriptor: file_internal_storage_fileset_index_index_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_storage_fileset_index_index_proto_goTypes,
		DependencyIndexes: file_internal_storage_fileset_index_index_proto_depIdxs,
		MessageInfos:      file_internal_storage_fileset_index_index_proto_msgTypes,
	}.Build()
	File_internal_storage_fileset_index_index_proto = out.File
	file_internal_storage_fileset_index_index_proto_rawDesc = nil
	file_internal_storage_fileset_index_index_proto_goTypes = nil
	file_internal_storage_fileset_index_index_proto_depIdxs = nil
}
