// package: pfs
// file: pfs/pfs.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as google_protobuf_wrappers_pb from "google-protobuf/google/protobuf/wrappers_pb";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";
import * as auth_auth_pb from "../auth/auth_pb";

export class Repo extends jspb.Message { 
    getName(): string;
    setName(value: string): Repo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Repo.AsObject;
    static toObject(includeInstance: boolean, msg: Repo): Repo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Repo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Repo;
    static deserializeBinaryFromReader(message: Repo, reader: jspb.BinaryReader): Repo;
}

export namespace Repo {
    export type AsObject = {
        name: string,
    }
}

export class Branch extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): Branch;

    getName(): string;
    setName(value: string): Branch;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Branch.AsObject;
    static toObject(includeInstance: boolean, msg: Branch): Branch.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Branch, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Branch;
    static deserializeBinaryFromReader(message: Branch, reader: jspb.BinaryReader): Branch;
}

export namespace Branch {
    export type AsObject = {
        repo?: Repo.AsObject,
        name: string,
    }
}

export class File extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): File;

    getPath(): string;
    setPath(value: string): File;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): File.AsObject;
    static toObject(includeInstance: boolean, msg: File): File.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: File, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): File;
    static deserializeBinaryFromReader(message: File, reader: jspb.BinaryReader): File;
}

export namespace File {
    export type AsObject = {
        commit?: Commit.AsObject,
        path: string,
    }
}

export class RepoInfo extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): RepoInfo;


    hasCreated(): boolean;
    clearCreated(): void;
    getCreated(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setCreated(value?: google_protobuf_timestamp_pb.Timestamp): RepoInfo;

    getSizeBytes(): number;
    setSizeBytes(value: number): RepoInfo;

    getDescription(): string;
    setDescription(value: string): RepoInfo;

    clearBranchesList(): void;
    getBranchesList(): Array<Branch>;
    setBranchesList(value: Array<Branch>): RepoInfo;
    addBranches(value?: Branch, index?: number): Branch;


    hasAuthInfo(): boolean;
    clearAuthInfo(): void;
    getAuthInfo(): RepoAuthInfo | undefined;
    setAuthInfo(value?: RepoAuthInfo): RepoInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RepoInfo.AsObject;
    static toObject(includeInstance: boolean, msg: RepoInfo): RepoInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RepoInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RepoInfo;
    static deserializeBinaryFromReader(message: RepoInfo, reader: jspb.BinaryReader): RepoInfo;
}

export namespace RepoInfo {
    export type AsObject = {
        repo?: Repo.AsObject,
        created?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        sizeBytes: number,
        description: string,
        branchesList: Array<Branch.AsObject>,
        authInfo?: RepoAuthInfo.AsObject,
    }
}

export class RepoAuthInfo extends jspb.Message { 
    getAccessLevel(): auth_auth_pb.Scope;
    setAccessLevel(value: auth_auth_pb.Scope): RepoAuthInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RepoAuthInfo.AsObject;
    static toObject(includeInstance: boolean, msg: RepoAuthInfo): RepoAuthInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RepoAuthInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RepoAuthInfo;
    static deserializeBinaryFromReader(message: RepoAuthInfo, reader: jspb.BinaryReader): RepoAuthInfo;
}

export namespace RepoAuthInfo {
    export type AsObject = {
        accessLevel: auth_auth_pb.Scope,
    }
}

export class BranchInfo extends jspb.Message { 

    hasBranch(): boolean;
    clearBranch(): void;
    getBranch(): Branch | undefined;
    setBranch(value?: Branch): BranchInfo;


    hasHead(): boolean;
    clearHead(): void;
    getHead(): Commit | undefined;
    setHead(value?: Commit): BranchInfo;

    clearProvenanceList(): void;
    getProvenanceList(): Array<Branch>;
    setProvenanceList(value: Array<Branch>): BranchInfo;
    addProvenance(value?: Branch, index?: number): Branch;

    clearSubvenanceList(): void;
    getSubvenanceList(): Array<Branch>;
    setSubvenanceList(value: Array<Branch>): BranchInfo;
    addSubvenance(value?: Branch, index?: number): Branch;

    clearDirectProvenanceList(): void;
    getDirectProvenanceList(): Array<Branch>;
    setDirectProvenanceList(value: Array<Branch>): BranchInfo;
    addDirectProvenance(value?: Branch, index?: number): Branch;


    hasTrigger(): boolean;
    clearTrigger(): void;
    getTrigger(): Trigger | undefined;
    setTrigger(value?: Trigger): BranchInfo;

    getName(): string;
    setName(value: string): BranchInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BranchInfo.AsObject;
    static toObject(includeInstance: boolean, msg: BranchInfo): BranchInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BranchInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BranchInfo;
    static deserializeBinaryFromReader(message: BranchInfo, reader: jspb.BinaryReader): BranchInfo;
}

export namespace BranchInfo {
    export type AsObject = {
        branch?: Branch.AsObject,
        head?: Commit.AsObject,
        provenanceList: Array<Branch.AsObject>,
        subvenanceList: Array<Branch.AsObject>,
        directProvenanceList: Array<Branch.AsObject>,
        trigger?: Trigger.AsObject,
        name: string,
    }
}

export class BranchInfos extends jspb.Message { 
    clearBranchInfoList(): void;
    getBranchInfoList(): Array<BranchInfo>;
    setBranchInfoList(value: Array<BranchInfo>): BranchInfos;
    addBranchInfo(value?: BranchInfo, index?: number): BranchInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BranchInfos.AsObject;
    static toObject(includeInstance: boolean, msg: BranchInfos): BranchInfos.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BranchInfos, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BranchInfos;
    static deserializeBinaryFromReader(message: BranchInfos, reader: jspb.BinaryReader): BranchInfos;
}

export namespace BranchInfos {
    export type AsObject = {
        branchInfoList: Array<BranchInfo.AsObject>,
    }
}

export class Trigger extends jspb.Message { 
    getBranch(): string;
    setBranch(value: string): Trigger;

    getAll(): boolean;
    setAll(value: boolean): Trigger;

    getCronSpec(): string;
    setCronSpec(value: string): Trigger;

    getSize(): string;
    setSize(value: string): Trigger;

    getCommits(): number;
    setCommits(value: number): Trigger;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Trigger.AsObject;
    static toObject(includeInstance: boolean, msg: Trigger): Trigger.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Trigger, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Trigger;
    static deserializeBinaryFromReader(message: Trigger, reader: jspb.BinaryReader): Trigger;
}

export namespace Trigger {
    export type AsObject = {
        branch: string,
        all: boolean,
        cronSpec: string,
        size: string,
        commits: number,
    }
}

export class CommitOrigin extends jspb.Message { 
    getKind(): OriginKind;
    setKind(value: OriginKind): CommitOrigin;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CommitOrigin.AsObject;
    static toObject(includeInstance: boolean, msg: CommitOrigin): CommitOrigin.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CommitOrigin, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CommitOrigin;
    static deserializeBinaryFromReader(message: CommitOrigin, reader: jspb.BinaryReader): CommitOrigin;
}

export namespace CommitOrigin {
    export type AsObject = {
        kind: OriginKind,
    }
}

export class Commit extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): Commit;

    getId(): string;
    setId(value: string): Commit;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Commit.AsObject;
    static toObject(includeInstance: boolean, msg: Commit): Commit.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Commit, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Commit;
    static deserializeBinaryFromReader(message: Commit, reader: jspb.BinaryReader): Commit;
}

export namespace Commit {
    export type AsObject = {
        repo?: Repo.AsObject,
        id: string,
    }
}

export class CommitRange extends jspb.Message { 

    hasLower(): boolean;
    clearLower(): void;
    getLower(): Commit | undefined;
    setLower(value?: Commit): CommitRange;


    hasUpper(): boolean;
    clearUpper(): void;
    getUpper(): Commit | undefined;
    setUpper(value?: Commit): CommitRange;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CommitRange.AsObject;
    static toObject(includeInstance: boolean, msg: CommitRange): CommitRange.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CommitRange, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CommitRange;
    static deserializeBinaryFromReader(message: CommitRange, reader: jspb.BinaryReader): CommitRange;
}

export namespace CommitRange {
    export type AsObject = {
        lower?: Commit.AsObject,
        upper?: Commit.AsObject,
    }
}

export class CommitProvenance extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): CommitProvenance;


    hasBranch(): boolean;
    clearBranch(): void;
    getBranch(): Branch | undefined;
    setBranch(value?: Branch): CommitProvenance;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CommitProvenance.AsObject;
    static toObject(includeInstance: boolean, msg: CommitProvenance): CommitProvenance.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CommitProvenance, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CommitProvenance;
    static deserializeBinaryFromReader(message: CommitProvenance, reader: jspb.BinaryReader): CommitProvenance;
}

export namespace CommitProvenance {
    export type AsObject = {
        commit?: Commit.AsObject,
        branch?: Branch.AsObject,
    }
}

export class CommitInfo extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): CommitInfo;


    hasBranch(): boolean;
    clearBranch(): void;
    getBranch(): Branch | undefined;
    setBranch(value?: Branch): CommitInfo;


    hasOrigin(): boolean;
    clearOrigin(): void;
    getOrigin(): CommitOrigin | undefined;
    setOrigin(value?: CommitOrigin): CommitInfo;

    getDescription(): string;
    setDescription(value: string): CommitInfo;


    hasParentCommit(): boolean;
    clearParentCommit(): void;
    getParentCommit(): Commit | undefined;
    setParentCommit(value?: Commit): CommitInfo;

    clearChildCommitsList(): void;
    getChildCommitsList(): Array<Commit>;
    setChildCommitsList(value: Array<Commit>): CommitInfo;
    addChildCommits(value?: Commit, index?: number): Commit;


    hasStarted(): boolean;
    clearStarted(): void;
    getStarted(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setStarted(value?: google_protobuf_timestamp_pb.Timestamp): CommitInfo;


    hasFinished(): boolean;
    clearFinished(): void;
    getFinished(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setFinished(value?: google_protobuf_timestamp_pb.Timestamp): CommitInfo;

    getSizeBytes(): number;
    setSizeBytes(value: number): CommitInfo;

    clearProvenanceList(): void;
    getProvenanceList(): Array<CommitProvenance>;
    setProvenanceList(value: Array<CommitProvenance>): CommitInfo;
    addProvenance(value?: CommitProvenance, index?: number): CommitProvenance;

    getReadyProvenance(): number;
    setReadyProvenance(value: number): CommitInfo;

    clearSubvenanceList(): void;
    getSubvenanceList(): Array<CommitRange>;
    setSubvenanceList(value: Array<CommitRange>): CommitInfo;
    addSubvenance(value?: CommitRange, index?: number): CommitRange;

    getSubvenantCommitsSuccess(): number;
    setSubvenantCommitsSuccess(value: number): CommitInfo;

    getSubvenantCommitsFailure(): number;
    setSubvenantCommitsFailure(value: number): CommitInfo;

    getSubvenantCommitsTotal(): number;
    setSubvenantCommitsTotal(value: number): CommitInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CommitInfo.AsObject;
    static toObject(includeInstance: boolean, msg: CommitInfo): CommitInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CommitInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CommitInfo;
    static deserializeBinaryFromReader(message: CommitInfo, reader: jspb.BinaryReader): CommitInfo;
}

export namespace CommitInfo {
    export type AsObject = {
        commit?: Commit.AsObject,
        branch?: Branch.AsObject,
        origin?: CommitOrigin.AsObject,
        description: string,
        parentCommit?: Commit.AsObject,
        childCommitsList: Array<Commit.AsObject>,
        started?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        finished?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        sizeBytes: number,
        provenanceList: Array<CommitProvenance.AsObject>,
        readyProvenance: number,
        subvenanceList: Array<CommitRange.AsObject>,
        subvenantCommitsSuccess: number,
        subvenantCommitsFailure: number,
        subvenantCommitsTotal: number,
    }
}

export class FileInfo extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): FileInfo;

    getFileType(): FileType;
    setFileType(value: FileType): FileInfo;

    getSizeBytes(): number;
    setSizeBytes(value: number): FileInfo;


    hasCommitted(): boolean;
    clearCommitted(): void;
    getCommitted(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setCommitted(value?: google_protobuf_timestamp_pb.Timestamp): FileInfo;

    getHash(): Uint8Array | string;
    getHash_asU8(): Uint8Array;
    getHash_asB64(): string;
    setHash(value: Uint8Array | string): FileInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FileInfo.AsObject;
    static toObject(includeInstance: boolean, msg: FileInfo): FileInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FileInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FileInfo;
    static deserializeBinaryFromReader(message: FileInfo, reader: jspb.BinaryReader): FileInfo;
}

export namespace FileInfo {
    export type AsObject = {
        file?: File.AsObject,
        fileType: FileType,
        sizeBytes: number,
        committed?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        hash: Uint8Array | string,
    }
}

export class CreateRepoRequest extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): CreateRepoRequest;

    getDescription(): string;
    setDescription(value: string): CreateRepoRequest;

    getUpdate(): boolean;
    setUpdate(value: boolean): CreateRepoRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateRepoRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateRepoRequest): CreateRepoRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateRepoRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateRepoRequest;
    static deserializeBinaryFromReader(message: CreateRepoRequest, reader: jspb.BinaryReader): CreateRepoRequest;
}

export namespace CreateRepoRequest {
    export type AsObject = {
        repo?: Repo.AsObject,
        description: string,
        update: boolean,
    }
}

export class InspectRepoRequest extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): InspectRepoRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectRepoRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectRepoRequest): InspectRepoRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectRepoRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectRepoRequest;
    static deserializeBinaryFromReader(message: InspectRepoRequest, reader: jspb.BinaryReader): InspectRepoRequest;
}

export namespace InspectRepoRequest {
    export type AsObject = {
        repo?: Repo.AsObject,
    }
}

export class ListRepoRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListRepoRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListRepoRequest): ListRepoRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListRepoRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListRepoRequest;
    static deserializeBinaryFromReader(message: ListRepoRequest, reader: jspb.BinaryReader): ListRepoRequest;
}

export namespace ListRepoRequest {
    export type AsObject = {
    }
}

export class ListRepoResponse extends jspb.Message { 
    clearRepoInfoList(): void;
    getRepoInfoList(): Array<RepoInfo>;
    setRepoInfoList(value: Array<RepoInfo>): ListRepoResponse;
    addRepoInfo(value?: RepoInfo, index?: number): RepoInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListRepoResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListRepoResponse): ListRepoResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListRepoResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListRepoResponse;
    static deserializeBinaryFromReader(message: ListRepoResponse, reader: jspb.BinaryReader): ListRepoResponse;
}

export namespace ListRepoResponse {
    export type AsObject = {
        repoInfoList: Array<RepoInfo.AsObject>,
    }
}

export class DeleteRepoRequest extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): DeleteRepoRequest;

    getForce(): boolean;
    setForce(value: boolean): DeleteRepoRequest;

    getAll(): boolean;
    setAll(value: boolean): DeleteRepoRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteRepoRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteRepoRequest): DeleteRepoRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteRepoRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteRepoRequest;
    static deserializeBinaryFromReader(message: DeleteRepoRequest, reader: jspb.BinaryReader): DeleteRepoRequest;
}

export namespace DeleteRepoRequest {
    export type AsObject = {
        repo?: Repo.AsObject,
        force: boolean,
        all: boolean,
    }
}

export class StartCommitRequest extends jspb.Message { 

    hasParent(): boolean;
    clearParent(): void;
    getParent(): Commit | undefined;
    setParent(value?: Commit): StartCommitRequest;

    getDescription(): string;
    setDescription(value: string): StartCommitRequest;

    getBranch(): string;
    setBranch(value: string): StartCommitRequest;

    clearProvenanceList(): void;
    getProvenanceList(): Array<CommitProvenance>;
    setProvenanceList(value: Array<CommitProvenance>): StartCommitRequest;
    addProvenance(value?: CommitProvenance, index?: number): CommitProvenance;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StartCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: StartCommitRequest): StartCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StartCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StartCommitRequest;
    static deserializeBinaryFromReader(message: StartCommitRequest, reader: jspb.BinaryReader): StartCommitRequest;
}

export namespace StartCommitRequest {
    export type AsObject = {
        parent?: Commit.AsObject,
        description: string,
        branch: string,
        provenanceList: Array<CommitProvenance.AsObject>,
    }
}

export class FinishCommitRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): FinishCommitRequest;

    getDescription(): string;
    setDescription(value: string): FinishCommitRequest;

    getSizeBytes(): number;
    setSizeBytes(value: number): FinishCommitRequest;

    getEmpty(): boolean;
    setEmpty(value: boolean): FinishCommitRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FinishCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: FinishCommitRequest): FinishCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FinishCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FinishCommitRequest;
    static deserializeBinaryFromReader(message: FinishCommitRequest, reader: jspb.BinaryReader): FinishCommitRequest;
}

export namespace FinishCommitRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
        description: string,
        sizeBytes: number,
        empty: boolean,
    }
}

export class InspectCommitRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): InspectCommitRequest;

    getBlockState(): CommitState;
    setBlockState(value: CommitState): InspectCommitRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectCommitRequest): InspectCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectCommitRequest;
    static deserializeBinaryFromReader(message: InspectCommitRequest, reader: jspb.BinaryReader): InspectCommitRequest;
}

export namespace InspectCommitRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
        blockState: CommitState,
    }
}

export class ListCommitRequest extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): ListCommitRequest;


    hasFrom(): boolean;
    clearFrom(): void;
    getFrom(): Commit | undefined;
    setFrom(value?: Commit): ListCommitRequest;


    hasTo(): boolean;
    clearTo(): void;
    getTo(): Commit | undefined;
    setTo(value?: Commit): ListCommitRequest;

    getNumber(): number;
    setNumber(value: number): ListCommitRequest;

    getReverse(): boolean;
    setReverse(value: boolean): ListCommitRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListCommitRequest): ListCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListCommitRequest;
    static deserializeBinaryFromReader(message: ListCommitRequest, reader: jspb.BinaryReader): ListCommitRequest;
}

export namespace ListCommitRequest {
    export type AsObject = {
        repo?: Repo.AsObject,
        from?: Commit.AsObject,
        to?: Commit.AsObject,
        number: number,
        reverse: boolean,
    }
}

export class CommitInfos extends jspb.Message { 
    clearCommitInfoList(): void;
    getCommitInfoList(): Array<CommitInfo>;
    setCommitInfoList(value: Array<CommitInfo>): CommitInfos;
    addCommitInfo(value?: CommitInfo, index?: number): CommitInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CommitInfos.AsObject;
    static toObject(includeInstance: boolean, msg: CommitInfos): CommitInfos.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CommitInfos, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CommitInfos;
    static deserializeBinaryFromReader(message: CommitInfos, reader: jspb.BinaryReader): CommitInfos;
}

export namespace CommitInfos {
    export type AsObject = {
        commitInfoList: Array<CommitInfo.AsObject>,
    }
}

export class SquashCommitRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): SquashCommitRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SquashCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SquashCommitRequest): SquashCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SquashCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SquashCommitRequest;
    static deserializeBinaryFromReader(message: SquashCommitRequest, reader: jspb.BinaryReader): SquashCommitRequest;
}

export namespace SquashCommitRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
    }
}

export class FlushCommitRequest extends jspb.Message { 
    clearCommitsList(): void;
    getCommitsList(): Array<Commit>;
    setCommitsList(value: Array<Commit>): FlushCommitRequest;
    addCommits(value?: Commit, index?: number): Commit;

    clearToReposList(): void;
    getToReposList(): Array<Repo>;
    setToReposList(value: Array<Repo>): FlushCommitRequest;
    addToRepos(value?: Repo, index?: number): Repo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FlushCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: FlushCommitRequest): FlushCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FlushCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FlushCommitRequest;
    static deserializeBinaryFromReader(message: FlushCommitRequest, reader: jspb.BinaryReader): FlushCommitRequest;
}

export namespace FlushCommitRequest {
    export type AsObject = {
        commitsList: Array<Commit.AsObject>,
        toReposList: Array<Repo.AsObject>,
    }
}

export class SubscribeCommitRequest extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): SubscribeCommitRequest;

    getBranch(): string;
    setBranch(value: string): SubscribeCommitRequest;


    hasProv(): boolean;
    clearProv(): void;
    getProv(): CommitProvenance | undefined;
    setProv(value?: CommitProvenance): SubscribeCommitRequest;


    hasFrom(): boolean;
    clearFrom(): void;
    getFrom(): Commit | undefined;
    setFrom(value?: Commit): SubscribeCommitRequest;

    getState(): CommitState;
    setState(value: CommitState): SubscribeCommitRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SubscribeCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SubscribeCommitRequest): SubscribeCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SubscribeCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SubscribeCommitRequest;
    static deserializeBinaryFromReader(message: SubscribeCommitRequest, reader: jspb.BinaryReader): SubscribeCommitRequest;
}

export namespace SubscribeCommitRequest {
    export type AsObject = {
        repo?: Repo.AsObject,
        branch: string,
        prov?: CommitProvenance.AsObject,
        from?: Commit.AsObject,
        state: CommitState,
    }
}

export class ClearCommitRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): ClearCommitRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ClearCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ClearCommitRequest): ClearCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ClearCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ClearCommitRequest;
    static deserializeBinaryFromReader(message: ClearCommitRequest, reader: jspb.BinaryReader): ClearCommitRequest;
}

export namespace ClearCommitRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
    }
}

export class CreateBranchRequest extends jspb.Message { 

    hasHead(): boolean;
    clearHead(): void;
    getHead(): Commit | undefined;
    setHead(value?: Commit): CreateBranchRequest;

    getSBranch(): string;
    setSBranch(value: string): CreateBranchRequest;


    hasBranch(): boolean;
    clearBranch(): void;
    getBranch(): Branch | undefined;
    setBranch(value?: Branch): CreateBranchRequest;

    clearProvenanceList(): void;
    getProvenanceList(): Array<Branch>;
    setProvenanceList(value: Array<Branch>): CreateBranchRequest;
    addProvenance(value?: Branch, index?: number): Branch;


    hasTrigger(): boolean;
    clearTrigger(): void;
    getTrigger(): Trigger | undefined;
    setTrigger(value?: Trigger): CreateBranchRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateBranchRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateBranchRequest): CreateBranchRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateBranchRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateBranchRequest;
    static deserializeBinaryFromReader(message: CreateBranchRequest, reader: jspb.BinaryReader): CreateBranchRequest;
}

export namespace CreateBranchRequest {
    export type AsObject = {
        head?: Commit.AsObject,
        sBranch: string,
        branch?: Branch.AsObject,
        provenanceList: Array<Branch.AsObject>,
        trigger?: Trigger.AsObject,
    }
}

export class InspectBranchRequest extends jspb.Message { 

    hasBranch(): boolean;
    clearBranch(): void;
    getBranch(): Branch | undefined;
    setBranch(value?: Branch): InspectBranchRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectBranchRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectBranchRequest): InspectBranchRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectBranchRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectBranchRequest;
    static deserializeBinaryFromReader(message: InspectBranchRequest, reader: jspb.BinaryReader): InspectBranchRequest;
}

export namespace InspectBranchRequest {
    export type AsObject = {
        branch?: Branch.AsObject,
    }
}

export class ListBranchRequest extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): ListBranchRequest;

    getReverse(): boolean;
    setReverse(value: boolean): ListBranchRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListBranchRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListBranchRequest): ListBranchRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListBranchRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListBranchRequest;
    static deserializeBinaryFromReader(message: ListBranchRequest, reader: jspb.BinaryReader): ListBranchRequest;
}

export namespace ListBranchRequest {
    export type AsObject = {
        repo?: Repo.AsObject,
        reverse: boolean,
    }
}

export class DeleteBranchRequest extends jspb.Message { 

    hasBranch(): boolean;
    clearBranch(): void;
    getBranch(): Branch | undefined;
    setBranch(value?: Branch): DeleteBranchRequest;

    getForce(): boolean;
    setForce(value: boolean): DeleteBranchRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteBranchRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteBranchRequest): DeleteBranchRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteBranchRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteBranchRequest;
    static deserializeBinaryFromReader(message: DeleteBranchRequest, reader: jspb.BinaryReader): DeleteBranchRequest;
}

export namespace DeleteBranchRequest {
    export type AsObject = {
        branch?: Branch.AsObject,
        force: boolean,
    }
}

export class AppendFile extends jspb.Message { 
    getOverwrite(): boolean;
    setOverwrite(value: boolean): AppendFile;

    getTag(): string;
    setTag(value: string): AppendFile;


    hasRawFileSource(): boolean;
    clearRawFileSource(): void;
    getRawFileSource(): RawFileSource | undefined;
    setRawFileSource(value?: RawFileSource): AppendFile;


    hasTarFileSource(): boolean;
    clearTarFileSource(): void;
    getTarFileSource(): TarFileSource | undefined;
    setTarFileSource(value?: TarFileSource): AppendFile;


    hasUrlFileSource(): boolean;
    clearUrlFileSource(): void;
    getUrlFileSource(): URLFileSource | undefined;
    setUrlFileSource(value?: URLFileSource): AppendFile;


    getSourceCase(): AppendFile.SourceCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AppendFile.AsObject;
    static toObject(includeInstance: boolean, msg: AppendFile): AppendFile.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AppendFile, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AppendFile;
    static deserializeBinaryFromReader(message: AppendFile, reader: jspb.BinaryReader): AppendFile;
}

export namespace AppendFile {
    export type AsObject = {
        overwrite: boolean,
        tag: string,
        rawFileSource?: RawFileSource.AsObject,
        tarFileSource?: TarFileSource.AsObject,
        urlFileSource?: URLFileSource.AsObject,
    }

    export enum SourceCase {
        SOURCE_NOT_SET = 0,
    
    RAW_FILE_SOURCE = 3,

    TAR_FILE_SOURCE = 4,

    URL_FILE_SOURCE = 5,

    }

}

export class RawFileSource extends jspb.Message { 
    getPath(): string;
    setPath(value: string): RawFileSource;

    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): RawFileSource;

    getEof(): boolean;
    setEof(value: boolean): RawFileSource;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RawFileSource.AsObject;
    static toObject(includeInstance: boolean, msg: RawFileSource): RawFileSource.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RawFileSource, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RawFileSource;
    static deserializeBinaryFromReader(message: RawFileSource, reader: jspb.BinaryReader): RawFileSource;
}

export namespace RawFileSource {
    export type AsObject = {
        path: string,
        data: Uint8Array | string,
        eof: boolean,
    }
}

export class TarFileSource extends jspb.Message { 
    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): TarFileSource;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TarFileSource.AsObject;
    static toObject(includeInstance: boolean, msg: TarFileSource): TarFileSource.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TarFileSource, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TarFileSource;
    static deserializeBinaryFromReader(message: TarFileSource, reader: jspb.BinaryReader): TarFileSource;
}

export namespace TarFileSource {
    export type AsObject = {
        data: Uint8Array | string,
    }
}

export class URLFileSource extends jspb.Message { 
    getPath(): string;
    setPath(value: string): URLFileSource;

    getUrl(): string;
    setUrl(value: string): URLFileSource;

    getRecursive(): boolean;
    setRecursive(value: boolean): URLFileSource;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): URLFileSource.AsObject;
    static toObject(includeInstance: boolean, msg: URLFileSource): URLFileSource.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: URLFileSource, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): URLFileSource;
    static deserializeBinaryFromReader(message: URLFileSource, reader: jspb.BinaryReader): URLFileSource;
}

export namespace URLFileSource {
    export type AsObject = {
        path: string,
        url: string,
        recursive: boolean,
    }
}

export class DeleteFile extends jspb.Message { 
    getFile(): string;
    setFile(value: string): DeleteFile;

    getTag(): string;
    setTag(value: string): DeleteFile;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteFile.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteFile): DeleteFile.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteFile, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteFile;
    static deserializeBinaryFromReader(message: DeleteFile, reader: jspb.BinaryReader): DeleteFile;
}

export namespace DeleteFile {
    export type AsObject = {
        file: string,
        tag: string,
    }
}

export class ModifyFileRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): ModifyFileRequest;


    hasAppendFile(): boolean;
    clearAppendFile(): void;
    getAppendFile(): AppendFile | undefined;
    setAppendFile(value?: AppendFile): ModifyFileRequest;


    hasDeleteFile(): boolean;
    clearDeleteFile(): void;
    getDeleteFile(): DeleteFile | undefined;
    setDeleteFile(value?: DeleteFile): ModifyFileRequest;


    getModificationCase(): ModifyFileRequest.ModificationCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ModifyFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ModifyFileRequest): ModifyFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ModifyFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ModifyFileRequest;
    static deserializeBinaryFromReader(message: ModifyFileRequest, reader: jspb.BinaryReader): ModifyFileRequest;
}

export namespace ModifyFileRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
        appendFile?: AppendFile.AsObject,
        deleteFile?: DeleteFile.AsObject,
    }

    export enum ModificationCase {
        MODIFICATION_NOT_SET = 0,
    
    APPEND_FILE = 2,

    DELETE_FILE = 3,

    }

}

export class CopyFileRequest extends jspb.Message { 

    hasSrc(): boolean;
    clearSrc(): void;
    getSrc(): File | undefined;
    setSrc(value?: File): CopyFileRequest;


    hasDst(): boolean;
    clearDst(): void;
    getDst(): File | undefined;
    setDst(value?: File): CopyFileRequest;

    getOverwrite(): boolean;
    setOverwrite(value: boolean): CopyFileRequest;

    getTag(): string;
    setTag(value: string): CopyFileRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CopyFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CopyFileRequest): CopyFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CopyFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CopyFileRequest;
    static deserializeBinaryFromReader(message: CopyFileRequest, reader: jspb.BinaryReader): CopyFileRequest;
}

export namespace CopyFileRequest {
    export type AsObject = {
        src?: File.AsObject,
        dst?: File.AsObject,
        overwrite: boolean,
        tag: string,
    }
}

export class GetFileRequest extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): GetFileRequest;

    getUrl(): string;
    setUrl(value: string): GetFileRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetFileRequest): GetFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetFileRequest;
    static deserializeBinaryFromReader(message: GetFileRequest, reader: jspb.BinaryReader): GetFileRequest;
}

export namespace GetFileRequest {
    export type AsObject = {
        file?: File.AsObject,
        url: string,
    }
}

export class InspectFileRequest extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): InspectFileRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectFileRequest): InspectFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectFileRequest;
    static deserializeBinaryFromReader(message: InspectFileRequest, reader: jspb.BinaryReader): InspectFileRequest;
}

export namespace InspectFileRequest {
    export type AsObject = {
        file?: File.AsObject,
    }
}

export class ListFileRequest extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): ListFileRequest;

    getFull(): boolean;
    setFull(value: boolean): ListFileRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListFileRequest): ListFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListFileRequest;
    static deserializeBinaryFromReader(message: ListFileRequest, reader: jspb.BinaryReader): ListFileRequest;
}

export namespace ListFileRequest {
    export type AsObject = {
        file?: File.AsObject,
        full: boolean,
    }
}

export class WalkFileRequest extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): WalkFileRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WalkFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: WalkFileRequest): WalkFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WalkFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WalkFileRequest;
    static deserializeBinaryFromReader(message: WalkFileRequest, reader: jspb.BinaryReader): WalkFileRequest;
}

export namespace WalkFileRequest {
    export type AsObject = {
        file?: File.AsObject,
    }
}

export class GlobFileRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): GlobFileRequest;

    getPattern(): string;
    setPattern(value: string): GlobFileRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GlobFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GlobFileRequest): GlobFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GlobFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GlobFileRequest;
    static deserializeBinaryFromReader(message: GlobFileRequest, reader: jspb.BinaryReader): GlobFileRequest;
}

export namespace GlobFileRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
        pattern: string,
    }
}

export class DiffFileRequest extends jspb.Message { 

    hasNewFile(): boolean;
    clearNewFile(): void;
    getNewFile(): File | undefined;
    setNewFile(value?: File): DiffFileRequest;


    hasOldFile(): boolean;
    clearOldFile(): void;
    getOldFile(): File | undefined;
    setOldFile(value?: File): DiffFileRequest;

    getShallow(): boolean;
    setShallow(value: boolean): DiffFileRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DiffFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DiffFileRequest): DiffFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DiffFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DiffFileRequest;
    static deserializeBinaryFromReader(message: DiffFileRequest, reader: jspb.BinaryReader): DiffFileRequest;
}

export namespace DiffFileRequest {
    export type AsObject = {
        newFile?: File.AsObject,
        oldFile?: File.AsObject,
        shallow: boolean,
    }
}

export class DiffFileResponse extends jspb.Message { 

    hasNewFile(): boolean;
    clearNewFile(): void;
    getNewFile(): FileInfo | undefined;
    setNewFile(value?: FileInfo): DiffFileResponse;


    hasOldFile(): boolean;
    clearOldFile(): void;
    getOldFile(): FileInfo | undefined;
    setOldFile(value?: FileInfo): DiffFileResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DiffFileResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DiffFileResponse): DiffFileResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DiffFileResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DiffFileResponse;
    static deserializeBinaryFromReader(message: DiffFileResponse, reader: jspb.BinaryReader): DiffFileResponse;
}

export namespace DiffFileResponse {
    export type AsObject = {
        newFile?: FileInfo.AsObject,
        oldFile?: FileInfo.AsObject,
    }
}

export class FsckRequest extends jspb.Message { 
    getFix(): boolean;
    setFix(value: boolean): FsckRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FsckRequest.AsObject;
    static toObject(includeInstance: boolean, msg: FsckRequest): FsckRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FsckRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FsckRequest;
    static deserializeBinaryFromReader(message: FsckRequest, reader: jspb.BinaryReader): FsckRequest;
}

export namespace FsckRequest {
    export type AsObject = {
        fix: boolean,
    }
}

export class FsckResponse extends jspb.Message { 
    getFix(): string;
    setFix(value: string): FsckResponse;

    getError(): string;
    setError(value: string): FsckResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FsckResponse.AsObject;
    static toObject(includeInstance: boolean, msg: FsckResponse): FsckResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FsckResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FsckResponse;
    static deserializeBinaryFromReader(message: FsckResponse, reader: jspb.BinaryReader): FsckResponse;
}

export namespace FsckResponse {
    export type AsObject = {
        fix: string,
        error: string,
    }
}

export class CreateFilesetResponse extends jspb.Message { 
    getFilesetId(): string;
    setFilesetId(value: string): CreateFilesetResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateFilesetResponse.AsObject;
    static toObject(includeInstance: boolean, msg: CreateFilesetResponse): CreateFilesetResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateFilesetResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateFilesetResponse;
    static deserializeBinaryFromReader(message: CreateFilesetResponse, reader: jspb.BinaryReader): CreateFilesetResponse;
}

export namespace CreateFilesetResponse {
    export type AsObject = {
        filesetId: string,
    }
}

export class RenewFilesetRequest extends jspb.Message { 
    getFilesetId(): string;
    setFilesetId(value: string): RenewFilesetRequest;

    getTtlSeconds(): number;
    setTtlSeconds(value: number): RenewFilesetRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RenewFilesetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RenewFilesetRequest): RenewFilesetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RenewFilesetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RenewFilesetRequest;
    static deserializeBinaryFromReader(message: RenewFilesetRequest, reader: jspb.BinaryReader): RenewFilesetRequest;
}

export namespace RenewFilesetRequest {
    export type AsObject = {
        filesetId: string,
        ttlSeconds: number,
    }
}

export class AddFilesetRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): AddFilesetRequest;

    getFilesetId(): string;
    setFilesetId(value: string): AddFilesetRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AddFilesetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: AddFilesetRequest): AddFilesetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AddFilesetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AddFilesetRequest;
    static deserializeBinaryFromReader(message: AddFilesetRequest, reader: jspb.BinaryReader): AddFilesetRequest;
}

export namespace AddFilesetRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
        filesetId: string,
    }
}

export class GetFilesetRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): GetFilesetRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetFilesetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetFilesetRequest): GetFilesetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetFilesetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetFilesetRequest;
    static deserializeBinaryFromReader(message: GetFilesetRequest, reader: jspb.BinaryReader): GetFilesetRequest;
}

export namespace GetFilesetRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
    }
}

export enum OriginKind {
    USER = 0,
    AUTO = 1,
    FSCK = 2,
}

export enum FileType {
    RESERVED = 0,
    FILE = 1,
    DIR = 2,
}

export enum CommitState {
    STARTED = 0,
    READY = 1,
    FINISHED = 2,
}

export enum Delimiter {
    NONE = 0,
    JSON = 1,
    LINE = 2,
    SQL = 3,
    CSV = 4,
}
