// package: pfs_v2
// file: pfs/pfs.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as google_protobuf_wrappers_pb from "google-protobuf/google/protobuf/wrappers_pb";
import * as google_protobuf_duration_pb from "google-protobuf/google/protobuf/duration_pb";
import * as google_protobuf_any_pb from "google-protobuf/google/protobuf/any_pb";
import * as auth_auth_pb from "../auth/auth_pb";
import * as task_task_pb from "../task/task_pb";
import * as protoextensions_validate_pb from "../protoextensions/validate_pb";

export class Repo extends jspb.Message { 
    getName(): string;
    setName(value: string): Repo;
    getType(): string;
    setType(value: string): Repo;

    hasProject(): boolean;
    clearProject(): void;
    getProject(): Project | undefined;
    setProject(value?: Project): Repo;

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
        project?: Project.AsObject,
    }
}

export class RepoPicker extends jspb.Message { 

    hasName(): boolean;
    clearName(): void;
    getName(): RepoPicker.RepoName | undefined;
    setName(value?: RepoPicker.RepoName): RepoPicker;

    getPickerCase(): RepoPicker.PickerCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RepoPicker.AsObject;
    static toObject(includeInstance: boolean, msg: RepoPicker): RepoPicker.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RepoPicker, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RepoPicker;
    static deserializeBinaryFromReader(message: RepoPicker, reader: jspb.BinaryReader): RepoPicker;
}

export namespace RepoPicker {
    export type AsObject = {
        name?: RepoPicker.RepoName.AsObject,
    }


    export class RepoName extends jspb.Message { 

        hasProject(): boolean;
        clearProject(): void;
        getProject(): ProjectPicker | undefined;
        setProject(value?: ProjectPicker): RepoName;
        getName(): string;
        setName(value: string): RepoName;
        getType(): string;
        setType(value: string): RepoName;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): RepoName.AsObject;
        static toObject(includeInstance: boolean, msg: RepoName): RepoName.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: RepoName, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): RepoName;
        static deserializeBinaryFromReader(message: RepoName, reader: jspb.BinaryReader): RepoName;
    }

    export namespace RepoName {
        export type AsObject = {
            project?: ProjectPicker.AsObject,
            name: string,
            type: string,
        }
    }


    export enum PickerCase {
        PICKER_NOT_SET = 0,
        NAME = 1,
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

export class BranchPicker extends jspb.Message { 

    hasName(): boolean;
    clearName(): void;
    getName(): BranchPicker.BranchName | undefined;
    setName(value?: BranchPicker.BranchName): BranchPicker;

    getPickerCase(): BranchPicker.PickerCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BranchPicker.AsObject;
    static toObject(includeInstance: boolean, msg: BranchPicker): BranchPicker.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BranchPicker, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BranchPicker;
    static deserializeBinaryFromReader(message: BranchPicker, reader: jspb.BinaryReader): BranchPicker;
}

export namespace BranchPicker {
    export type AsObject = {
        name?: BranchPicker.BranchName.AsObject,
    }


    export class BranchName extends jspb.Message { 

        hasRepo(): boolean;
        clearRepo(): void;
        getRepo(): RepoPicker | undefined;
        setRepo(value?: RepoPicker): BranchName;
        getName(): string;
        setName(value: string): BranchName;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): BranchName.AsObject;
        static toObject(includeInstance: boolean, msg: BranchName): BranchName.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: BranchName, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): BranchName;
        static deserializeBinaryFromReader(message: BranchName, reader: jspb.BinaryReader): BranchName;
    }

    export namespace BranchName {
        export type AsObject = {
            repo?: RepoPicker.AsObject,
            name: string,
        }
    }


    export enum PickerCase {
        PICKER_NOT_SET = 0,
        NAME = 1,
    }

}

export class File extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): File;
    getPath(): string;
    setPath(value: string): File;
    getDatum(): string;
    setDatum(value: string): File;

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
        datum: string,
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
    getAuthInfo(): AuthInfo | undefined;
    setAuthInfo(value?: AuthInfo): RepoInfo;

    hasDetails(): boolean;
    clearDetails(): void;
    getDetails(): RepoInfo.Details | undefined;
    setDetails(value?: RepoInfo.Details): RepoInfo;

    getMetadataMap(): jspb.Map<string, string>;
    clearMetadataMap(): void;

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
        authInfo?: AuthInfo.AsObject,
        details?: RepoInfo.Details.AsObject,

        metadataMap: Array<[string, string]>,
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

export class AuthInfo extends jspb.Message { 
    clearPermissionsList(): void;
    getPermissionsList(): Array<auth_auth_pb.Permission>;
    setPermissionsList(value: Array<auth_auth_pb.Permission>): AuthInfo;
    addPermissions(value: auth_auth_pb.Permission, index?: number): auth_auth_pb.Permission;
    clearRolesList(): void;
    getRolesList(): Array<string>;
    setRolesList(value: Array<string>): AuthInfo;
    addRoles(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AuthInfo.AsObject;
    static toObject(includeInstance: boolean, msg: AuthInfo): AuthInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AuthInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AuthInfo;
    static deserializeBinaryFromReader(message: AuthInfo, reader: jspb.BinaryReader): AuthInfo;
}

export namespace AuthInfo {
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

    getMetadataMap(): jspb.Map<string, string>;
    clearMetadataMap(): void;

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

        metadataMap: Array<[string, string]>,
    }
}

export class Trigger extends jspb.Message { 
    getBranch(): string;
    setBranch(value: string): Trigger;
    getAll(): boolean;
    setAll(value: boolean): Trigger;
    getRateLimitSpec(): string;
    setRateLimitSpec(value: string): Trigger;
    getSize(): string;
    setSize(value: string): Trigger;
    getCommits(): number;
    setCommits(value: number): Trigger;
    getCronSpec(): string;
    setCronSpec(value: string): Trigger;

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
        rateLimitSpec: string,
        size: string,
        commits: number,
        cronSpec: string,
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

    hasBranch(): boolean;
    clearBranch(): void;
    getBranch(): Branch | undefined;
    setBranch(value?: Branch): Commit;

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
        branch?: Branch.AsObject,
    }
}

export class CommitPicker extends jspb.Message { 

    hasBranchHead(): boolean;
    clearBranchHead(): void;
    getBranchHead(): BranchPicker | undefined;
    setBranchHead(value?: BranchPicker): CommitPicker;

    hasId(): boolean;
    clearId(): void;
    getId(): CommitPicker.CommitByGlobalId | undefined;
    setId(value?: CommitPicker.CommitByGlobalId): CommitPicker;

    hasAncestor(): boolean;
    clearAncestor(): void;
    getAncestor(): CommitPicker.AncestorOf | undefined;
    setAncestor(value?: CommitPicker.AncestorOf): CommitPicker;

    hasBranchRoot(): boolean;
    clearBranchRoot(): void;
    getBranchRoot(): CommitPicker.BranchRoot | undefined;
    setBranchRoot(value?: CommitPicker.BranchRoot): CommitPicker;

    getPickerCase(): CommitPicker.PickerCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CommitPicker.AsObject;
    static toObject(includeInstance: boolean, msg: CommitPicker): CommitPicker.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CommitPicker, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CommitPicker;
    static deserializeBinaryFromReader(message: CommitPicker, reader: jspb.BinaryReader): CommitPicker;
}

export namespace CommitPicker {
    export type AsObject = {
        branchHead?: BranchPicker.AsObject,
        id?: CommitPicker.CommitByGlobalId.AsObject,
        ancestor?: CommitPicker.AncestorOf.AsObject,
        branchRoot?: CommitPicker.BranchRoot.AsObject,
    }


    export class CommitByGlobalId extends jspb.Message { 

        hasRepo(): boolean;
        clearRepo(): void;
        getRepo(): RepoPicker | undefined;
        setRepo(value?: RepoPicker): CommitByGlobalId;
        getId(): string;
        setId(value: string): CommitByGlobalId;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): CommitByGlobalId.AsObject;
        static toObject(includeInstance: boolean, msg: CommitByGlobalId): CommitByGlobalId.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: CommitByGlobalId, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): CommitByGlobalId;
        static deserializeBinaryFromReader(message: CommitByGlobalId, reader: jspb.BinaryReader): CommitByGlobalId;
    }

    export namespace CommitByGlobalId {
        export type AsObject = {
            repo?: RepoPicker.AsObject,
            id: string,
        }
    }

    export class BranchRoot extends jspb.Message { 
        getOffset(): number;
        setOffset(value: number): BranchRoot;

        hasBranch(): boolean;
        clearBranch(): void;
        getBranch(): BranchPicker | undefined;
        setBranch(value?: BranchPicker): BranchRoot;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): BranchRoot.AsObject;
        static toObject(includeInstance: boolean, msg: BranchRoot): BranchRoot.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: BranchRoot, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): BranchRoot;
        static deserializeBinaryFromReader(message: BranchRoot, reader: jspb.BinaryReader): BranchRoot;
    }

    export namespace BranchRoot {
        export type AsObject = {
            offset: number,
            branch?: BranchPicker.AsObject,
        }
    }

    export class AncestorOf extends jspb.Message { 
        getOffset(): number;
        setOffset(value: number): AncestorOf;

        hasStart(): boolean;
        clearStart(): void;
        getStart(): CommitPicker | undefined;
        setStart(value?: CommitPicker): AncestorOf;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): AncestorOf.AsObject;
        static toObject(includeInstance: boolean, msg: AncestorOf): AncestorOf.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: AncestorOf, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): AncestorOf;
        static deserializeBinaryFromReader(message: AncestorOf, reader: jspb.BinaryReader): AncestorOf;
    }

    export namespace AncestorOf {
        export type AsObject = {
            offset: number,
            start?: CommitPicker.AsObject,
        }
    }


    export enum PickerCase {
        PICKER_NOT_SET = 0,
        BRANCH_HEAD = 1,
        ID = 2,
        ANCESTOR = 3,
        BRANCH_ROOT = 4,
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

    hasFinishing(): boolean;
    clearFinishing(): void;
    getFinishing(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setFinishing(value?: google_protobuf_timestamp_pb.Timestamp): CommitInfo;

    hasFinished(): boolean;
    clearFinished(): void;
    getFinished(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setFinished(value?: google_protobuf_timestamp_pb.Timestamp): CommitInfo;
    clearDirectProvenanceList(): void;
    getDirectProvenanceList(): Array<Commit>;
    setDirectProvenanceList(value: Array<Commit>): CommitInfo;
    addDirectProvenance(value?: Commit, index?: number): Commit;
    clearDirectSubvenanceList(): void;
    getDirectSubvenanceList(): Array<Commit>;
    setDirectSubvenanceList(value: Array<Commit>): CommitInfo;
    addDirectSubvenance(value?: Commit, index?: number): Commit;
    getError(): string;
    setError(value: string): CommitInfo;
    getSizeBytesUpperBound(): number;
    setSizeBytesUpperBound(value: number): CommitInfo;

    hasDetails(): boolean;
    clearDetails(): void;
    getDetails(): CommitInfo.Details | undefined;
    setDetails(value?: CommitInfo.Details): CommitInfo;

    getMetadataMap(): jspb.Map<string, string>;
    clearMetadataMap(): void;

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
        finishing?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        finished?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        directProvenanceList: Array<Commit.AsObject>,
        directSubvenanceList: Array<Commit.AsObject>,
        error: string,
        sizeBytesUpperBound: number,
        details?: CommitInfo.Details.AsObject,

        metadataMap: Array<[string, string]>,
    }


    export class Details extends jspb.Message { 
        getSizeBytes(): number;
        setSizeBytes(value: number): Details;

        hasCompactingTime(): boolean;
        clearCompactingTime(): void;
        getCompactingTime(): google_protobuf_duration_pb.Duration | undefined;
        setCompactingTime(value?: google_protobuf_duration_pb.Duration): Details;

        hasValidatingTime(): boolean;
        clearValidatingTime(): void;
        getValidatingTime(): google_protobuf_duration_pb.Duration | undefined;
        setValidatingTime(value?: google_protobuf_duration_pb.Duration): Details;

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
            compactingTime?: google_protobuf_duration_pb.Duration.AsObject,
            validatingTime?: google_protobuf_duration_pb.Duration.AsObject,
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

export class Project extends jspb.Message { 
    getName(): string;
    setName(value: string): Project;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Project.AsObject;
    static toObject(includeInstance: boolean, msg: Project): Project.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Project, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Project;
    static deserializeBinaryFromReader(message: Project, reader: jspb.BinaryReader): Project;
}

export namespace Project {
    export type AsObject = {
        name: string,
    }
}

export class ProjectInfo extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): Project | undefined;
    setProject(value?: Project): ProjectInfo;
    getDescription(): string;
    setDescription(value: string): ProjectInfo;

    hasAuthInfo(): boolean;
    clearAuthInfo(): void;
    getAuthInfo(): AuthInfo | undefined;
    setAuthInfo(value?: AuthInfo): ProjectInfo;

    hasCreatedAt(): boolean;
    clearCreatedAt(): void;
    getCreatedAt(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setCreatedAt(value?: google_protobuf_timestamp_pb.Timestamp): ProjectInfo;

    getMetadataMap(): jspb.Map<string, string>;
    clearMetadataMap(): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProjectInfo.AsObject;
    static toObject(includeInstance: boolean, msg: ProjectInfo): ProjectInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProjectInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProjectInfo;
    static deserializeBinaryFromReader(message: ProjectInfo, reader: jspb.BinaryReader): ProjectInfo;
}

export namespace ProjectInfo {
    export type AsObject = {
        project?: Project.AsObject,
        description: string,
        authInfo?: AuthInfo.AsObject,
        createdAt?: google_protobuf_timestamp_pb.Timestamp.AsObject,

        metadataMap: Array<[string, string]>,
    }
}

export class ProjectPicker extends jspb.Message { 

    hasName(): boolean;
    clearName(): void;
    getName(): string;
    setName(value: string): ProjectPicker;

    getPickerCase(): ProjectPicker.PickerCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProjectPicker.AsObject;
    static toObject(includeInstance: boolean, msg: ProjectPicker): ProjectPicker.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProjectPicker, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProjectPicker;
    static deserializeBinaryFromReader(message: ProjectPicker, reader: jspb.BinaryReader): ProjectPicker;
}

export namespace ProjectPicker {
    export type AsObject = {
        name: string,
    }

    export enum PickerCase {
        PICKER_NOT_SET = 0,
        NAME = 1,
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
    clearProjectsList(): void;
    getProjectsList(): Array<Project>;
    setProjectsList(value: Array<Project>): ListRepoRequest;
    addProjects(value?: Project, index?: number): Project;

    hasPage(): boolean;
    clearPage(): void;
    getPage(): RepoPage | undefined;
    setPage(value?: RepoPage): ListRepoRequest;

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
        projectsList: Array<Project.AsObject>,
        page?: RepoPage.AsObject,
    }
}

export class RepoPage extends jspb.Message { 
    getOrder(): RepoPage.Ordering;
    setOrder(value: RepoPage.Ordering): RepoPage;
    getPageSize(): number;
    setPageSize(value: number): RepoPage;
    getPageIndex(): number;
    setPageIndex(value: number): RepoPage;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RepoPage.AsObject;
    static toObject(includeInstance: boolean, msg: RepoPage): RepoPage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RepoPage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RepoPage;
    static deserializeBinaryFromReader(message: RepoPage, reader: jspb.BinaryReader): RepoPage;
}

export namespace RepoPage {
    export type AsObject = {
        order: RepoPage.Ordering,
        pageSize: number,
        pageIndex: number,
    }

    export enum Ordering {
    PROJECT_REPO = 0,
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

export class DeleteReposRequest extends jspb.Message { 
    clearProjectsList(): void;
    getProjectsList(): Array<Project>;
    setProjectsList(value: Array<Project>): DeleteReposRequest;
    addProjects(value?: Project, index?: number): Project;
    getForce(): boolean;
    setForce(value: boolean): DeleteReposRequest;
    getAll(): boolean;
    setAll(value: boolean): DeleteReposRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteReposRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteReposRequest): DeleteReposRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteReposRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteReposRequest;
    static deserializeBinaryFromReader(message: DeleteReposRequest, reader: jspb.BinaryReader): DeleteReposRequest;
}

export namespace DeleteReposRequest {
    export type AsObject = {
        projectsList: Array<Project.AsObject>,
        force: boolean,
        all: boolean,
    }
}

export class DeleteRepoResponse extends jspb.Message { 
    getDeleted(): boolean;
    setDeleted(value: boolean): DeleteRepoResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteRepoResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteRepoResponse): DeleteRepoResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteRepoResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteRepoResponse;
    static deserializeBinaryFromReader(message: DeleteRepoResponse, reader: jspb.BinaryReader): DeleteRepoResponse;
}

export namespace DeleteRepoResponse {
    export type AsObject = {
        deleted: boolean,
    }
}

export class DeleteReposResponse extends jspb.Message { 
    clearReposList(): void;
    getReposList(): Array<Repo>;
    setReposList(value: Array<Repo>): DeleteReposResponse;
    addRepos(value?: Repo, index?: number): Repo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteReposResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteReposResponse): DeleteReposResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteReposResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteReposResponse;
    static deserializeBinaryFromReader(message: DeleteReposResponse, reader: jspb.BinaryReader): DeleteReposResponse;
}

export namespace DeleteReposResponse {
    export type AsObject = {
        reposList: Array<Repo.AsObject>,
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
    getError(): string;
    setError(value: string): FinishCommitRequest;
    getForce(): boolean;
    setForce(value: boolean): FinishCommitRequest;

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
        error: string,
        force: boolean,
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
    getAll(): boolean;
    setAll(value: boolean): ListCommitRequest;
    getOriginKind(): OriginKind;
    setOriginKind(value: OriginKind): ListCommitRequest;

    hasStartedTime(): boolean;
    clearStartedTime(): void;
    getStartedTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setStartedTime(value?: google_protobuf_timestamp_pb.Timestamp): ListCommitRequest;

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
        all: boolean,
        originKind: OriginKind,
        startedTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
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

    hasProject(): boolean;
    clearProject(): void;
    getProject(): Project | undefined;
    setProject(value?: Project): ListCommitSetRequest;

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
        project?: Project.AsObject,
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

export class DropCommitSetRequest extends jspb.Message { 

    hasCommitSet(): boolean;
    clearCommitSet(): void;
    getCommitSet(): CommitSet | undefined;
    setCommitSet(value?: CommitSet): DropCommitSetRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DropCommitSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DropCommitSetRequest): DropCommitSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DropCommitSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DropCommitSetRequest;
    static deserializeBinaryFromReader(message: DropCommitSetRequest, reader: jspb.BinaryReader): DropCommitSetRequest;
}

export namespace DropCommitSetRequest {
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
    getAll(): boolean;
    setAll(value: boolean): SubscribeCommitRequest;
    getOriginKind(): OriginKind;
    setOriginKind(value: OriginKind): SubscribeCommitRequest;

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
        all: boolean,
        originKind: OriginKind,
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

export class SquashCommitRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): SquashCommitRequest;
    getRecursive(): boolean;
    setRecursive(value: boolean): SquashCommitRequest;

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
        recursive: boolean,
    }
}

export class SquashCommitResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SquashCommitResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SquashCommitResponse): SquashCommitResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SquashCommitResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SquashCommitResponse;
    static deserializeBinaryFromReader(message: SquashCommitResponse, reader: jspb.BinaryReader): SquashCommitResponse;
}

export namespace SquashCommitResponse {
    export type AsObject = {
    }
}

export class DropCommitRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): DropCommitRequest;
    getRecursive(): boolean;
    setRecursive(value: boolean): DropCommitRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DropCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DropCommitRequest): DropCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DropCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DropCommitRequest;
    static deserializeBinaryFromReader(message: DropCommitRequest, reader: jspb.BinaryReader): DropCommitRequest;
}

export namespace DropCommitRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
        recursive: boolean,
    }
}

export class WalkCommitProvenanceRequest extends jspb.Message { 
    clearStartList(): void;
    getStartList(): Array<CommitPicker>;
    setStartList(value: Array<CommitPicker>): WalkCommitProvenanceRequest;
    addStart(value?: CommitPicker, index?: number): CommitPicker;
    getMaxCommits(): number;
    setMaxCommits(value: number): WalkCommitProvenanceRequest;
    getMaxDepth(): number;
    setMaxDepth(value: number): WalkCommitProvenanceRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WalkCommitProvenanceRequest.AsObject;
    static toObject(includeInstance: boolean, msg: WalkCommitProvenanceRequest): WalkCommitProvenanceRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WalkCommitProvenanceRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WalkCommitProvenanceRequest;
    static deserializeBinaryFromReader(message: WalkCommitProvenanceRequest, reader: jspb.BinaryReader): WalkCommitProvenanceRequest;
}

export namespace WalkCommitProvenanceRequest {
    export type AsObject = {
        startList: Array<CommitPicker.AsObject>,
        maxCommits: number,
        maxDepth: number,
    }
}

export class WalkCommitSubvenanceRequest extends jspb.Message { 
    clearStartList(): void;
    getStartList(): Array<CommitPicker>;
    setStartList(value: Array<CommitPicker>): WalkCommitSubvenanceRequest;
    addStart(value?: CommitPicker, index?: number): CommitPicker;
    getMaxCommits(): number;
    setMaxCommits(value: number): WalkCommitSubvenanceRequest;
    getMaxDepth(): number;
    setMaxDepth(value: number): WalkCommitSubvenanceRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WalkCommitSubvenanceRequest.AsObject;
    static toObject(includeInstance: boolean, msg: WalkCommitSubvenanceRequest): WalkCommitSubvenanceRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WalkCommitSubvenanceRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WalkCommitSubvenanceRequest;
    static deserializeBinaryFromReader(message: WalkCommitSubvenanceRequest, reader: jspb.BinaryReader): WalkCommitSubvenanceRequest;
}

export namespace WalkCommitSubvenanceRequest {
    export type AsObject = {
        startList: Array<CommitPicker.AsObject>,
        maxCommits: number,
        maxDepth: number,
    }
}

export class WalkBranchProvenanceRequest extends jspb.Message { 
    clearStartList(): void;
    getStartList(): Array<BranchPicker>;
    setStartList(value: Array<BranchPicker>): WalkBranchProvenanceRequest;
    addStart(value?: BranchPicker, index?: number): BranchPicker;
    getMaxBranches(): number;
    setMaxBranches(value: number): WalkBranchProvenanceRequest;
    getMaxDepth(): number;
    setMaxDepth(value: number): WalkBranchProvenanceRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WalkBranchProvenanceRequest.AsObject;
    static toObject(includeInstance: boolean, msg: WalkBranchProvenanceRequest): WalkBranchProvenanceRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WalkBranchProvenanceRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WalkBranchProvenanceRequest;
    static deserializeBinaryFromReader(message: WalkBranchProvenanceRequest, reader: jspb.BinaryReader): WalkBranchProvenanceRequest;
}

export namespace WalkBranchProvenanceRequest {
    export type AsObject = {
        startList: Array<BranchPicker.AsObject>,
        maxBranches: number,
        maxDepth: number,
    }
}

export class WalkBranchSubvenanceRequest extends jspb.Message { 
    clearStartList(): void;
    getStartList(): Array<BranchPicker>;
    setStartList(value: Array<BranchPicker>): WalkBranchSubvenanceRequest;
    addStart(value?: BranchPicker, index?: number): BranchPicker;
    getMaxBranches(): number;
    setMaxBranches(value: number): WalkBranchSubvenanceRequest;
    getMaxDepth(): number;
    setMaxDepth(value: number): WalkBranchSubvenanceRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WalkBranchSubvenanceRequest.AsObject;
    static toObject(includeInstance: boolean, msg: WalkBranchSubvenanceRequest): WalkBranchSubvenanceRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WalkBranchSubvenanceRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WalkBranchSubvenanceRequest;
    static deserializeBinaryFromReader(message: WalkBranchSubvenanceRequest, reader: jspb.BinaryReader): WalkBranchSubvenanceRequest;
}

export namespace WalkBranchSubvenanceRequest {
    export type AsObject = {
        startList: Array<BranchPicker.AsObject>,
        maxBranches: number,
        maxDepth: number,
    }
}

export class DropCommitResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DropCommitResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DropCommitResponse): DropCommitResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DropCommitResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DropCommitResponse;
    static deserializeBinaryFromReader(message: DropCommitResponse, reader: jspb.BinaryReader): DropCommitResponse;
}

export namespace DropCommitResponse {
    export type AsObject = {
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

export class FindCommitsRequest extends jspb.Message { 

    hasStart(): boolean;
    clearStart(): void;
    getStart(): Commit | undefined;
    setStart(value?: Commit): FindCommitsRequest;
    getFilePath(): string;
    setFilePath(value: string): FindCommitsRequest;
    getLimit(): number;
    setLimit(value: number): FindCommitsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FindCommitsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: FindCommitsRequest): FindCommitsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FindCommitsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FindCommitsRequest;
    static deserializeBinaryFromReader(message: FindCommitsRequest, reader: jspb.BinaryReader): FindCommitsRequest;
}

export namespace FindCommitsRequest {
    export type AsObject = {
        start?: Commit.AsObject,
        filePath: string,
        limit: number,
    }
}

export class FindCommitsResponse extends jspb.Message { 

    hasFoundCommit(): boolean;
    clearFoundCommit(): void;
    getFoundCommit(): Commit | undefined;
    setFoundCommit(value?: Commit): FindCommitsResponse;

    hasLastSearchedCommit(): boolean;
    clearLastSearchedCommit(): void;
    getLastSearchedCommit(): Commit | undefined;
    setLastSearchedCommit(value?: Commit): FindCommitsResponse;
    getCommitsSearched(): number;
    setCommitsSearched(value: number): FindCommitsResponse;

    getResultCase(): FindCommitsResponse.ResultCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FindCommitsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: FindCommitsResponse): FindCommitsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FindCommitsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FindCommitsResponse;
    static deserializeBinaryFromReader(message: FindCommitsResponse, reader: jspb.BinaryReader): FindCommitsResponse;
}

export namespace FindCommitsResponse {
    export type AsObject = {
        foundCommit?: Commit.AsObject,
        lastSearchedCommit?: Commit.AsObject,
        commitsSearched: number,
    }

    export enum ResultCase {
        RESULT_NOT_SET = 0,
        FOUND_COMMIT = 1,
        LAST_SEARCHED_COMMIT = 2,
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

export class CreateProjectRequest extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): Project | undefined;
    setProject(value?: Project): CreateProjectRequest;
    getDescription(): string;
    setDescription(value: string): CreateProjectRequest;
    getUpdate(): boolean;
    setUpdate(value: boolean): CreateProjectRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateProjectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateProjectRequest): CreateProjectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateProjectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateProjectRequest;
    static deserializeBinaryFromReader(message: CreateProjectRequest, reader: jspb.BinaryReader): CreateProjectRequest;
}

export namespace CreateProjectRequest {
    export type AsObject = {
        project?: Project.AsObject,
        description: string,
        update: boolean,
    }
}

export class InspectProjectRequest extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): Project | undefined;
    setProject(value?: Project): InspectProjectRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectProjectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectProjectRequest): InspectProjectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectProjectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectProjectRequest;
    static deserializeBinaryFromReader(message: InspectProjectRequest, reader: jspb.BinaryReader): InspectProjectRequest;
}

export namespace InspectProjectRequest {
    export type AsObject = {
        project?: Project.AsObject,
    }
}

export class InspectProjectV2Request extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): Project | undefined;
    setProject(value?: Project): InspectProjectV2Request;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectProjectV2Request.AsObject;
    static toObject(includeInstance: boolean, msg: InspectProjectV2Request): InspectProjectV2Request.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectProjectV2Request, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectProjectV2Request;
    static deserializeBinaryFromReader(message: InspectProjectV2Request, reader: jspb.BinaryReader): InspectProjectV2Request;
}

export namespace InspectProjectV2Request {
    export type AsObject = {
        project?: Project.AsObject,
    }
}

export class InspectProjectV2Response extends jspb.Message { 

    hasInfo(): boolean;
    clearInfo(): void;
    getInfo(): ProjectInfo | undefined;
    setInfo(value?: ProjectInfo): InspectProjectV2Response;
    getDefaultsJson(): string;
    setDefaultsJson(value: string): InspectProjectV2Response;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectProjectV2Response.AsObject;
    static toObject(includeInstance: boolean, msg: InspectProjectV2Response): InspectProjectV2Response.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectProjectV2Response, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectProjectV2Response;
    static deserializeBinaryFromReader(message: InspectProjectV2Response, reader: jspb.BinaryReader): InspectProjectV2Response;
}

export namespace InspectProjectV2Response {
    export type AsObject = {
        info?: ProjectInfo.AsObject,
        defaultsJson: string,
    }
}

export class ListProjectRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListProjectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListProjectRequest): ListProjectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListProjectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListProjectRequest;
    static deserializeBinaryFromReader(message: ListProjectRequest, reader: jspb.BinaryReader): ListProjectRequest;
}

export namespace ListProjectRequest {
    export type AsObject = {
    }
}

export class DeleteProjectRequest extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): Project | undefined;
    setProject(value?: Project): DeleteProjectRequest;
    getForce(): boolean;
    setForce(value: boolean): DeleteProjectRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteProjectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteProjectRequest): DeleteProjectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteProjectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteProjectRequest;
    static deserializeBinaryFromReader(message: DeleteProjectRequest, reader: jspb.BinaryReader): DeleteProjectRequest;
}

export namespace DeleteProjectRequest {
    export type AsObject = {
        project?: Project.AsObject,
        force: boolean,
    }
}

export class AddFile extends jspb.Message { 
    getPath(): string;
    setPath(value: string): AddFile;
    getDatum(): string;
    setDatum(value: string): AddFile;

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
        datum: string,
        raw?: google_protobuf_wrappers_pb.BytesValue.AsObject,
        url?: AddFile.URLSource.AsObject,
    }


    export class URLSource extends jspb.Message { 
        getUrl(): string;
        setUrl(value: string): URLSource;
        getRecursive(): boolean;
        setRecursive(value: boolean): URLSource;
        getConcurrency(): number;
        setConcurrency(value: number): URLSource;

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
            concurrency: number,
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
    getDatum(): string;
    setDatum(value: string): DeleteFile;

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
        datum: string,
    }
}

export class CopyFile extends jspb.Message { 
    getDst(): string;
    setDst(value: string): CopyFile;
    getDatum(): string;
    setDatum(value: string): CopyFile;

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
        datum: string,
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
    getOffset(): number;
    setOffset(value: number): GetFileRequest;

    hasPathRange(): boolean;
    clearPathRange(): void;
    getPathRange(): PathRange | undefined;
    setPathRange(value?: PathRange): GetFileRequest;

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
        offset: number,
        pathRange?: PathRange.AsObject,
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

    hasPaginationmarker(): boolean;
    clearPaginationmarker(): void;
    getPaginationmarker(): File | undefined;
    setPaginationmarker(value?: File): ListFileRequest;
    getNumber(): number;
    setNumber(value: number): ListFileRequest;
    getReverse(): boolean;
    setReverse(value: boolean): ListFileRequest;

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
        paginationmarker?: File.AsObject,
        number: number,
        reverse: boolean,
    }
}

export class WalkFileRequest extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): WalkFileRequest;

    hasPaginationmarker(): boolean;
    clearPaginationmarker(): void;
    getPaginationmarker(): File | undefined;
    setPaginationmarker(value?: File): WalkFileRequest;
    getNumber(): number;
    setNumber(value: number): WalkFileRequest;
    getReverse(): boolean;
    setReverse(value: boolean): WalkFileRequest;

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
        paginationmarker?: File.AsObject,
        number: number,
        reverse: boolean,
    }
}

export class GlobFileRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): GlobFileRequest;
    getPattern(): string;
    setPattern(value: string): GlobFileRequest;

    hasPathRange(): boolean;
    clearPathRange(): void;
    getPathRange(): PathRange | undefined;
    setPathRange(value?: PathRange): GlobFileRequest;

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
        pathRange?: PathRange.AsObject,
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

    hasZombieTarget(): boolean;
    clearZombieTarget(): void;
    getZombieTarget(): Commit | undefined;
    setZombieTarget(value?: Commit): FsckRequest;

    hasZombieAll(): boolean;
    clearZombieAll(): void;
    getZombieAll(): boolean;
    setZombieAll(value: boolean): FsckRequest;

    getZombieCheckCase(): FsckRequest.ZombieCheckCase;

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
        zombieTarget?: Commit.AsObject,
        zombieAll: boolean,
    }

    export enum ZombieCheckCase {
        ZOMBIE_CHECK_NOT_SET = 0,
        ZOMBIE_TARGET = 2,
        ZOMBIE_ALL = 3,
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
    getType(): GetFileSetRequest.FileSetType;
    setType(value: GetFileSetRequest.FileSetType): GetFileSetRequest;

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
        type: GetFileSetRequest.FileSetType,
    }

    export enum FileSetType {
    TOTAL = 0,
    DIFF = 1,
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

export class ComposeFileSetRequest extends jspb.Message { 
    clearFileSetIdsList(): void;
    getFileSetIdsList(): Array<string>;
    setFileSetIdsList(value: Array<string>): ComposeFileSetRequest;
    addFileSetIds(value: string, index?: number): string;
    getTtlSeconds(): number;
    setTtlSeconds(value: number): ComposeFileSetRequest;
    getCompact(): boolean;
    setCompact(value: boolean): ComposeFileSetRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ComposeFileSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ComposeFileSetRequest): ComposeFileSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ComposeFileSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ComposeFileSetRequest;
    static deserializeBinaryFromReader(message: ComposeFileSetRequest, reader: jspb.BinaryReader): ComposeFileSetRequest;
}

export namespace ComposeFileSetRequest {
    export type AsObject = {
        fileSetIdsList: Array<string>,
        ttlSeconds: number,
        compact: boolean,
    }
}

export class ShardFileSetRequest extends jspb.Message { 
    getFileSetId(): string;
    setFileSetId(value: string): ShardFileSetRequest;
    getNumFiles(): number;
    setNumFiles(value: number): ShardFileSetRequest;
    getSizeBytes(): number;
    setSizeBytes(value: number): ShardFileSetRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ShardFileSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ShardFileSetRequest): ShardFileSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ShardFileSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ShardFileSetRequest;
    static deserializeBinaryFromReader(message: ShardFileSetRequest, reader: jspb.BinaryReader): ShardFileSetRequest;
}

export namespace ShardFileSetRequest {
    export type AsObject = {
        fileSetId: string,
        numFiles: number,
        sizeBytes: number,
    }
}

export class PathRange extends jspb.Message { 
    getLower(): string;
    setLower(value: string): PathRange;
    getUpper(): string;
    setUpper(value: string): PathRange;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PathRange.AsObject;
    static toObject(includeInstance: boolean, msg: PathRange): PathRange.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PathRange, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PathRange;
    static deserializeBinaryFromReader(message: PathRange, reader: jspb.BinaryReader): PathRange;
}

export namespace PathRange {
    export type AsObject = {
        lower: string,
        upper: string,
    }
}

export class ShardFileSetResponse extends jspb.Message { 
    clearShardsList(): void;
    getShardsList(): Array<PathRange>;
    setShardsList(value: Array<PathRange>): ShardFileSetResponse;
    addShards(value?: PathRange, index?: number): PathRange;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ShardFileSetResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ShardFileSetResponse): ShardFileSetResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ShardFileSetResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ShardFileSetResponse;
    static deserializeBinaryFromReader(message: ShardFileSetResponse, reader: jspb.BinaryReader): ShardFileSetResponse;
}

export namespace ShardFileSetResponse {
    export type AsObject = {
        shardsList: Array<PathRange.AsObject>,
    }
}

export class CheckStorageRequest extends jspb.Message { 
    getReadChunkData(): boolean;
    setReadChunkData(value: boolean): CheckStorageRequest;
    getChunkBegin(): Uint8Array | string;
    getChunkBegin_asU8(): Uint8Array;
    getChunkBegin_asB64(): string;
    setChunkBegin(value: Uint8Array | string): CheckStorageRequest;
    getChunkEnd(): Uint8Array | string;
    getChunkEnd_asU8(): Uint8Array;
    getChunkEnd_asB64(): string;
    setChunkEnd(value: Uint8Array | string): CheckStorageRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CheckStorageRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CheckStorageRequest): CheckStorageRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CheckStorageRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CheckStorageRequest;
    static deserializeBinaryFromReader(message: CheckStorageRequest, reader: jspb.BinaryReader): CheckStorageRequest;
}

export namespace CheckStorageRequest {
    export type AsObject = {
        readChunkData: boolean,
        chunkBegin: Uint8Array | string,
        chunkEnd: Uint8Array | string,
    }
}

export class CheckStorageResponse extends jspb.Message { 
    getChunkObjectCount(): number;
    setChunkObjectCount(value: number): CheckStorageResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CheckStorageResponse.AsObject;
    static toObject(includeInstance: boolean, msg: CheckStorageResponse): CheckStorageResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CheckStorageResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CheckStorageResponse;
    static deserializeBinaryFromReader(message: CheckStorageResponse, reader: jspb.BinaryReader): CheckStorageResponse;
}

export namespace CheckStorageResponse {
    export type AsObject = {
        chunkObjectCount: number,
    }
}

export class PutCacheRequest extends jspb.Message { 
    getKey(): string;
    setKey(value: string): PutCacheRequest;

    hasValue(): boolean;
    clearValue(): void;
    getValue(): google_protobuf_any_pb.Any | undefined;
    setValue(value?: google_protobuf_any_pb.Any): PutCacheRequest;
    clearFileSetIdsList(): void;
    getFileSetIdsList(): Array<string>;
    setFileSetIdsList(value: Array<string>): PutCacheRequest;
    addFileSetIds(value: string, index?: number): string;
    getTag(): string;
    setTag(value: string): PutCacheRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PutCacheRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PutCacheRequest): PutCacheRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PutCacheRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PutCacheRequest;
    static deserializeBinaryFromReader(message: PutCacheRequest, reader: jspb.BinaryReader): PutCacheRequest;
}

export namespace PutCacheRequest {
    export type AsObject = {
        key: string,
        value?: google_protobuf_any_pb.Any.AsObject,
        fileSetIdsList: Array<string>,
        tag: string,
    }
}

export class GetCacheRequest extends jspb.Message { 
    getKey(): string;
    setKey(value: string): GetCacheRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetCacheRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetCacheRequest): GetCacheRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetCacheRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetCacheRequest;
    static deserializeBinaryFromReader(message: GetCacheRequest, reader: jspb.BinaryReader): GetCacheRequest;
}

export namespace GetCacheRequest {
    export type AsObject = {
        key: string,
    }
}

export class GetCacheResponse extends jspb.Message { 

    hasValue(): boolean;
    clearValue(): void;
    getValue(): google_protobuf_any_pb.Any | undefined;
    setValue(value?: google_protobuf_any_pb.Any): GetCacheResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetCacheResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetCacheResponse): GetCacheResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetCacheResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetCacheResponse;
    static deserializeBinaryFromReader(message: GetCacheResponse, reader: jspb.BinaryReader): GetCacheResponse;
}

export namespace GetCacheResponse {
    export type AsObject = {
        value?: google_protobuf_any_pb.Any.AsObject,
    }
}

export class ClearCacheRequest extends jspb.Message { 
    getTagPrefix(): string;
    setTagPrefix(value: string): ClearCacheRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ClearCacheRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ClearCacheRequest): ClearCacheRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ClearCacheRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ClearCacheRequest;
    static deserializeBinaryFromReader(message: ClearCacheRequest, reader: jspb.BinaryReader): ClearCacheRequest;
}

export namespace ClearCacheRequest {
    export type AsObject = {
        tagPrefix: string,
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

export class ObjectStorageEgress extends jspb.Message { 
    getUrl(): string;
    setUrl(value: string): ObjectStorageEgress;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ObjectStorageEgress.AsObject;
    static toObject(includeInstance: boolean, msg: ObjectStorageEgress): ObjectStorageEgress.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ObjectStorageEgress, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ObjectStorageEgress;
    static deserializeBinaryFromReader(message: ObjectStorageEgress, reader: jspb.BinaryReader): ObjectStorageEgress;
}

export namespace ObjectStorageEgress {
    export type AsObject = {
        url: string,
    }
}

export class SQLDatabaseEgress extends jspb.Message { 
    getUrl(): string;
    setUrl(value: string): SQLDatabaseEgress;

    hasFileFormat(): boolean;
    clearFileFormat(): void;
    getFileFormat(): SQLDatabaseEgress.FileFormat | undefined;
    setFileFormat(value?: SQLDatabaseEgress.FileFormat): SQLDatabaseEgress;

    hasSecret(): boolean;
    clearSecret(): void;
    getSecret(): SQLDatabaseEgress.Secret | undefined;
    setSecret(value?: SQLDatabaseEgress.Secret): SQLDatabaseEgress;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SQLDatabaseEgress.AsObject;
    static toObject(includeInstance: boolean, msg: SQLDatabaseEgress): SQLDatabaseEgress.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SQLDatabaseEgress, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SQLDatabaseEgress;
    static deserializeBinaryFromReader(message: SQLDatabaseEgress, reader: jspb.BinaryReader): SQLDatabaseEgress;
}

export namespace SQLDatabaseEgress {
    export type AsObject = {
        url: string,
        fileFormat?: SQLDatabaseEgress.FileFormat.AsObject,
        secret?: SQLDatabaseEgress.Secret.AsObject,
    }


    export class FileFormat extends jspb.Message { 
        getType(): SQLDatabaseEgress.FileFormat.Type;
        setType(value: SQLDatabaseEgress.FileFormat.Type): FileFormat;
        clearColumnsList(): void;
        getColumnsList(): Array<string>;
        setColumnsList(value: Array<string>): FileFormat;
        addColumns(value: string, index?: number): string;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): FileFormat.AsObject;
        static toObject(includeInstance: boolean, msg: FileFormat): FileFormat.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: FileFormat, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): FileFormat;
        static deserializeBinaryFromReader(message: FileFormat, reader: jspb.BinaryReader): FileFormat;
    }

    export namespace FileFormat {
        export type AsObject = {
            type: SQLDatabaseEgress.FileFormat.Type,
            columnsList: Array<string>,
        }

        export enum Type {
    UNKNOWN = 0,
    CSV = 1,
    JSON = 2,
    PARQUET = 3,
        }

    }

    export class Secret extends jspb.Message { 
        getName(): string;
        setName(value: string): Secret;
        getKey(): string;
        setKey(value: string): Secret;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): Secret.AsObject;
        static toObject(includeInstance: boolean, msg: Secret): Secret.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: Secret, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): Secret;
        static deserializeBinaryFromReader(message: Secret, reader: jspb.BinaryReader): Secret;
    }

    export namespace Secret {
        export type AsObject = {
            name: string,
            key: string,
        }
    }

}

export class EgressRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): EgressRequest;

    hasObjectStorage(): boolean;
    clearObjectStorage(): void;
    getObjectStorage(): ObjectStorageEgress | undefined;
    setObjectStorage(value?: ObjectStorageEgress): EgressRequest;

    hasSqlDatabase(): boolean;
    clearSqlDatabase(): void;
    getSqlDatabase(): SQLDatabaseEgress | undefined;
    setSqlDatabase(value?: SQLDatabaseEgress): EgressRequest;

    getTargetCase(): EgressRequest.TargetCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EgressRequest.AsObject;
    static toObject(includeInstance: boolean, msg: EgressRequest): EgressRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EgressRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EgressRequest;
    static deserializeBinaryFromReader(message: EgressRequest, reader: jspb.BinaryReader): EgressRequest;
}

export namespace EgressRequest {
    export type AsObject = {
        commit?: Commit.AsObject,
        objectStorage?: ObjectStorageEgress.AsObject,
        sqlDatabase?: SQLDatabaseEgress.AsObject,
    }

    export enum TargetCase {
        TARGET_NOT_SET = 0,
        OBJECT_STORAGE = 2,
        SQL_DATABASE = 3,
    }

}

export class EgressResponse extends jspb.Message { 

    hasObjectStorage(): boolean;
    clearObjectStorage(): void;
    getObjectStorage(): EgressResponse.ObjectStorageResult | undefined;
    setObjectStorage(value?: EgressResponse.ObjectStorageResult): EgressResponse;

    hasSqlDatabase(): boolean;
    clearSqlDatabase(): void;
    getSqlDatabase(): EgressResponse.SQLDatabaseResult | undefined;
    setSqlDatabase(value?: EgressResponse.SQLDatabaseResult): EgressResponse;

    getResultCase(): EgressResponse.ResultCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EgressResponse.AsObject;
    static toObject(includeInstance: boolean, msg: EgressResponse): EgressResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EgressResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EgressResponse;
    static deserializeBinaryFromReader(message: EgressResponse, reader: jspb.BinaryReader): EgressResponse;
}

export namespace EgressResponse {
    export type AsObject = {
        objectStorage?: EgressResponse.ObjectStorageResult.AsObject,
        sqlDatabase?: EgressResponse.SQLDatabaseResult.AsObject,
    }


    export class ObjectStorageResult extends jspb.Message { 
        getBytesWritten(): number;
        setBytesWritten(value: number): ObjectStorageResult;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): ObjectStorageResult.AsObject;
        static toObject(includeInstance: boolean, msg: ObjectStorageResult): ObjectStorageResult.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: ObjectStorageResult, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): ObjectStorageResult;
        static deserializeBinaryFromReader(message: ObjectStorageResult, reader: jspb.BinaryReader): ObjectStorageResult;
    }

    export namespace ObjectStorageResult {
        export type AsObject = {
            bytesWritten: number,
        }
    }

    export class SQLDatabaseResult extends jspb.Message { 

        getRowsWrittenMap(): jspb.Map<string, number>;
        clearRowsWrittenMap(): void;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): SQLDatabaseResult.AsObject;
        static toObject(includeInstance: boolean, msg: SQLDatabaseResult): SQLDatabaseResult.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: SQLDatabaseResult, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): SQLDatabaseResult;
        static deserializeBinaryFromReader(message: SQLDatabaseResult, reader: jspb.BinaryReader): SQLDatabaseResult;
    }

    export namespace SQLDatabaseResult {
        export type AsObject = {

            rowsWrittenMap: Array<[string, number]>,
        }
    }


    export enum ResultCase {
        RESULT_NOT_SET = 0,
        OBJECT_STORAGE = 1,
        SQL_DATABASE = 2,
    }

}

export class ReposSummaryRequest extends jspb.Message { 
    clearProjectsList(): void;
    getProjectsList(): Array<ProjectPicker>;
    setProjectsList(value: Array<ProjectPicker>): ReposSummaryRequest;
    addProjects(value?: ProjectPicker, index?: number): ProjectPicker;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ReposSummaryRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ReposSummaryRequest): ReposSummaryRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ReposSummaryRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ReposSummaryRequest;
    static deserializeBinaryFromReader(message: ReposSummaryRequest, reader: jspb.BinaryReader): ReposSummaryRequest;
}

export namespace ReposSummaryRequest {
    export type AsObject = {
        projectsList: Array<ProjectPicker.AsObject>,
    }
}

export class ReposSummary extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): Project | undefined;
    setProject(value?: Project): ReposSummary;
    getUserRepoCount(): number;
    setUserRepoCount(value: number): ReposSummary;
    getSizeBytes(): number;
    setSizeBytes(value: number): ReposSummary;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ReposSummary.AsObject;
    static toObject(includeInstance: boolean, msg: ReposSummary): ReposSummary.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ReposSummary, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ReposSummary;
    static deserializeBinaryFromReader(message: ReposSummary, reader: jspb.BinaryReader): ReposSummary;
}

export namespace ReposSummary {
    export type AsObject = {
        project?: Project.AsObject,
        userRepoCount: number,
        sizeBytes: number,
    }
}

export class ReposSummaryResponse extends jspb.Message { 
    clearSummariesList(): void;
    getSummariesList(): Array<ReposSummary>;
    setSummariesList(value: Array<ReposSummary>): ReposSummaryResponse;
    addSummaries(value?: ReposSummary, index?: number): ReposSummary;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ReposSummaryResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ReposSummaryResponse): ReposSummaryResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ReposSummaryResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ReposSummaryResponse;
    static deserializeBinaryFromReader(message: ReposSummaryResponse, reader: jspb.BinaryReader): ReposSummaryResponse;
}

export namespace ReposSummaryResponse {
    export type AsObject = {
        summariesList: Array<ReposSummary.AsObject>,
    }
}

export enum OriginKind {
    ORIGIN_KIND_UNKNOWN = 0,
    USER = 1,
    AUTO = 2,
    FSCK = 3,
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
    FINISHING = 3,
    FINISHED = 4,
}

export enum Delimiter {
    NONE = 0,
    JSON = 1,
    LINE = 2,
    SQL = 3,
    CSV = 4,
}
