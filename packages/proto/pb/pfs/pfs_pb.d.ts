// package: pfs_v2
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
    getType(): string;
    setType(value: string): Repo;

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
        type: string,
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
    getTag(): string;
    setTag(value: string): File;

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
        tag: string,
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
    getSizeBytesUpperBound(): number;
    setSizeBytesUpperBound(value: number): RepoInfo;
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

    hasDetails(): boolean;
    clearDetails(): void;
    getDetails(): RepoInfo.Details | undefined;
    setDetails(value?: RepoInfo.Details): RepoInfo;

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
        sizeBytesUpperBound: number,
        description: string,
        branchesList: Array<Branch.AsObject>,
        authInfo?: RepoAuthInfo.AsObject,
        details?: RepoInfo.Details.AsObject,
    }


    export class Details extends jspb.Message { 
        getSizeBytes(): number;
        setSizeBytes(value: number): Details;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): Details.AsObject;
        static toObject(includeInstance: boolean, msg: Details): Details.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: Details, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): Details;
        static deserializeBinaryFromReader(message: Details, reader: jspb.BinaryReader): Details;
    }

    export namespace Details {
        export type AsObject = {
            sizeBytes: number,
        }
    }

}

export class RepoAuthInfo extends jspb.Message { 
    clearPermissionsList(): void;
    getPermissionsList(): Array<auth_auth_pb.Permission>;
    setPermissionsList(value: Array<auth_auth_pb.Permission>): RepoAuthInfo;
    addPermissions(value: auth_auth_pb.Permission, index?: number): auth_auth_pb.Permission;
    clearRolesList(): void;
    getRolesList(): Array<string>;
    setRolesList(value: Array<string>): RepoAuthInfo;
    addRoles(value: string, index?: number): string;

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
        permissionsList: Array<auth_auth_pb.Permission>,
        rolesList: Array<string>,
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

    hasBranch(): boolean;
    clearBranch(): void;
    getBranch(): Branch | undefined;
    setBranch(value?: Branch): Commit;
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
        branch?: Branch.AsObject,
        id: string,
    }
}

export class CommitInfo extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): CommitInfo;

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
    clearDirectProvenanceList(): void;
    getDirectProvenanceList(): Array<Branch>;
    setDirectProvenanceList(value: Array<Branch>): CommitInfo;
    addDirectProvenance(value?: Branch, index?: number): Branch;
    getError(): boolean;
    setError(value: boolean): CommitInfo;
    getSizeBytesUpperBound(): number;
    setSizeBytesUpperBound(value: number): CommitInfo;

    hasDetails(): boolean;
    clearDetails(): void;
    getDetails(): CommitInfo.Details | undefined;
    setDetails(value?: CommitInfo.Details): CommitInfo;

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
        origin?: CommitOrigin.AsObject,
        description: string,
        parentCommit?: Commit.AsObject,
        childCommitsList: Array<Commit.AsObject>,
        started?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        finished?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        directProvenanceList: Array<Branch.AsObject>,
        error: boolean,
        sizeBytesUpperBound: number,
        details?: CommitInfo.Details.AsObject,
    }


    export class Details extends jspb.Message { 
        getSizeBytes(): number;
        setSizeBytes(value: number): Details;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): Details.AsObject;
        static toObject(includeInstance: boolean, msg: Details): Details.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: Details, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): Details;
        static deserializeBinaryFromReader(message: Details, reader: jspb.BinaryReader): Details;
    }

    export namespace Details {
        export type AsObject = {
            sizeBytes: number,
        }
    }

}

export class CommitSet extends jspb.Message { 
    getId(): string;
    setId(value: string): CommitSet;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CommitSet.AsObject;
    static toObject(includeInstance: boolean, msg: CommitSet): CommitSet.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CommitSet, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CommitSet;
    static deserializeBinaryFromReader(message: CommitSet, reader: jspb.BinaryReader): CommitSet;
}

export namespace CommitSet {
    export type AsObject = {
        id: string,
    }
}

export class CommitSetInfo extends jspb.Message { 

    hasCommitSet(): boolean;
    clearCommitSet(): void;
    getCommitSet(): CommitSet | undefined;
    setCommitSet(value?: CommitSet): CommitSetInfo;
    clearCommitsList(): void;
    getCommitsList(): Array<CommitInfo>;
    setCommitsList(value: Array<CommitInfo>): CommitSetInfo;
    addCommits(value?: CommitInfo, index?: number): CommitInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CommitSetInfo.AsObject;
    static toObject(includeInstance: boolean, msg: CommitSetInfo): CommitSetInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CommitSetInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CommitSetInfo;
    static deserializeBinaryFromReader(message: CommitSetInfo, reader: jspb.BinaryReader): CommitSetInfo;
}

export namespace CommitSetInfo {
    export type AsObject = {
        commitSet?: CommitSet.AsObject,
        commitsList: Array<CommitInfo.AsObject>,
    }
}

export class FileInfo extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): FileInfo;
    getFileType(): FileType;
    setFileType(value: FileType): FileInfo;

    hasCommitted(): boolean;
    clearCommitted(): void;
    getCommitted(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setCommitted(value?: google_protobuf_timestamp_pb.Timestamp): FileInfo;
    getSizeBytes(): number;
    setSizeBytes(value: number): FileInfo;
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
        committed?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        sizeBytes: number,
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
    getType(): string;
    setType(value: string): ListRepoRequest;

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
        type: string,
    }
}

export class DeleteRepoRequest extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): DeleteRepoRequest;
    getForce(): boolean;
    setForce(value: boolean): DeleteRepoRequest;

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
    }
}

export class StartCommitRequest extends jspb.Message { 

    hasParent(): boolean;
    clearParent(): void;
    getParent(): Commit | undefined;
    setParent(value?: Commit): StartCommitRequest;
    getDescription(): string;
    setDescription(value: string): StartCommitRequest;

    hasBranch(): boolean;
    clearBranch(): void;
    getBranch(): Branch | undefined;
    setBranch(value?: Branch): StartCommitRequest;

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
        branch?: Branch.AsObject,
    }
}

export class FinishCommitRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): FinishCommitRequest;
    getDescription(): string;
    setDescription(value: string): FinishCommitRequest;
    getError(): boolean;
    setError(value: boolean): FinishCommitRequest;

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
        error: boolean,
    }
}

export class InspectCommitRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): InspectCommitRequest;
    getWait(): CommitState;
    setWait(value: CommitState): InspectCommitRequest;

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
        wait: CommitState,
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

export class InspectCommitSetRequest extends jspb.Message { 

    hasCommitSet(): boolean;
    clearCommitSet(): void;
    getCommitSet(): CommitSet | undefined;
    setCommitSet(value?: CommitSet): InspectCommitSetRequest;
    getWait(): boolean;
    setWait(value: boolean): InspectCommitSetRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectCommitSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectCommitSetRequest): InspectCommitSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectCommitSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectCommitSetRequest;
    static deserializeBinaryFromReader(message: InspectCommitSetRequest, reader: jspb.BinaryReader): InspectCommitSetRequest;
}

export namespace InspectCommitSetRequest {
    export type AsObject = {
        commitSet?: CommitSet.AsObject,
        wait: boolean,
    }
}

export class ListCommitSetRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListCommitSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListCommitSetRequest): ListCommitSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListCommitSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListCommitSetRequest;
    static deserializeBinaryFromReader(message: ListCommitSetRequest, reader: jspb.BinaryReader): ListCommitSetRequest;
}

export namespace ListCommitSetRequest {
    export type AsObject = {
    }
}

export class SquashCommitSetRequest extends jspb.Message { 

    hasCommitSet(): boolean;
    clearCommitSet(): void;
    getCommitSet(): CommitSet | undefined;
    setCommitSet(value?: CommitSet): SquashCommitSetRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SquashCommitSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SquashCommitSetRequest): SquashCommitSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SquashCommitSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SquashCommitSetRequest;
    static deserializeBinaryFromReader(message: SquashCommitSetRequest, reader: jspb.BinaryReader): SquashCommitSetRequest;
}

export namespace SquashCommitSetRequest {
    export type AsObject = {
        commitSet?: CommitSet.AsObject,
    }
}

export class SubscribeCommitRequest extends jspb.Message { 

    hasRepo(): boolean;
    clearRepo(): void;
    getRepo(): Repo | undefined;
    setRepo(value?: Repo): SubscribeCommitRequest;
    getBranch(): string;
    setBranch(value: string): SubscribeCommitRequest;

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
    getNewCommitSet(): boolean;
    setNewCommitSet(value: boolean): CreateBranchRequest;

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
        branch?: Branch.AsObject,
        provenanceList: Array<Branch.AsObject>,
        trigger?: Trigger.AsObject,
        newCommitSet: boolean,
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

export class AddFile extends jspb.Message { 
    getPath(): string;
    setPath(value: string): AddFile;
    getTag(): string;
    setTag(value: string): AddFile;

    hasRaw(): boolean;
    clearRaw(): void;
    getRaw(): google_protobuf_wrappers_pb.BytesValue | undefined;
    setRaw(value?: google_protobuf_wrappers_pb.BytesValue): AddFile;

    hasUrl(): boolean;
    clearUrl(): void;
    getUrl(): AddFile.URLSource | undefined;
    setUrl(value?: AddFile.URLSource): AddFile;

    getSourceCase(): AddFile.SourceCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AddFile.AsObject;
    static toObject(includeInstance: boolean, msg: AddFile): AddFile.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AddFile, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AddFile;
    static deserializeBinaryFromReader(message: AddFile, reader: jspb.BinaryReader): AddFile;
}

export namespace AddFile {
    export type AsObject = {
        path: string,
        tag: string,
        raw?: google_protobuf_wrappers_pb.BytesValue.AsObject,
        url?: AddFile.URLSource.AsObject,
    }


    export class URLSource extends jspb.Message { 
        getUrl(): string;
        setUrl(value: string): URLSource;
        getRecursive(): boolean;
        setRecursive(value: boolean): URLSource;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): URLSource.AsObject;
        static toObject(includeInstance: boolean, msg: URLSource): URLSource.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: URLSource, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): URLSource;
        static deserializeBinaryFromReader(message: URLSource, reader: jspb.BinaryReader): URLSource;
    }

    export namespace URLSource {
        export type AsObject = {
            url: string,
            recursive: boolean,
        }
    }


    export enum SourceCase {
        SOURCE_NOT_SET = 0,
        RAW = 3,
        URL = 4,
    }

}

export class DeleteFile extends jspb.Message { 
    getPath(): string;
    setPath(value: string): DeleteFile;
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
        path: string,
        tag: string,
    }
}

export class CopyFile extends jspb.Message { 
    getDst(): string;
    setDst(value: string): CopyFile;
    getTag(): string;
    setTag(value: string): CopyFile;

    hasSrc(): boolean;
    clearSrc(): void;
    getSrc(): File | undefined;
    setSrc(value?: File): CopyFile;
    getAppend(): boolean;
    setAppend(value: boolean): CopyFile;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CopyFile.AsObject;
    static toObject(includeInstance: boolean, msg: CopyFile): CopyFile.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CopyFile, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CopyFile;
    static deserializeBinaryFromReader(message: CopyFile, reader: jspb.BinaryReader): CopyFile;
}

export namespace CopyFile {
    export type AsObject = {
        dst: string,
        tag: string,
        src?: File.AsObject,
        append: boolean,
    }
}

export class ModifyFileRequest extends jspb.Message { 

    hasSetCommit(): boolean;
    clearSetCommit(): void;
    getSetCommit(): Commit | undefined;
    setSetCommit(value?: Commit): ModifyFileRequest;

    hasAddFile(): boolean;
    clearAddFile(): void;
    getAddFile(): AddFile | undefined;
    setAddFile(value?: AddFile): ModifyFileRequest;

    hasDeleteFile(): boolean;
    clearDeleteFile(): void;
    getDeleteFile(): DeleteFile | undefined;
    setDeleteFile(value?: DeleteFile): ModifyFileRequest;

    hasCopyFile(): boolean;
    clearCopyFile(): void;
    getCopyFile(): CopyFile | undefined;
    setCopyFile(value?: CopyFile): ModifyFileRequest;

    getBodyCase(): ModifyFileRequest.BodyCase;

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
        setCommit?: Commit.AsObject,
        addFile?: AddFile.AsObject,
        deleteFile?: DeleteFile.AsObject,
        copyFile?: CopyFile.AsObject,
    }

    export enum BodyCase {
        BODY_NOT_SET = 0,
        SET_COMMIT = 1,
        ADD_FILE = 2,
        DELETE_FILE = 3,
        COPY_FILE = 4,
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
    getDetails(): boolean;
    setDetails(value: boolean): ListFileRequest;

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
        details: boolean,
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

export class CreateFileSetResponse extends jspb.Message { 
    getFileSetId(): string;
    setFileSetId(value: string): CreateFileSetResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateFileSetResponse.AsObject;
    static toObject(includeInstance: boolean, msg: CreateFileSetResponse): CreateFileSetResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateFileSetResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateFileSetResponse;
    static deserializeBinaryFromReader(message: CreateFileSetResponse, reader: jspb.BinaryReader): CreateFileSetResponse;
}

export namespace CreateFileSetResponse {
    export type AsObject = {
        fileSetId: string,
    }
}

export class GetFileSetRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): GetFileSetRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetFileSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetFileSetRequest): GetFileSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetFileSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetFileSetRequest;
    static deserializeBinaryFromReader(message: GetFileSetRequest, reader: jspb.BinaryReader): GetFileSetRequest;
}

export namespace GetFileSetRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
    }
}

export class AddFileSetRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): AddFileSetRequest;
    getFileSetId(): string;
    setFileSetId(value: string): AddFileSetRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AddFileSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: AddFileSetRequest): AddFileSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AddFileSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AddFileSetRequest;
    static deserializeBinaryFromReader(message: AddFileSetRequest, reader: jspb.BinaryReader): AddFileSetRequest;
}

export namespace AddFileSetRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
        fileSetId: string,
    }
}

export class RenewFileSetRequest extends jspb.Message { 
    getFileSetId(): string;
    setFileSetId(value: string): RenewFileSetRequest;
    getTtlSeconds(): number;
    setTtlSeconds(value: number): RenewFileSetRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RenewFileSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RenewFileSetRequest): RenewFileSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RenewFileSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RenewFileSetRequest;
    static deserializeBinaryFromReader(message: RenewFileSetRequest, reader: jspb.BinaryReader): RenewFileSetRequest;
}

export namespace RenewFileSetRequest {
    export type AsObject = {
        fileSetId: string,
        ttlSeconds: number,
    }
}

export class ActivateAuthRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ActivateAuthRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ActivateAuthRequest): ActivateAuthRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ActivateAuthRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ActivateAuthRequest;
    static deserializeBinaryFromReader(message: ActivateAuthRequest, reader: jspb.BinaryReader): ActivateAuthRequest;
}

export namespace ActivateAuthRequest {
    export type AsObject = {
    }
}

export class ActivateAuthResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ActivateAuthResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ActivateAuthResponse): ActivateAuthResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ActivateAuthResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ActivateAuthResponse;
    static deserializeBinaryFromReader(message: ActivateAuthResponse, reader: jspb.BinaryReader): ActivateAuthResponse;
}

export namespace ActivateAuthResponse {
    export type AsObject = {
    }
}

export class RunLoadTestRequest extends jspb.Message { 
    getSpec(): Uint8Array | string;
    getSpec_asU8(): Uint8Array;
    getSpec_asB64(): string;
    setSpec(value: Uint8Array | string): RunLoadTestRequest;
    getSeed(): number;
    setSeed(value: number): RunLoadTestRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RunLoadTestRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RunLoadTestRequest): RunLoadTestRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RunLoadTestRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RunLoadTestRequest;
    static deserializeBinaryFromReader(message: RunLoadTestRequest, reader: jspb.BinaryReader): RunLoadTestRequest;
}

export namespace RunLoadTestRequest {
    export type AsObject = {
        spec: Uint8Array | string,
        seed: number,
    }
}

export class RunLoadTestResponse extends jspb.Message { 

    hasBranch(): boolean;
    clearBranch(): void;
    getBranch(): Branch | undefined;
    setBranch(value?: Branch): RunLoadTestResponse;
    getSeed(): number;
    setSeed(value: number): RunLoadTestResponse;
    getError(): string;
    setError(value: string): RunLoadTestResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RunLoadTestResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RunLoadTestResponse): RunLoadTestResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RunLoadTestResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RunLoadTestResponse;
    static deserializeBinaryFromReader(message: RunLoadTestResponse, reader: jspb.BinaryReader): RunLoadTestResponse;
}

export namespace RunLoadTestResponse {
    export type AsObject = {
        branch?: Branch.AsObject,
        seed: number,
        error: string,
    }
}

export enum OriginKind {
    ORIGIN_KIND_UNKNOWN = 0,
    USER = 1,
    AUTO = 2,
    FSCK = 3,
    ALIAS = 4,
}

export enum FileType {
    RESERVED = 0,
    FILE = 1,
    DIR = 2,
}

export enum CommitState {
    COMMIT_STATE_UNKNOWN = 0,
    STARTED = 1,
    READY = 2,
    FINISHED = 3,
}

export enum Delimiter {
    NONE = 0,
    JSON = 1,
    LINE = 2,
    SQL = 3,
    CSV = 4,
}
