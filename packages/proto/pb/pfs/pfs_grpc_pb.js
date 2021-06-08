// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var pfs_pfs_pb = require('../pfs/pfs_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var google_protobuf_wrappers_pb = require('google-protobuf/google/protobuf/wrappers_pb.js');
var gogoproto_gogo_pb = require('../gogoproto/gogo_pb.js');
var auth_auth_pb = require('../auth/auth_pb.js');

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

function serialize_pfs_v2_AddFilesetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.AddFilesetRequest)) {
    throw new Error('Expected argument of type pfs_v2.AddFilesetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_AddFilesetRequest(buffer_arg) {
  return pfs_pfs_pb.AddFilesetRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pfs_v2_BranchInfos(arg) {
  if (!(arg instanceof pfs_pfs_pb.BranchInfos)) {
    throw new Error('Expected argument of type pfs_v2.BranchInfos');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_BranchInfos(buffer_arg) {
  return pfs_pfs_pb.BranchInfos.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pfs_v2_CreateBranchRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.CreateBranchRequest)) {
    throw new Error('Expected argument of type pfs_v2.CreateBranchRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_CreateBranchRequest(buffer_arg) {
  return pfs_pfs_pb.CreateBranchRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_CreateFilesetResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.CreateFilesetResponse)) {
    throw new Error('Expected argument of type pfs_v2.CreateFilesetResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_CreateFilesetResponse(buffer_arg) {
  return pfs_pfs_pb.CreateFilesetResponse.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pfs_v2_FlushCommitRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.FlushCommitRequest)) {
    throw new Error('Expected argument of type pfs_v2.FlushCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_FlushCommitRequest(buffer_arg) {
  return pfs_pfs_pb.FlushCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pfs_v2_GetFileRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.GetFileRequest)) {
    throw new Error('Expected argument of type pfs_v2.GetFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_GetFileRequest(buffer_arg) {
  return pfs_pfs_pb.GetFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pfs_v2_GetFilesetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.GetFilesetRequest)) {
    throw new Error('Expected argument of type pfs_v2.GetFilesetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_GetFilesetRequest(buffer_arg) {
  return pfs_pfs_pb.GetFilesetRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pfs_v2_InspectFileRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.InspectFileRequest)) {
    throw new Error('Expected argument of type pfs_v2.InspectFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_InspectFileRequest(buffer_arg) {
  return pfs_pfs_pb.InspectFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pfs_v2_ListFileRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.ListFileRequest)) {
    throw new Error('Expected argument of type pfs_v2.ListFileRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ListFileRequest(buffer_arg) {
  return pfs_pfs_pb.ListFileRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pfs_v2_ListRepoResponse(arg) {
  if (!(arg instanceof pfs_pfs_pb.ListRepoResponse)) {
    throw new Error('Expected argument of type pfs_v2.ListRepoResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_ListRepoResponse(buffer_arg) {
  return pfs_pfs_pb.ListRepoResponse.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pfs_v2_RenewFilesetRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.RenewFilesetRequest)) {
    throw new Error('Expected argument of type pfs_v2.RenewFilesetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_RenewFilesetRequest(buffer_arg) {
  return pfs_pfs_pb.RenewFilesetRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pfs_v2_SquashCommitRequest(arg) {
  if (!(arg instanceof pfs_pfs_pb.SquashCommitRequest)) {
    throw new Error('Expected argument of type pfs_v2.SquashCommitRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pfs_v2_SquashCommitRequest(buffer_arg) {
  return pfs_pfs_pb.SquashCommitRequest.deserializeBinary(new Uint8Array(buffer_arg));
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
    responseStream: false,
    requestType: pfs_pfs_pb.ListRepoRequest,
    responseType: pfs_pfs_pb.ListRepoResponse,
    requestSerialize: serialize_pfs_v2_ListRepoRequest,
    requestDeserialize: deserialize_pfs_v2_ListRepoRequest,
    responseSerialize: serialize_pfs_v2_ListRepoResponse,
    responseDeserialize: deserialize_pfs_v2_ListRepoResponse,
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
  // SquashCommit squashes a commit into it's parent.
squashCommit: {
    path: '/pfs_v2.API/SquashCommit',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.SquashCommitRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_SquashCommitRequest,
    requestDeserialize: deserialize_pfs_v2_SquashCommitRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // FlushCommit waits for downstream commits to finish.
flushCommit: {
    path: '/pfs_v2.API/FlushCommit',
    requestStream: false,
    responseStream: true,
    requestType: pfs_pfs_pb.FlushCommitRequest,
    responseType: pfs_pfs_pb.CommitInfo,
    requestSerialize: serialize_pfs_v2_FlushCommitRequest,
    requestDeserialize: deserialize_pfs_v2_FlushCommitRequest,
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
    responseStream: false,
    requestType: pfs_pfs_pb.ListBranchRequest,
    responseType: pfs_pfs_pb.BranchInfos,
    requestSerialize: serialize_pfs_v2_ListBranchRequest,
    requestDeserialize: deserialize_pfs_v2_ListBranchRequest,
    responseSerialize: serialize_pfs_v2_BranchInfos,
    responseDeserialize: deserialize_pfs_v2_BranchInfos,
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
  // Fileset API
// CreateFileset creates a new fileset.
createFileset: {
    path: '/pfs_v2.API/CreateFileset',
    requestStream: true,
    responseStream: false,
    requestType: pfs_pfs_pb.ModifyFileRequest,
    responseType: pfs_pfs_pb.CreateFilesetResponse,
    requestSerialize: serialize_pfs_v2_ModifyFileRequest,
    requestDeserialize: deserialize_pfs_v2_ModifyFileRequest,
    responseSerialize: serialize_pfs_v2_CreateFilesetResponse,
    responseDeserialize: deserialize_pfs_v2_CreateFilesetResponse,
  },
  // GetFileset returns a fileset with the data from a commit
getFileset: {
    path: '/pfs_v2.API/GetFileset',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.GetFilesetRequest,
    responseType: pfs_pfs_pb.CreateFilesetResponse,
    requestSerialize: serialize_pfs_v2_GetFilesetRequest,
    requestDeserialize: deserialize_pfs_v2_GetFilesetRequest,
    responseSerialize: serialize_pfs_v2_CreateFilesetResponse,
    responseDeserialize: deserialize_pfs_v2_CreateFilesetResponse,
  },
  // AddFileset associates a fileset with a commit
addFileset: {
    path: '/pfs_v2.API/AddFileset',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.AddFilesetRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_AddFilesetRequest,
    requestDeserialize: deserialize_pfs_v2_AddFilesetRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // RenewFileset prevents a fileset from being deleted for a set amount of time.
renewFileset: {
    path: '/pfs_v2.API/RenewFileset',
    requestStream: false,
    responseStream: false,
    requestType: pfs_pfs_pb.RenewFilesetRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pfs_v2_RenewFilesetRequest,
    requestDeserialize: deserialize_pfs_v2_RenewFilesetRequest,
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
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
