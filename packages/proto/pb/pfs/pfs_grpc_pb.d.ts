// package: pfs
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
    inspectCommit: IAPIService_IInspectCommit;
    listCommit: IAPIService_IListCommit;
    squashCommit: IAPIService_ISquashCommit;
    flushCommit: IAPIService_IFlushCommit;
    subscribeCommit: IAPIService_ISubscribeCommit;
    clearCommit: IAPIService_IClearCommit;
    createBranch: IAPIService_ICreateBranch;
    inspectBranch: IAPIService_IInspectBranch;
    listBranch: IAPIService_IListBranch;
    deleteBranch: IAPIService_IDeleteBranch;
    modifyFile: IAPIService_IModifyFile;
    getFile: IAPIService_IGetFile;
    inspectFile: IAPIService_IInspectFile;
    listFile: IAPIService_IListFile;
    walkFile: IAPIService_IWalkFile;
    globFile: IAPIService_IGlobFile;
    diffFile: IAPIService_IDiffFile;
    activateAuth: IAPIService_IActivateAuth;
    deleteAll: IAPIService_IDeleteAll;
    fsck: IAPIService_IFsck;
    createFileset: IAPIService_ICreateFileset;
    getFileset: IAPIService_IGetFileset;
    addFileset: IAPIService_IAddFileset;
    renewFileset: IAPIService_IRenewFileset;
}

interface IAPIService_ICreateRepo extends grpc.MethodDefinition<pfs_pfs_pb.CreateRepoRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/CreateRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.CreateRepoRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.CreateRepoRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectRepo extends grpc.MethodDefinition<pfs_pfs_pb.InspectRepoRequest, pfs_pfs_pb.RepoInfo> {
    path: "/pfs.API/InspectRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.InspectRepoRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.InspectRepoRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.RepoInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.RepoInfo>;
}
interface IAPIService_IListRepo extends grpc.MethodDefinition<pfs_pfs_pb.ListRepoRequest, pfs_pfs_pb.ListRepoResponse> {
    path: "/pfs.API/ListRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ListRepoRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ListRepoRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.ListRepoResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.ListRepoResponse>;
}
interface IAPIService_IDeleteRepo extends grpc.MethodDefinition<pfs_pfs_pb.DeleteRepoRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/DeleteRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.DeleteRepoRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.DeleteRepoRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IStartCommit extends grpc.MethodDefinition<pfs_pfs_pb.StartCommitRequest, pfs_pfs_pb.Commit> {
    path: "/pfs.API/StartCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.StartCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.StartCommitRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.Commit>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.Commit>;
}
interface IAPIService_IFinishCommit extends grpc.MethodDefinition<pfs_pfs_pb.FinishCommitRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/FinishCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.FinishCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.FinishCommitRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectCommit extends grpc.MethodDefinition<pfs_pfs_pb.InspectCommitRequest, pfs_pfs_pb.CommitInfo> {
    path: "/pfs.API/InspectCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.InspectCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.InspectCommitRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_IListCommit extends grpc.MethodDefinition<pfs_pfs_pb.ListCommitRequest, pfs_pfs_pb.CommitInfo> {
    path: "/pfs.API/ListCommit";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ListCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ListCommitRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_ISquashCommit extends grpc.MethodDefinition<pfs_pfs_pb.SquashCommitRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/SquashCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.SquashCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.SquashCommitRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IFlushCommit extends grpc.MethodDefinition<pfs_pfs_pb.FlushCommitRequest, pfs_pfs_pb.CommitInfo> {
    path: "/pfs.API/FlushCommit";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.FlushCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.FlushCommitRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_ISubscribeCommit extends grpc.MethodDefinition<pfs_pfs_pb.SubscribeCommitRequest, pfs_pfs_pb.CommitInfo> {
    path: "/pfs.API/SubscribeCommit";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.SubscribeCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.SubscribeCommitRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_IClearCommit extends grpc.MethodDefinition<pfs_pfs_pb.ClearCommitRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/ClearCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ClearCommitRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ClearCommitRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_ICreateBranch extends grpc.MethodDefinition<pfs_pfs_pb.CreateBranchRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/CreateBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.CreateBranchRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.CreateBranchRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectBranch extends grpc.MethodDefinition<pfs_pfs_pb.InspectBranchRequest, pfs_pfs_pb.BranchInfo> {
    path: "/pfs.API/InspectBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.InspectBranchRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.InspectBranchRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.BranchInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.BranchInfo>;
}
interface IAPIService_IListBranch extends grpc.MethodDefinition<pfs_pfs_pb.ListBranchRequest, pfs_pfs_pb.BranchInfos> {
    path: "/pfs.API/ListBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ListBranchRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ListBranchRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.BranchInfos>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.BranchInfos>;
}
interface IAPIService_IDeleteBranch extends grpc.MethodDefinition<pfs_pfs_pb.DeleteBranchRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/DeleteBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.DeleteBranchRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.DeleteBranchRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IModifyFile extends grpc.MethodDefinition<pfs_pfs_pb.ModifyFileRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/ModifyFile";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ModifyFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ModifyFileRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IGetFile extends grpc.MethodDefinition<pfs_pfs_pb.GetFileRequest, google_protobuf_wrappers_pb.BytesValue> {
    path: "/pfs.API/GetFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.GetFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.GetFileRequest>;
    responseSerialize: grpc.serialize<google_protobuf_wrappers_pb.BytesValue>;
    responseDeserialize: grpc.deserialize<google_protobuf_wrappers_pb.BytesValue>;
}
interface IAPIService_IInspectFile extends grpc.MethodDefinition<pfs_pfs_pb.InspectFileRequest, pfs_pfs_pb.FileInfo> {
    path: "/pfs.API/InspectFile";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.InspectFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.InspectFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IListFile extends grpc.MethodDefinition<pfs_pfs_pb.ListFileRequest, pfs_pfs_pb.FileInfo> {
    path: "/pfs.API/ListFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ListFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ListFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IWalkFile extends grpc.MethodDefinition<pfs_pfs_pb.WalkFileRequest, pfs_pfs_pb.FileInfo> {
    path: "/pfs.API/WalkFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.WalkFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.WalkFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IGlobFile extends grpc.MethodDefinition<pfs_pfs_pb.GlobFileRequest, pfs_pfs_pb.FileInfo> {
    path: "/pfs.API/GlobFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.GlobFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.GlobFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IDiffFile extends grpc.MethodDefinition<pfs_pfs_pb.DiffFileRequest, pfs_pfs_pb.DiffFileResponse> {
    path: "/pfs.API/DiffFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.DiffFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.DiffFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.DiffFileResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.DiffFileResponse>;
}
interface IAPIService_IActivateAuth extends grpc.MethodDefinition<pfs_pfs_pb.ActivateAuthRequest, pfs_pfs_pb.ActivateAuthResponse> {
    path: "/pfs.API/ActivateAuth";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ActivateAuthRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ActivateAuthRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.ActivateAuthResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.ActivateAuthResponse>;
}
interface IAPIService_IDeleteAll extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/DeleteAll";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IFsck extends grpc.MethodDefinition<pfs_pfs_pb.FsckRequest, pfs_pfs_pb.FsckResponse> {
    path: "/pfs.API/Fsck";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pfs_pfs_pb.FsckRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.FsckRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.FsckResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.FsckResponse>;
}
interface IAPIService_ICreateFileset extends grpc.MethodDefinition<pfs_pfs_pb.ModifyFileRequest, pfs_pfs_pb.CreateFilesetResponse> {
    path: "/pfs.API/CreateFileset";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.ModifyFileRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.ModifyFileRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CreateFilesetResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CreateFilesetResponse>;
}
interface IAPIService_IGetFileset extends grpc.MethodDefinition<pfs_pfs_pb.GetFilesetRequest, pfs_pfs_pb.CreateFilesetResponse> {
    path: "/pfs.API/GetFileset";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.GetFilesetRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.GetFilesetRequest>;
    responseSerialize: grpc.serialize<pfs_pfs_pb.CreateFilesetResponse>;
    responseDeserialize: grpc.deserialize<pfs_pfs_pb.CreateFilesetResponse>;
}
interface IAPIService_IAddFileset extends grpc.MethodDefinition<pfs_pfs_pb.AddFilesetRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/AddFileset";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.AddFilesetRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.AddFilesetRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IRenewFileset extends grpc.MethodDefinition<pfs_pfs_pb.RenewFilesetRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/RenewFileset";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pfs_pfs_pb.RenewFilesetRequest>;
    requestDeserialize: grpc.deserialize<pfs_pfs_pb.RenewFilesetRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}

export const APIService: IAPIService;

export interface IAPIServer extends grpc.UntypedServiceImplementation {
    createRepo: grpc.handleUnaryCall<pfs_pfs_pb.CreateRepoRequest, google_protobuf_empty_pb.Empty>;
    inspectRepo: grpc.handleUnaryCall<pfs_pfs_pb.InspectRepoRequest, pfs_pfs_pb.RepoInfo>;
    listRepo: grpc.handleUnaryCall<pfs_pfs_pb.ListRepoRequest, pfs_pfs_pb.ListRepoResponse>;
    deleteRepo: grpc.handleUnaryCall<pfs_pfs_pb.DeleteRepoRequest, google_protobuf_empty_pb.Empty>;
    startCommit: grpc.handleUnaryCall<pfs_pfs_pb.StartCommitRequest, pfs_pfs_pb.Commit>;
    finishCommit: grpc.handleUnaryCall<pfs_pfs_pb.FinishCommitRequest, google_protobuf_empty_pb.Empty>;
    inspectCommit: grpc.handleUnaryCall<pfs_pfs_pb.InspectCommitRequest, pfs_pfs_pb.CommitInfo>;
    listCommit: grpc.handleServerStreamingCall<pfs_pfs_pb.ListCommitRequest, pfs_pfs_pb.CommitInfo>;
    squashCommit: grpc.handleUnaryCall<pfs_pfs_pb.SquashCommitRequest, google_protobuf_empty_pb.Empty>;
    flushCommit: grpc.handleServerStreamingCall<pfs_pfs_pb.FlushCommitRequest, pfs_pfs_pb.CommitInfo>;
    subscribeCommit: grpc.handleServerStreamingCall<pfs_pfs_pb.SubscribeCommitRequest, pfs_pfs_pb.CommitInfo>;
    clearCommit: grpc.handleUnaryCall<pfs_pfs_pb.ClearCommitRequest, google_protobuf_empty_pb.Empty>;
    createBranch: grpc.handleUnaryCall<pfs_pfs_pb.CreateBranchRequest, google_protobuf_empty_pb.Empty>;
    inspectBranch: grpc.handleUnaryCall<pfs_pfs_pb.InspectBranchRequest, pfs_pfs_pb.BranchInfo>;
    listBranch: grpc.handleUnaryCall<pfs_pfs_pb.ListBranchRequest, pfs_pfs_pb.BranchInfos>;
    deleteBranch: grpc.handleUnaryCall<pfs_pfs_pb.DeleteBranchRequest, google_protobuf_empty_pb.Empty>;
    modifyFile: handleClientStreamingCall<pfs_pfs_pb.ModifyFileRequest, google_protobuf_empty_pb.Empty>;
    getFile: grpc.handleServerStreamingCall<pfs_pfs_pb.GetFileRequest, google_protobuf_wrappers_pb.BytesValue>;
    inspectFile: grpc.handleUnaryCall<pfs_pfs_pb.InspectFileRequest, pfs_pfs_pb.FileInfo>;
    listFile: grpc.handleServerStreamingCall<pfs_pfs_pb.ListFileRequest, pfs_pfs_pb.FileInfo>;
    walkFile: grpc.handleServerStreamingCall<pfs_pfs_pb.WalkFileRequest, pfs_pfs_pb.FileInfo>;
    globFile: grpc.handleServerStreamingCall<pfs_pfs_pb.GlobFileRequest, pfs_pfs_pb.FileInfo>;
    diffFile: grpc.handleServerStreamingCall<pfs_pfs_pb.DiffFileRequest, pfs_pfs_pb.DiffFileResponse>;
    activateAuth: grpc.handleUnaryCall<pfs_pfs_pb.ActivateAuthRequest, pfs_pfs_pb.ActivateAuthResponse>;
    deleteAll: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty>;
    fsck: grpc.handleServerStreamingCall<pfs_pfs_pb.FsckRequest, pfs_pfs_pb.FsckResponse>;
    createFileset: handleClientStreamingCall<pfs_pfs_pb.ModifyFileRequest, pfs_pfs_pb.CreateFilesetResponse>;
    getFileset: grpc.handleUnaryCall<pfs_pfs_pb.GetFilesetRequest, pfs_pfs_pb.CreateFilesetResponse>;
    addFileset: grpc.handleUnaryCall<pfs_pfs_pb.AddFilesetRequest, google_protobuf_empty_pb.Empty>;
    renewFileset: grpc.handleUnaryCall<pfs_pfs_pb.RenewFilesetRequest, google_protobuf_empty_pb.Empty>;
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
    inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    listCommit(request: pfs_pfs_pb.ListCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    listCommit(request: pfs_pfs_pb.ListCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    squashCommit(request: pfs_pfs_pb.SquashCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    squashCommit(request: pfs_pfs_pb.SquashCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    squashCommit(request: pfs_pfs_pb.SquashCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    flushCommit(request: pfs_pfs_pb.FlushCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    flushCommit(request: pfs_pfs_pb.FlushCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    subscribeCommit(request: pfs_pfs_pb.SubscribeCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    subscribeCommit(request: pfs_pfs_pb.SubscribeCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    clearCommit(request: pfs_pfs_pb.ClearCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    clearCommit(request: pfs_pfs_pb.ClearCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    clearCommit(request: pfs_pfs_pb.ClearCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
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
    getFile(request: pfs_pfs_pb.GetFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getFile(request: pfs_pfs_pb.GetFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
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
    createFileset(callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    createFileset(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    createFileset(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    createFileset(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    getFileset(request: pfs_pfs_pb.GetFilesetRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientUnaryCall;
    getFileset(request: pfs_pfs_pb.GetFilesetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientUnaryCall;
    getFileset(request: pfs_pfs_pb.GetFilesetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientUnaryCall;
    addFileset(request: pfs_pfs_pb.AddFilesetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    addFileset(request: pfs_pfs_pb.AddFilesetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    addFileset(request: pfs_pfs_pb.AddFilesetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    renewFileset(request: pfs_pfs_pb.RenewFilesetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    renewFileset(request: pfs_pfs_pb.RenewFilesetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    renewFileset(request: pfs_pfs_pb.RenewFilesetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
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
    public inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    public inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    public inspectCommit(request: pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    public listCommit(request: pfs_pfs_pb.ListCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public listCommit(request: pfs_pfs_pb.ListCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public squashCommit(request: pfs_pfs_pb.SquashCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public squashCommit(request: pfs_pfs_pb.SquashCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public squashCommit(request: pfs_pfs_pb.SquashCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public flushCommit(request: pfs_pfs_pb.FlushCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public flushCommit(request: pfs_pfs_pb.FlushCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public subscribeCommit(request: pfs_pfs_pb.SubscribeCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public subscribeCommit(request: pfs_pfs_pb.SubscribeCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pfs_pfs_pb.CommitInfo>;
    public clearCommit(request: pfs_pfs_pb.ClearCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public clearCommit(request: pfs_pfs_pb.ClearCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public clearCommit(request: pfs_pfs_pb.ClearCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
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
    public getFile(request: pfs_pfs_pb.GetFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getFile(request: pfs_pfs_pb.GetFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
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
    public createFileset(callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public createFileset(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public createFileset(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public createFileset(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientWritableStream<pfs_pfs_pb.ModifyFileRequest>;
    public getFileset(request: pfs_pfs_pb.GetFilesetRequest, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientUnaryCall;
    public getFileset(request: pfs_pfs_pb.GetFilesetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientUnaryCall;
    public getFileset(request: pfs_pfs_pb.GetFilesetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pfs_pfs_pb.CreateFilesetResponse) => void): grpc.ClientUnaryCall;
    public addFileset(request: pfs_pfs_pb.AddFilesetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public addFileset(request: pfs_pfs_pb.AddFilesetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public addFileset(request: pfs_pfs_pb.AddFilesetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public renewFileset(request: pfs_pfs_pb.RenewFilesetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public renewFileset(request: pfs_pfs_pb.RenewFilesetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public renewFileset(request: pfs_pfs_pb.RenewFilesetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
}
