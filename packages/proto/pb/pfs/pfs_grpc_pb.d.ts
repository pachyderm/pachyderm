// package: pfs_v2
// file: pfs/pfs.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as pfs_pfs_pb from "../pfs/pfs_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as google_protobuf_wrappers_pb from "google-protobuf/google/protobuf/wrappers_pb";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";
import * as auth_auth_pb from "../auth/auth_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    createRepo: IAPIService_ICreateRepo;
    inspectRepo: IAPIService_IInspectRepo;
    listRepo: IAPIService_IListRepo;
    deleteRepo: IAPIService_IDeleteRepo;
    startCommit: IAPIService_IStartCommit;
    finishCommit: IAPIService_IFinishCommit;
    clearCommit: IAPIService_IClearCommit;
    inspectCommit: IAPIService_IInspectCommit;
    listCommit: IAPIService_IListCommit;
    subscribeCommit: IAPIService_ISubscribeCommit;
    inspectCommitSet: IAPIService_IInspectCommitSet;
    squashCommitSet: IAPIService_ISquashCommitSet;
    createBranch: IAPIService_ICreateBranch;
    inspectBranch: IAPIService_IInspectBranch;
    listBranch: IAPIService_IListBranch;
    deleteBranch: IAPIService_IDeleteBranch;
    modifyFile: IAPIService_IModifyFile;
    getFileTAR: IAPIService_IGetFileTAR;
    inspectFile: IAPIService_IInspectFile;
    listFile: IAPIService_IListFile;
    walkFile: IAPIService_IWalkFile;
    globFile: IAPIService_IGlobFile;
    diffFile: IAPIService_IDiffFile;
    activateAuth: IAPIService_IActivateAuth;
    deleteAll: IAPIService_IDeleteAll;
    fsck: IAPIService_IFsck;
    createFileSet: IAPIService_ICreateFileSet;
    getFileSet: IAPIService_IGetFileSet;
    addFileSet: IAPIService_IAddFileSet;
    renewFileSet: IAPIService_IRenewFileSet;
    runLoadTest: IAPIService_IRunLoadTest;
}

interface IAPIService_ICreateRepo extends grpc.MethodDefinition<pfs_pfs_pb.CreateRepoRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/CreateRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.CreateRepoRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.CreateRepoRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectRepo extends grpc.MethodDefinition<pfs_pfs_pb.InspectRepoRequest, pfs_pfs_pb.RepoInfo> {
    path: "/pfs_v2.API/InspectRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.InspectRepoRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.InspectRepoRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.RepoInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.RepoInfo>;
}
interface IAPIService_IListRepo extends grpc.MethodDefinition<pfs_pfs_pb.ListRepoRequest, pfs_pfs_pb.ListRepoResponse> {
    path: "/pfs_v2.API/ListRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ListRepoRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ListRepoRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.ListRepoResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.ListRepoResponse>;
}
interface IAPIService_IDeleteRepo extends grpc.MethodDefinition<pfs_pfs_pb.DeleteRepoRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/DeleteRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.DeleteRepoRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.DeleteRepoRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IStartCommit extends grpc.MethodDefinition<pfs_pfs_pb.StartCommitRequest, pfs_pfs_pb.Commit> {
    path: "/pfs_v2.API/StartCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.StartCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.StartCommitRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.Commit>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.Commit>;
}
interface IAPIService_IFinishCommit extends grpc.MethodDefinition<pfs_pfs_pb.FinishCommitRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/FinishCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.FinishCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.FinishCommitRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IClearCommit extends grpc.MethodDefinition<pfs_pfs_pb.ClearCommitRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/ClearCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ClearCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ClearCommitRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectCommit extends grpc.MethodDefinition<pfs_pfs_pb.InspectCommitRequest, pfs_pfs_pb.CommitInfo> {
    path: "/pfs_v2.API/InspectCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.InspectCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.InspectCommitRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_IListCommit extends grpc.MethodDefinition<pfs_pfs_pb.ListCommitRequest, pfs_pfs_pb.CommitInfo> {
    path: "/pfs_v2.API/ListCommit";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ListCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ListCommitRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_ISubscribeCommit extends grpc.MethodDefinition<pfs_pfs_pb.SubscribeCommitRequest, pfs_pfs_pb.CommitInfo> {
    path: "/pfs_v2.API/SubscribeCommit";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.SubscribeCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.SubscribeCommitRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_IInspectCommitSet extends grpc.MethodDefinition<pfs_pfs_pb.InspectCommitSetRequest, pfs_pfs_pb.CommitInfo> {
    path: "/pfs_v2.API/InspectCommitSet";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.InspectCommitSetRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.InspectCommitSetRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_ISquashCommitSet extends grpc.MethodDefinition<pfs_pfs_pb.SquashCommitSetRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/SquashCommitSet";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.SquashCommitSetRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.SquashCommitSetRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_ICreateBranch extends grpc.MethodDefinition<pfs_pfs_pb.CreateBranchRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/CreateBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.CreateBranchRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.CreateBranchRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectBranch extends grpc.MethodDefinition<pfs_pfs_pb.InspectBranchRequest, pfs_pfs_pb.BranchInfo> {
    path: "/pfs_v2.API/InspectBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.InspectBranchRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.InspectBranchRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.BranchInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.BranchInfo>;
}
interface IAPIService_IListBranch extends grpc.MethodDefinition<pfs_pfs_pb.ListBranchRequest, pfs_pfs_pb.BranchInfos> {
    path: "/pfs_v2.API/ListBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ListBranchRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ListBranchRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.BranchInfos>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.BranchInfos>;
}
interface IAPIService_IDeleteBranch extends grpc.MethodDefinition<pfs_pfs_pb.DeleteBranchRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/DeleteBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.DeleteBranchRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.DeleteBranchRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IModifyFile extends grpc.MethodDefinition<pfs_pfs_pb.ModifyFileRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/ModifyFile";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ModifyFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ModifyFileRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IGetFileTAR extends grpc.MethodDefinition<pfs_pfs_pb.GetFileRequest, google_protobuf_wrappers_pb.BytesValue> {
    path: "/pfs_v2.API/GetFileTAR";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.GetFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.GetFileRequest>;
    responseSerialize: grpc.serialize<google_protobuf_wrappers_pb.BytesValue>;
    responseDeserialize: grpc.deserialize<google_protobuf_wrappers_pb.BytesValue>;
}
interface IAPIService_IInspectFile extends grpc.MethodDefinition<pfs_pfs_pb.InspectFileRequest, pfs_pfs_pb.FileInfo> {
    path: "/pfs_v2.API/InspectFile";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.InspectFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.InspectFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IListFile extends grpc.MethodDefinition<pfs_pfs_pb.ListFileRequest, pfs_pfs_pb.FileInfo> {
    path: "/pfs_v2.API/ListFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ListFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ListFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IWalkFile extends grpc.MethodDefinition<pfs_pfs_pb.WalkFileRequest, pfs_pfs_pb.FileInfo> {
    path: "/pfs_v2.API/WalkFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.WalkFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.WalkFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IGlobFile extends grpc.MethodDefinition<pfs_pfs_pb.GlobFileRequest, pfs_pfs_pb.FileInfo> {
    path: "/pfs_v2.API/GlobFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.GlobFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.GlobFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IDiffFile extends grpc.MethodDefinition<pfs_pfs_pb.DiffFileRequest, pfs_pfs_pb.DiffFileResponse> {
    path: "/pfs_v2.API/DiffFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.DiffFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.DiffFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.DiffFileResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.DiffFileResponse>;
}
interface IAPIService_IActivateAuth extends grpc.MethodDefinition<pfs_pfs_pb.ActivateAuthRequest, pfs_pfs_pb.ActivateAuthResponse> {
    path: "/pfs_v2.API/ActivateAuth";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ActivateAuthRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ActivateAuthRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.ActivateAuthResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.ActivateAuthResponse>;
}
interface IAPIService_IDeleteAll extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/DeleteAll";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IFsck extends grpc.MethodDefinition<pfs_pfs_pb.FsckRequest, pfs_pfs_pb.FsckResponse> {
    path: "/pfs_v2.API/Fsck";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.FsckRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.FsckRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.FsckResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.FsckResponse>;
}
interface IAPIService_ICreateFileSet extends grpc.MethodDefinition<pfs_pfs_pb.ModifyFileRequest, pfs_pfs_pb.CreateFileSetResponse> {
    path: "/pfs_v2.API/CreateFileSet";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ModifyFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ModifyFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CreateFileSetResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CreateFileSetResponse>;
}
interface IAPIService_IGetFileSet extends grpc.MethodDefinition<pfs_pfs_pb.GetFileSetRequest, pfs_pfs_pb.CreateFileSetResponse> {
    path: "/pfs_v2.API/GetFileSet";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.GetFileSetRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.GetFileSetRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CreateFileSetResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CreateFileSetResponse>;
}
interface IAPIService_IAddFileSet extends grpc.MethodDefinition<pfs_pfs_pb.AddFileSetRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/AddFileSet";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.AddFileSetRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.AddFileSetRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IRenewFileSet extends grpc.MethodDefinition<pfs_pfs_pb.RenewFileSetRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs_v2.API/RenewFileSet";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.RenewFileSetRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.RenewFileSetRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IRunLoadTest extends grpc.MethodDefinition<pfs_pfs_pb.RunLoadTestRequest, pfs_pfs_pb.RunLoadTestResponse> {
    path: "/pfs_v2.API/RunLoadTest";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.RunLoadTestRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.RunLoadTestRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.RunLoadTestResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.RunLoadTestResponse>;
}

export const APIService: IAPIService;

export interface IAPIServer extends grpc.UntypedServiceImplementation {
    createRepo: grpc.handleUnaryCall<pfs_pfs_pb.CreateRepoRequest, google_protobuf_empty_pb.Empty>;
    inspectRepo: grpc.handleUnaryCall<pfs_pfs_pb.InspectRepoRequest, pfs_pfs_pb.RepoInfo>;
    listRepo: grpc.handleUnaryCall<pfs_pfs_pb.ListRepoRequest, pfs_pfs_pb.ListRepoResponse>;
    deleteRepo: grpc.handleUnaryCall<pfs_pfs_pb.DeleteRepoRequest, google_protobuf_empty_pb.Empty>;
    startCommit: grpc.handleUnaryCall<pfs_pfs_pb.StartCommitRequest, pfs_pfs_pb.Commit>;
    finishCommit: grpc.handleUnaryCall<pfs_pfs_pb.FinishCommitRequest, google_protobuf_empty_pb.Empty>;
    clearCommit: grpc.handleUnaryCall<pfs_pfs_pb.ClearCommitRequest, google_protobuf_empty_pb.Empty>;
    inspectCommit: grpc.handleUnaryCall<pfs_pfs_pb.InspectCommitRequest, pfs_pfs_pb.CommitInfo>;
    listCommit: grpc.handleServerStreamingCall<pfs_pfs_pb.ListCommitRequest, pfs_pfs_pb.CommitInfo>;
    subscribeCommit: grpc.handleServerStreamingCall<pfs_pfs_pb.SubscribeCommitRequest, pfs_pfs_pb.CommitInfo>;
    inspectCommitSet: grpc.handleServerStreamingCall<pfs_pfs_pb.InspectCommitSetRequest, pfs_pfs_pb.CommitInfo>;
    squashCommitSet: grpc.handleUnaryCall<pfs_pfs_pb.SquashCommitSetRequest, google_protobuf_empty_pb.Empty>;
    createBranch: grpc.handleUnaryCall<pfs_pfs_pb.CreateBranchRequest, google_protobuf_empty_pb.Empty>;
    inspectBranch: grpc.handleUnaryCall<pfs_pfs_pb.InspectBranchRequest, pfs_pfs_pb.BranchInfo>;
    listBranch: grpc.handleUnaryCall<pfs_pfs_pb.ListBranchRequest, pfs_pfs_pb.BranchInfos>;
    deleteBranch: grpc.handleUnaryCall<pfs_pfs_pb.DeleteBranchRequest, google_protobuf_empty_pb.Empty>;
    modifyFile: handleClientStreamingCall<pfs_pfs_pb.ModifyFileRequest, google_protobuf_empty_pb.Empty>;
    getFileTAR: grpc.handleServerStreamingCall<pfs_pfs_pb.GetFileRequest, google_protobuf_wrappers_pb.BytesValue>;
    inspectFile: grpc.handleUnaryCall<pfs_pfs_pb.InspectFileRequest, pfs_pfs_pb.FileInfo>;
    listFile: grpc.handleServerStreamingCall<pfs_pfs_pb.ListFileRequest, pfs_pfs_pb.FileInfo>;
    walkFile: grpc.handleServerStreamingCall<pfs_pfs_pb.WalkFileRequest, pfs_pfs_pb.FileInfo>;
    globFile: grpc.handleServerStreamingCall<pfs_pfs_pb.GlobFileRequest, pfs_pfs_pb.FileInfo>;
    diffFile: grpc.handleServerStreamingCall<pfs_pfs_pb.DiffFileRequest, pfs_pfs_pb.DiffFileResponse>;
    activateAuth: grpc.handleUnaryCall<pfs_pfs_pb.ActivateAuthRequest, pfs_pfs_pb.ActivateAuthResponse>;
    deleteAll: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty>;
    fsck: grpc.handleServerStreamingCall<pfs_pfs_pb.FsckRequest, pfs_pfs_pb.FsckResponse>;
    createFileSet: handleClientStreamingCall<pfs_pfs_pb.ModifyFileRequest, pfs_pfs_pb.CreateFileSetResponse>;
    getFileSet: grpc.handleUnaryCall<pfs_pfs_pb.GetFileSetRequest, pfs_pfs_pb.CreateFileSetResponse>;
    addFileSet: grpc.handleUnaryCall<pfs_pfs_pb.AddFileSetRequest, google_protobuf_empty_pb.Empty>;
    renewFileSet: grpc.handleUnaryCall<pfs_pfs_pb.RenewFileSetRequest, google_protobuf_empty_pb.Empty>;
    runLoadTest: grpc.handleUnaryCall<pfs_pfs_pb.RunLoadTestRequest, pfs_pfs_pb.RunLoadTestResponse>;
}

export interface IAPIClient {
    createRepo(request: pfs_pfs_pb.CreateRepoRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createRepo(request: pfs_pfs_pb.CreateRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createRepo(request: pfs_pfs_pb.CreateRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    inspectRepo(request: pfs_pfs_pb.InspectRepoRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    inspectRepo(request: pfs_pfs_pb.InspectRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    inspectRepo(request: pfs_pfs_pb.InspectRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    listRepo(request: pfs_pfs_pb.ListRepoRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    listRepo(request: pfs_pfs_pb.ListRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    listRepo(request: pfs_pfs_pb.ListRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    deleteRepo(request: pfs_pfs_pb.DeleteRepoRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteRepo(request: pfs_pfs_pb.DeleteRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteRepo(request: pfs_pfs_pb.DeleteRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    startCommit(request: pfs_pfs_pb.StartCommitRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    startCommit(request: pfs_pfs_pb.StartCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    startCommit(request: pfs_pfs_pb.StartCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    finishCommit(request: pfs_pfs_pb.FinishCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    finishCommit(request: pfs_pfs_pb.FinishCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    finishCommit(request: pfs_pfs_pb.FinishCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    clearCommit(request: pfs_pfs_pb.ClearCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    clearCommit(request: pfs_pfs_pb.ClearCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    clearCommit(request: pfs_pfs_pb.ClearCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    listCommit(request: pfs_pfs_pb.ListCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    listCommit(request: pfs_pfs_pb.ListCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    subscribeCommit(request: pfs_pfs_pb.SubscribeCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    subscribeCommit(request: pfs_pfs_pb.SubscribeCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    inspectCommitSet(request: pfs_pfs_pb.InspectCommitSetRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    inspectCommitSet(request: pfs_pfs_pb.InspectCommitSetRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    squashCommitSet(request: pfs_pfs_pb.SquashCommitSetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    squashCommitSet(request: pfs_pfs_pb.SquashCommitSetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    squashCommitSet(request: pfs_pfs_pb.SquashCommitSetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createBranch(request: pfs_pfs_pb.CreateBranchRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createBranch(request: pfs_pfs_pb.CreateBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createBranch(request: pfs_pfs_pb.CreateBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    inspectBranch(request: pfs_pfs_pb.InspectBranchRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    inspectBranch(request: pfs_pfs_pb.InspectBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    inspectBranch(request: pfs_pfs_pb.InspectBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    listBranch(request: pfs_pfs_pb.ListBranchRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    listBranch(request: pfs_pfs_pb.ListBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    listBranch(request: pfs_pfs_pb.ListBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    deleteBranch(request: pfs_pfs_pb.DeleteBranchRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteBranch(request: pfs_pfs_pb.DeleteBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteBranch(request: pfs_pfs_pb.DeleteBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    modifyFile(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    modifyFile(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    modifyFile(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    modifyFile(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    getFileTAR(request: pfs_pfs_pb.GetFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getFileTAR(request: pfs_pfs_pb.GetFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    inspectFile(request: pfs_pfs_pb.InspectFileRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    inspectFile(request: pfs_pfs_pb.InspectFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    inspectFile(request: pfs_pfs_pb.InspectFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    listFile(request: pfs_pfs_pb.ListFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    listFile(request: pfs_pfs_pb.ListFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    walkFile(request: pfs_pfs_pb.WalkFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    walkFile(request: pfs_pfs_pb.WalkFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    globFile(request: pfs_pfs_pb.GlobFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    globFile(request: pfs_pfs_pb.GlobFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    diffFile(request: pfs_pfs_pb.DiffFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.DiffFileResponse>;
    diffFile(request: pfs_pfs_pb.DiffFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.DiffFileResponse>;
    activateAuth(request: pfs_pfs_pb.ActivateAuthRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    activateAuth(request: pfs_pfs_pb.ActivateAuthRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    activateAuth(request: pfs_pfs_pb.ActivateAuthRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    deleteAll(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    fsck(request: pfs_pfs_pb.FsckRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FsckResponse>;
    fsck(request: pfs_pfs_pb.FsckRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FsckResponse>;
    createFileSet(callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    createFileSet(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    createFileSet(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    createFileSet(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    getFileSet(request: pfs_pfs_pb.GetFileSetRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientUnaryCall;
    getFileSet(request: pfs_pfs_pb.GetFileSetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientUnaryCall;
    getFileSet(request: pfs_pfs_pb.GetFileSetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientUnaryCall;
    addFileSet(request: pfs_pfs_pb.AddFileSetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    addFileSet(request: pfs_pfs_pb.AddFileSetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    addFileSet(request: pfs_pfs_pb.AddFileSetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    renewFileSet(request: pfs_pfs_pb.RenewFileSetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    renewFileSet(request: pfs_pfs_pb.RenewFileSetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    renewFileSet(request: pfs_pfs_pb.RenewFileSetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    runLoadTest(request: pfs_pfs_pb.RunLoadTestRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    runLoadTest(request: pfs_pfs_pb.RunLoadTestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    runLoadTest(request: pfs_pfs_pb.RunLoadTestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public createRepo(request: pfs_pfs_pb.CreateRepoRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createRepo(request: pfs_pfs_pb.CreateRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createRepo(request: pfs_pfs_pb.CreateRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public inspectRepo(request: pfs_pfs_pb.InspectRepoRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    public inspectRepo(request: pfs_pfs_pb.InspectRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    public inspectRepo(request: pfs_pfs_pb.InspectRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    public listRepo(request: pfs_pfs_pb.ListRepoRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    public listRepo(request: pfs_pfs_pb.ListRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    public listRepo(request: pfs_pfs_pb.ListRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    public deleteRepo(request: pfs_pfs_pb.DeleteRepoRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteRepo(request: pfs_pfs_pb.DeleteRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteRepo(request: pfs_pfs_pb.DeleteRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public startCommit(request: pfs_pfs_pb.StartCommitRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    public startCommit(request: pfs_pfs_pb.StartCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    public startCommit(request: pfs_pfs_pb.StartCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    public finishCommit(request: pfs_pfs_pb.FinishCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public finishCommit(request: pfs_pfs_pb.FinishCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public finishCommit(request: pfs_pfs_pb.FinishCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public clearCommit(request: pfs_pfs_pb.ClearCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public clearCommit(request: pfs_pfs_pb.ClearCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public clearCommit(request: pfs_pfs_pb.ClearCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    public inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    public inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    public listCommit(request: pfs_pfs_pb.ListCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public listCommit(request: pfs_pfs_pb.ListCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public subscribeCommit(request: pfs_pfs_pb.SubscribeCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public subscribeCommit(request: pfs_pfs_pb.SubscribeCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public inspectCommitSet(request: pfs_pfs_pb.InspectCommitSetRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public inspectCommitSet(request: pfs_pfs_pb.InspectCommitSetRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public squashCommitSet(request: pfs_pfs_pb.SquashCommitSetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public squashCommitSet(request: pfs_pfs_pb.SquashCommitSetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public squashCommitSet(request: pfs_pfs_pb.SquashCommitSetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createBranch(request: pfs_pfs_pb.CreateBranchRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createBranch(request: pfs_pfs_pb.CreateBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createBranch(request: pfs_pfs_pb.CreateBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public inspectBranch(request: pfs_pfs_pb.InspectBranchRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    public inspectBranch(request: pfs_pfs_pb.InspectBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    public inspectBranch(request: pfs_pfs_pb.InspectBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    public listBranch(request: pfs_pfs_pb.ListBranchRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    public listBranch(request: pfs_pfs_pb.ListBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    public listBranch(request: pfs_pfs_pb.ListBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    public deleteBranch(request: pfs_pfs_pb.DeleteBranchRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteBranch(request: pfs_pfs_pb.DeleteBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteBranch(request: pfs_pfs_pb.DeleteBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public modifyFile(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public modifyFile(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public modifyFile(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public modifyFile(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public getFileTAR(request: pfs_pfs_pb.GetFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getFileTAR(request: pfs_pfs_pb.GetFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public inspectFile(request: pfs_pfs_pb.InspectFileRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    public inspectFile(request: pfs_pfs_pb.InspectFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    public inspectFile(request: pfs_pfs_pb.InspectFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    public listFile(request: pfs_pfs_pb.ListFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    public listFile(request: pfs_pfs_pb.ListFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    public walkFile(request: pfs_pfs_pb.WalkFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    public walkFile(request: pfs_pfs_pb.WalkFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    public globFile(request: pfs_pfs_pb.GlobFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    public globFile(request: pfs_pfs_pb.GlobFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FileInfo>;
    public diffFile(request: pfs_pfs_pb.DiffFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.DiffFileResponse>;
    public diffFile(request: pfs_pfs_pb.DiffFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.DiffFileResponse>;
    public activateAuth(request: pfs_pfs_pb.ActivateAuthRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    public activateAuth(request: pfs_pfs_pb.ActivateAuthRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    public activateAuth(request: pfs_pfs_pb.ActivateAuthRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    public deleteAll(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public fsck(request: pfs_pfs_pb.FsckRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FsckResponse>;
    public fsck(request: pfs_pfs_pb.FsckRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.FsckResponse>;
    public createFileSet(callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public createFileSet(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public createFileSet(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public createFileSet(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public getFileSet(request: pfs_pfs_pb.GetFileSetRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientUnaryCall;
    public getFileSet(request: pfs_pfs_pb.GetFileSetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientUnaryCall;
    public getFileSet(request: pfs_pfs_pb.GetFileSetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFileSetResponse) => void): grpc.ClientUnaryCall;
    public addFileSet(request: pfs_pfs_pb.AddFileSetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public addFileSet(request: pfs_pfs_pb.AddFileSetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public addFileSet(request: pfs_pfs_pb.AddFileSetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public renewFileSet(request: pfs_pfs_pb.RenewFileSetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public renewFileSet(request: pfs_pfs_pb.RenewFileSetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public renewFileSet(request: pfs_pfs_pb.RenewFileSetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public runLoadTest(request: pfs_pfs_pb.RunLoadTestRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    public runLoadTest(request: pfs_pfs_pb.RunLoadTestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    public runLoadTest(request: pfs_pfs_pb.RunLoadTestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
}
