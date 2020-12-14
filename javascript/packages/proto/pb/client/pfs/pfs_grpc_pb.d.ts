// package: pfs
// file: client/pfs/pfs.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as client_pfs_pfs_pb from "../../client/pfs/pfs_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as google_protobuf_wrappers_pb from "google-protobuf/google/protobuf/wrappers_pb";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";
import * as client_auth_auth_pb from "../../client/auth/auth_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    createRepo: IAPIService_ICreateRepo;
    inspectRepo: IAPIService_IInspectRepo;
    listRepo: IAPIService_IListRepo;
    deleteRepo: IAPIService_IDeleteRepo;
    startCommit: IAPIService_IStartCommit;
    finishCommit: IAPIService_IFinishCommit;
    inspectCommit: IAPIService_IInspectCommit;
    listCommit: IAPIService_IListCommit;
    listCommitStream: IAPIService_IListCommitStream;
    deleteCommit: IAPIService_IDeleteCommit;
    flushCommit: IAPIService_IFlushCommit;
    subscribeCommit: IAPIService_ISubscribeCommit;
    buildCommit: IAPIService_IBuildCommit;
    createBranch: IAPIService_ICreateBranch;
    inspectBranch: IAPIService_IInspectBranch;
    listBranch: IAPIService_IListBranch;
    deleteBranch: IAPIService_IDeleteBranch;
    putFile: IAPIService_IPutFile;
    copyFile: IAPIService_ICopyFile;
    getFile: IAPIService_IGetFile;
    inspectFile: IAPIService_IInspectFile;
    listFile: IAPIService_IListFile;
    listFileStream: IAPIService_IListFileStream;
    walkFile: IAPIService_IWalkFile;
    globFile: IAPIService_IGlobFile;
    globFileStream: IAPIService_IGlobFileStream;
    diffFile: IAPIService_IDiffFile;
    deleteFile: IAPIService_IDeleteFile;
    deleteAll: IAPIService_IDeleteAll;
    fsck: IAPIService_IFsck;
    fileOperationV2: IAPIService_IFileOperationV2;
    getTarV2: IAPIService_IGetTarV2;
    diffFileV2: IAPIService_IDiffFileV2;
    createTmpFileSet: IAPIService_ICreateTmpFileSet;
    renewTmpFileSet: IAPIService_IRenewTmpFileSet;
    clearCommitV2: IAPIService_IClearCommitV2;
}

interface IAPIService_ICreateRepo extends grpc.MethodDefinition<client_pfs_pfs_pb.CreateRepoRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/CreateRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.CreateRepoRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.CreateRepoRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectRepo extends grpc.MethodDefinition<client_pfs_pfs_pb.InspectRepoRequest, client_pfs_pfs_pb.RepoInfo> {
    path: "/pfs.API/InspectRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.InspectRepoRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.InspectRepoRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.RepoInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.RepoInfo>;
}
interface IAPIService_IListRepo extends grpc.MethodDefinition<client_pfs_pfs_pb.ListRepoRequest, client_pfs_pfs_pb.ListRepoResponse> {
    path: "/pfs.API/ListRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.ListRepoRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListRepoRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.ListRepoResponse>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListRepoResponse>;
}
interface IAPIService_IDeleteRepo extends grpc.MethodDefinition<client_pfs_pfs_pb.DeleteRepoRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/DeleteRepo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.DeleteRepoRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.DeleteRepoRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IStartCommit extends grpc.MethodDefinition<client_pfs_pfs_pb.StartCommitRequest, client_pfs_pfs_pb.Commit> {
    path: "/pfs.API/StartCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.StartCommitRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.StartCommitRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.Commit>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.Commit>;
}
interface IAPIService_IFinishCommit extends grpc.MethodDefinition<client_pfs_pfs_pb.FinishCommitRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/FinishCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.FinishCommitRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.FinishCommitRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectCommit extends grpc.MethodDefinition<client_pfs_pfs_pb.InspectCommitRequest, client_pfs_pfs_pb.CommitInfo> {
    path: "/pfs.API/InspectCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.InspectCommitRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.InspectCommitRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_IListCommit extends grpc.MethodDefinition<client_pfs_pfs_pb.ListCommitRequest, client_pfs_pfs_pb.CommitInfos> {
    path: "/pfs.API/ListCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.ListCommitRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListCommitRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.CommitInfos>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.CommitInfos>;
}
interface IAPIService_IListCommitStream extends grpc.MethodDefinition<client_pfs_pfs_pb.ListCommitRequest, client_pfs_pfs_pb.CommitInfo> {
    path: "/pfs.API/ListCommitStream";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.ListCommitRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListCommitRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_IDeleteCommit extends grpc.MethodDefinition<client_pfs_pfs_pb.DeleteCommitRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/DeleteCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.DeleteCommitRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.DeleteCommitRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IFlushCommit extends grpc.MethodDefinition<client_pfs_pfs_pb.FlushCommitRequest, client_pfs_pfs_pb.CommitInfo> {
    path: "/pfs.API/FlushCommit";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.FlushCommitRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.FlushCommitRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_ISubscribeCommit extends grpc.MethodDefinition<client_pfs_pfs_pb.SubscribeCommitRequest, client_pfs_pfs_pb.CommitInfo> {
    path: "/pfs.API/SubscribeCommit";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.SubscribeCommitRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.SubscribeCommitRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.CommitInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.CommitInfo>;
}
interface IAPIService_IBuildCommit extends grpc.MethodDefinition<client_pfs_pfs_pb.BuildCommitRequest, client_pfs_pfs_pb.Commit> {
    path: "/pfs.API/BuildCommit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.BuildCommitRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.BuildCommitRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.Commit>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.Commit>;
}
interface IAPIService_ICreateBranch extends grpc.MethodDefinition<client_pfs_pfs_pb.CreateBranchRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/CreateBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.CreateBranchRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.CreateBranchRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectBranch extends grpc.MethodDefinition<client_pfs_pfs_pb.InspectBranchRequest, client_pfs_pfs_pb.BranchInfo> {
    path: "/pfs.API/InspectBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.InspectBranchRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.InspectBranchRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.BranchInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.BranchInfo>;
}
interface IAPIService_IListBranch extends grpc.MethodDefinition<client_pfs_pfs_pb.ListBranchRequest, client_pfs_pfs_pb.BranchInfos> {
    path: "/pfs.API/ListBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.ListBranchRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListBranchRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.BranchInfos>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.BranchInfos>;
}
interface IAPIService_IDeleteBranch extends grpc.MethodDefinition<client_pfs_pfs_pb.DeleteBranchRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/DeleteBranch";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.DeleteBranchRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.DeleteBranchRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IPutFile extends grpc.MethodDefinition<client_pfs_pfs_pb.PutFileRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/PutFile";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.PutFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.PutFileRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_ICopyFile extends grpc.MethodDefinition<client_pfs_pfs_pb.CopyFileRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/CopyFile";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.CopyFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.CopyFileRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IGetFile extends grpc.MethodDefinition<client_pfs_pfs_pb.GetFileRequest, google_protobuf_wrappers_pb.BytesValue> {
    path: "/pfs.API/GetFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.GetFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.GetFileRequest>;
    responseSerialize: grpc.serialize<google_protobuf_wrappers_pb.BytesValue>;
    responseDeserialize: grpc.deserialize<google_protobuf_wrappers_pb.BytesValue>;
}
interface IAPIService_IInspectFile extends grpc.MethodDefinition<client_pfs_pfs_pb.InspectFileRequest, client_pfs_pfs_pb.FileInfo> {
    path: "/pfs.API/InspectFile";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.InspectFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.InspectFileRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IListFile extends grpc.MethodDefinition<client_pfs_pfs_pb.ListFileRequest, client_pfs_pfs_pb.FileInfos> {
    path: "/pfs.API/ListFile";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.ListFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListFileRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.FileInfos>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.FileInfos>;
}
interface IAPIService_IListFileStream extends grpc.MethodDefinition<client_pfs_pfs_pb.ListFileRequest, client_pfs_pfs_pb.FileInfo> {
    path: "/pfs.API/ListFileStream";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.ListFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListFileRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IWalkFile extends grpc.MethodDefinition<client_pfs_pfs_pb.WalkFileRequest, client_pfs_pfs_pb.FileInfo> {
    path: "/pfs.API/WalkFile";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.WalkFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.WalkFileRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IGlobFile extends grpc.MethodDefinition<client_pfs_pfs_pb.GlobFileRequest, client_pfs_pfs_pb.FileInfos> {
    path: "/pfs.API/GlobFile";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.GlobFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.GlobFileRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.FileInfos>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.FileInfos>;
}
interface IAPIService_IGlobFileStream extends grpc.MethodDefinition<client_pfs_pfs_pb.GlobFileRequest, client_pfs_pfs_pb.FileInfo> {
    path: "/pfs.API/GlobFileStream";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.GlobFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.GlobFileRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.FileInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.FileInfo>;
}
interface IAPIService_IDiffFile extends grpc.MethodDefinition<client_pfs_pfs_pb.DiffFileRequest, client_pfs_pfs_pb.DiffFileResponse> {
    path: "/pfs.API/DiffFile";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.DiffFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.DiffFileRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.DiffFileResponse>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.DiffFileResponse>;
}
interface IAPIService_IDeleteFile extends grpc.MethodDefinition<client_pfs_pfs_pb.DeleteFileRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/DeleteFile";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.DeleteFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.DeleteFileRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
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
interface IAPIService_IFsck extends grpc.MethodDefinition<client_pfs_pfs_pb.FsckRequest, client_pfs_pfs_pb.FsckResponse> {
    path: "/pfs.API/Fsck";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.FsckRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.FsckRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.FsckResponse>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.FsckResponse>;
}
interface IAPIService_IFileOperationV2 extends grpc.MethodDefinition<client_pfs_pfs_pb.FileOperationRequestV2, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/FileOperationV2";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.FileOperationRequestV2>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.FileOperationRequestV2>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IGetTarV2 extends grpc.MethodDefinition<client_pfs_pfs_pb.GetTarRequestV2, google_protobuf_wrappers_pb.BytesValue> {
    path: "/pfs.API/GetTarV2";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.GetTarRequestV2>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.GetTarRequestV2>;
    responseSerialize: grpc.serialize<google_protobuf_wrappers_pb.BytesValue>;
    responseDeserialize: grpc.deserialize<google_protobuf_wrappers_pb.BytesValue>;
}
interface IAPIService_IDiffFileV2 extends grpc.MethodDefinition<client_pfs_pfs_pb.DiffFileRequest, client_pfs_pfs_pb.DiffFileResponseV2> {
    path: "/pfs.API/DiffFileV2";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.DiffFileRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.DiffFileRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.DiffFileResponseV2>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.DiffFileResponseV2>;
}
interface IAPIService_ICreateTmpFileSet extends grpc.MethodDefinition<client_pfs_pfs_pb.FileOperationRequestV2, client_pfs_pfs_pb.CreateTmpFileSetResponse> {
    path: "/pfs.API/CreateTmpFileSet";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.FileOperationRequestV2>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.FileOperationRequestV2>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.CreateTmpFileSetResponse>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.CreateTmpFileSetResponse>;
}
interface IAPIService_IRenewTmpFileSet extends grpc.MethodDefinition<client_pfs_pfs_pb.RenewTmpFileSetRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/RenewTmpFileSet";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.RenewTmpFileSetRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.RenewTmpFileSetRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IClearCommitV2 extends grpc.MethodDefinition<client_pfs_pfs_pb.ClearCommitRequestV2, google_protobuf_empty_pb.Empty> {
    path: "/pfs.API/ClearCommitV2";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.ClearCommitRequestV2>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.ClearCommitRequestV2>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}

export const APIService: IAPIService;

export interface IAPIServer {
    createRepo: grpc.handleUnaryCall<client_pfs_pfs_pb.CreateRepoRequest, google_protobuf_empty_pb.Empty>;
    inspectRepo: grpc.handleUnaryCall<client_pfs_pfs_pb.InspectRepoRequest, client_pfs_pfs_pb.RepoInfo>;
    listRepo: grpc.handleUnaryCall<client_pfs_pfs_pb.ListRepoRequest, client_pfs_pfs_pb.ListRepoResponse>;
    deleteRepo: grpc.handleUnaryCall<client_pfs_pfs_pb.DeleteRepoRequest, google_protobuf_empty_pb.Empty>;
    startCommit: grpc.handleUnaryCall<client_pfs_pfs_pb.StartCommitRequest, client_pfs_pfs_pb.Commit>;
    finishCommit: grpc.handleUnaryCall<client_pfs_pfs_pb.FinishCommitRequest, google_protobuf_empty_pb.Empty>;
    inspectCommit: grpc.handleUnaryCall<client_pfs_pfs_pb.InspectCommitRequest, client_pfs_pfs_pb.CommitInfo>;
    listCommit: grpc.handleUnaryCall<client_pfs_pfs_pb.ListCommitRequest, client_pfs_pfs_pb.CommitInfos>;
    listCommitStream: grpc.handleServerStreamingCall<client_pfs_pfs_pb.ListCommitRequest, client_pfs_pfs_pb.CommitInfo>;
    deleteCommit: grpc.handleUnaryCall<client_pfs_pfs_pb.DeleteCommitRequest, google_protobuf_empty_pb.Empty>;
    flushCommit: grpc.handleServerStreamingCall<client_pfs_pfs_pb.FlushCommitRequest, client_pfs_pfs_pb.CommitInfo>;
    subscribeCommit: grpc.handleServerStreamingCall<client_pfs_pfs_pb.SubscribeCommitRequest, client_pfs_pfs_pb.CommitInfo>;
    buildCommit: grpc.handleUnaryCall<client_pfs_pfs_pb.BuildCommitRequest, client_pfs_pfs_pb.Commit>;
    createBranch: grpc.handleUnaryCall<client_pfs_pfs_pb.CreateBranchRequest, google_protobuf_empty_pb.Empty>;
    inspectBranch: grpc.handleUnaryCall<client_pfs_pfs_pb.InspectBranchRequest, client_pfs_pfs_pb.BranchInfo>;
    listBranch: grpc.handleUnaryCall<client_pfs_pfs_pb.ListBranchRequest, client_pfs_pfs_pb.BranchInfos>;
    deleteBranch: grpc.handleUnaryCall<client_pfs_pfs_pb.DeleteBranchRequest, google_protobuf_empty_pb.Empty>;
    putFile: handleClientStreamingCall<client_pfs_pfs_pb.PutFileRequest, google_protobuf_empty_pb.Empty>;
    copyFile: grpc.handleUnaryCall<client_pfs_pfs_pb.CopyFileRequest, google_protobuf_empty_pb.Empty>;
    getFile: grpc.handleServerStreamingCall<client_pfs_pfs_pb.GetFileRequest, google_protobuf_wrappers_pb.BytesValue>;
    inspectFile: grpc.handleUnaryCall<client_pfs_pfs_pb.InspectFileRequest, client_pfs_pfs_pb.FileInfo>;
    listFile: grpc.handleUnaryCall<client_pfs_pfs_pb.ListFileRequest, client_pfs_pfs_pb.FileInfos>;
    listFileStream: grpc.handleServerStreamingCall<client_pfs_pfs_pb.ListFileRequest, client_pfs_pfs_pb.FileInfo>;
    walkFile: grpc.handleServerStreamingCall<client_pfs_pfs_pb.WalkFileRequest, client_pfs_pfs_pb.FileInfo>;
    globFile: grpc.handleUnaryCall<client_pfs_pfs_pb.GlobFileRequest, client_pfs_pfs_pb.FileInfos>;
    globFileStream: grpc.handleServerStreamingCall<client_pfs_pfs_pb.GlobFileRequest, client_pfs_pfs_pb.FileInfo>;
    diffFile: grpc.handleUnaryCall<client_pfs_pfs_pb.DiffFileRequest, client_pfs_pfs_pb.DiffFileResponse>;
    deleteFile: grpc.handleUnaryCall<client_pfs_pfs_pb.DeleteFileRequest, google_protobuf_empty_pb.Empty>;
    deleteAll: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty>;
    fsck: grpc.handleServerStreamingCall<client_pfs_pfs_pb.FsckRequest, client_pfs_pfs_pb.FsckResponse>;
    fileOperationV2: handleClientStreamingCall<client_pfs_pfs_pb.FileOperationRequestV2, google_protobuf_empty_pb.Empty>;
    getTarV2: grpc.handleServerStreamingCall<client_pfs_pfs_pb.GetTarRequestV2, google_protobuf_wrappers_pb.BytesValue>;
    diffFileV2: grpc.handleServerStreamingCall<client_pfs_pfs_pb.DiffFileRequest, client_pfs_pfs_pb.DiffFileResponseV2>;
    createTmpFileSet: handleClientStreamingCall<client_pfs_pfs_pb.FileOperationRequestV2, client_pfs_pfs_pb.CreateTmpFileSetResponse>;
    renewTmpFileSet: grpc.handleUnaryCall<client_pfs_pfs_pb.RenewTmpFileSetRequest, google_protobuf_empty_pb.Empty>;
    clearCommitV2: grpc.handleUnaryCall<client_pfs_pfs_pb.ClearCommitRequestV2, google_protobuf_empty_pb.Empty>;
}

export interface IAPIClient {
    createRepo(request: client_pfs_pfs_pb.CreateRepoRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createRepo(request: client_pfs_pfs_pb.CreateRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createRepo(request: client_pfs_pfs_pb.CreateRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    inspectRepo(request: client_pfs_pfs_pb.InspectRepoRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    inspectRepo(request: client_pfs_pfs_pb.InspectRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    inspectRepo(request: client_pfs_pfs_pb.InspectRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    listRepo(request: client_pfs_pfs_pb.ListRepoRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    listRepo(request: client_pfs_pfs_pb.ListRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    listRepo(request: client_pfs_pfs_pb.ListRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    deleteRepo(request: client_pfs_pfs_pb.DeleteRepoRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteRepo(request: client_pfs_pfs_pb.DeleteRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteRepo(request: client_pfs_pfs_pb.DeleteRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    startCommit(request: client_pfs_pfs_pb.StartCommitRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    startCommit(request: client_pfs_pfs_pb.StartCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    startCommit(request: client_pfs_pfs_pb.StartCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    finishCommit(request: client_pfs_pfs_pb.FinishCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    finishCommit(request: client_pfs_pfs_pb.FinishCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    finishCommit(request: client_pfs_pfs_pb.FinishCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    inspectCommit(request: client_pfs_pfs_pb.InspectCommitRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    inspectCommit(request: client_pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    inspectCommit(request: client_pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    listCommit(request: client_pfs_pfs_pb.ListCommitRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfos) => void): grpc.ClientUnaryCall;
    listCommit(request: client_pfs_pfs_pb.ListCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfos) => void): grpc.ClientUnaryCall;
    listCommit(request: client_pfs_pfs_pb.ListCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfos) => void): grpc.ClientUnaryCall;
    listCommitStream(request: client_pfs_pfs_pb.ListCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    listCommitStream(request: client_pfs_pfs_pb.ListCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    deleteCommit(request: client_pfs_pfs_pb.DeleteCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteCommit(request: client_pfs_pfs_pb.DeleteCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteCommit(request: client_pfs_pfs_pb.DeleteCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    flushCommit(request: client_pfs_pfs_pb.FlushCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    flushCommit(request: client_pfs_pfs_pb.FlushCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    subscribeCommit(request: client_pfs_pfs_pb.SubscribeCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    subscribeCommit(request: client_pfs_pfs_pb.SubscribeCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    buildCommit(request: client_pfs_pfs_pb.BuildCommitRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    buildCommit(request: client_pfs_pfs_pb.BuildCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    buildCommit(request: client_pfs_pfs_pb.BuildCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    createBranch(request: client_pfs_pfs_pb.CreateBranchRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createBranch(request: client_pfs_pfs_pb.CreateBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createBranch(request: client_pfs_pfs_pb.CreateBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    inspectBranch(request: client_pfs_pfs_pb.InspectBranchRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    inspectBranch(request: client_pfs_pfs_pb.InspectBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    inspectBranch(request: client_pfs_pfs_pb.InspectBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    listBranch(request: client_pfs_pfs_pb.ListBranchRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    listBranch(request: client_pfs_pfs_pb.ListBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    listBranch(request: client_pfs_pfs_pb.ListBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    deleteBranch(request: client_pfs_pfs_pb.DeleteBranchRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteBranch(request: client_pfs_pfs_pb.DeleteBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteBranch(request: client_pfs_pfs_pb.DeleteBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    putFile(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutFileRequest>;
    putFile(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutFileRequest>;
    putFile(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutFileRequest>;
    putFile(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutFileRequest>;
    copyFile(request: client_pfs_pfs_pb.CopyFileRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    copyFile(request: client_pfs_pfs_pb.CopyFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    copyFile(request: client_pfs_pfs_pb.CopyFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    getFile(request: client_pfs_pfs_pb.GetFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getFile(request: client_pfs_pfs_pb.GetFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    inspectFile(request: client_pfs_pfs_pb.InspectFileRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    inspectFile(request: client_pfs_pfs_pb.InspectFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    inspectFile(request: client_pfs_pfs_pb.InspectFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    listFile(request: client_pfs_pfs_pb.ListFileRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    listFile(request: client_pfs_pfs_pb.ListFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    listFile(request: client_pfs_pfs_pb.ListFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    listFileStream(request: client_pfs_pfs_pb.ListFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    listFileStream(request: client_pfs_pfs_pb.ListFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    walkFile(request: client_pfs_pfs_pb.WalkFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    walkFile(request: client_pfs_pfs_pb.WalkFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    globFile(request: client_pfs_pfs_pb.GlobFileRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    globFile(request: client_pfs_pfs_pb.GlobFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    globFile(request: client_pfs_pfs_pb.GlobFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    globFileStream(request: client_pfs_pfs_pb.GlobFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    globFileStream(request: client_pfs_pfs_pb.GlobFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    diffFile(request: client_pfs_pfs_pb.DiffFileRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DiffFileResponse) => void): grpc.ClientUnaryCall;
    diffFile(request: client_pfs_pfs_pb.DiffFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DiffFileResponse) => void): grpc.ClientUnaryCall;
    diffFile(request: client_pfs_pfs_pb.DiffFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DiffFileResponse) => void): grpc.ClientUnaryCall;
    deleteFile(request: client_pfs_pfs_pb.DeleteFileRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteFile(request: client_pfs_pfs_pb.DeleteFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteFile(request: client_pfs_pfs_pb.DeleteFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteAll(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    fsck(request: client_pfs_pfs_pb.FsckRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FsckResponse>;
    fsck(request: client_pfs_pfs_pb.FsckRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FsckResponse>;
    fileOperationV2(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    fileOperationV2(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    fileOperationV2(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    fileOperationV2(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    getTarV2(request: client_pfs_pfs_pb.GetTarRequestV2, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getTarV2(request: client_pfs_pfs_pb.GetTarRequestV2, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    diffFileV2(request: client_pfs_pfs_pb.DiffFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.DiffFileResponseV2>;
    diffFileV2(request: client_pfs_pfs_pb.DiffFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.DiffFileResponseV2>;
    createTmpFileSet(callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CreateTmpFileSetResponse) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    createTmpFileSet(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CreateTmpFileSetResponse) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    createTmpFileSet(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CreateTmpFileSetResponse) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    createTmpFileSet(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CreateTmpFileSetResponse) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    renewTmpFileSet(request: client_pfs_pfs_pb.RenewTmpFileSetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    renewTmpFileSet(request: client_pfs_pfs_pb.RenewTmpFileSetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    renewTmpFileSet(request: client_pfs_pfs_pb.RenewTmpFileSetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    clearCommitV2(request: client_pfs_pfs_pb.ClearCommitRequestV2, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    clearCommitV2(request: client_pfs_pfs_pb.ClearCommitRequestV2, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    clearCommitV2(request: client_pfs_pfs_pb.ClearCommitRequestV2, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public createRepo(request: client_pfs_pfs_pb.CreateRepoRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createRepo(request: client_pfs_pfs_pb.CreateRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createRepo(request: client_pfs_pfs_pb.CreateRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public inspectRepo(request: client_pfs_pfs_pb.InspectRepoRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    public inspectRepo(request: client_pfs_pfs_pb.InspectRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    public inspectRepo(request: client_pfs_pfs_pb.InspectRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.RepoInfo) => void): grpc.ClientUnaryCall;
    public listRepo(request: client_pfs_pfs_pb.ListRepoRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    public listRepo(request: client_pfs_pfs_pb.ListRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    public listRepo(request: client_pfs_pfs_pb.ListRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ListRepoResponse) => void): grpc.ClientUnaryCall;
    public deleteRepo(request: client_pfs_pfs_pb.DeleteRepoRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteRepo(request: client_pfs_pfs_pb.DeleteRepoRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteRepo(request: client_pfs_pfs_pb.DeleteRepoRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public startCommit(request: client_pfs_pfs_pb.StartCommitRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    public startCommit(request: client_pfs_pfs_pb.StartCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    public startCommit(request: client_pfs_pfs_pb.StartCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    public finishCommit(request: client_pfs_pfs_pb.FinishCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public finishCommit(request: client_pfs_pfs_pb.FinishCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public finishCommit(request: client_pfs_pfs_pb.FinishCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public inspectCommit(request: client_pfs_pfs_pb.InspectCommitRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    public inspectCommit(request: client_pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    public inspectCommit(request: client_pfs_pfs_pb.InspectCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfo) => void): grpc.ClientUnaryCall;
    public listCommit(request: client_pfs_pfs_pb.ListCommitRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfos) => void): grpc.ClientUnaryCall;
    public listCommit(request: client_pfs_pfs_pb.ListCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfos) => void): grpc.ClientUnaryCall;
    public listCommit(request: client_pfs_pfs_pb.ListCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CommitInfos) => void): grpc.ClientUnaryCall;
    public listCommitStream(request: client_pfs_pfs_pb.ListCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    public listCommitStream(request: client_pfs_pfs_pb.ListCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    public deleteCommit(request: client_pfs_pfs_pb.DeleteCommitRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteCommit(request: client_pfs_pfs_pb.DeleteCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteCommit(request: client_pfs_pfs_pb.DeleteCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public flushCommit(request: client_pfs_pfs_pb.FlushCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    public flushCommit(request: client_pfs_pfs_pb.FlushCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    public subscribeCommit(request: client_pfs_pfs_pb.SubscribeCommitRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    public subscribeCommit(request: client_pfs_pfs_pb.SubscribeCommitRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.CommitInfo>;
    public buildCommit(request: client_pfs_pfs_pb.BuildCommitRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    public buildCommit(request: client_pfs_pfs_pb.BuildCommitRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    public buildCommit(request: client_pfs_pfs_pb.BuildCommitRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Commit) => void): grpc.ClientUnaryCall;
    public createBranch(request: client_pfs_pfs_pb.CreateBranchRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createBranch(request: client_pfs_pfs_pb.CreateBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createBranch(request: client_pfs_pfs_pb.CreateBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public inspectBranch(request: client_pfs_pfs_pb.InspectBranchRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    public inspectBranch(request: client_pfs_pfs_pb.InspectBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    public inspectBranch(request: client_pfs_pfs_pb.InspectBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfo) => void): grpc.ClientUnaryCall;
    public listBranch(request: client_pfs_pfs_pb.ListBranchRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    public listBranch(request: client_pfs_pfs_pb.ListBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    public listBranch(request: client_pfs_pfs_pb.ListBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.BranchInfos) => void): grpc.ClientUnaryCall;
    public deleteBranch(request: client_pfs_pfs_pb.DeleteBranchRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteBranch(request: client_pfs_pfs_pb.DeleteBranchRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteBranch(request: client_pfs_pfs_pb.DeleteBranchRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public putFile(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutFileRequest>;
    public putFile(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutFileRequest>;
    public putFile(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutFileRequest>;
    public putFile(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutFileRequest>;
    public copyFile(request: client_pfs_pfs_pb.CopyFileRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public copyFile(request: client_pfs_pfs_pb.CopyFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public copyFile(request: client_pfs_pfs_pb.CopyFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public getFile(request: client_pfs_pfs_pb.GetFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getFile(request: client_pfs_pfs_pb.GetFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public inspectFile(request: client_pfs_pfs_pb.InspectFileRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    public inspectFile(request: client_pfs_pfs_pb.InspectFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    public inspectFile(request: client_pfs_pfs_pb.InspectFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfo) => void): grpc.ClientUnaryCall;
    public listFile(request: client_pfs_pfs_pb.ListFileRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    public listFile(request: client_pfs_pfs_pb.ListFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    public listFile(request: client_pfs_pfs_pb.ListFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    public listFileStream(request: client_pfs_pfs_pb.ListFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    public listFileStream(request: client_pfs_pfs_pb.ListFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    public walkFile(request: client_pfs_pfs_pb.WalkFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    public walkFile(request: client_pfs_pfs_pb.WalkFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    public globFile(request: client_pfs_pfs_pb.GlobFileRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    public globFile(request: client_pfs_pfs_pb.GlobFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    public globFile(request: client_pfs_pfs_pb.GlobFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.FileInfos) => void): grpc.ClientUnaryCall;
    public globFileStream(request: client_pfs_pfs_pb.GlobFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    public globFileStream(request: client_pfs_pfs_pb.GlobFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FileInfo>;
    public diffFile(request: client_pfs_pfs_pb.DiffFileRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DiffFileResponse) => void): grpc.ClientUnaryCall;
    public diffFile(request: client_pfs_pfs_pb.DiffFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DiffFileResponse) => void): grpc.ClientUnaryCall;
    public diffFile(request: client_pfs_pfs_pb.DiffFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DiffFileResponse) => void): grpc.ClientUnaryCall;
    public deleteFile(request: client_pfs_pfs_pb.DeleteFileRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteFile(request: client_pfs_pfs_pb.DeleteFileRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteFile(request: client_pfs_pfs_pb.DeleteFileRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteAll(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public fsck(request: client_pfs_pfs_pb.FsckRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FsckResponse>;
    public fsck(request: client_pfs_pfs_pb.FsckRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.FsckResponse>;
    public fileOperationV2(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    public fileOperationV2(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    public fileOperationV2(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    public fileOperationV2(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    public getTarV2(request: client_pfs_pfs_pb.GetTarRequestV2, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getTarV2(request: client_pfs_pfs_pb.GetTarRequestV2, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public diffFileV2(request: client_pfs_pfs_pb.DiffFileRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.DiffFileResponseV2>;
    public diffFileV2(request: client_pfs_pfs_pb.DiffFileRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.DiffFileResponseV2>;
    public createTmpFileSet(callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CreateTmpFileSetResponse) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    public createTmpFileSet(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CreateTmpFileSetResponse) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    public createTmpFileSet(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CreateTmpFileSetResponse) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    public createTmpFileSet(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CreateTmpFileSetResponse) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.FileOperationRequestV2>;
    public renewTmpFileSet(request: client_pfs_pfs_pb.RenewTmpFileSetRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public renewTmpFileSet(request: client_pfs_pfs_pb.RenewTmpFileSetRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public renewTmpFileSet(request: client_pfs_pfs_pb.RenewTmpFileSetRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public clearCommitV2(request: client_pfs_pfs_pb.ClearCommitRequestV2, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public clearCommitV2(request: client_pfs_pfs_pb.ClearCommitRequestV2, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public clearCommitV2(request: client_pfs_pfs_pb.ClearCommitRequestV2, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
}

interface IObjectAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    putObject: IObjectAPIService_IPutObject;
    putObjectSplit: IObjectAPIService_IPutObjectSplit;
    putObjects: IObjectAPIService_IPutObjects;
    createObject: IObjectAPIService_ICreateObject;
    getObject: IObjectAPIService_IGetObject;
    getObjects: IObjectAPIService_IGetObjects;
    putBlock: IObjectAPIService_IPutBlock;
    getBlock: IObjectAPIService_IGetBlock;
    getBlocks: IObjectAPIService_IGetBlocks;
    listBlock: IObjectAPIService_IListBlock;
    tagObject: IObjectAPIService_ITagObject;
    inspectObject: IObjectAPIService_IInspectObject;
    checkObject: IObjectAPIService_ICheckObject;
    listObjects: IObjectAPIService_IListObjects;
    deleteObjects: IObjectAPIService_IDeleteObjects;
    getTag: IObjectAPIService_IGetTag;
    inspectTag: IObjectAPIService_IInspectTag;
    listTags: IObjectAPIService_IListTags;
    deleteTags: IObjectAPIService_IDeleteTags;
    compact: IObjectAPIService_ICompact;
    putObjDirect: IObjectAPIService_IPutObjDirect;
    getObjDirect: IObjectAPIService_IGetObjDirect;
    deleteObjDirect: IObjectAPIService_IDeleteObjDirect;
}

interface IObjectAPIService_IPutObject extends grpc.MethodDefinition<client_pfs_pfs_pb.PutObjectRequest, client_pfs_pfs_pb.Object> {
    path: "/pfs.ObjectAPI/PutObject";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.PutObjectRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.PutObjectRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.Object>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.Object>;
}
interface IObjectAPIService_IPutObjectSplit extends grpc.MethodDefinition<client_pfs_pfs_pb.PutObjectRequest, client_pfs_pfs_pb.Objects> {
    path: "/pfs.ObjectAPI/PutObjectSplit";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.PutObjectRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.PutObjectRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.Objects>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.Objects>;
}
interface IObjectAPIService_IPutObjects extends grpc.MethodDefinition<client_pfs_pfs_pb.PutObjectRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.ObjectAPI/PutObjects";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.PutObjectRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.PutObjectRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IObjectAPIService_ICreateObject extends grpc.MethodDefinition<client_pfs_pfs_pb.CreateObjectRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.ObjectAPI/CreateObject";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.CreateObjectRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.CreateObjectRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IObjectAPIService_IGetObject extends grpc.MethodDefinition<client_pfs_pfs_pb.Object, google_protobuf_wrappers_pb.BytesValue> {
    path: "/pfs.ObjectAPI/GetObject";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.Object>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.Object>;
    responseSerialize: grpc.serialize<google_protobuf_wrappers_pb.BytesValue>;
    responseDeserialize: grpc.deserialize<google_protobuf_wrappers_pb.BytesValue>;
}
interface IObjectAPIService_IGetObjects extends grpc.MethodDefinition<client_pfs_pfs_pb.GetObjectsRequest, google_protobuf_wrappers_pb.BytesValue> {
    path: "/pfs.ObjectAPI/GetObjects";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.GetObjectsRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.GetObjectsRequest>;
    responseSerialize: grpc.serialize<google_protobuf_wrappers_pb.BytesValue>;
    responseDeserialize: grpc.deserialize<google_protobuf_wrappers_pb.BytesValue>;
}
interface IObjectAPIService_IPutBlock extends grpc.MethodDefinition<client_pfs_pfs_pb.PutBlockRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.ObjectAPI/PutBlock";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.PutBlockRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.PutBlockRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IObjectAPIService_IGetBlock extends grpc.MethodDefinition<client_pfs_pfs_pb.GetBlockRequest, google_protobuf_wrappers_pb.BytesValue> {
    path: "/pfs.ObjectAPI/GetBlock";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.GetBlockRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.GetBlockRequest>;
    responseSerialize: grpc.serialize<google_protobuf_wrappers_pb.BytesValue>;
    responseDeserialize: grpc.deserialize<google_protobuf_wrappers_pb.BytesValue>;
}
interface IObjectAPIService_IGetBlocks extends grpc.MethodDefinition<client_pfs_pfs_pb.GetBlocksRequest, google_protobuf_wrappers_pb.BytesValue> {
    path: "/pfs.ObjectAPI/GetBlocks";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.GetBlocksRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.GetBlocksRequest>;
    responseSerialize: grpc.serialize<google_protobuf_wrappers_pb.BytesValue>;
    responseDeserialize: grpc.deserialize<google_protobuf_wrappers_pb.BytesValue>;
}
interface IObjectAPIService_IListBlock extends grpc.MethodDefinition<client_pfs_pfs_pb.ListBlockRequest, client_pfs_pfs_pb.Block> {
    path: "/pfs.ObjectAPI/ListBlock";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.ListBlockRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListBlockRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.Block>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.Block>;
}
interface IObjectAPIService_ITagObject extends grpc.MethodDefinition<client_pfs_pfs_pb.TagObjectRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.ObjectAPI/TagObject";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.TagObjectRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.TagObjectRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IObjectAPIService_IInspectObject extends grpc.MethodDefinition<client_pfs_pfs_pb.Object, client_pfs_pfs_pb.ObjectInfo> {
    path: "/pfs.ObjectAPI/InspectObject";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.Object>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.Object>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.ObjectInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.ObjectInfo>;
}
interface IObjectAPIService_ICheckObject extends grpc.MethodDefinition<client_pfs_pfs_pb.CheckObjectRequest, client_pfs_pfs_pb.CheckObjectResponse> {
    path: "/pfs.ObjectAPI/CheckObject";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.CheckObjectRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.CheckObjectRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.CheckObjectResponse>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.CheckObjectResponse>;
}
interface IObjectAPIService_IListObjects extends grpc.MethodDefinition<client_pfs_pfs_pb.ListObjectsRequest, client_pfs_pfs_pb.ObjectInfo> {
    path: "/pfs.ObjectAPI/ListObjects";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.ListObjectsRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListObjectsRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.ObjectInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.ObjectInfo>;
}
interface IObjectAPIService_IDeleteObjects extends grpc.MethodDefinition<client_pfs_pfs_pb.DeleteObjectsRequest, client_pfs_pfs_pb.DeleteObjectsResponse> {
    path: "/pfs.ObjectAPI/DeleteObjects";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.DeleteObjectsRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.DeleteObjectsRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.DeleteObjectsResponse>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.DeleteObjectsResponse>;
}
interface IObjectAPIService_IGetTag extends grpc.MethodDefinition<client_pfs_pfs_pb.Tag, google_protobuf_wrappers_pb.BytesValue> {
    path: "/pfs.ObjectAPI/GetTag";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.Tag>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.Tag>;
    responseSerialize: grpc.serialize<google_protobuf_wrappers_pb.BytesValue>;
    responseDeserialize: grpc.deserialize<google_protobuf_wrappers_pb.BytesValue>;
}
interface IObjectAPIService_IInspectTag extends grpc.MethodDefinition<client_pfs_pfs_pb.Tag, client_pfs_pfs_pb.ObjectInfo> {
    path: "/pfs.ObjectAPI/InspectTag";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.Tag>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.Tag>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.ObjectInfo>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.ObjectInfo>;
}
interface IObjectAPIService_IListTags extends grpc.MethodDefinition<client_pfs_pfs_pb.ListTagsRequest, client_pfs_pfs_pb.ListTagsResponse> {
    path: "/pfs.ObjectAPI/ListTags";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.ListTagsRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListTagsRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.ListTagsResponse>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.ListTagsResponse>;
}
interface IObjectAPIService_IDeleteTags extends grpc.MethodDefinition<client_pfs_pfs_pb.DeleteTagsRequest, client_pfs_pfs_pb.DeleteTagsResponse> {
    path: "/pfs.ObjectAPI/DeleteTags";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.DeleteTagsRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.DeleteTagsRequest>;
    responseSerialize: grpc.serialize<client_pfs_pfs_pb.DeleteTagsResponse>;
    responseDeserialize: grpc.deserialize<client_pfs_pfs_pb.DeleteTagsResponse>;
}
interface IObjectAPIService_ICompact extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty> {
    path: "/pfs.ObjectAPI/Compact";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IObjectAPIService_IPutObjDirect extends grpc.MethodDefinition<client_pfs_pfs_pb.PutObjDirectRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.ObjectAPI/PutObjDirect";
    requestStream: true;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.PutObjDirectRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.PutObjDirectRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IObjectAPIService_IGetObjDirect extends grpc.MethodDefinition<client_pfs_pfs_pb.GetObjDirectRequest, google_protobuf_wrappers_pb.BytesValue> {
    path: "/pfs.ObjectAPI/GetObjDirect";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.GetObjDirectRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.GetObjDirectRequest>;
    responseSerialize: grpc.serialize<google_protobuf_wrappers_pb.BytesValue>;
    responseDeserialize: grpc.deserialize<google_protobuf_wrappers_pb.BytesValue>;
}
interface IObjectAPIService_IDeleteObjDirect extends grpc.MethodDefinition<client_pfs_pfs_pb.DeleteObjDirectRequest, google_protobuf_empty_pb.Empty> {
    path: "/pfs.ObjectAPI/DeleteObjDirect";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_pfs_pfs_pb.DeleteObjDirectRequest>;
    requestDeserialize: grpc.deserialize<client_pfs_pfs_pb.DeleteObjDirectRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}

export const ObjectAPIService: IObjectAPIService;

export interface IObjectAPIServer {
    putObject: handleClientStreamingCall<client_pfs_pfs_pb.PutObjectRequest, client_pfs_pfs_pb.Object>;
    putObjectSplit: handleClientStreamingCall<client_pfs_pfs_pb.PutObjectRequest, client_pfs_pfs_pb.Objects>;
    putObjects: handleClientStreamingCall<client_pfs_pfs_pb.PutObjectRequest, google_protobuf_empty_pb.Empty>;
    createObject: grpc.handleUnaryCall<client_pfs_pfs_pb.CreateObjectRequest, google_protobuf_empty_pb.Empty>;
    getObject: grpc.handleServerStreamingCall<client_pfs_pfs_pb.Object, google_protobuf_wrappers_pb.BytesValue>;
    getObjects: grpc.handleServerStreamingCall<client_pfs_pfs_pb.GetObjectsRequest, google_protobuf_wrappers_pb.BytesValue>;
    putBlock: handleClientStreamingCall<client_pfs_pfs_pb.PutBlockRequest, google_protobuf_empty_pb.Empty>;
    getBlock: grpc.handleServerStreamingCall<client_pfs_pfs_pb.GetBlockRequest, google_protobuf_wrappers_pb.BytesValue>;
    getBlocks: grpc.handleServerStreamingCall<client_pfs_pfs_pb.GetBlocksRequest, google_protobuf_wrappers_pb.BytesValue>;
    listBlock: grpc.handleServerStreamingCall<client_pfs_pfs_pb.ListBlockRequest, client_pfs_pfs_pb.Block>;
    tagObject: grpc.handleUnaryCall<client_pfs_pfs_pb.TagObjectRequest, google_protobuf_empty_pb.Empty>;
    inspectObject: grpc.handleUnaryCall<client_pfs_pfs_pb.Object, client_pfs_pfs_pb.ObjectInfo>;
    checkObject: grpc.handleUnaryCall<client_pfs_pfs_pb.CheckObjectRequest, client_pfs_pfs_pb.CheckObjectResponse>;
    listObjects: grpc.handleServerStreamingCall<client_pfs_pfs_pb.ListObjectsRequest, client_pfs_pfs_pb.ObjectInfo>;
    deleteObjects: grpc.handleUnaryCall<client_pfs_pfs_pb.DeleteObjectsRequest, client_pfs_pfs_pb.DeleteObjectsResponse>;
    getTag: grpc.handleServerStreamingCall<client_pfs_pfs_pb.Tag, google_protobuf_wrappers_pb.BytesValue>;
    inspectTag: grpc.handleUnaryCall<client_pfs_pfs_pb.Tag, client_pfs_pfs_pb.ObjectInfo>;
    listTags: grpc.handleServerStreamingCall<client_pfs_pfs_pb.ListTagsRequest, client_pfs_pfs_pb.ListTagsResponse>;
    deleteTags: grpc.handleUnaryCall<client_pfs_pfs_pb.DeleteTagsRequest, client_pfs_pfs_pb.DeleteTagsResponse>;
    compact: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty>;
    putObjDirect: handleClientStreamingCall<client_pfs_pfs_pb.PutObjDirectRequest, google_protobuf_empty_pb.Empty>;
    getObjDirect: grpc.handleServerStreamingCall<client_pfs_pfs_pb.GetObjDirectRequest, google_protobuf_wrappers_pb.BytesValue>;
    deleteObjDirect: grpc.handleUnaryCall<client_pfs_pfs_pb.DeleteObjDirectRequest, google_protobuf_empty_pb.Empty>;
}

export interface IObjectAPIClient {
    putObject(callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Object) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObject(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Object) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObject(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Object) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObject(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Object) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObjectSplit(callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Objects) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObjectSplit(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Objects) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObjectSplit(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Objects) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObjectSplit(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Objects) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObjects(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObjects(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObjects(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    putObjects(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    createObject(request: client_pfs_pfs_pb.CreateObjectRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createObject(request: client_pfs_pfs_pb.CreateObjectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createObject(request: client_pfs_pfs_pb.CreateObjectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    getObject(request: client_pfs_pfs_pb.Object, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getObject(request: client_pfs_pfs_pb.Object, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getObjects(request: client_pfs_pfs_pb.GetObjectsRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getObjects(request: client_pfs_pfs_pb.GetObjectsRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    putBlock(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutBlockRequest>;
    putBlock(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutBlockRequest>;
    putBlock(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutBlockRequest>;
    putBlock(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutBlockRequest>;
    getBlock(request: client_pfs_pfs_pb.GetBlockRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getBlock(request: client_pfs_pfs_pb.GetBlockRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getBlocks(request: client_pfs_pfs_pb.GetBlocksRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getBlocks(request: client_pfs_pfs_pb.GetBlocksRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    listBlock(request: client_pfs_pfs_pb.ListBlockRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.Block>;
    listBlock(request: client_pfs_pfs_pb.ListBlockRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.Block>;
    tagObject(request: client_pfs_pfs_pb.TagObjectRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    tagObject(request: client_pfs_pfs_pb.TagObjectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    tagObject(request: client_pfs_pfs_pb.TagObjectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    inspectObject(request: client_pfs_pfs_pb.Object, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    inspectObject(request: client_pfs_pfs_pb.Object, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    inspectObject(request: client_pfs_pfs_pb.Object, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    checkObject(request: client_pfs_pfs_pb.CheckObjectRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CheckObjectResponse) => void): grpc.ClientUnaryCall;
    checkObject(request: client_pfs_pfs_pb.CheckObjectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CheckObjectResponse) => void): grpc.ClientUnaryCall;
    checkObject(request: client_pfs_pfs_pb.CheckObjectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CheckObjectResponse) => void): grpc.ClientUnaryCall;
    listObjects(request: client_pfs_pfs_pb.ListObjectsRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.ObjectInfo>;
    listObjects(request: client_pfs_pfs_pb.ListObjectsRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.ObjectInfo>;
    deleteObjects(request: client_pfs_pfs_pb.DeleteObjectsRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteObjectsResponse) => void): grpc.ClientUnaryCall;
    deleteObjects(request: client_pfs_pfs_pb.DeleteObjectsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteObjectsResponse) => void): grpc.ClientUnaryCall;
    deleteObjects(request: client_pfs_pfs_pb.DeleteObjectsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteObjectsResponse) => void): grpc.ClientUnaryCall;
    getTag(request: client_pfs_pfs_pb.Tag, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getTag(request: client_pfs_pfs_pb.Tag, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    inspectTag(request: client_pfs_pfs_pb.Tag, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    inspectTag(request: client_pfs_pfs_pb.Tag, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    inspectTag(request: client_pfs_pfs_pb.Tag, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    listTags(request: client_pfs_pfs_pb.ListTagsRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.ListTagsResponse>;
    listTags(request: client_pfs_pfs_pb.ListTagsRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.ListTagsResponse>;
    deleteTags(request: client_pfs_pfs_pb.DeleteTagsRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteTagsResponse) => void): grpc.ClientUnaryCall;
    deleteTags(request: client_pfs_pfs_pb.DeleteTagsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteTagsResponse) => void): grpc.ClientUnaryCall;
    deleteTags(request: client_pfs_pfs_pb.DeleteTagsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteTagsResponse) => void): grpc.ClientUnaryCall;
    compact(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    compact(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    compact(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    putObjDirect(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjDirectRequest>;
    putObjDirect(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjDirectRequest>;
    putObjDirect(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjDirectRequest>;
    putObjDirect(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjDirectRequest>;
    getObjDirect(request: client_pfs_pfs_pb.GetObjDirectRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    getObjDirect(request: client_pfs_pfs_pb.GetObjDirectRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    deleteObjDirect(request: client_pfs_pfs_pb.DeleteObjDirectRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteObjDirect(request: client_pfs_pfs_pb.DeleteObjDirectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteObjDirect(request: client_pfs_pfs_pb.DeleteObjDirectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
}

export class ObjectAPIClient extends grpc.Client implements IObjectAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public putObject(callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Object) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObject(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Object) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObject(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Object) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObject(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Object) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObjectSplit(callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Objects) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObjectSplit(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Objects) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObjectSplit(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Objects) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObjectSplit(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.Objects) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObjects(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObjects(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObjects(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public putObjects(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjectRequest>;
    public createObject(request: client_pfs_pfs_pb.CreateObjectRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createObject(request: client_pfs_pfs_pb.CreateObjectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createObject(request: client_pfs_pfs_pb.CreateObjectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public getObject(request: client_pfs_pfs_pb.Object, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getObject(request: client_pfs_pfs_pb.Object, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getObjects(request: client_pfs_pfs_pb.GetObjectsRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getObjects(request: client_pfs_pfs_pb.GetObjectsRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public putBlock(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutBlockRequest>;
    public putBlock(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutBlockRequest>;
    public putBlock(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutBlockRequest>;
    public putBlock(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutBlockRequest>;
    public getBlock(request: client_pfs_pfs_pb.GetBlockRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getBlock(request: client_pfs_pfs_pb.GetBlockRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getBlocks(request: client_pfs_pfs_pb.GetBlocksRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getBlocks(request: client_pfs_pfs_pb.GetBlocksRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public listBlock(request: client_pfs_pfs_pb.ListBlockRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.Block>;
    public listBlock(request: client_pfs_pfs_pb.ListBlockRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.Block>;
    public tagObject(request: client_pfs_pfs_pb.TagObjectRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public tagObject(request: client_pfs_pfs_pb.TagObjectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public tagObject(request: client_pfs_pfs_pb.TagObjectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public inspectObject(request: client_pfs_pfs_pb.Object, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    public inspectObject(request: client_pfs_pfs_pb.Object, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    public inspectObject(request: client_pfs_pfs_pb.Object, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    public checkObject(request: client_pfs_pfs_pb.CheckObjectRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CheckObjectResponse) => void): grpc.ClientUnaryCall;
    public checkObject(request: client_pfs_pfs_pb.CheckObjectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CheckObjectResponse) => void): grpc.ClientUnaryCall;
    public checkObject(request: client_pfs_pfs_pb.CheckObjectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.CheckObjectResponse) => void): grpc.ClientUnaryCall;
    public listObjects(request: client_pfs_pfs_pb.ListObjectsRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.ObjectInfo>;
    public listObjects(request: client_pfs_pfs_pb.ListObjectsRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.ObjectInfo>;
    public deleteObjects(request: client_pfs_pfs_pb.DeleteObjectsRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteObjectsResponse) => void): grpc.ClientUnaryCall;
    public deleteObjects(request: client_pfs_pfs_pb.DeleteObjectsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteObjectsResponse) => void): grpc.ClientUnaryCall;
    public deleteObjects(request: client_pfs_pfs_pb.DeleteObjectsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteObjectsResponse) => void): grpc.ClientUnaryCall;
    public getTag(request: client_pfs_pfs_pb.Tag, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getTag(request: client_pfs_pfs_pb.Tag, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public inspectTag(request: client_pfs_pfs_pb.Tag, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    public inspectTag(request: client_pfs_pfs_pb.Tag, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    public inspectTag(request: client_pfs_pfs_pb.Tag, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.ObjectInfo) => void): grpc.ClientUnaryCall;
    public listTags(request: client_pfs_pfs_pb.ListTagsRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.ListTagsResponse>;
    public listTags(request: client_pfs_pfs_pb.ListTagsRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<client_pfs_pfs_pb.ListTagsResponse>;
    public deleteTags(request: client_pfs_pfs_pb.DeleteTagsRequest, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteTagsResponse) => void): grpc.ClientUnaryCall;
    public deleteTags(request: client_pfs_pfs_pb.DeleteTagsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteTagsResponse) => void): grpc.ClientUnaryCall;
    public deleteTags(request: client_pfs_pfs_pb.DeleteTagsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_pfs_pfs_pb.DeleteTagsResponse) => void): grpc.ClientUnaryCall;
    public compact(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public compact(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public compact(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public putObjDirect(callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjDirectRequest>;
    public putObjDirect(metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjDirectRequest>;
    public putObjDirect(options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjDirectRequest>;
    public putObjDirect(metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientWritableStream<client_pfs_pfs_pb.PutObjDirectRequest>;
    public getObjDirect(request: client_pfs_pfs_pb.GetObjDirectRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public getObjDirect(request: client_pfs_pfs_pb.GetObjDirectRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<google_protobuf_wrappers_pb.BytesValue>;
    public deleteObjDirect(request: client_pfs_pfs_pb.DeleteObjDirectRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteObjDirect(request: client_pfs_pfs_pb.DeleteObjDirectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteObjDirect(request: client_pfs_pfs_pb.DeleteObjDirectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
}
