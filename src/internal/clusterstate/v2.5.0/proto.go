// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.11.4
// source: pfs/pfs.proto

package v2_5_0

import (
	reflect "reflect"
	sync "sync"

	duration "github.com/golang/protobuf/ptypes/duration"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)


// CommitInfo is the main data structure representing a commit in etcd
type CommitInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Commit *pfs.Commit       `protobuf:"bytes,1,opt,name=commit,proto3" json:"commit,omitempty"`
	Origin *pfs.CommitOrigin `protobuf:"bytes,2,opt,name=origin,proto3" json:"origin,omitempty"`
	// description is a user-provided script describing this commit
	Description         string               `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	ParentCommit        *pfs.Commit              `protobuf:"bytes,4,opt,name=parent_commit,json=parentCommit,proto3" json:"parent_commit,omitempty"`
	ChildCommits        []*pfs.Commit            `protobuf:"bytes,5,rep,name=child_commits,json=childCommits,proto3" json:"child_commits,omitempty"`
	Started             *timestamp.Timestamp `protobuf:"bytes,6,opt,name=started,proto3" json:"started,omitempty"`
	Finishing           *timestamp.Timestamp `protobuf:"bytes,7,opt,name=finishing,proto3" json:"finishing,omitempty"`
	Finished            *timestamp.Timestamp `protobuf:"bytes,8,opt,name=finished,proto3" json:"finished,omitempty"`
	DirectProvenance    []*pfs.Branch            `protobuf:"bytes,9,rep,name=direct_provenance,json=directProvenance,proto3" json:"direct_provenance,omitempty"`
	Error               string               `protobuf:"bytes,10,opt,name=error,proto3" json:"error,omitempty"`
	SizeBytesUpperBound int64                `protobuf:"varint,11,opt,name=size_bytes_upper_bound,json=sizeBytesUpperBound,proto3" json:"size_bytes_upper_bound,omitempty"`
	Details             *pfs.CommitInfo_Details  `protobuf:"bytes,12,opt,name=details,proto3" json:"details,omitempty"`
}

func (x *CommitInfo) Reset() {
	*x = CommitInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pfs_pfs_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CommitInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CommitInfo) ProtoMessage() {}

func (x *CommitInfo) ProtoReflect() protoreflect.Message {
	mi := &file_pfs_pfs_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CommitInfo.ProtoReflect.Descriptor instead.
func (*CommitInfo) Descriptor() ([]byte, []int) {
	return file_pfs_pfs_proto_rawDescGZIP(), []int{4}
}

func (x *CommitInfo) GetCommit() *pfs.Commit {
	if x != nil {
		return x.Commit
	}
	return nil
}

func (x *CommitInfo) GetOrigin() *pfs.CommitOrigin {
	if x != nil {
		return x.Origin
	}
	return nil
}

func (x *CommitInfo) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *CommitInfo) GetParentCommit() *pfs.Commit {
	if x != nil {
		return x.ParentCommit
	}
	return nil
}

func (x *CommitInfo) GetChildCommits() []*pfs.Commit {
	if x != nil {
		return x.ChildCommits
	}
	return nil
}

func (x *CommitInfo) GetStarted() *timestamp.Timestamp {
	if x != nil {
		return x.Started
	}
	return nil
}

func (x *CommitInfo) GetFinishing() *timestamp.Timestamp {
	if x != nil {
		return x.Finishing
	}
	return nil
}

func (x *CommitInfo) GetFinished() *timestamp.Timestamp {
	if x != nil {
		return x.Finished
	}
	return nil
}

func (x *CommitInfo) GetDirectProvenance() []*pfs.Branch {
	if x != nil {
		return x.DirectProvenance
	}
	return nil
}

func (x *CommitInfo) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *CommitInfo) GetSizeBytesUpperBound() int64 {
	if x != nil {
		return x.SizeBytesUpperBound
	}
	return 0
}

func (x *CommitInfo) GetDetails() *pfs.CommitInfo_Details {
	if x != nil {
		return x.Details
	}
	return nil
}

var File_pfs_pfs_proto protoreflect.FileDescriptor

var file_pfs_pfs_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x70, 0x66, 0x73, 0x2f, 0x70, 0x66, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x06, 0x70, 0x66, 0x73, 0x5f, 0x76, 0x32, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x59, 0x0a, 0x04, 0x52, 0x65, 0x70, 0x6f,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x29, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x66, 0x73, 0x5f,
	0x76, 0x32, 0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x22, 0x3e, 0x0a, 0x06, 0x42, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x12, 0x20, 0x0a,
	0x04, 0x72, 0x65, 0x70, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70, 0x66,
	0x73, 0x5f, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x52, 0x04, 0x72, 0x65, 0x70, 0x6f, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x22, 0x36, 0x0a, 0x0c, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x4f, 0x72, 0x69,
	0x67, 0x69, 0x6e, 0x12, 0x26, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x12, 0x2e, 0x70, 0x66, 0x73, 0x5f, 0x76, 0x32, 0x2e, 0x4f, 0x72, 0x69, 0x67, 0x69,
	0x6e, 0x4b, 0x69, 0x6e, 0x64, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x22, 0x40, 0x0a, 0x06, 0x43,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x26, 0x0a, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x66, 0x73, 0x5f, 0x76, 0x32, 0x2e, 0x42,
	0x72, 0x61, 0x6e, 0x63, 0x68, 0x52, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x87, 0x06,
	0x0a, 0x0a, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x26, 0x0a, 0x06,
	0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70,
	0x66, 0x73, 0x5f, 0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x52, 0x06, 0x63, 0x6f,
	0x6d, 0x6d, 0x69, 0x74, 0x12, 0x2c, 0x0a, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x70, 0x66, 0x73, 0x5f, 0x76, 0x32, 0x2e, 0x43, 0x6f,
	0x6d, 0x6d, 0x69, 0x74, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x52, 0x06, 0x6f, 0x72, 0x69, 0x67,
	0x69, 0x6e, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x33, 0x0a, 0x0d, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x63,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x66,
	0x73, 0x5f, 0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x52, 0x0c, 0x70, 0x61, 0x72,
	0x65, 0x6e, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12, 0x33, 0x0a, 0x0d, 0x63, 0x68, 0x69,
	0x6c, 0x64, 0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x0e, 0x2e, 0x70, 0x66, 0x73, 0x5f, 0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74,
	0x52, 0x0c, 0x63, 0x68, 0x69, 0x6c, 0x64, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x73, 0x12, 0x34,
	0x0a, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x07, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x65, 0x64, 0x12, 0x38, 0x0a, 0x09, 0x66, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x69, 0x6e,
	0x67, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x09, 0x66, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x69, 0x6e, 0x67, 0x12, 0x36,
	0x0a, 0x08, 0x66, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x65, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x08, 0x66, 0x69,
	0x6e, 0x69, 0x73, 0x68, 0x65, 0x64, 0x12, 0x3b, 0x0a, 0x11, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74,
	0x5f, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x18, 0x09, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x66, 0x73, 0x5f, 0x76, 0x32, 0x2e, 0x42, 0x72, 0x61, 0x6e, 0x63,
	0x68, 0x52, 0x10, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x50, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61,
	0x6e, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x33, 0x0a, 0x16, 0x73, 0x69, 0x7a,
	0x65, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x5f, 0x75, 0x70, 0x70, 0x65, 0x72, 0x5f, 0x62, 0x6f,
	0x75, 0x6e, 0x64, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x03, 0x52, 0x13, 0x73, 0x69, 0x7a, 0x65, 0x42,
	0x79, 0x74, 0x65, 0x73, 0x55, 0x70, 0x70, 0x65, 0x72, 0x42, 0x6f, 0x75, 0x6e, 0x64, 0x12, 0x34,
	0x0a, 0x07, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x70, 0x66, 0x73, 0x5f, 0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x49,
	0x6e, 0x66, 0x6f, 0x2e, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x07, 0x64, 0x65, 0x74,
	0x61, 0x69, 0x6c, 0x73, 0x1a, 0xb0, 0x01, 0x0a, 0x07, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73,
	0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x69, 0x7a, 0x65, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73, 0x69, 0x7a, 0x65, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12,
	0x42, 0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x63, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x0e, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x63, 0x74, 0x69, 0x6e, 0x67, 0x54,
	0x69, 0x6d, 0x65, 0x12, 0x42, 0x0a, 0x0f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6e,
	0x67, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x69, 0x6e, 0x67, 0x54, 0x69, 0x6d, 0x65, 0x22, 0x1d, 0x0a, 0x07, 0x50, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x2a, 0x4e, 0x0a, 0x0a, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e,
	0x4b, 0x69, 0x6e, 0x64, 0x12, 0x17, 0x0a, 0x13, 0x4f, 0x52, 0x49, 0x47, 0x49, 0x4e, 0x5f, 0x4b,
	0x49, 0x4e, 0x44, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x08, 0x0a,
	0x04, 0x55, 0x53, 0x45, 0x52, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x41, 0x55, 0x54, 0x4f, 0x10,
	0x02, 0x12, 0x08, 0x0a, 0x04, 0x46, 0x53, 0x43, 0x4b, 0x10, 0x03, 0x12, 0x09, 0x0a, 0x05, 0x41,
	0x4c, 0x49, 0x41, 0x53, 0x10, 0x04, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x61, 0x63, 0x68, 0x79, 0x64, 0x65, 0x72, 0x6d, 0x2f, 0x70,
	0x61, 0x63, 0x68, 0x79, 0x64, 0x65, 0x72, 0x6d, 0x2f, 0x76, 0x32, 0x2f, 0x73, 0x72, 0x63, 0x2f,
	0x70, 0x66, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pfs_pfs_proto_rawDescOnce sync.Once
	file_pfs_pfs_proto_rawDescData = file_pfs_pfs_proto_rawDesc
)

func file_pfs_pfs_proto_rawDescGZIP() []byte {
	file_pfs_pfs_proto_rawDescOnce.Do(func() {
		file_pfs_pfs_proto_rawDescData = protoimpl.X.CompressGZIP(file_pfs_pfs_proto_rawDescData)
	})
	return file_pfs_pfs_proto_rawDescData
}

var file_pfs_pfs_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_pfs_pfs_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_pfs_pfs_proto_goTypes = []interface{}{
	(OriginKind)(0),             // 0: pfs_v2.OriginKind
	(*Repo)(nil),                // 1: pfs_v2.Repo
	(*Branch)(nil),              // 2: pfs_v2.Branch
	(*CommitOrigin)(nil),        // 3: pfs_v2.CommitOrigin
	(*Commit)(nil),              // 4: pfs_v2.Commit
	(*CommitInfo)(nil),          // 5: pfs_v2.CommitInfo
	(*Project)(nil),             // 6: pfs_v2.Project
	(*CommitInfo_Details)(nil),  // 7: pfs_v2.CommitInfo.Details
	(*timestamp.Timestamp)(nil), // 8: google.protobuf.Timestamp
	(*duration.Duration)(nil),   // 9: google.protobuf.Duration
}
var file_pfs_pfs_proto_depIdxs = []int32{
	6,  // 0: pfs_v2.Repo.project:type_name -> pfs_v2.Project
	1,  // 1: pfs_v2.Branch.repo:type_name -> pfs_v2.Repo
	0,  // 2: pfs_v2.CommitOrigin.kind:type_name -> pfs_v2.OriginKind
	2,  // 3: pfs_v2.Commit.branch:type_name -> pfs_v2.Branch
	4,  // 4: pfs_v2.CommitInfo.commit:type_name -> pfs_v2.Commit
	3,  // 5: pfs_v2.CommitInfo.origin:type_name -> pfs_v2.CommitOrigin
	4,  // 6: pfs_v2.CommitInfo.parent_commit:type_name -> pfs_v2.Commit
	4,  // 7: pfs_v2.CommitInfo.child_commits:type_name -> pfs_v2.Commit
	8,  // 8: pfs_v2.CommitInfo.started:type_name -> google.protobuf.Timestamp
	8,  // 9: pfs_v2.CommitInfo.finishing:type_name -> google.protobuf.Timestamp
	8,  // 10: pfs_v2.CommitInfo.finished:type_name -> google.protobuf.Timestamp
	2,  // 11: pfs_v2.CommitInfo.direct_provenance:type_name -> pfs_v2.Branch
	7,  // 12: pfs_v2.CommitInfo.details:type_name -> pfs_v2.CommitInfo.Details
	9,  // 13: pfs_v2.CommitInfo.Details.compacting_time:type_name -> google.protobuf.Duration
	9,  // 14: pfs_v2.CommitInfo.Details.validating_time:type_name -> google.protobuf.Duration
	15, // [15:15] is the sub-list for method output_type
	15, // [15:15] is the sub-list for method input_type
	15, // [15:15] is the sub-list for extension type_name
	15, // [15:15] is the sub-list for extension extendee
	0,  // [0:15] is the sub-list for field type_name
}
