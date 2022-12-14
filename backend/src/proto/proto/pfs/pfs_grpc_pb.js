// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var pfs_pfs_pb = require('../pfs/pfs_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var google_protobuf_wrappers_pb = require('google-protobuf/google/protobuf/wrappers_pb.js');
var google_protobuf_duration_pb = require('google-protobuf/google/protobuf/duration_pb.js');
var google_protobuf_any_pb = require('google-protobuf/google/protobuf/any_pb.js');
var gogoproto_gogo_pb = require('../gogoproto/gogo_pb.js');
var auth_auth_pb = require('../auth/auth_pb.js');
var task_task_pb = require('../task/task_pb.js');

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

function serialize_pfs_v2_ActivateAuthRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ActivateAuthRequest)) {
    throw new Error('Expected argument of type pfs_v2.ActivateAuthRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ActivateAuthRequest(buffer_arg) {
  return pfs_pfs_pb.ActivateAuthRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ActivateAuthResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.ActivateAuthResponse)) {
    throw new Error('Expected argument of type pfs_v2.ActivateAuthResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ActivateAuthResponse(buffer_arg) {
  return pfs_pfs_pb.ActivateAuthResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_AddFileSetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.AddFileSetRequest)) {
    throw new Error('Expected argument of type pfs_v2.AddFileSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_AddFileSetRequest(buffer_arg) {
  return pfs_pfs_pb.AddFileSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_BranchInfo(arg) {
  if (!(arg instanceof pfs_pfs_pb.BranchInfo)) {
    throw new Error('Expected argument of type pfs_v2.BranchInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_BranchInfo(buffer_arg) {
  return pfs_pfs_pb.BranchInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_CheckStorageRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.CheckStorageRequest)) {
    throw new Error('Expected argument of type pfs_v2.CheckStorageRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_CheckStorageRequest(buffer_arg) {
  return pfs_pfs_pb.CheckStorageRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_CheckStorageResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.CheckStorageResponse)) {
    throw new Error('Expected argument of type pfs_v2.CheckStorageResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_CheckStorageResponse(buffer_arg) {
  return pfs_pfs_pb.CheckStorageResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ClearCacheRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ClearCacheRequest)) {
    throw new Error('Expected argument of type pfs_v2.ClearCacheRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ClearCacheRequest(buffer_arg) {
  return pfs_pfs_pb.ClearCacheRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ClearCommitRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ClearCommitRequest)) {
    throw new Error('Expected argument of type pfs_v2.ClearCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ClearCommitRequest(buffer_arg) {
  return pfs_pfs_pb.ClearCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_Commit(arg) {
  if (!(arg instanceof pfs_pfs_pb.Commit)) {
    throw new Error('Expected argument of type pfs_v2.Commit');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_Commit(buffer_arg) {
  return pfs_pfs_pb.Commit.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_CommitInfo(arg) {
  if (!(arg instanceof pfs_pfs_pb.CommitInfo)) {
    throw new Error('Expected argument of type pfs_v2.CommitInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_CommitInfo(buffer_arg) {
  return pfs_pfs_pb.CommitInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_CommitSetInfo(arg) {
  if (!(arg instanceof pfs_pfs_pb.CommitSetInfo)) {
    throw new Error('Expected argument of type pfs_v2.CommitSetInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_CommitSetInfo(buffer_arg) {
  return pfs_pfs_pb.CommitSetInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ComposeFileSetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ComposeFileSetRequest)) {
    throw new Error('Expected argument of type pfs_v2.ComposeFileSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ComposeFileSetRequest(buffer_arg) {
  return pfs_pfs_pb.ComposeFileSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_CreateBranchRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.CreateBranchRequest)) {
    throw new Error('Expected argument of type pfs_v2.CreateBranchRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_CreateBranchRequest(buffer_arg) {
  return pfs_pfs_pb.CreateBranchRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_CreateFileSetResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.CreateFileSetResponse)) {
    throw new Error('Expected argument of type pfs_v2.CreateFileSetResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_CreateFileSetResponse(buffer_arg) {
  return pfs_pfs_pb.CreateFileSetResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_CreateProjectRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.CreateProjectRequest)) {
    throw new Error('Expected argument of type pfs_v2.CreateProjectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_CreateProjectRequest(buffer_arg) {
  return pfs_pfs_pb.CreateProjectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_CreateRepoRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.CreateRepoRequest)) {
    throw new Error('Expected argument of type pfs_v2.CreateRepoRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_CreateRepoRequest(buffer_arg) {
  return pfs_pfs_pb.CreateRepoRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_DeleteBranchRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.DeleteBranchRequest)) {
    throw new Error('Expected argument of type pfs_v2.DeleteBranchRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_DeleteBranchRequest(buffer_arg) {
  return pfs_pfs_pb.DeleteBranchRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_DeleteProjectRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.DeleteProjectRequest)) {
    throw new Error('Expected argument of type pfs_v2.DeleteProjectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_DeleteProjectRequest(buffer_arg) {
  return pfs_pfs_pb.DeleteProjectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_DeleteRepoRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.DeleteRepoRequest)) {
    throw new Error('Expected argument of type pfs_v2.DeleteRepoRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_DeleteRepoRequest(buffer_arg) {
  return pfs_pfs_pb.DeleteRepoRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_DiffFileRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.DiffFileRequest)) {
    throw new Error('Expected argument of type pfs_v2.DiffFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_DiffFileRequest(buffer_arg) {
  return pfs_pfs_pb.DiffFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_DiffFileResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.DiffFileResponse)) {
    throw new Error('Expected argument of type pfs_v2.DiffFileResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_DiffFileResponse(buffer_arg) {
  return pfs_pfs_pb.DiffFileResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_DropCommitSetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.DropCommitSetRequest)) {
    throw new Error('Expected argument of type pfs_v2.DropCommitSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_DropCommitSetRequest(buffer_arg) {
  return pfs_pfs_pb.DropCommitSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_EgressRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.EgressRequest)) {
    throw new Error('Expected argument of type pfs_v2.EgressRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_EgressRequest(buffer_arg) {
  return pfs_pfs_pb.EgressRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_EgressResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.EgressResponse)) {
    throw new Error('Expected argument of type pfs_v2.EgressResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_EgressResponse(buffer_arg) {
  return pfs_pfs_pb.EgressResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_FileInfo(arg) {
  if (!(arg instanceof pfs_pfs_pb.FileInfo)) {
    throw new Error('Expected argument of type pfs_v2.FileInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_FileInfo(buffer_arg) {
  return pfs_pfs_pb.FileInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_FinishCommitRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.FinishCommitRequest)) {
    throw new Error('Expected argument of type pfs_v2.FinishCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_FinishCommitRequest(buffer_arg) {
  return pfs_pfs_pb.FinishCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_FsckRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.FsckRequest)) {
    throw new Error('Expected argument of type pfs_v2.FsckRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_FsckRequest(buffer_arg) {
  return pfs_pfs_pb.FsckRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_FsckResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.FsckResponse)) {
    throw new Error('Expected argument of type pfs_v2.FsckResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_FsckResponse(buffer_arg) {
  return pfs_pfs_pb.FsckResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_GetCacheRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.GetCacheRequest)) {
    throw new Error('Expected argument of type pfs_v2.GetCacheRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_GetCacheRequest(buffer_arg) {
  return pfs_pfs_pb.GetCacheRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_GetCacheResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.GetCacheResponse)) {
    throw new Error('Expected argument of type pfs_v2.GetCacheResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_GetCacheResponse(buffer_arg) {
  return pfs_pfs_pb.GetCacheResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_GetFileRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.GetFileRequest)) {
    throw new Error('Expected argument of type pfs_v2.GetFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_GetFileRequest(buffer_arg) {
  return pfs_pfs_pb.GetFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_GetFileSetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.GetFileSetRequest)) {
    throw new Error('Expected argument of type pfs_v2.GetFileSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_GetFileSetRequest(buffer_arg) {
  return pfs_pfs_pb.GetFileSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_GlobFileRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.GlobFileRequest)) {
    throw new Error('Expected argument of type pfs_v2.GlobFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_GlobFileRequest(buffer_arg) {
  return pfs_pfs_pb.GlobFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_InspectBranchRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.InspectBranchRequest)) {
    throw new Error('Expected argument of type pfs_v2.InspectBranchRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_InspectBranchRequest(buffer_arg) {
  return pfs_pfs_pb.InspectBranchRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_InspectCommitRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.InspectCommitRequest)) {
    throw new Error('Expected argument of type pfs_v2.InspectCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_InspectCommitRequest(buffer_arg) {
  return pfs_pfs_pb.InspectCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_InspectCommitSetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.InspectCommitSetRequest)) {
    throw new Error('Expected argument of type pfs_v2.InspectCommitSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_InspectCommitSetRequest(buffer_arg) {
  return pfs_pfs_pb.InspectCommitSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_InspectFileRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.InspectFileRequest)) {
    throw new Error('Expected argument of type pfs_v2.InspectFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_InspectFileRequest(buffer_arg) {
  return pfs_pfs_pb.InspectFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_InspectProjectRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.InspectProjectRequest)) {
    throw new Error('Expected argument of type pfs_v2.InspectProjectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_InspectProjectRequest(buffer_arg) {
  return pfs_pfs_pb.InspectProjectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_InspectRepoRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.InspectRepoRequest)) {
    throw new Error('Expected argument of type pfs_v2.InspectRepoRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_InspectRepoRequest(buffer_arg) {
  return pfs_pfs_pb.InspectRepoRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ListBranchRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ListBranchRequest)) {
    throw new Error('Expected argument of type pfs_v2.ListBranchRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ListBranchRequest(buffer_arg) {
  return pfs_pfs_pb.ListBranchRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ListCommitRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ListCommitRequest)) {
    throw new Error('Expected argument of type pfs_v2.ListCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ListCommitRequest(buffer_arg) {
  return pfs_pfs_pb.ListCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ListCommitSetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ListCommitSetRequest)) {
    throw new Error('Expected argument of type pfs_v2.ListCommitSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ListCommitSetRequest(buffer_arg) {
  return pfs_pfs_pb.ListCommitSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ListFileRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ListFileRequest)) {
    throw new Error('Expected argument of type pfs_v2.ListFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ListFileRequest(buffer_arg) {
  return pfs_pfs_pb.ListFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ListProjectRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ListProjectRequest)) {
    throw new Error('Expected argument of type pfs_v2.ListProjectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ListProjectRequest(buffer_arg) {
  return pfs_pfs_pb.ListProjectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ListRepoRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ListRepoRequest)) {
    throw new Error('Expected argument of type pfs_v2.ListRepoRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ListRepoRequest(buffer_arg) {
  return pfs_pfs_pb.ListRepoRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ModifyFileRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ModifyFileRequest)) {
    throw new Error('Expected argument of type pfs_v2.ModifyFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ModifyFileRequest(buffer_arg) {
  return pfs_pfs_pb.ModifyFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ProjectInfo(arg) {
  if (!(arg instanceof pfs_pfs_pb.ProjectInfo)) {
    throw new Error('Expected argument of type pfs_v2.ProjectInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ProjectInfo(buffer_arg) {
  return pfs_pfs_pb.ProjectInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_PutCacheRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.PutCacheRequest)) {
    throw new Error('Expected argument of type pfs_v2.PutCacheRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_PutCacheRequest(buffer_arg) {
  return pfs_pfs_pb.PutCacheRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_RenewFileSetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.RenewFileSetRequest)) {
    throw new Error('Expected argument of type pfs_v2.RenewFileSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_RenewFileSetRequest(buffer_arg) {
  return pfs_pfs_pb.RenewFileSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_RepoInfo(arg) {
  if (!(arg instanceof pfs_pfs_pb.RepoInfo)) {
    throw new Error('Expected argument of type pfs_v2.RepoInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_RepoInfo(buffer_arg) {
  return pfs_pfs_pb.RepoInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_RunLoadTestRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.RunLoadTestRequest)) {
    throw new Error('Expected argument of type pfs_v2.RunLoadTestRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_RunLoadTestRequest(buffer_arg) {
  return pfs_pfs_pb.RunLoadTestRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_RunLoadTestResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.RunLoadTestResponse)) {
    throw new Error('Expected argument of type pfs_v2.RunLoadTestResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_RunLoadTestResponse(buffer_arg) {
  return pfs_pfs_pb.RunLoadTestResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ShardFileSetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ShardFileSetRequest)) {
    throw new Error('Expected argument of type pfs_v2.ShardFileSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ShardFileSetRequest(buffer_arg) {
  return pfs_pfs_pb.ShardFileSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_ShardFileSetResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.ShardFileSetResponse)) {
    throw new Error('Expected argument of type pfs_v2.ShardFileSetResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ShardFileSetResponse(buffer_arg) {
  return pfs_pfs_pb.ShardFileSetResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_SquashCommitSetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.SquashCommitSetRequest)) {
    throw new Error('Expected argument of type pfs_v2.SquashCommitSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_SquashCommitSetRequest(buffer_arg) {
  return pfs_pfs_pb.SquashCommitSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_StartCommitRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.StartCommitRequest)) {
    throw new Error('Expected argument of type pfs_v2.StartCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_StartCommitRequest(buffer_arg) {
  return pfs_pfs_pb.StartCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_SubscribeCommitRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.SubscribeCommitRequest)) {
    throw new Error('Expected argument of type pfs_v2.SubscribeCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_SubscribeCommitRequest(buffer_arg) {
  return pfs_pfs_pb.SubscribeCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_WalkFileRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.WalkFileRequest)) {
    throw new Error('Expected argument of type pfs_v2.WalkFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_WalkFileRequest(buffer_arg) {
  return pfs_pfs_pb.WalkFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_taskapi_ListTaskRequest(arg) {
  if (!(arg instanceof task_task_pb.ListTaskRequest)) {
    throw new Error('Expected argument of type taskapi.ListTaskRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_taskapi_ListTaskRequest(buffer_arg) {
  return task_task_pb.ListTaskRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_taskapi_TaskInfo(arg) {
  if (!(arg instanceof task_task_pb.TaskInfo)) {
    throw new Error('Expected argument of type taskapi.TaskInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_taskapi_TaskInfo(buffer_arg) {
  return task_task_pb.TaskInfo.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  // CreateRepo creates a new repo.
createRepo: {
    path: '/pfs_v2.API/CreateRepo',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.CreateRepoRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_CreateRepoRequest,
    requestDeserialize: deserialize_pfs_v2_CreateRepoRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // InspectRepo returns info about a repo.
inspectRepo: {
    path: '/pfs_v2.API/InspectRepo',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.InspectRepoRequest,
    responseType: pfs_pfs_pb.RepoInfo,
    requestSerialize: serialize_pfs_v2_InspectRepoRequest,
    requestDeserialize: deserialize_pfs_v2_InspectRepoRequest,
    responseSerialize: serialize_pfs_v2_RepoInfo,
    responseDeserialize: deserialize_pfs_v2_RepoInfo,
  },
  // ListRepo returns info about all repos.
listRepo: {
    path: '/pfs_v2.API/ListRepo',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.ListRepoRequest,
    responseType: pfs_pfs_pb.RepoInfo,
    requestSerialize: serialize_pfs_v2_ListRepoRequest,
    requestDeserialize: deserialize_pfs_v2_ListRepoRequest,
    responseSerialize: serialize_pfs_v2_RepoInfo,
    responseDeserialize: deserialize_pfs_v2_RepoInfo,
  },
  // DeleteRepo deletes a repo.
deleteRepo: {
    path: '/pfs_v2.API/DeleteRepo',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.DeleteRepoRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_DeleteRepoRequest,
    requestDeserialize: deserialize_pfs_v2_DeleteRepoRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // StartCommit creates a new write commit from a parent commit.
startCommit: {
    path: '/pfs_v2.API/StartCommit',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.StartCommitRequest,
    responseType: pfs_pfs_pb.Commit,
    requestSerialize: serialize_pfs_v2_StartCommitRequest,
    requestDeserialize: deserialize_pfs_v2_StartCommitRequest,
    responseSerialize: serialize_pfs_v2_Commit,
    responseDeserialize: deserialize_pfs_v2_Commit,
  },
  // FinishCommit turns a write commit into a read commit.
finishCommit: {
    path: '/pfs_v2.API/FinishCommit',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.FinishCommitRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_FinishCommitRequest,
    requestDeserialize: deserialize_pfs_v2_FinishCommitRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // ClearCommit removes all data from the commit.
clearCommit: {
    path: '/pfs_v2.API/ClearCommit',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.ClearCommitRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_ClearCommitRequest,
    requestDeserialize: deserialize_pfs_v2_ClearCommitRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // InspectCommit returns the info about a commit.
inspectCommit: {
    path: '/pfs_v2.API/InspectCommit',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.InspectCommitRequest,
    responseType: pfs_pfs_pb.CommitInfo,
    requestSerialize: serialize_pfs_v2_InspectCommitRequest,
    requestDeserialize: deserialize_pfs_v2_InspectCommitRequest,
    responseSerialize: serialize_pfs_v2_CommitInfo,
    responseDeserialize: deserialize_pfs_v2_CommitInfo,
  },
  // ListCommit returns info about all commits.
listCommit: {
    path: '/pfs_v2.API/ListCommit',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.ListCommitRequest,
    responseType: pfs_pfs_pb.CommitInfo,
    requestSerialize: serialize_pfs_v2_ListCommitRequest,
    requestDeserialize: deserialize_pfs_v2_ListCommitRequest,
    responseSerialize: serialize_pfs_v2_CommitInfo,
    responseDeserialize: deserialize_pfs_v2_CommitInfo,
  },
  // SubscribeCommit subscribes for new commits on a given branch.
subscribeCommit: {
    path: '/pfs_v2.API/SubscribeCommit',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.SubscribeCommitRequest,
    responseType: pfs_pfs_pb.CommitInfo,
    requestSerialize: serialize_pfs_v2_SubscribeCommitRequest,
    requestDeserialize: deserialize_pfs_v2_SubscribeCommitRequest,
    responseSerialize: serialize_pfs_v2_CommitInfo,
    responseDeserialize: deserialize_pfs_v2_CommitInfo,
  },
  // InspectCommitSet returns the info about a CommitSet.
inspectCommitSet: {
    path: '/pfs_v2.API/InspectCommitSet',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.InspectCommitSetRequest,
    responseType: pfs_pfs_pb.CommitInfo,
    requestSerialize: serialize_pfs_v2_InspectCommitSetRequest,
    requestDeserialize: deserialize_pfs_v2_InspectCommitSetRequest,
    responseSerialize: serialize_pfs_v2_CommitInfo,
    responseDeserialize: deserialize_pfs_v2_CommitInfo,
  },
  // ListCommitSet returns info about all CommitSets.
listCommitSet: {
    path: '/pfs_v2.API/ListCommitSet',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.ListCommitSetRequest,
    responseType: pfs_pfs_pb.CommitSetInfo,
    requestSerialize: serialize_pfs_v2_ListCommitSetRequest,
    requestDeserialize: deserialize_pfs_v2_ListCommitSetRequest,
    responseSerialize: serialize_pfs_v2_CommitSetInfo,
    responseDeserialize: deserialize_pfs_v2_CommitSetInfo,
  },
  // SquashCommitSet squashes the commits of a CommitSet into their children.
squashCommitSet: {
    path: '/pfs_v2.API/SquashCommitSet',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.SquashCommitSetRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_SquashCommitSetRequest,
    requestDeserialize: deserialize_pfs_v2_SquashCommitSetRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // DropCommitSet drops the commits of a CommitSet and all data included in the commits.
dropCommitSet: {
    path: '/pfs_v2.API/DropCommitSet',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.DropCommitSetRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_DropCommitSetRequest,
    requestDeserialize: deserialize_pfs_v2_DropCommitSetRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // CreateBranch creates a new branch.
createBranch: {
    path: '/pfs_v2.API/CreateBranch',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.CreateBranchRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_CreateBranchRequest,
    requestDeserialize: deserialize_pfs_v2_CreateBranchRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // InspectBranch returns info about a branch.
inspectBranch: {
    path: '/pfs_v2.API/InspectBranch',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.InspectBranchRequest,
    responseType: pfs_pfs_pb.BranchInfo,
    requestSerialize: serialize_pfs_v2_InspectBranchRequest,
    requestDeserialize: deserialize_pfs_v2_InspectBranchRequest,
    responseSerialize: serialize_pfs_v2_BranchInfo,
    responseDeserialize: deserialize_pfs_v2_BranchInfo,
  },
  // ListBranch returns info about the heads of branches.
listBranch: {
    path: '/pfs_v2.API/ListBranch',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.ListBranchRequest,
    responseType: pfs_pfs_pb.BranchInfo,
    requestSerialize: serialize_pfs_v2_ListBranchRequest,
    requestDeserialize: deserialize_pfs_v2_ListBranchRequest,
    responseSerialize: serialize_pfs_v2_BranchInfo,
    responseDeserialize: deserialize_pfs_v2_BranchInfo,
  },
  // DeleteBranch deletes a branch; note that the commits still exist.
deleteBranch: {
    path: '/pfs_v2.API/DeleteBranch',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.DeleteBranchRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_DeleteBranchRequest,
    requestDeserialize: deserialize_pfs_v2_DeleteBranchRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // ModifyFile performs modifications on a set of files.
modifyFile: {
    path: '/pfs_v2.API/ModifyFile',
    requestStream: true,
    responseStream: false,
    requestType: pfs_pfs_pb.ModifyFileRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_ModifyFileRequest,
    requestDeserialize: deserialize_pfs_v2_ModifyFileRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // GetFile returns the contents of a single file
getFile: {
    path: '/pfs_v2.API/GetFile',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.GetFileRequest,
    responseType: google_protobuf_wrappers_pb.BytesValue,
    requestSerialize: serialize_pfs_v2_GetFileRequest,
    requestDeserialize: deserialize_pfs_v2_GetFileRequest,
    responseSerialize: serialize_google_protobuf_BytesValue,
    responseDeserialize: deserialize_google_protobuf_BytesValue,
  },
  // GetFileTAR returns a TAR stream of the contents matched by the request
getFileTAR: {
    path: '/pfs_v2.API/GetFileTAR',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.GetFileRequest,
    responseType: google_protobuf_wrappers_pb.BytesValue,
    requestSerialize: serialize_pfs_v2_GetFileRequest,
    requestDeserialize: deserialize_pfs_v2_GetFileRequest,
    responseSerialize: serialize_google_protobuf_BytesValue,
    responseDeserialize: deserialize_google_protobuf_BytesValue,
  },
  // InspectFile returns info about a file.
inspectFile: {
    path: '/pfs_v2.API/InspectFile',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.InspectFileRequest,
    responseType: pfs_pfs_pb.FileInfo,
    requestSerialize: serialize_pfs_v2_InspectFileRequest,
    requestDeserialize: deserialize_pfs_v2_InspectFileRequest,
    responseSerialize: serialize_pfs_v2_FileInfo,
    responseDeserialize: deserialize_pfs_v2_FileInfo,
  },
  // ListFile returns info about all files.
listFile: {
    path: '/pfs_v2.API/ListFile',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.ListFileRequest,
    responseType: pfs_pfs_pb.FileInfo,
    requestSerialize: serialize_pfs_v2_ListFileRequest,
    requestDeserialize: deserialize_pfs_v2_ListFileRequest,
    responseSerialize: serialize_pfs_v2_FileInfo,
    responseDeserialize: deserialize_pfs_v2_FileInfo,
  },
  // WalkFile walks over all the files under a directory, including children of children.
walkFile: {
    path: '/pfs_v2.API/WalkFile',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.WalkFileRequest,
    responseType: pfs_pfs_pb.FileInfo,
    requestSerialize: serialize_pfs_v2_WalkFileRequest,
    requestDeserialize: deserialize_pfs_v2_WalkFileRequest,
    responseSerialize: serialize_pfs_v2_FileInfo,
    responseDeserialize: deserialize_pfs_v2_FileInfo,
  },
  // GlobFile returns info about all files.
globFile: {
    path: '/pfs_v2.API/GlobFile',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.GlobFileRequest,
    responseType: pfs_pfs_pb.FileInfo,
    requestSerialize: serialize_pfs_v2_GlobFileRequest,
    requestDeserialize: deserialize_pfs_v2_GlobFileRequest,
    responseSerialize: serialize_pfs_v2_FileInfo,
    responseDeserialize: deserialize_pfs_v2_FileInfo,
  },
  // DiffFile returns the differences between 2 paths at 2 commits.
diffFile: {
    path: '/pfs_v2.API/DiffFile',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.DiffFileRequest,
    responseType: pfs_pfs_pb.DiffFileResponse,
    requestSerialize: serialize_pfs_v2_DiffFileRequest,
    requestDeserialize: deserialize_pfs_v2_DiffFileRequest,
    responseSerialize: serialize_pfs_v2_DiffFileResponse,
    responseDeserialize: deserialize_pfs_v2_DiffFileResponse,
  },
  // ActivateAuth creates a role binding for all existing repos
activateAuth: {
    path: '/pfs_v2.API/ActivateAuth',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.ActivateAuthRequest,
    responseType: pfs_pfs_pb.ActivateAuthResponse,
    requestSerialize: serialize_pfs_v2_ActivateAuthRequest,
    requestDeserialize: deserialize_pfs_v2_ActivateAuthRequest,
    responseSerialize: serialize_pfs_v2_ActivateAuthResponse,
    responseDeserialize: deserialize_pfs_v2_ActivateAuthResponse,
  },
  // DeleteAll deletes everything.
deleteAll: {
    path: '/pfs_v2.API/DeleteAll',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Fsck does a file system consistency check for pfs.
fsck: {
    path: '/pfs_v2.API/Fsck',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.FsckRequest,
    responseType: pfs_pfs_pb.FsckResponse,
    requestSerialize: serialize_pfs_v2_FsckRequest,
    requestDeserialize: deserialize_pfs_v2_FsckRequest,
    responseSerialize: serialize_pfs_v2_FsckResponse,
    responseDeserialize: deserialize_pfs_v2_FsckResponse,
  },
  // FileSet API
// CreateFileSet creates a new file set.
createFileSet: {
    path: '/pfs_v2.API/CreateFileSet',
    requestStream: true,
    responseStream: false,
    requestType: pfs_pfs_pb.ModifyFileRequest,
    responseType: pfs_pfs_pb.CreateFileSetResponse,
    requestSerialize: serialize_pfs_v2_ModifyFileRequest,
    requestDeserialize: deserialize_pfs_v2_ModifyFileRequest,
    responseSerialize: serialize_pfs_v2_CreateFileSetResponse,
    responseDeserialize: deserialize_pfs_v2_CreateFileSetResponse,
  },
  // GetFileSet returns a file set with the data from a commit
getFileSet: {
    path: '/pfs_v2.API/GetFileSet',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.GetFileSetRequest,
    responseType: pfs_pfs_pb.CreateFileSetResponse,
    requestSerialize: serialize_pfs_v2_GetFileSetRequest,
    requestDeserialize: deserialize_pfs_v2_GetFileSetRequest,
    responseSerialize: serialize_pfs_v2_CreateFileSetResponse,
    responseDeserialize: deserialize_pfs_v2_CreateFileSetResponse,
  },
  // AddFileSet associates a file set with a commit
addFileSet: {
    path: '/pfs_v2.API/AddFileSet',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.AddFileSetRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_AddFileSetRequest,
    requestDeserialize: deserialize_pfs_v2_AddFileSetRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // RenewFileSet prevents a file set from being deleted for a set amount of time.
renewFileSet: {
    path: '/pfs_v2.API/RenewFileSet',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.RenewFileSetRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_RenewFileSetRequest,
    requestDeserialize: deserialize_pfs_v2_RenewFileSetRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // ComposeFileSet composes a file set from a list of file sets.
composeFileSet: {
    path: '/pfs_v2.API/ComposeFileSet',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.ComposeFileSetRequest,
    responseType: pfs_pfs_pb.CreateFileSetResponse,
    requestSerialize: serialize_pfs_v2_ComposeFileSetRequest,
    requestDeserialize: deserialize_pfs_v2_ComposeFileSetRequest,
    responseSerialize: serialize_pfs_v2_CreateFileSetResponse,
    responseDeserialize: deserialize_pfs_v2_CreateFileSetResponse,
  },
  shardFileSet: {
    path: '/pfs_v2.API/ShardFileSet',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.ShardFileSetRequest,
    responseType: pfs_pfs_pb.ShardFileSetResponse,
    requestSerialize: serialize_pfs_v2_ShardFileSetRequest,
    requestDeserialize: deserialize_pfs_v2_ShardFileSetRequest,
    responseSerialize: serialize_pfs_v2_ShardFileSetResponse,
    responseDeserialize: deserialize_pfs_v2_ShardFileSetResponse,
  },
  // CheckStorage runs integrity checks for the storage layer.
checkStorage: {
    path: '/pfs_v2.API/CheckStorage',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.CheckStorageRequest,
    responseType: pfs_pfs_pb.CheckStorageResponse,
    requestSerialize: serialize_pfs_v2_CheckStorageRequest,
    requestDeserialize: deserialize_pfs_v2_CheckStorageRequest,
    responseSerialize: serialize_pfs_v2_CheckStorageResponse,
    responseDeserialize: deserialize_pfs_v2_CheckStorageResponse,
  },
  putCache: {
    path: '/pfs_v2.API/PutCache',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.PutCacheRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_PutCacheRequest,
    requestDeserialize: deserialize_pfs_v2_PutCacheRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  getCache: {
    path: '/pfs_v2.API/GetCache',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.GetCacheRequest,
    responseType: pfs_pfs_pb.GetCacheResponse,
    requestSerialize: serialize_pfs_v2_GetCacheRequest,
    requestDeserialize: deserialize_pfs_v2_GetCacheRequest,
    responseSerialize: serialize_pfs_v2_GetCacheResponse,
    responseDeserialize: deserialize_pfs_v2_GetCacheResponse,
  },
  clearCache: {
    path: '/pfs_v2.API/ClearCache',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.ClearCacheRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_ClearCacheRequest,
    requestDeserialize: deserialize_pfs_v2_ClearCacheRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // RunLoadTest runs a load test.
runLoadTest: {
    path: '/pfs_v2.API/RunLoadTest',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.RunLoadTestRequest,
    responseType: pfs_pfs_pb.RunLoadTestResponse,
    requestSerialize: serialize_pfs_v2_RunLoadTestRequest,
    requestDeserialize: deserialize_pfs_v2_RunLoadTestRequest,
    responseSerialize: serialize_pfs_v2_RunLoadTestResponse,
    responseDeserialize: deserialize_pfs_v2_RunLoadTestResponse,
  },
  // RunLoadTestDefault runs the default load tests.
runLoadTestDefault: {
    path: '/pfs_v2.API/RunLoadTestDefault',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: pfs_pfs_pb.RunLoadTestResponse,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_pfs_v2_RunLoadTestResponse,
    responseDeserialize: deserialize_pfs_v2_RunLoadTestResponse,
  },
  // ListTask lists PFS tasks
listTask: {
    path: '/pfs_v2.API/ListTask',
    requestStream: false,
    responseStream: true,
    requestType: task_task_pb.ListTaskRequest,
    responseType: task_task_pb.TaskInfo,
    requestSerialize: serialize_taskapi_ListTaskRequest,
    requestDeserialize: deserialize_taskapi_ListTaskRequest,
    responseSerialize: serialize_taskapi_TaskInfo,
    responseDeserialize: deserialize_taskapi_TaskInfo,
  },
  // Egress writes data from a commit to an external system
egress: {
    path: '/pfs_v2.API/Egress',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.EgressRequest,
    responseType: pfs_pfs_pb.EgressResponse,
    requestSerialize: serialize_pfs_v2_EgressRequest,
    requestDeserialize: deserialize_pfs_v2_EgressRequest,
    responseSerialize: serialize_pfs_v2_EgressResponse,
    responseDeserialize: deserialize_pfs_v2_EgressResponse,
  },
  // Project API
// CreateProject creates a new project.
createProject: {
    path: '/pfs_v2.API/CreateProject',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.CreateProjectRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_CreateProjectRequest,
    requestDeserialize: deserialize_pfs_v2_CreateProjectRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // InspectProject returns info about a project.
inspectProject: {
    path: '/pfs_v2.API/InspectProject',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.InspectProjectRequest,
    responseType: pfs_pfs_pb.ProjectInfo,
    requestSerialize: serialize_pfs_v2_InspectProjectRequest,
    requestDeserialize: deserialize_pfs_v2_InspectProjectRequest,
    responseSerialize: serialize_pfs_v2_ProjectInfo,
    responseDeserialize: deserialize_pfs_v2_ProjectInfo,
  },
  // ListProject returns info about all projects.
listProject: {
    path: '/pfs_v2.API/ListProject',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.ListProjectRequest,
    responseType: pfs_pfs_pb.ProjectInfo,
    requestSerialize: serialize_pfs_v2_ListProjectRequest,
    requestDeserialize: deserialize_pfs_v2_ListProjectRequest,
    responseSerialize: serialize_pfs_v2_ProjectInfo,
    responseDeserialize: deserialize_pfs_v2_ProjectInfo,
  },
  // DeleteProject deletes a project.
deleteProject: {
    path: '/pfs_v2.API/DeleteProject',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.DeleteProjectRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_DeleteProjectRequest,
    requestDeserialize: deserialize_pfs_v2_DeleteProjectRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
