// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var client_pfs_pfs_pb = require('../../client/pfs/pfs_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var google_protobuf_wrappers_pb = require('google-protobuf/google/protobuf/wrappers_pb.js');
var gogoproto_gogo_pb = require('../../gogoproto/gogo_pb.js');
var client_auth_auth_pb = require('../../client/auth/auth_pb.js');

function serialize_google_protobuf_BytesValue(arg) {
  if (!(arg instanceof google_protobuf_wrappers_pb.BytesValue)) {
    throw new Error('Expected argument of type google.protobuf.BytesValue');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_google_protobuf_BytesValue(buffer_arg) {
  return google_protobuf_wrappers_pb.BytesValue.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_google_protobuf_Empty(arg) {
  if (!(arg instanceof google_protobuf_empty_pb.Empty)) {
    throw new Error('Expected argument of type google.protobuf.Empty');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_google_protobuf_Empty(buffer_arg) {
  return google_protobuf_empty_pb.Empty.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_Block(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.Block)) {
    throw new Error('Expected argument of type pfs.Block');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_Block(buffer_arg) {
  return client_pfs_pfs_pb.Block.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_BranchInfo(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.BranchInfo)) {
    throw new Error('Expected argument of type pfs.BranchInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_BranchInfo(buffer_arg) {
  return client_pfs_pfs_pb.BranchInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_BranchInfos(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.BranchInfos)) {
    throw new Error('Expected argument of type pfs.BranchInfos');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_BranchInfos(buffer_arg) {
  return client_pfs_pfs_pb.BranchInfos.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_BuildCommitRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.BuildCommitRequest)) {
    throw new Error('Expected argument of type pfs.BuildCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_BuildCommitRequest(buffer_arg) {
  return client_pfs_pfs_pb.BuildCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_CheckObjectRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.CheckObjectRequest)) {
    throw new Error('Expected argument of type pfs.CheckObjectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_CheckObjectRequest(buffer_arg) {
  return client_pfs_pfs_pb.CheckObjectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_CheckObjectResponse(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.CheckObjectResponse)) {
    throw new Error('Expected argument of type pfs.CheckObjectResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_CheckObjectResponse(buffer_arg) {
  return client_pfs_pfs_pb.CheckObjectResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ClearCommitRequestV2(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ClearCommitRequestV2)) {
    throw new Error('Expected argument of type pfs.ClearCommitRequestV2');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ClearCommitRequestV2(buffer_arg) {
  return client_pfs_pfs_pb.ClearCommitRequestV2.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_Commit(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.Commit)) {
    throw new Error('Expected argument of type pfs.Commit');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_Commit(buffer_arg) {
  return client_pfs_pfs_pb.Commit.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_CommitInfo(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.CommitInfo)) {
    throw new Error('Expected argument of type pfs.CommitInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_CommitInfo(buffer_arg) {
  return client_pfs_pfs_pb.CommitInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_CommitInfos(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.CommitInfos)) {
    throw new Error('Expected argument of type pfs.CommitInfos');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_CommitInfos(buffer_arg) {
  return client_pfs_pfs_pb.CommitInfos.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_CopyFileRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.CopyFileRequest)) {
    throw new Error('Expected argument of type pfs.CopyFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_CopyFileRequest(buffer_arg) {
  return client_pfs_pfs_pb.CopyFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_CreateBranchRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.CreateBranchRequest)) {
    throw new Error('Expected argument of type pfs.CreateBranchRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_CreateBranchRequest(buffer_arg) {
  return client_pfs_pfs_pb.CreateBranchRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_CreateObjectRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.CreateObjectRequest)) {
    throw new Error('Expected argument of type pfs.CreateObjectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_CreateObjectRequest(buffer_arg) {
  return client_pfs_pfs_pb.CreateObjectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_CreateRepoRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.CreateRepoRequest)) {
    throw new Error('Expected argument of type pfs.CreateRepoRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_CreateRepoRequest(buffer_arg) {
  return client_pfs_pfs_pb.CreateRepoRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_CreateTmpFileSetResponse(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.CreateTmpFileSetResponse)) {
    throw new Error('Expected argument of type pfs.CreateTmpFileSetResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_CreateTmpFileSetResponse(buffer_arg) {
  return client_pfs_pfs_pb.CreateTmpFileSetResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DeleteBranchRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DeleteBranchRequest)) {
    throw new Error('Expected argument of type pfs.DeleteBranchRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DeleteBranchRequest(buffer_arg) {
  return client_pfs_pfs_pb.DeleteBranchRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DeleteCommitRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DeleteCommitRequest)) {
    throw new Error('Expected argument of type pfs.DeleteCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DeleteCommitRequest(buffer_arg) {
  return client_pfs_pfs_pb.DeleteCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DeleteFileRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DeleteFileRequest)) {
    throw new Error('Expected argument of type pfs.DeleteFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DeleteFileRequest(buffer_arg) {
  return client_pfs_pfs_pb.DeleteFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DeleteObjDirectRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DeleteObjDirectRequest)) {
    throw new Error('Expected argument of type pfs.DeleteObjDirectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DeleteObjDirectRequest(buffer_arg) {
  return client_pfs_pfs_pb.DeleteObjDirectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DeleteObjectsRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DeleteObjectsRequest)) {
    throw new Error('Expected argument of type pfs.DeleteObjectsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DeleteObjectsRequest(buffer_arg) {
  return client_pfs_pfs_pb.DeleteObjectsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DeleteObjectsResponse(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DeleteObjectsResponse)) {
    throw new Error('Expected argument of type pfs.DeleteObjectsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DeleteObjectsResponse(buffer_arg) {
  return client_pfs_pfs_pb.DeleteObjectsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DeleteRepoRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DeleteRepoRequest)) {
    throw new Error('Expected argument of type pfs.DeleteRepoRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DeleteRepoRequest(buffer_arg) {
  return client_pfs_pfs_pb.DeleteRepoRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DeleteTagsRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DeleteTagsRequest)) {
    throw new Error('Expected argument of type pfs.DeleteTagsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DeleteTagsRequest(buffer_arg) {
  return client_pfs_pfs_pb.DeleteTagsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DeleteTagsResponse(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DeleteTagsResponse)) {
    throw new Error('Expected argument of type pfs.DeleteTagsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DeleteTagsResponse(buffer_arg) {
  return client_pfs_pfs_pb.DeleteTagsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DiffFileRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DiffFileRequest)) {
    throw new Error('Expected argument of type pfs.DiffFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DiffFileRequest(buffer_arg) {
  return client_pfs_pfs_pb.DiffFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DiffFileResponse(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DiffFileResponse)) {
    throw new Error('Expected argument of type pfs.DiffFileResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DiffFileResponse(buffer_arg) {
  return client_pfs_pfs_pb.DiffFileResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_DiffFileResponseV2(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.DiffFileResponseV2)) {
    throw new Error('Expected argument of type pfs.DiffFileResponseV2');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_DiffFileResponseV2(buffer_arg) {
  return client_pfs_pfs_pb.DiffFileResponseV2.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_FileInfo(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.FileInfo)) {
    throw new Error('Expected argument of type pfs.FileInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_FileInfo(buffer_arg) {
  return client_pfs_pfs_pb.FileInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_FileInfos(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.FileInfos)) {
    throw new Error('Expected argument of type pfs.FileInfos');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_FileInfos(buffer_arg) {
  return client_pfs_pfs_pb.FileInfos.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_FileOperationRequestV2(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.FileOperationRequestV2)) {
    throw new Error('Expected argument of type pfs.FileOperationRequestV2');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_FileOperationRequestV2(buffer_arg) {
  return client_pfs_pfs_pb.FileOperationRequestV2.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_FinishCommitRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.FinishCommitRequest)) {
    throw new Error('Expected argument of type pfs.FinishCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_FinishCommitRequest(buffer_arg) {
  return client_pfs_pfs_pb.FinishCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_FlushCommitRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.FlushCommitRequest)) {
    throw new Error('Expected argument of type pfs.FlushCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_FlushCommitRequest(buffer_arg) {
  return client_pfs_pfs_pb.FlushCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_FsckRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.FsckRequest)) {
    throw new Error('Expected argument of type pfs.FsckRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_FsckRequest(buffer_arg) {
  return client_pfs_pfs_pb.FsckRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_FsckResponse(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.FsckResponse)) {
    throw new Error('Expected argument of type pfs.FsckResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_FsckResponse(buffer_arg) {
  return client_pfs_pfs_pb.FsckResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_GetBlockRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.GetBlockRequest)) {
    throw new Error('Expected argument of type pfs.GetBlockRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_GetBlockRequest(buffer_arg) {
  return client_pfs_pfs_pb.GetBlockRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_GetBlocksRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.GetBlocksRequest)) {
    throw new Error('Expected argument of type pfs.GetBlocksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_GetBlocksRequest(buffer_arg) {
  return client_pfs_pfs_pb.GetBlocksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_GetFileRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.GetFileRequest)) {
    throw new Error('Expected argument of type pfs.GetFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_GetFileRequest(buffer_arg) {
  return client_pfs_pfs_pb.GetFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_GetObjDirectRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.GetObjDirectRequest)) {
    throw new Error('Expected argument of type pfs.GetObjDirectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_GetObjDirectRequest(buffer_arg) {
  return client_pfs_pfs_pb.GetObjDirectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_GetObjectsRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.GetObjectsRequest)) {
    throw new Error('Expected argument of type pfs.GetObjectsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_GetObjectsRequest(buffer_arg) {
  return client_pfs_pfs_pb.GetObjectsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_GetTarRequestV2(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.GetTarRequestV2)) {
    throw new Error('Expected argument of type pfs.GetTarRequestV2');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_GetTarRequestV2(buffer_arg) {
  return client_pfs_pfs_pb.GetTarRequestV2.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_GlobFileRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.GlobFileRequest)) {
    throw new Error('Expected argument of type pfs.GlobFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_GlobFileRequest(buffer_arg) {
  return client_pfs_pfs_pb.GlobFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_InspectBranchRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.InspectBranchRequest)) {
    throw new Error('Expected argument of type pfs.InspectBranchRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_InspectBranchRequest(buffer_arg) {
  return client_pfs_pfs_pb.InspectBranchRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_InspectCommitRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.InspectCommitRequest)) {
    throw new Error('Expected argument of type pfs.InspectCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_InspectCommitRequest(buffer_arg) {
  return client_pfs_pfs_pb.InspectCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_InspectFileRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.InspectFileRequest)) {
    throw new Error('Expected argument of type pfs.InspectFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_InspectFileRequest(buffer_arg) {
  return client_pfs_pfs_pb.InspectFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_InspectRepoRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.InspectRepoRequest)) {
    throw new Error('Expected argument of type pfs.InspectRepoRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_InspectRepoRequest(buffer_arg) {
  return client_pfs_pfs_pb.InspectRepoRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ListBlockRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ListBlockRequest)) {
    throw new Error('Expected argument of type pfs.ListBlockRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ListBlockRequest(buffer_arg) {
  return client_pfs_pfs_pb.ListBlockRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ListBranchRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ListBranchRequest)) {
    throw new Error('Expected argument of type pfs.ListBranchRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ListBranchRequest(buffer_arg) {
  return client_pfs_pfs_pb.ListBranchRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ListCommitRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ListCommitRequest)) {
    throw new Error('Expected argument of type pfs.ListCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ListCommitRequest(buffer_arg) {
  return client_pfs_pfs_pb.ListCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ListFileRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ListFileRequest)) {
    throw new Error('Expected argument of type pfs.ListFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ListFileRequest(buffer_arg) {
  return client_pfs_pfs_pb.ListFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ListObjectsRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ListObjectsRequest)) {
    throw new Error('Expected argument of type pfs.ListObjectsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ListObjectsRequest(buffer_arg) {
  return client_pfs_pfs_pb.ListObjectsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ListRepoRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ListRepoRequest)) {
    throw new Error('Expected argument of type pfs.ListRepoRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ListRepoRequest(buffer_arg) {
  return client_pfs_pfs_pb.ListRepoRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ListRepoResponse(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ListRepoResponse)) {
    throw new Error('Expected argument of type pfs.ListRepoResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ListRepoResponse(buffer_arg) {
  return client_pfs_pfs_pb.ListRepoResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ListTagsRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ListTagsRequest)) {
    throw new Error('Expected argument of type pfs.ListTagsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ListTagsRequest(buffer_arg) {
  return client_pfs_pfs_pb.ListTagsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ListTagsResponse(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ListTagsResponse)) {
    throw new Error('Expected argument of type pfs.ListTagsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ListTagsResponse(buffer_arg) {
  return client_pfs_pfs_pb.ListTagsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_Object(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.Object)) {
    throw new Error('Expected argument of type pfs.Object');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_Object(buffer_arg) {
  return client_pfs_pfs_pb.Object.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_ObjectInfo(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.ObjectInfo)) {
    throw new Error('Expected argument of type pfs.ObjectInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_ObjectInfo(buffer_arg) {
  return client_pfs_pfs_pb.ObjectInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_Objects(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.Objects)) {
    throw new Error('Expected argument of type pfs.Objects');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_Objects(buffer_arg) {
  return client_pfs_pfs_pb.Objects.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_PutBlockRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.PutBlockRequest)) {
    throw new Error('Expected argument of type pfs.PutBlockRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_PutBlockRequest(buffer_arg) {
  return client_pfs_pfs_pb.PutBlockRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_PutFileRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.PutFileRequest)) {
    throw new Error('Expected argument of type pfs.PutFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_PutFileRequest(buffer_arg) {
  return client_pfs_pfs_pb.PutFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_PutObjDirectRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.PutObjDirectRequest)) {
    throw new Error('Expected argument of type pfs.PutObjDirectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_PutObjDirectRequest(buffer_arg) {
  return client_pfs_pfs_pb.PutObjDirectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_PutObjectRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.PutObjectRequest)) {
    throw new Error('Expected argument of type pfs.PutObjectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_PutObjectRequest(buffer_arg) {
  return client_pfs_pfs_pb.PutObjectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_RenewTmpFileSetRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.RenewTmpFileSetRequest)) {
    throw new Error('Expected argument of type pfs.RenewTmpFileSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_RenewTmpFileSetRequest(buffer_arg) {
  return client_pfs_pfs_pb.RenewTmpFileSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_RepoInfo(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.RepoInfo)) {
    throw new Error('Expected argument of type pfs.RepoInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_RepoInfo(buffer_arg) {
  return client_pfs_pfs_pb.RepoInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_StartCommitRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.StartCommitRequest)) {
    throw new Error('Expected argument of type pfs.StartCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_StartCommitRequest(buffer_arg) {
  return client_pfs_pfs_pb.StartCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_SubscribeCommitRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.SubscribeCommitRequest)) {
    throw new Error('Expected argument of type pfs.SubscribeCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_SubscribeCommitRequest(buffer_arg) {
  return client_pfs_pfs_pb.SubscribeCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_Tag(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.Tag)) {
    throw new Error('Expected argument of type pfs.Tag');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_Tag(buffer_arg) {
  return client_pfs_pfs_pb.Tag.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_TagObjectRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.TagObjectRequest)) {
    throw new Error('Expected argument of type pfs.TagObjectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_TagObjectRequest(buffer_arg) {
  return client_pfs_pfs_pb.TagObjectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_WalkFileRequest(arg) {
  if (!(arg instanceof client_pfs_pfs_pb.WalkFileRequest)) {
    throw new Error('Expected argument of type pfs.WalkFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_WalkFileRequest(buffer_arg) {
  return client_pfs_pfs_pb.WalkFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  // Repo rpcs
// CreateRepo creates a new repo.
// An error is returned if the repo already exists.
createRepo: {
    path: '/pfs.API/CreateRepo',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.CreateRepoRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_CreateRepoRequest,
    requestDeserialize: deserialize_pfs_CreateRepoRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // InspectRepo returns info about a repo.
inspectRepo: {
    path: '/pfs.API/InspectRepo',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.InspectRepoRequest,
    responseType: client_pfs_pfs_pb.RepoInfo,
    requestSerialize: serialize_pfs_InspectRepoRequest,
    requestDeserialize: deserialize_pfs_InspectRepoRequest,
    responseSerialize: serialize_pfs_RepoInfo,
    responseDeserialize: deserialize_pfs_RepoInfo,
  },
  // ListRepo returns info about all repos.
listRepo: {
    path: '/pfs.API/ListRepo',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.ListRepoRequest,
    responseType: client_pfs_pfs_pb.ListRepoResponse,
    requestSerialize: serialize_pfs_ListRepoRequest,
    requestDeserialize: deserialize_pfs_ListRepoRequest,
    responseSerialize: serialize_pfs_ListRepoResponse,
    responseDeserialize: deserialize_pfs_ListRepoResponse,
  },
  // DeleteRepo deletes a repo.
deleteRepo: {
    path: '/pfs.API/DeleteRepo',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.DeleteRepoRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_DeleteRepoRequest,
    requestDeserialize: deserialize_pfs_DeleteRepoRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Commit rpcs
// StartCommit creates a new write commit from a parent commit.
startCommit: {
    path: '/pfs.API/StartCommit',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.StartCommitRequest,
    responseType: client_pfs_pfs_pb.Commit,
    requestSerialize: serialize_pfs_StartCommitRequest,
    requestDeserialize: deserialize_pfs_StartCommitRequest,
    responseSerialize: serialize_pfs_Commit,
    responseDeserialize: deserialize_pfs_Commit,
  },
  // FinishCommit turns a write commit into a read commit.
finishCommit: {
    path: '/pfs.API/FinishCommit',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.FinishCommitRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_FinishCommitRequest,
    requestDeserialize: deserialize_pfs_FinishCommitRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // InspectCommit returns the info about a commit.
inspectCommit: {
    path: '/pfs.API/InspectCommit',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.InspectCommitRequest,
    responseType: client_pfs_pfs_pb.CommitInfo,
    requestSerialize: serialize_pfs_InspectCommitRequest,
    requestDeserialize: deserialize_pfs_InspectCommitRequest,
    responseSerialize: serialize_pfs_CommitInfo,
    responseDeserialize: deserialize_pfs_CommitInfo,
  },
  // ListCommit returns info about all commits. This is deprecated in favor of
// ListCommitStream.
listCommit: {
    path: '/pfs.API/ListCommit',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.ListCommitRequest,
    responseType: client_pfs_pfs_pb.CommitInfos,
    requestSerialize: serialize_pfs_ListCommitRequest,
    requestDeserialize: deserialize_pfs_ListCommitRequest,
    responseSerialize: serialize_pfs_CommitInfos,
    responseDeserialize: deserialize_pfs_CommitInfos,
  },
  // ListCommitStream is like ListCommit, but returns its results in a GRPC stream
listCommitStream: {
    path: '/pfs.API/ListCommitStream',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.ListCommitRequest,
    responseType: client_pfs_pfs_pb.CommitInfo,
    requestSerialize: serialize_pfs_ListCommitRequest,
    requestDeserialize: deserialize_pfs_ListCommitRequest,
    responseSerialize: serialize_pfs_CommitInfo,
    responseDeserialize: deserialize_pfs_CommitInfo,
  },
  // DeleteCommit deletes a commit.
deleteCommit: {
    path: '/pfs.API/DeleteCommit',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.DeleteCommitRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_DeleteCommitRequest,
    requestDeserialize: deserialize_pfs_DeleteCommitRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // FlushCommit waits for downstream commits to finish
flushCommit: {
    path: '/pfs.API/FlushCommit',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.FlushCommitRequest,
    responseType: client_pfs_pfs_pb.CommitInfo,
    requestSerialize: serialize_pfs_FlushCommitRequest,
    requestDeserialize: deserialize_pfs_FlushCommitRequest,
    responseSerialize: serialize_pfs_CommitInfo,
    responseDeserialize: deserialize_pfs_CommitInfo,
  },
  // SubscribeCommit subscribes for new commits on a given branch
subscribeCommit: {
    path: '/pfs.API/SubscribeCommit',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.SubscribeCommitRequest,
    responseType: client_pfs_pfs_pb.CommitInfo,
    requestSerialize: serialize_pfs_SubscribeCommitRequest,
    requestDeserialize: deserialize_pfs_SubscribeCommitRequest,
    responseSerialize: serialize_pfs_CommitInfo,
    responseDeserialize: deserialize_pfs_CommitInfo,
  },
  // BuildCommit builds a commit that's backed by the given tree
buildCommit: {
    path: '/pfs.API/BuildCommit',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.BuildCommitRequest,
    responseType: client_pfs_pfs_pb.Commit,
    requestSerialize: serialize_pfs_BuildCommitRequest,
    requestDeserialize: deserialize_pfs_BuildCommitRequest,
    responseSerialize: serialize_pfs_Commit,
    responseDeserialize: deserialize_pfs_Commit,
  },
  // CreateBranch creates a new branch
createBranch: {
    path: '/pfs.API/CreateBranch',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.CreateBranchRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_CreateBranchRequest,
    requestDeserialize: deserialize_pfs_CreateBranchRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // InspectBranch returns info about a branch.
inspectBranch: {
    path: '/pfs.API/InspectBranch',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.InspectBranchRequest,
    responseType: client_pfs_pfs_pb.BranchInfo,
    requestSerialize: serialize_pfs_InspectBranchRequest,
    requestDeserialize: deserialize_pfs_InspectBranchRequest,
    responseSerialize: serialize_pfs_BranchInfo,
    responseDeserialize: deserialize_pfs_BranchInfo,
  },
  // ListBranch returns info about the heads of branches.
listBranch: {
    path: '/pfs.API/ListBranch',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.ListBranchRequest,
    responseType: client_pfs_pfs_pb.BranchInfos,
    requestSerialize: serialize_pfs_ListBranchRequest,
    requestDeserialize: deserialize_pfs_ListBranchRequest,
    responseSerialize: serialize_pfs_BranchInfos,
    responseDeserialize: deserialize_pfs_BranchInfos,
  },
  // DeleteBranch deletes a branch; note that the commits still exist.
deleteBranch: {
    path: '/pfs.API/DeleteBranch',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.DeleteBranchRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_DeleteBranchRequest,
    requestDeserialize: deserialize_pfs_DeleteBranchRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // File rpcs
// PutFile writes the specified file to pfs.
putFile: {
    path: '/pfs.API/PutFile',
    requestStream: true,
    responseStream: false,
    requestType: client_pfs_pfs_pb.PutFileRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_PutFileRequest,
    requestDeserialize: deserialize_pfs_PutFileRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // CopyFile copies the contents of one file to another.
copyFile: {
    path: '/pfs.API/CopyFile',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.CopyFileRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_CopyFileRequest,
    requestDeserialize: deserialize_pfs_CopyFileRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // GetFile returns a byte stream of the contents of the file.
getFile: {
    path: '/pfs.API/GetFile',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.GetFileRequest,
    responseType: google_protobuf_wrappers_pb.BytesValue,
    requestSerialize: serialize_pfs_GetFileRequest,
    requestDeserialize: deserialize_pfs_GetFileRequest,
    responseSerialize: serialize_google_protobuf_BytesValue,
    responseDeserialize: deserialize_google_protobuf_BytesValue,
  },
  // InspectFile returns info about a file.
inspectFile: {
    path: '/pfs.API/InspectFile',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.InspectFileRequest,
    responseType: client_pfs_pfs_pb.FileInfo,
    requestSerialize: serialize_pfs_InspectFileRequest,
    requestDeserialize: deserialize_pfs_InspectFileRequest,
    responseSerialize: serialize_pfs_FileInfo,
    responseDeserialize: deserialize_pfs_FileInfo,
  },
  // ListFile returns info about all files. This is deprecated in favor of
// ListFileStream
listFile: {
    path: '/pfs.API/ListFile',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.ListFileRequest,
    responseType: client_pfs_pfs_pb.FileInfos,
    requestSerialize: serialize_pfs_ListFileRequest,
    requestDeserialize: deserialize_pfs_ListFileRequest,
    responseSerialize: serialize_pfs_FileInfos,
    responseDeserialize: deserialize_pfs_FileInfos,
  },
  // ListFileStream is a streaming version of ListFile
// TODO(msteffen): When the dash has been updated to use ListFileStream,
// replace ListFile with this RPC (https://github.com/pachyderm/dash/issues/201)
listFileStream: {
    path: '/pfs.API/ListFileStream',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.ListFileRequest,
    responseType: client_pfs_pfs_pb.FileInfo,
    requestSerialize: serialize_pfs_ListFileRequest,
    requestDeserialize: deserialize_pfs_ListFileRequest,
    responseSerialize: serialize_pfs_FileInfo,
    responseDeserialize: deserialize_pfs_FileInfo,
  },
  // WalkFile walks over all the files under a directory, including children of children.
walkFile: {
    path: '/pfs.API/WalkFile',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.WalkFileRequest,
    responseType: client_pfs_pfs_pb.FileInfo,
    requestSerialize: serialize_pfs_WalkFileRequest,
    requestDeserialize: deserialize_pfs_WalkFileRequest,
    responseSerialize: serialize_pfs_FileInfo,
    responseDeserialize: deserialize_pfs_FileInfo,
  },
  // GlobFile returns info about all files. This is deprecated in favor of
// GlobFileStream
globFile: {
    path: '/pfs.API/GlobFile',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.GlobFileRequest,
    responseType: client_pfs_pfs_pb.FileInfos,
    requestSerialize: serialize_pfs_GlobFileRequest,
    requestDeserialize: deserialize_pfs_GlobFileRequest,
    responseSerialize: serialize_pfs_FileInfos,
    responseDeserialize: deserialize_pfs_FileInfos,
  },
  // GlobFileStream is a streaming version of GlobFile
// TODO(msteffen): When the dash has been updated to use GlobFileStream,
// replace GlobFile with this RPC (https://github.com/pachyderm/dash/issues/201)
globFileStream: {
    path: '/pfs.API/GlobFileStream',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.GlobFileRequest,
    responseType: client_pfs_pfs_pb.FileInfo,
    requestSerialize: serialize_pfs_GlobFileRequest,
    requestDeserialize: deserialize_pfs_GlobFileRequest,
    responseSerialize: serialize_pfs_FileInfo,
    responseDeserialize: deserialize_pfs_FileInfo,
  },
  // DiffFile returns the differences between 2 paths at 2 commits.
diffFile: {
    path: '/pfs.API/DiffFile',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.DiffFileRequest,
    responseType: client_pfs_pfs_pb.DiffFileResponse,
    requestSerialize: serialize_pfs_DiffFileRequest,
    requestDeserialize: deserialize_pfs_DiffFileRequest,
    responseSerialize: serialize_pfs_DiffFileResponse,
    responseDeserialize: deserialize_pfs_DiffFileResponse,
  },
  // DeleteFile deletes a file.
deleteFile: {
    path: '/pfs.API/DeleteFile',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.DeleteFileRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_DeleteFileRequest,
    requestDeserialize: deserialize_pfs_DeleteFileRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // DeleteAll deletes everything
deleteAll: {
    path: '/pfs.API/DeleteAll',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Fsck does a file system consistency check for pfs
fsck: {
    path: '/pfs.API/Fsck',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.FsckRequest,
    responseType: client_pfs_pfs_pb.FsckResponse,
    requestSerialize: serialize_pfs_FsckRequest,
    requestDeserialize: deserialize_pfs_FsckRequest,
    responseSerialize: serialize_pfs_FsckResponse,
    responseDeserialize: deserialize_pfs_FsckResponse,
  },
  // RPCs specific to Pachyderm 2.
fileOperationV2: {
    path: '/pfs.API/FileOperationV2',
    requestStream: true,
    responseStream: false,
    requestType: client_pfs_pfs_pb.FileOperationRequestV2,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_FileOperationRequestV2,
    requestDeserialize: deserialize_pfs_FileOperationRequestV2,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  getTarV2: {
    path: '/pfs.API/GetTarV2',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.GetTarRequestV2,
    responseType: google_protobuf_wrappers_pb.BytesValue,
    requestSerialize: serialize_pfs_GetTarRequestV2,
    requestDeserialize: deserialize_pfs_GetTarRequestV2,
    responseSerialize: serialize_google_protobuf_BytesValue,
    responseDeserialize: deserialize_google_protobuf_BytesValue,
  },
  // DiffFileV2 returns the differences between 2 paths at 2 commits.
// it streams back one file at a time which is either from the new path, or the old path
diffFileV2: {
    path: '/pfs.API/DiffFileV2',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.DiffFileRequest,
    responseType: client_pfs_pfs_pb.DiffFileResponseV2,
    requestSerialize: serialize_pfs_DiffFileRequest,
    requestDeserialize: deserialize_pfs_DiffFileRequest,
    responseSerialize: serialize_pfs_DiffFileResponseV2,
    responseDeserialize: deserialize_pfs_DiffFileResponseV2,
  },
  // CreateTmpFileSet creates a new temp fileset
createTmpFileSet: {
    path: '/pfs.API/CreateTmpFileSet',
    requestStream: true,
    responseStream: false,
    requestType: client_pfs_pfs_pb.FileOperationRequestV2,
    responseType: client_pfs_pfs_pb.CreateTmpFileSetResponse,
    requestSerialize: serialize_pfs_FileOperationRequestV2,
    requestDeserialize: deserialize_pfs_FileOperationRequestV2,
    responseSerialize: serialize_pfs_CreateTmpFileSetResponse,
    responseDeserialize: deserialize_pfs_CreateTmpFileSetResponse,
  },
  // RenewTmpFileSet prevents the temporary fileset from being deleted for a set amount of time
renewTmpFileSet: {
    path: '/pfs.API/RenewTmpFileSet',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.RenewTmpFileSetRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_RenewTmpFileSetRequest,
    requestDeserialize: deserialize_pfs_RenewTmpFileSetRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // ClearCommitV2 removes all data from the commit.
clearCommitV2: {
    path: '/pfs.API/ClearCommitV2',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.ClearCommitRequestV2,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_ClearCommitRequestV2,
    requestDeserialize: deserialize_pfs_ClearCommitRequestV2,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
var ObjectAPIService = exports.ObjectAPIService = {
  putObject: {
    path: '/pfs.ObjectAPI/PutObject',
    requestStream: true,
    responseStream: false,
    requestType: client_pfs_pfs_pb.PutObjectRequest,
    responseType: client_pfs_pfs_pb.Object,
    requestSerialize: serialize_pfs_PutObjectRequest,
    requestDeserialize: deserialize_pfs_PutObjectRequest,
    responseSerialize: serialize_pfs_Object,
    responseDeserialize: deserialize_pfs_Object,
  },
  putObjectSplit: {
    path: '/pfs.ObjectAPI/PutObjectSplit',
    requestStream: true,
    responseStream: false,
    requestType: client_pfs_pfs_pb.PutObjectRequest,
    responseType: client_pfs_pfs_pb.Objects,
    requestSerialize: serialize_pfs_PutObjectRequest,
    requestDeserialize: deserialize_pfs_PutObjectRequest,
    responseSerialize: serialize_pfs_Objects,
    responseDeserialize: deserialize_pfs_Objects,
  },
  putObjects: {
    path: '/pfs.ObjectAPI/PutObjects',
    requestStream: true,
    responseStream: false,
    requestType: client_pfs_pfs_pb.PutObjectRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_PutObjectRequest,
    requestDeserialize: deserialize_pfs_PutObjectRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  createObject: {
    path: '/pfs.ObjectAPI/CreateObject',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.CreateObjectRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_CreateObjectRequest,
    requestDeserialize: deserialize_pfs_CreateObjectRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  getObject: {
    path: '/pfs.ObjectAPI/GetObject',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.Object,
    responseType: google_protobuf_wrappers_pb.BytesValue,
    requestSerialize: serialize_pfs_Object,
    requestDeserialize: deserialize_pfs_Object,
    responseSerialize: serialize_google_protobuf_BytesValue,
    responseDeserialize: deserialize_google_protobuf_BytesValue,
  },
  getObjects: {
    path: '/pfs.ObjectAPI/GetObjects',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.GetObjectsRequest,
    responseType: google_protobuf_wrappers_pb.BytesValue,
    requestSerialize: serialize_pfs_GetObjectsRequest,
    requestDeserialize: deserialize_pfs_GetObjectsRequest,
    responseSerialize: serialize_google_protobuf_BytesValue,
    responseDeserialize: deserialize_google_protobuf_BytesValue,
  },
  putBlock: {
    path: '/pfs.ObjectAPI/PutBlock',
    requestStream: true,
    responseStream: false,
    requestType: client_pfs_pfs_pb.PutBlockRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_PutBlockRequest,
    requestDeserialize: deserialize_pfs_PutBlockRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  getBlock: {
    path: '/pfs.ObjectAPI/GetBlock',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.GetBlockRequest,
    responseType: google_protobuf_wrappers_pb.BytesValue,
    requestSerialize: serialize_pfs_GetBlockRequest,
    requestDeserialize: deserialize_pfs_GetBlockRequest,
    responseSerialize: serialize_google_protobuf_BytesValue,
    responseDeserialize: deserialize_google_protobuf_BytesValue,
  },
  getBlocks: {
    path: '/pfs.ObjectAPI/GetBlocks',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.GetBlocksRequest,
    responseType: google_protobuf_wrappers_pb.BytesValue,
    requestSerialize: serialize_pfs_GetBlocksRequest,
    requestDeserialize: deserialize_pfs_GetBlocksRequest,
    responseSerialize: serialize_google_protobuf_BytesValue,
    responseDeserialize: deserialize_google_protobuf_BytesValue,
  },
  listBlock: {
    path: '/pfs.ObjectAPI/ListBlock',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.ListBlockRequest,
    responseType: client_pfs_pfs_pb.Block,
    requestSerialize: serialize_pfs_ListBlockRequest,
    requestDeserialize: deserialize_pfs_ListBlockRequest,
    responseSerialize: serialize_pfs_Block,
    responseDeserialize: deserialize_pfs_Block,
  },
  tagObject: {
    path: '/pfs.ObjectAPI/TagObject',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.TagObjectRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_TagObjectRequest,
    requestDeserialize: deserialize_pfs_TagObjectRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  inspectObject: {
    path: '/pfs.ObjectAPI/InspectObject',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.Object,
    responseType: client_pfs_pfs_pb.ObjectInfo,
    requestSerialize: serialize_pfs_Object,
    requestDeserialize: deserialize_pfs_Object,
    responseSerialize: serialize_pfs_ObjectInfo,
    responseDeserialize: deserialize_pfs_ObjectInfo,
  },
  // CheckObject checks if an object exists in the blob store without
// actually reading the object.
checkObject: {
    path: '/pfs.ObjectAPI/CheckObject',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.CheckObjectRequest,
    responseType: client_pfs_pfs_pb.CheckObjectResponse,
    requestSerialize: serialize_pfs_CheckObjectRequest,
    requestDeserialize: deserialize_pfs_CheckObjectRequest,
    responseSerialize: serialize_pfs_CheckObjectResponse,
    responseDeserialize: deserialize_pfs_CheckObjectResponse,
  },
  listObjects: {
    path: '/pfs.ObjectAPI/ListObjects',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.ListObjectsRequest,
    responseType: client_pfs_pfs_pb.ObjectInfo,
    requestSerialize: serialize_pfs_ListObjectsRequest,
    requestDeserialize: deserialize_pfs_ListObjectsRequest,
    responseSerialize: serialize_pfs_ObjectInfo,
    responseDeserialize: deserialize_pfs_ObjectInfo,
  },
  deleteObjects: {
    path: '/pfs.ObjectAPI/DeleteObjects',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.DeleteObjectsRequest,
    responseType: client_pfs_pfs_pb.DeleteObjectsResponse,
    requestSerialize: serialize_pfs_DeleteObjectsRequest,
    requestDeserialize: deserialize_pfs_DeleteObjectsRequest,
    responseSerialize: serialize_pfs_DeleteObjectsResponse,
    responseDeserialize: deserialize_pfs_DeleteObjectsResponse,
  },
  getTag: {
    path: '/pfs.ObjectAPI/GetTag',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.Tag,
    responseType: google_protobuf_wrappers_pb.BytesValue,
    requestSerialize: serialize_pfs_Tag,
    requestDeserialize: deserialize_pfs_Tag,
    responseSerialize: serialize_google_protobuf_BytesValue,
    responseDeserialize: deserialize_google_protobuf_BytesValue,
  },
  inspectTag: {
    path: '/pfs.ObjectAPI/InspectTag',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.Tag,
    responseType: client_pfs_pfs_pb.ObjectInfo,
    requestSerialize: serialize_pfs_Tag,
    requestDeserialize: deserialize_pfs_Tag,
    responseSerialize: serialize_pfs_ObjectInfo,
    responseDeserialize: deserialize_pfs_ObjectInfo,
  },
  listTags: {
    path: '/pfs.ObjectAPI/ListTags',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.ListTagsRequest,
    responseType: client_pfs_pfs_pb.ListTagsResponse,
    requestSerialize: serialize_pfs_ListTagsRequest,
    requestDeserialize: deserialize_pfs_ListTagsRequest,
    responseSerialize: serialize_pfs_ListTagsResponse,
    responseDeserialize: deserialize_pfs_ListTagsResponse,
  },
  deleteTags: {
    path: '/pfs.ObjectAPI/DeleteTags',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.DeleteTagsRequest,
    responseType: client_pfs_pfs_pb.DeleteTagsResponse,
    requestSerialize: serialize_pfs_DeleteTagsRequest,
    requestDeserialize: deserialize_pfs_DeleteTagsRequest,
    responseSerialize: serialize_pfs_DeleteTagsResponse,
    responseDeserialize: deserialize_pfs_DeleteTagsResponse,
  },
  compact: {
    path: '/pfs.ObjectAPI/Compact',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // PutObjDirect puts an obj directly into object store, bypassing the content
// addressing layer.
putObjDirect: {
    path: '/pfs.ObjectAPI/PutObjDirect',
    requestStream: true,
    responseStream: false,
    requestType: client_pfs_pfs_pb.PutObjDirectRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_PutObjDirectRequest,
    requestDeserialize: deserialize_pfs_PutObjDirectRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // GetObjDirect gets an obj directly out of object store, bypassing the
// content addressing layer.
getObjDirect: {
    path: '/pfs.ObjectAPI/GetObjDirect',
    requestStream: false,
    responseStream: true,
    requestType: client_pfs_pfs_pb.GetObjDirectRequest,
    responseType: google_protobuf_wrappers_pb.BytesValue,
    requestSerialize: serialize_pfs_GetObjDirectRequest,
    requestDeserialize: deserialize_pfs_GetObjDirectRequest,
    responseSerialize: serialize_google_protobuf_BytesValue,
    responseDeserialize: deserialize_google_protobuf_BytesValue,
  },
  deleteObjDirect: {
    path: '/pfs.ObjectAPI/DeleteObjDirect',
    requestStream: false,
    responseStream: false,
    requestType: client_pfs_pfs_pb.DeleteObjDirectRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_DeleteObjDirectRequest,
    requestDeserialize: deserialize_pfs_DeleteObjDirectRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
};

exports.ObjectAPIClient = grpc.makeGenericClientConstructor(ObjectAPIService);
