// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.11.4
// source: internal/pfsload/pfsload.proto

package pfsload

import (
	pfs "github.com/pachyderm/pachyderm/v2/src/pfs"
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

type CommitSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count         int64               `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	Modifications []*ModificationSpec `protobuf:"bytes,2,rep,name=modifications,proto3" json:"modifications,omitempty"`
	FileSources   []*FileSourceSpec   `protobuf:"bytes,3,rep,name=file_sources,json=fileSources,proto3" json:"file_sources,omitempty"`
	Validator     *ValidatorSpec      `protobuf:"bytes,4,opt,name=validator,proto3" json:"validator,omitempty"`
}

func (x *CommitSpec) Reset() {
	*x = CommitSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommitSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommitSpec) ProtoMessage() {}

func (x *CommitSpec) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommitSpec.ProtoReflect.Descriptor instead.
func (*CommitSpec) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{0}
}

func (x *CommitSpec) GetCount() int64 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *CommitSpec) GetModifications() []*ModificationSpec {
	if x != nil {
		return x.Modifications
	}
	return nil
}

func (x *CommitSpec) GetFileSources() []*FileSourceSpec {
	if x != nil {
		return x.FileSources
	}
	return nil
}

func (x *CommitSpec) GetValidator() *ValidatorSpec {
	if x != nil {
		return x.Validator
	}
	return nil
}

type ModificationSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count   int64        `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	PutFile *PutFileSpec `protobuf:"bytes,2,opt,name=put_file,json=putFile,proto3" json:"put_file,omitempty"`
}

func (x *ModificationSpec) Reset() {
	*x = ModificationSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ModificationSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModificationSpec) ProtoMessage() {}

func (x *ModificationSpec) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModificationSpec.ProtoReflect.Descriptor instead.
func (*ModificationSpec) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{1}
}

func (x *ModificationSpec) GetCount() int64 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *ModificationSpec) GetPutFile() *PutFileSpec {
	if x != nil {
		return x.PutFile
	}
	return nil
}

type PutFileSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count  int64  `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	Source string `protobuf:"bytes,2,opt,name=source,proto3" json:"source,omitempty"`
}

func (x *PutFileSpec) Reset() {
	*x = PutFileSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutFileSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutFileSpec) ProtoMessage() {}

func (x *PutFileSpec) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutFileSpec.ProtoReflect.Descriptor instead.
func (*PutFileSpec) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{2}
}

func (x *PutFileSpec) GetCount() int64 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *PutFileSpec) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

type PutFileTask struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count      int64           `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	FileSource *FileSourceSpec `protobuf:"bytes,2,opt,name=file_source,json=fileSource,proto3" json:"file_source,omitempty"`
	Seed       int64           `protobuf:"varint,3,opt,name=seed,proto3" json:"seed,omitempty"`
	AuthToken  string          `protobuf:"bytes,4,opt,name=auth_token,json=authToken,proto3" json:"auth_token,omitempty"`
}

func (x *PutFileTask) Reset() {
	*x = PutFileTask{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutFileTask) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutFileTask) ProtoMessage() {}

func (x *PutFileTask) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutFileTask.ProtoReflect.Descriptor instead.
func (*PutFileTask) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{3}
}

func (x *PutFileTask) GetCount() int64 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *PutFileTask) GetFileSource() *FileSourceSpec {
	if x != nil {
		return x.FileSource
	}
	return nil
}

func (x *PutFileTask) GetSeed() int64 {
	if x != nil {
		return x.Seed
	}
	return 0
}

func (x *PutFileTask) GetAuthToken() string {
	if x != nil {
		return x.AuthToken
	}
	return ""
}

type PutFileTaskResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FileSetId string `protobuf:"bytes,1,opt,name=file_set_id,json=fileSetId,proto3" json:"file_set_id,omitempty"`
	Hash      []byte `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *PutFileTaskResult) Reset() {
	*x = PutFileTaskResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutFileTaskResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutFileTaskResult) ProtoMessage() {}

func (x *PutFileTaskResult) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutFileTaskResult.ProtoReflect.Descriptor instead.
func (*PutFileTaskResult) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{4}
}

func (x *PutFileTaskResult) GetFileSetId() string {
	if x != nil {
		return x.FileSetId
	}
	return ""
}

func (x *PutFileTaskResult) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

type FileSourceSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string                `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Random *RandomFileSourceSpec `protobuf:"bytes,2,opt,name=random,proto3" json:"random,omitempty"`
}

func (x *FileSourceSpec) Reset() {
	*x = FileSourceSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileSourceSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileSourceSpec) ProtoMessage() {}

func (x *FileSourceSpec) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileSourceSpec.ProtoReflect.Descriptor instead.
func (*FileSourceSpec) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{5}
}

func (x *FileSourceSpec) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *FileSourceSpec) GetRandom() *RandomFileSourceSpec {
	if x != nil {
		return x.Random
	}
	return nil
}

type RandomFileSourceSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Directory     *RandomDirectorySpec `protobuf:"bytes,1,opt,name=directory,proto3" json:"directory,omitempty"`
	Sizes         []*SizeSpec          `protobuf:"bytes,2,rep,name=sizes,proto3" json:"sizes,omitempty"`
	IncrementPath bool                 `protobuf:"varint,3,opt,name=increment_path,json=incrementPath,proto3" json:"increment_path,omitempty"`
}

func (x *RandomFileSourceSpec) Reset() {
	*x = RandomFileSourceSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RandomFileSourceSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RandomFileSourceSpec) ProtoMessage() {}

func (x *RandomFileSourceSpec) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RandomFileSourceSpec.ProtoReflect.Descriptor instead.
func (*RandomFileSourceSpec) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{6}
}

func (x *RandomFileSourceSpec) GetDirectory() *RandomDirectorySpec {
	if x != nil {
		return x.Directory
	}
	return nil
}

func (x *RandomFileSourceSpec) GetSizes() []*SizeSpec {
	if x != nil {
		return x.Sizes
	}
	return nil
}

func (x *RandomFileSourceSpec) GetIncrementPath() bool {
	if x != nil {
		return x.IncrementPath
	}
	return false
}

type RandomDirectorySpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Depth *SizeSpec `protobuf:"bytes,1,opt,name=depth,proto3" json:"depth,omitempty"`
	Run   int64     `protobuf:"varint,2,opt,name=run,proto3" json:"run,omitempty"`
}

func (x *RandomDirectorySpec) Reset() {
	*x = RandomDirectorySpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RandomDirectorySpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RandomDirectorySpec) ProtoMessage() {}

func (x *RandomDirectorySpec) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RandomDirectorySpec.ProtoReflect.Descriptor instead.
func (*RandomDirectorySpec) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{7}
}

func (x *RandomDirectorySpec) GetDepth() *SizeSpec {
	if x != nil {
		return x.Depth
	}
	return nil
}

func (x *RandomDirectorySpec) GetRun() int64 {
	if x != nil {
		return x.Run
	}
	return 0
}

type SizeSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MinSize int64 `protobuf:"varint,1,opt,name=min_size,json=min,proto3" json:"min_size,omitempty"`
	MaxSize int64 `protobuf:"varint,2,opt,name=max_size,json=max,proto3" json:"max_size,omitempty"`
	Prob    int64 `protobuf:"varint,3,opt,name=prob,proto3" json:"prob,omitempty"`
}

func (x *SizeSpec) Reset() {
	*x = SizeSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SizeSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SizeSpec) ProtoMessage() {}

func (x *SizeSpec) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SizeSpec.ProtoReflect.Descriptor instead.
func (*SizeSpec) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{8}
}

func (x *SizeSpec) GetMinSize() int64 {
	if x != nil {
		return x.MinSize
	}
	return 0
}

func (x *SizeSpec) GetMaxSize() int64 {
	if x != nil {
		return x.MaxSize
	}
	return 0
}

func (x *SizeSpec) GetProb() int64 {
	if x != nil {
		return x.Prob
	}
	return 0
}

type ValidatorSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Frequency *FrequencySpec `protobuf:"bytes,1,opt,name=frequency,proto3" json:"frequency,omitempty"`
}

func (x *ValidatorSpec) Reset() {
	*x = ValidatorSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValidatorSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValidatorSpec) ProtoMessage() {}

func (x *ValidatorSpec) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValidatorSpec.ProtoReflect.Descriptor instead.
func (*ValidatorSpec) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{9}
}

func (x *ValidatorSpec) GetFrequency() *FrequencySpec {
	if x != nil {
		return x.Frequency
	}
	return nil
}

type FrequencySpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count int64 `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	Prob  int64 `protobuf:"varint,2,opt,name=prob,proto3" json:"prob,omitempty"`
}

func (x *FrequencySpec) Reset() {
	*x = FrequencySpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FrequencySpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FrequencySpec) ProtoMessage() {}

func (x *FrequencySpec) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FrequencySpec.ProtoReflect.Descriptor instead.
func (*FrequencySpec) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{10}
}

func (x *FrequencySpec) GetCount() int64 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *FrequencySpec) GetProb() int64 {
	if x != nil {
		return x.Prob
	}
	return 0
}

type State struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Commits []*State_Commit `protobuf:"bytes,1,rep,name=commits,proto3" json:"commits,omitempty"`
}

func (x *State) Reset() {
	*x = State{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *State) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*State) ProtoMessage() {}

func (x *State) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use State.ProtoReflect.Descriptor instead.
func (*State) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{11}
}

func (x *State) GetCommits() []*State_Commit {
	if x != nil {
		return x.Commits
	}
	return nil
}

type State_Commit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Commit *pfs.Commit `protobuf:"bytes,1,opt,name=commit,proto3" json:"commit,omitempty"`
	Hash   []byte      `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *State_Commit) Reset() {
	*x = State_Commit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_pfsload_pfsload_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *State_Commit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*State_Commit) ProtoMessage() {}

func (x *State_Commit) ProtoReflect() protoreflect.Message {
	mi := &file_internal_pfsload_pfsload_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use State_Commit.ProtoReflect.Descriptor instead.
func (*State_Commit) Descriptor() ([]byte, []int) {
	return file_internal_pfsload_pfsload_proto_rawDescGZIP(), []int{11, 0}
}

func (x *State_Commit) GetCommit() *pfs.Commit {
	if x != nil {
		return x.Commit
	}
	return nil
}

func (x *State_Commit) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

var File_internal_pfsload_pfsload_proto protoreflect.FileDescriptor

var file_internal_pfsload_pfsload_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x66, 0x73, 0x6c, 0x6f,
	0x61, 0x64, 0x2f, 0x70, 0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x07, 0x70, 0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x1a, 0x0d, 0x70, 0x66, 0x73, 0x2f, 0x70,
	0x66, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd5, 0x01, 0x0a, 0x0a, 0x43, 0x6f, 0x6d,
	0x6d, 0x69, 0x74, 0x53, 0x70, 0x65, 0x63, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x3f, 0x0a,
	0x0d, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x70, 0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x4d,
	0x6f, 0x64, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x70, 0x65, 0x63, 0x52,
	0x0d, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x3a,
	0x0a, 0x0c, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x46,
	0x69, 0x6c, 0x65, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x70, 0x65, 0x63, 0x52, 0x0b, 0x66,
	0x69, 0x6c, 0x65, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x12, 0x34, 0x0a, 0x09, 0x76, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e,
	0x70, 0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f,
	0x72, 0x53, 0x70, 0x65, 0x63, 0x52, 0x09, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72,
	0x22, 0x59, 0x0a, 0x10, 0x4d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x53, 0x70, 0x65, 0x63, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x2f, 0x0a, 0x08, 0x70, 0x75,
	0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x70,
	0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x50, 0x75, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x53, 0x70,
	0x65, 0x63, 0x52, 0x07, 0x70, 0x75, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x22, 0x3b, 0x0a, 0x0b, 0x50,
	0x75, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x53, 0x70, 0x65, 0x63, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x22, 0x90, 0x01, 0x0a, 0x0b, 0x50, 0x75, 0x74,
	0x46, 0x69, 0x6c, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x38,
	0x0a, 0x0b, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x46, 0x69,
	0x6c, 0x65, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x70, 0x65, 0x63, 0x52, 0x0a, 0x66, 0x69,
	0x6c, 0x65, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x65, 0x65, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x65, 0x65, 0x64, 0x12, 0x1d, 0x0a, 0x0a,
	0x61, 0x75, 0x74, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x61, 0x75, 0x74, 0x68, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x47, 0x0a, 0x11, 0x50,
	0x75, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x12, 0x1e, 0x0a, 0x0b, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x73, 0x65, 0x74, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x53, 0x65, 0x74, 0x49, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x68, 0x61, 0x73, 0x68, 0x22, 0x5b, 0x0a, 0x0e, 0x46, 0x69, 0x6c, 0x65, 0x53, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x53, 0x70, 0x65, 0x63, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x35, 0x0a, 0x06, 0x72, 0x61,
	0x6e, 0x64, 0x6f, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x70, 0x66, 0x73,
	0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x46, 0x69, 0x6c, 0x65, 0x53,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x70, 0x65, 0x63, 0x52, 0x06, 0x72, 0x61, 0x6e, 0x64, 0x6f,
	0x6d, 0x22, 0xa2, 0x01, 0x0a, 0x14, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x46, 0x69, 0x6c, 0x65,
	0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x70, 0x65, 0x63, 0x12, 0x3a, 0x0a, 0x09, 0x64, 0x69,
	0x72, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e,
	0x70, 0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x44, 0x69,
	0x72, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x53, 0x70, 0x65, 0x63, 0x52, 0x09, 0x64, 0x69, 0x72,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x27, 0x0a, 0x05, 0x73, 0x69, 0x7a, 0x65, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x70, 0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x2e,
	0x53, 0x69, 0x7a, 0x65, 0x53, 0x70, 0x65, 0x63, 0x52, 0x05, 0x73, 0x69, 0x7a, 0x65, 0x73, 0x12,
	0x25, 0x0a, 0x0e, 0x69, 0x6e, 0x63, 0x72, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x61, 0x74,
	0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x69, 0x6e, 0x63, 0x72, 0x65, 0x6d, 0x65,
	0x6e, 0x74, 0x50, 0x61, 0x74, 0x68, 0x22, 0x50, 0x0a, 0x13, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d,
	0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x53, 0x70, 0x65, 0x63, 0x12, 0x27, 0x0a,
	0x05, 0x64, 0x65, 0x70, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x70,
	0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x53, 0x69, 0x7a, 0x65, 0x53, 0x70, 0x65, 0x63, 0x52,
	0x05, 0x64, 0x65, 0x70, 0x74, 0x68, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x75, 0x6e, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x03, 0x72, 0x75, 0x6e, 0x22, 0x4c, 0x0a, 0x08, 0x53, 0x69, 0x7a, 0x65,
	0x53, 0x70, 0x65, 0x63, 0x12, 0x15, 0x0a, 0x08, 0x6d, 0x69, 0x6e, 0x5f, 0x73, 0x69, 0x7a, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6d, 0x69, 0x6e, 0x12, 0x15, 0x0a, 0x08, 0x6d,
	0x61, 0x78, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6d,
	0x61, 0x78, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x72, 0x6f, 0x62, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x04, 0x70, 0x72, 0x6f, 0x62, 0x22, 0x45, 0x0a, 0x0d, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x6f, 0x72, 0x53, 0x70, 0x65, 0x63, 0x12, 0x34, 0x0a, 0x09, 0x66, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x6e, 0x63, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x70, 0x66, 0x73,
	0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x46, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x53, 0x70,
	0x65, 0x63, 0x52, 0x09, 0x66, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x22, 0x39, 0x0a,
	0x0d, 0x46, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x53, 0x70, 0x65, 0x63, 0x12, 0x14,
	0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x72, 0x6f, 0x62, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x04, 0x70, 0x72, 0x6f, 0x62, 0x22, 0x7e, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x12, 0x2f, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x15, 0x2e, 0x70, 0x66, 0x73, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x69,
	0x74, 0x73, 0x1a, 0x44, 0x0a, 0x06, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x26, 0x0a, 0x06,
	0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70,
	0x66, 0x73, 0x5f, 0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x52, 0x06, 0x63, 0x6f,
	0x6d, 0x6d, 0x69, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x42, 0x38, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x61, 0x63, 0x68, 0x79, 0x64, 0x65, 0x72, 0x6d,
	0x2f, 0x70, 0x61, 0x63, 0x68, 0x79, 0x64, 0x65, 0x72, 0x6d, 0x2f, 0x76, 0x32, 0x2f, 0x73, 0x72,
	0x63, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x66, 0x73, 0x6c, 0x6f,
	0x61, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_pfsload_pfsload_proto_rawDescOnce sync.Once
	file_internal_pfsload_pfsload_proto_rawDescData = file_internal_pfsload_pfsload_proto_rawDesc
)

func file_internal_pfsload_pfsload_proto_rawDescGZIP() []byte {
	file_internal_pfsload_pfsload_proto_rawDescOnce.Do(func() {
		file_internal_pfsload_pfsload_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_pfsload_pfsload_proto_rawDescData)
	})
	return file_internal_pfsload_pfsload_proto_rawDescData
}

var file_internal_pfsload_pfsload_proto_msgTypes = make([]protoimpl.MessageInfo, 13)
var file_internal_pfsload_pfsload_proto_goTypes = []interface{}{
	(*CommitSpec)(nil),           // 0: pfsload.CommitSpec
	(*ModificationSpec)(nil),     // 1: pfsload.ModificationSpec
	(*PutFileSpec)(nil),          // 2: pfsload.PutFileSpec
	(*PutFileTask)(nil),          // 3: pfsload.PutFileTask
	(*PutFileTaskResult)(nil),    // 4: pfsload.PutFileTaskResult
	(*FileSourceSpec)(nil),       // 5: pfsload.FileSourceSpec
	(*RandomFileSourceSpec)(nil), // 6: pfsload.RandomFileSourceSpec
	(*RandomDirectorySpec)(nil),  // 7: pfsload.RandomDirectorySpec
	(*SizeSpec)(nil),             // 8: pfsload.SizeSpec
	(*ValidatorSpec)(nil),        // 9: pfsload.ValidatorSpec
	(*FrequencySpec)(nil),        // 10: pfsload.FrequencySpec
	(*State)(nil),                // 11: pfsload.State
	(*State_Commit)(nil),         // 12: pfsload.State.Commit
	(*pfs.Commit)(nil),           // 13: pfs_v2.Commit
}
var file_internal_pfsload_pfsload_proto_depIdxs = []int32{
	1,  // 0: pfsload.CommitSpec.modifications:type_name -> pfsload.ModificationSpec
	5,  // 1: pfsload.CommitSpec.file_sources:type_name -> pfsload.FileSourceSpec
	9,  // 2: pfsload.CommitSpec.validator:type_name -> pfsload.ValidatorSpec
	2,  // 3: pfsload.ModificationSpec.put_file:type_name -> pfsload.PutFileSpec
	5,  // 4: pfsload.PutFileTask.file_source:type_name -> pfsload.FileSourceSpec
	6,  // 5: pfsload.FileSourceSpec.random:type_name -> pfsload.RandomFileSourceSpec
	7,  // 6: pfsload.RandomFileSourceSpec.directory:type_name -> pfsload.RandomDirectorySpec
	8,  // 7: pfsload.RandomFileSourceSpec.sizes:type_name -> pfsload.SizeSpec
	8,  // 8: pfsload.RandomDirectorySpec.depth:type_name -> pfsload.SizeSpec
	10, // 9: pfsload.ValidatorSpec.frequency:type_name -> pfsload.FrequencySpec
	12, // 10: pfsload.State.commits:type_name -> pfsload.State.Commit
	13, // 11: pfsload.State.Commit.commit:type_name -> pfs_v2.Commit
	12, // [12:12] is the sub-list for method output_type
	12, // [12:12] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_internal_pfsload_pfsload_proto_init() }
func file_internal_pfsload_pfsload_proto_init() {
	if File_internal_pfsload_pfsload_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_pfsload_pfsload_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CommitSpec); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ModificationSpec); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutFileSpec); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutFileTask); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutFileTaskResult); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileSourceSpec); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RandomFileSourceSpec); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RandomDirectorySpec); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SizeSpec); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValidatorSpec); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FrequencySpec); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*State); i {
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
		file_internal_pfsload_pfsload_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*State_Commit); i {
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
			RawDescriptor: file_internal_pfsload_pfsload_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   13,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_pfsload_pfsload_proto_goTypes,
		DependencyIndexes: file_internal_pfsload_pfsload_proto_depIdxs,
		MessageInfos:      file_internal_pfsload_pfsload_proto_msgTypes,
	}.Build()
	File_internal_pfsload_pfsload_proto = out.File
	file_internal_pfsload_pfsload_proto_rawDesc = nil
	file_internal_pfsload_pfsload_proto_goTypes = nil
	file_internal_pfsload_pfsload_proto_depIdxs = nil
}
