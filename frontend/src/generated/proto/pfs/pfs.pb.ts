/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as Auth_v2Auth from "../auth/auth.pb";
import * as fm from "../fetch.pb";
import * as GoogleProtobufAny from "../google/protobuf/any.pb";
import * as GoogleProtobufDuration from "../google/protobuf/duration.pb";
import * as GoogleProtobufEmpty from "../google/protobuf/empty.pb";
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb";
import * as GoogleProtobufWrappers from "../google/protobuf/wrappers.pb";
import * as TaskapiTask from "../task/task.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
    | { [k in keyof T]?: undefined }
    | (
        keyof T extends infer K ?
        (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
            : never)
        : never);

export enum OriginKind {
    ORIGIN_KIND_UNKNOWN = "ORIGIN_KIND_UNKNOWN",
    USER = "USER",
    AUTO = "AUTO",
    FSCK = "FSCK",
}

export enum FileType {
    RESERVED = "RESERVED",
    FILE = "FILE",
    DIR = "DIR",
}

export enum CommitState {
    COMMIT_STATE_UNKNOWN = "COMMIT_STATE_UNKNOWN",
    STARTED = "STARTED",
    READY = "READY",
    FINISHING = "FINISHING",
    FINISHED = "FINISHED",
}

export enum Delimiter {
    NONE = "NONE",
    JSON = "JSON",
    LINE = "LINE",
    SQL = "SQL",
    CSV = "CSV",
}

export enum RepoPageOrdering {
    PROJECT_REPO = "PROJECT_REPO",
}

export enum GetFileSetRequestFileSetType {
    TOTAL = "TOTAL",
    DIFF = "DIFF",
}

export enum SQLDatabaseEgressFileFormatType {
    UNKNOWN = "UNKNOWN",
    CSV = "CSV",
    JSON = "JSON",
    PARQUET = "PARQUET",
}

export type Repo = {
    __typename?: "Repo";
    name?: string;
    type?: string;
    project?: Project;
};

export type RepoPickerRepoName = {
    __typename?: "RepoPickerRepoName";
    project?: ProjectPicker;
    name?: string;
    type?: string;
};


type BaseRepoPicker = {
    __typename?: "BaseRepoPicker";
};

export type RepoPicker = BaseRepoPicker
    & OneOf<{ name: RepoPickerRepoName; }>;

export type Branch = {
    __typename?: "Branch";
    repo?: Repo;
    name?: string;
};

export type BranchPickerBranchName = {
    __typename?: "BranchPickerBranchName";
    repo?: RepoPicker;
    name?: string;
};


type BaseBranchPicker = {
    __typename?: "BaseBranchPicker";
};

export type BranchPicker = BaseBranchPicker
    & OneOf<{ name: BranchPickerBranchName; }>;

export type File = {
    __typename?: "File";
    commit?: Commit;
    path?: string;
    datum?: string;
};

export type RepoInfoDetails = {
    __typename?: "RepoInfoDetails";
    sizeBytes?: string;
};

export type RepoInfo = {
    __typename?: "RepoInfo";
    repo?: Repo;
    created?: GoogleProtobufTimestamp.Timestamp;
    sizeBytesUpperBound?: string;
    description?: string;
    branches?: Branch[];
    authInfo?: AuthInfo;
    details?: RepoInfoDetails;
};

export type AuthInfo = {
    __typename?: "AuthInfo";
    permissions?: Auth_v2Auth.Permission[];
    roles?: string[];
};

export type BranchInfo = {
    __typename?: "BranchInfo";
    branch?: Branch;
    head?: Commit;
    provenance?: Branch[];
    subvenance?: Branch[];
    directProvenance?: Branch[];
    trigger?: Trigger;
};

export type Trigger = {
    __typename?: "Trigger";
    branch?: string;
    all?: boolean;
    rateLimitSpec?: string;
    size?: string;
    commits?: string;
    cronSpec?: string;
};

export type CommitOrigin = {
    __typename?: "CommitOrigin";
    kind?: OriginKind;
};

export type Commit = {
    __typename?: "Commit";
    repo?: Repo;
    id?: string;
    branch?: Branch;
};

export type CommitPickerCommitByGlobalId = {
    __typename?: "CommitPickerCommitByGlobalId";
    repo?: RepoPicker;
    id?: string;
};

export type CommitPickerBranchRoot = {
    __typename?: "CommitPickerBranchRoot";
    offset?: number;
    branch?: BranchPicker;
};

export type CommitPickerAncestorOf = {
    __typename?: "CommitPickerAncestorOf";
    offset?: number;
    start?: CommitPicker;
};


type BaseCommitPicker = {
    __typename?: "BaseCommitPicker";
};

export type CommitPicker = BaseCommitPicker
    & OneOf<{ branchHead: BranchPicker; id: CommitPickerCommitByGlobalId; ancestor: CommitPickerAncestorOf; branchRoot: CommitPickerBranchRoot; }>;

export type CommitInfoDetails = {
    __typename?: "CommitInfoDetails";
    sizeBytes?: string;
    compactingTime?: GoogleProtobufDuration.Duration;
    validatingTime?: GoogleProtobufDuration.Duration;
};

export type CommitInfo = {
    __typename?: "CommitInfo";
    commit?: Commit;
    origin?: CommitOrigin;
    description?: string;
    parentCommit?: Commit;
    childCommits?: Commit[];
    started?: GoogleProtobufTimestamp.Timestamp;
    finishing?: GoogleProtobufTimestamp.Timestamp;
    finished?: GoogleProtobufTimestamp.Timestamp;
    directProvenance?: Commit[];
    error?: string;
    sizeBytesUpperBound?: string;
    details?: CommitInfoDetails;
};

export type CommitSet = {
    __typename?: "CommitSet";
    id?: string;
};

export type CommitSetInfo = {
    __typename?: "CommitSetInfo";
    commitSet?: CommitSet;
    commits?: CommitInfo[];
};

export type FileInfo = {
    __typename?: "FileInfo";
    file?: File;
    fileType?: FileType;
    committed?: GoogleProtobufTimestamp.Timestamp;
    sizeBytes?: string;
    hash?: Uint8Array;
};

export type Project = {
    __typename?: "Project";
    name?: string;
};

export type ProjectInfo = {
    __typename?: "ProjectInfo";
    project?: Project;
    description?: string;
    authInfo?: AuthInfo;
    createdAt?: GoogleProtobufTimestamp.Timestamp;
    metadata?: { [key: string]: string; };
};


type BaseProjectPicker = {
    __typename?: "BaseProjectPicker";
};

export type ProjectPicker = BaseProjectPicker
    & OneOf<{ name: string; }>;

export type CreateRepoRequest = {
    __typename?: "CreateRepoRequest";
    repo?: Repo;
    description?: string;
    update?: boolean;
};

export type InspectRepoRequest = {
    __typename?: "InspectRepoRequest";
    repo?: Repo;
};

export type ListRepoRequest = {
    __typename?: "ListRepoRequest";
    type?: string;
    projects?: Project[];
    page?: RepoPage;
};

export type RepoPage = {
    __typename?: "RepoPage";
    order?: RepoPageOrdering;
    pageSize?: string;
    pageIndex?: string;
};

export type DeleteRepoRequest = {
    __typename?: "DeleteRepoRequest";
    repo?: Repo;
    force?: boolean;
};

export type DeleteReposRequest = {
    __typename?: "DeleteReposRequest";
    projects?: Project[];
    force?: boolean;
    all?: boolean;
};

export type DeleteRepoResponse = {
    __typename?: "DeleteRepoResponse";
    deleted?: boolean;
};

export type DeleteReposResponse = {
    __typename?: "DeleteReposResponse";
    repos?: Repo[];
};

export type StartCommitRequest = {
    __typename?: "StartCommitRequest";
    parent?: Commit;
    description?: string;
    branch?: Branch;
};

export type FinishCommitRequest = {
    __typename?: "FinishCommitRequest";
    commit?: Commit;
    description?: string;
    error?: string;
    force?: boolean;
};

export type InspectCommitRequest = {
    __typename?: "InspectCommitRequest";
    commit?: Commit;
    wait?: CommitState;
};

export type ListCommitRequest = {
    __typename?: "ListCommitRequest";
    repo?: Repo;
    from?: Commit;
    to?: Commit;
    number?: string;
    reverse?: boolean;
    all?: boolean;
    originKind?: OriginKind;
    startedTime?: GoogleProtobufTimestamp.Timestamp;
};

export type InspectCommitSetRequest = {
    __typename?: "InspectCommitSetRequest";
    commitSet?: CommitSet;
    wait?: boolean;
};

export type ListCommitSetRequest = {
    __typename?: "ListCommitSetRequest";
    project?: Project;
};

export type SquashCommitSetRequest = {
    __typename?: "SquashCommitSetRequest";
    commitSet?: CommitSet;
};

export type DropCommitSetRequest = {
    __typename?: "DropCommitSetRequest";
    commitSet?: CommitSet;
};

export type SubscribeCommitRequest = {
    __typename?: "SubscribeCommitRequest";
    repo?: Repo;
    branch?: string;
    from?: Commit;
    state?: CommitState;
    all?: boolean;
    originKind?: OriginKind;
};

export type ClearCommitRequest = {
    __typename?: "ClearCommitRequest";
    commit?: Commit;
};

export type SquashCommitRequest = {
    __typename?: "SquashCommitRequest";
    commit?: Commit;
    recursive?: boolean;
};

export type SquashCommitResponse = {
    __typename?: "SquashCommitResponse";
};

export type DropCommitRequest = {
    __typename?: "DropCommitRequest";
    commit?: Commit;
    recursive?: boolean;
};

export type WalkCommitProvenanceRequest = {
    __typename?: "WalkCommitProvenanceRequest";
    start?: CommitPicker[];
    maxCommits?: string;
    maxDepth?: string;
};

export type WalkCommitSubvenanceRequest = {
    __typename?: "WalkCommitSubvenanceRequest";
    start?: CommitPicker[];
    maxCommits?: string;
    maxDepth?: string;
};

export type WalkBranchProvenanceRequest = {
    __typename?: "WalkBranchProvenanceRequest";
    start?: BranchPicker[];
    maxBranches?: string;
    maxDepth?: string;
};

export type WalkBranchSubvenanceRequest = {
    __typename?: "WalkBranchSubvenanceRequest";
    start?: BranchPicker[];
    maxBranches?: string;
    maxDepth?: string;
};

export type DropCommitResponse = {
    __typename?: "DropCommitResponse";
};

export type CreateBranchRequest = {
    __typename?: "CreateBranchRequest";
    head?: Commit;
    branch?: Branch;
    provenance?: Branch[];
    trigger?: Trigger;
    newCommitSet?: boolean;
};

export type FindCommitsRequest = {
    __typename?: "FindCommitsRequest";
    start?: Commit;
    filePath?: string;
    limit?: number;
};


type BaseFindCommitsResponse = {
    __typename?: "BaseFindCommitsResponse";
    commitsSearched?: number;
};

export type FindCommitsResponse = BaseFindCommitsResponse
    & OneOf<{ foundCommit: Commit; lastSearchedCommit: Commit; }>;

export type InspectBranchRequest = {
    __typename?: "InspectBranchRequest";
    branch?: Branch;
};

export type ListBranchRequest = {
    __typename?: "ListBranchRequest";
    repo?: Repo;
    reverse?: boolean;
};

export type DeleteBranchRequest = {
    __typename?: "DeleteBranchRequest";
    branch?: Branch;
    force?: boolean;
};

export type CreateProjectRequest = {
    __typename?: "CreateProjectRequest";
    project?: Project;
    description?: string;
    update?: boolean;
};

export type InspectProjectRequest = {
    __typename?: "InspectProjectRequest";
    project?: Project;
};

export type InspectProjectV2Request = {
    __typename?: "InspectProjectV2Request";
    project?: Project;
};

export type InspectProjectV2Response = {
    __typename?: "InspectProjectV2Response";
    info?: ProjectInfo;
    defaultsJson?: string;
};

export type ListProjectRequest = {
    __typename?: "ListProjectRequest";
};

export type DeleteProjectRequest = {
    __typename?: "DeleteProjectRequest";
    project?: Project;
    force?: boolean;
};

export type AddFileURLSource = {
    __typename?: "AddFileURLSource";
    uRL?: string;
    recursive?: boolean;
    concurrency?: number;
};


type BaseAddFile = {
    __typename?: "BaseAddFile";
    path?: string;
    datum?: string;
};

export type AddFile = BaseAddFile
    & OneOf<{ raw: GoogleProtobufWrappers.BytesValue; url: AddFileURLSource; }>;

export type DeleteFile = {
    __typename?: "DeleteFile";
    path?: string;
    datum?: string;
};

export type CopyFile = {
    __typename?: "CopyFile";
    dst?: string;
    datum?: string;
    src?: File;
    append?: boolean;
};


type BaseModifyFileRequest = {
    __typename?: "BaseModifyFileRequest";
};

export type ModifyFileRequest = BaseModifyFileRequest
    & OneOf<{ setCommit: Commit; addFile: AddFile; deleteFile: DeleteFile; copyFile: CopyFile; }>;

export type GetFileRequest = {
    __typename?: "GetFileRequest";
    file?: File;
    uRL?: string;
    offset?: string;
    pathRange?: PathRange;
};

export type InspectFileRequest = {
    __typename?: "InspectFileRequest";
    file?: File;
};

export type ListFileRequest = {
    __typename?: "ListFileRequest";
    file?: File;
    paginationMarker?: File;
    number?: string;
    reverse?: boolean;
};

export type WalkFileRequest = {
    __typename?: "WalkFileRequest";
    file?: File;
    paginationMarker?: File;
    number?: string;
    reverse?: boolean;
};

export type GlobFileRequest = {
    __typename?: "GlobFileRequest";
    commit?: Commit;
    pattern?: string;
    pathRange?: PathRange;
};

export type DiffFileRequest = {
    __typename?: "DiffFileRequest";
    newFile?: File;
    oldFile?: File;
    shallow?: boolean;
};

export type DiffFileResponse = {
    __typename?: "DiffFileResponse";
    newFile?: FileInfo;
    oldFile?: FileInfo;
};


type BaseFsckRequest = {
    __typename?: "BaseFsckRequest";
    fix?: boolean;
};

export type FsckRequest = BaseFsckRequest
    & OneOf<{ zombieTarget: Commit; zombieAll: boolean; }>;

export type FsckResponse = {
    __typename?: "FsckResponse";
    fix?: string;
    error?: string;
};

export type CreateFileSetResponse = {
    __typename?: "CreateFileSetResponse";
    fileSetId?: string;
};

export type GetFileSetRequest = {
    __typename?: "GetFileSetRequest";
    commit?: Commit;
    type?: GetFileSetRequestFileSetType;
};

export type AddFileSetRequest = {
    __typename?: "AddFileSetRequest";
    commit?: Commit;
    fileSetId?: string;
};

export type RenewFileSetRequest = {
    __typename?: "RenewFileSetRequest";
    fileSetId?: string;
    ttlSeconds?: string;
};

export type ComposeFileSetRequest = {
    __typename?: "ComposeFileSetRequest";
    fileSetIds?: string[];
    ttlSeconds?: string;
    compact?: boolean;
};

export type ShardFileSetRequest = {
    __typename?: "ShardFileSetRequest";
    fileSetId?: string;
    numFiles?: string;
    sizeBytes?: string;
};

export type PathRange = {
    __typename?: "PathRange";
    lower?: string;
    upper?: string;
};

export type ShardFileSetResponse = {
    __typename?: "ShardFileSetResponse";
    shards?: PathRange[];
};

export type CheckStorageRequest = {
    __typename?: "CheckStorageRequest";
    readChunkData?: boolean;
    chunkBegin?: Uint8Array;
    chunkEnd?: Uint8Array;
};

export type CheckStorageResponse = {
    __typename?: "CheckStorageResponse";
    chunkObjectCount?: string;
};

export type PutCacheRequest = {
    __typename?: "PutCacheRequest";
    key?: string;
    value?: GoogleProtobufAny.Any;
    fileSetIds?: string[];
    tag?: string;
};

export type GetCacheRequest = {
    __typename?: "GetCacheRequest";
    key?: string;
};

export type GetCacheResponse = {
    __typename?: "GetCacheResponse";
    value?: GoogleProtobufAny.Any;
};

export type ClearCacheRequest = {
    __typename?: "ClearCacheRequest";
    tagPrefix?: string;
};

export type ActivateAuthRequest = {
    __typename?: "ActivateAuthRequest";
};

export type ActivateAuthResponse = {
    __typename?: "ActivateAuthResponse";
};

export type ObjectStorageEgress = {
    __typename?: "ObjectStorageEgress";
    url?: string;
};

export type SQLDatabaseEgressFileFormat = {
    __typename?: "SQLDatabaseEgressFileFormat";
    type?: SQLDatabaseEgressFileFormatType;
    columns?: string[];
};

export type SQLDatabaseEgressSecret = {
    __typename?: "SQLDatabaseEgressSecret";
    name?: string;
    key?: string;
};

export type SQLDatabaseEgress = {
    __typename?: "SQLDatabaseEgress";
    url?: string;
    fileFormat?: SQLDatabaseEgressFileFormat;
    secret?: SQLDatabaseEgressSecret;
};


type BaseEgressRequest = {
    __typename?: "BaseEgressRequest";
    commit?: Commit;
};

export type EgressRequest = BaseEgressRequest
    & OneOf<{ objectStorage: ObjectStorageEgress; sqlDatabase: SQLDatabaseEgress; }>;

export type EgressResponseObjectStorageResult = {
    __typename?: "EgressResponseObjectStorageResult";
    bytesWritten?: string;
};

export type EgressResponseSQLDatabaseResult = {
    __typename?: "EgressResponseSQLDatabaseResult";
    rowsWritten?: { [key: string]: string; };
};


type BaseEgressResponse = {
    __typename?: "BaseEgressResponse";
};

export type EgressResponse = BaseEgressResponse
    & OneOf<{ objectStorage: EgressResponseObjectStorageResult; sqlDatabase: EgressResponseSQLDatabaseResult; }>;

export class API {
    static CreateRepo(req: CreateRepoRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<CreateRepoRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/CreateRepo`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectRepo(req: InspectRepoRequest, initReq?: fm.InitReq): Promise<RepoInfo> {
        return fm.fetchReq<InspectRepoRequest, RepoInfo>(`/pfs_v2.API/InspectRepo`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListRepo(req: ListRepoRequest, entityNotifier?: fm.NotifyStreamEntityArrival<RepoInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ListRepoRequest, RepoInfo>(`/pfs_v2.API/ListRepo`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteRepo(req: DeleteRepoRequest, initReq?: fm.InitReq): Promise<DeleteRepoResponse> {
        return fm.fetchReq<DeleteRepoRequest, DeleteRepoResponse>(`/pfs_v2.API/DeleteRepo`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteRepos(req: DeleteReposRequest, initReq?: fm.InitReq): Promise<DeleteReposResponse> {
        return fm.fetchReq<DeleteReposRequest, DeleteReposResponse>(`/pfs_v2.API/DeleteRepos`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static StartCommit(req: StartCommitRequest, initReq?: fm.InitReq): Promise<Commit> {
        return fm.fetchReq<StartCommitRequest, Commit>(`/pfs_v2.API/StartCommit`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static FinishCommit(req: FinishCommitRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<FinishCommitRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/FinishCommit`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ClearCommit(req: ClearCommitRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<ClearCommitRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/ClearCommit`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectCommit(req: InspectCommitRequest, initReq?: fm.InitReq): Promise<CommitInfo> {
        return fm.fetchReq<InspectCommitRequest, CommitInfo>(`/pfs_v2.API/InspectCommit`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListCommit(req: ListCommitRequest, entityNotifier?: fm.NotifyStreamEntityArrival<CommitInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ListCommitRequest, CommitInfo>(`/pfs_v2.API/ListCommit`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static SubscribeCommit(req: SubscribeCommitRequest, entityNotifier?: fm.NotifyStreamEntityArrival<CommitInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<SubscribeCommitRequest, CommitInfo>(`/pfs_v2.API/SubscribeCommit`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static SquashCommit(req: SquashCommitRequest, initReq?: fm.InitReq): Promise<SquashCommitResponse> {
        return fm.fetchReq<SquashCommitRequest, SquashCommitResponse>(`/pfs_v2.API/SquashCommit`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DropCommit(req: DropCommitRequest, initReq?: fm.InitReq): Promise<DropCommitResponse> {
        return fm.fetchReq<DropCommitRequest, DropCommitResponse>(`/pfs_v2.API/DropCommit`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectCommitSet(req: InspectCommitSetRequest, entityNotifier?: fm.NotifyStreamEntityArrival<CommitInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<InspectCommitSetRequest, CommitInfo>(`/pfs_v2.API/InspectCommitSet`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListCommitSet(req: ListCommitSetRequest, entityNotifier?: fm.NotifyStreamEntityArrival<CommitSetInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ListCommitSetRequest, CommitSetInfo>(`/pfs_v2.API/ListCommitSet`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static SquashCommitSet(req: SquashCommitSetRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<SquashCommitSetRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/SquashCommitSet`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DropCommitSet(req: DropCommitSetRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<DropCommitSetRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/DropCommitSet`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static FindCommits(req: FindCommitsRequest, entityNotifier?: fm.NotifyStreamEntityArrival<FindCommitsResponse>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<FindCommitsRequest, FindCommitsResponse>(`/pfs_v2.API/FindCommits`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static WalkCommitProvenance(req: WalkCommitProvenanceRequest, entityNotifier?: fm.NotifyStreamEntityArrival<CommitInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<WalkCommitProvenanceRequest, CommitInfo>(`/pfs_v2.API/WalkCommitProvenance`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static WalkCommitSubvenance(req: WalkCommitSubvenanceRequest, entityNotifier?: fm.NotifyStreamEntityArrival<CommitInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<WalkCommitSubvenanceRequest, CommitInfo>(`/pfs_v2.API/WalkCommitSubvenance`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static CreateBranch(req: CreateBranchRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<CreateBranchRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/CreateBranch`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectBranch(req: InspectBranchRequest, initReq?: fm.InitReq): Promise<BranchInfo> {
        return fm.fetchReq<InspectBranchRequest, BranchInfo>(`/pfs_v2.API/InspectBranch`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListBranch(req: ListBranchRequest, entityNotifier?: fm.NotifyStreamEntityArrival<BranchInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ListBranchRequest, BranchInfo>(`/pfs_v2.API/ListBranch`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteBranch(req: DeleteBranchRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<DeleteBranchRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/DeleteBranch`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static WalkBranchProvenance(req: WalkBranchProvenanceRequest, entityNotifier?: fm.NotifyStreamEntityArrival<BranchInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<WalkBranchProvenanceRequest, BranchInfo>(`/pfs_v2.API/WalkBranchProvenance`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static WalkBranchSubvenance(req: WalkBranchSubvenanceRequest, entityNotifier?: fm.NotifyStreamEntityArrival<BranchInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<WalkBranchSubvenanceRequest, BranchInfo>(`/pfs_v2.API/WalkBranchSubvenance`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetFile(req: GetFileRequest, entityNotifier?: fm.NotifyStreamEntityArrival<GoogleProtobufWrappers.BytesValue>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<GetFileRequest, GoogleProtobufWrappers.BytesValue>(`/pfs_v2.API/GetFile`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetFileTAR(req: GetFileRequest, entityNotifier?: fm.NotifyStreamEntityArrival<GoogleProtobufWrappers.BytesValue>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<GetFileRequest, GoogleProtobufWrappers.BytesValue>(`/pfs_v2.API/GetFileTAR`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectFile(req: InspectFileRequest, initReq?: fm.InitReq): Promise<FileInfo> {
        return fm.fetchReq<InspectFileRequest, FileInfo>(`/pfs_v2.API/InspectFile`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListFile(req: ListFileRequest, entityNotifier?: fm.NotifyStreamEntityArrival<FileInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ListFileRequest, FileInfo>(`/pfs_v2.API/ListFile`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static WalkFile(req: WalkFileRequest, entityNotifier?: fm.NotifyStreamEntityArrival<FileInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<WalkFileRequest, FileInfo>(`/pfs_v2.API/WalkFile`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GlobFile(req: GlobFileRequest, entityNotifier?: fm.NotifyStreamEntityArrival<FileInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<GlobFileRequest, FileInfo>(`/pfs_v2.API/GlobFile`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DiffFile(req: DiffFileRequest, entityNotifier?: fm.NotifyStreamEntityArrival<DiffFileResponse>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<DiffFileRequest, DiffFileResponse>(`/pfs_v2.API/DiffFile`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ActivateAuth(req: ActivateAuthRequest, initReq?: fm.InitReq): Promise<ActivateAuthResponse> {
        return fm.fetchReq<ActivateAuthRequest, ActivateAuthResponse>(`/pfs_v2.API/ActivateAuth`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteAll(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<GoogleProtobufEmpty.Empty, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/DeleteAll`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Fsck(req: FsckRequest, entityNotifier?: fm.NotifyStreamEntityArrival<FsckResponse>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<FsckRequest, FsckResponse>(`/pfs_v2.API/Fsck`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetFileSet(req: GetFileSetRequest, initReq?: fm.InitReq): Promise<CreateFileSetResponse> {
        return fm.fetchReq<GetFileSetRequest, CreateFileSetResponse>(`/pfs_v2.API/GetFileSet`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static AddFileSet(req: AddFileSetRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<AddFileSetRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/AddFileSet`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RenewFileSet(req: RenewFileSetRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<RenewFileSetRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/RenewFileSet`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ComposeFileSet(req: ComposeFileSetRequest, initReq?: fm.InitReq): Promise<CreateFileSetResponse> {
        return fm.fetchReq<ComposeFileSetRequest, CreateFileSetResponse>(`/pfs_v2.API/ComposeFileSet`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ShardFileSet(req: ShardFileSetRequest, initReq?: fm.InitReq): Promise<ShardFileSetResponse> {
        return fm.fetchReq<ShardFileSetRequest, ShardFileSetResponse>(`/pfs_v2.API/ShardFileSet`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static CheckStorage(req: CheckStorageRequest, initReq?: fm.InitReq): Promise<CheckStorageResponse> {
        return fm.fetchReq<CheckStorageRequest, CheckStorageResponse>(`/pfs_v2.API/CheckStorage`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static PutCache(req: PutCacheRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<PutCacheRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/PutCache`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetCache(req: GetCacheRequest, initReq?: fm.InitReq): Promise<GetCacheResponse> {
        return fm.fetchReq<GetCacheRequest, GetCacheResponse>(`/pfs_v2.API/GetCache`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ClearCache(req: ClearCacheRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<ClearCacheRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/ClearCache`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListTask(req: TaskapiTask.ListTaskRequest, entityNotifier?: fm.NotifyStreamEntityArrival<TaskapiTask.TaskInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<TaskapiTask.ListTaskRequest, TaskapiTask.TaskInfo>(`/pfs_v2.API/ListTask`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Egress(req: EgressRequest, initReq?: fm.InitReq): Promise<EgressResponse> {
        return fm.fetchReq<EgressRequest, EgressResponse>(`/pfs_v2.API/Egress`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static CreateProject(req: CreateProjectRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<CreateProjectRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/CreateProject`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectProject(req: InspectProjectRequest, initReq?: fm.InitReq): Promise<ProjectInfo> {
        return fm.fetchReq<InspectProjectRequest, ProjectInfo>(`/pfs_v2.API/InspectProject`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectProjectV2(req: InspectProjectV2Request, initReq?: fm.InitReq): Promise<InspectProjectV2Response> {
        return fm.fetchReq<InspectProjectV2Request, InspectProjectV2Response>(`/pfs_v2.API/InspectProjectV2`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListProject(req: ListProjectRequest, entityNotifier?: fm.NotifyStreamEntityArrival<ProjectInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ListProjectRequest, ProjectInfo>(`/pfs_v2.API/ListProject`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteProject(req: DeleteProjectRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<DeleteProjectRequest, GoogleProtobufEmpty.Empty>(`/pfs_v2.API/DeleteProject`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
}
