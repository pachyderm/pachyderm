// package: pfs
// file: client/pfs/pfs.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as google_protobuf_wrappers_pb from "google-protobuf/google/protobuf/wrappers_pb";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";
import * as client_auth_auth_pb from "../../client/auth/auth_pb";

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

export class Block extends jspb.Message { 
    getHash(): string;
    setHash(value: string): Block;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Block.AsObject;
    static toObject(includeInstance: boolean, msg: Block): Block.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Block, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Block;
    static deserializeBinaryFromReader(message: Block, reader: jspb.BinaryReader): Block;
}

export namespace Block {
    export type AsObject = {
        hash: string,
    }
}

export class Object extends jspb.Message { 
    getHash(): string;
    setHash(value: string): Object;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Object.AsObject;
    static toObject(includeInstance: boolean, msg: Object): Object.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Object, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Object;
    static deserializeBinaryFromReader(message: Object, reader: jspb.BinaryReader): Object;
}

export namespace Object {
    export type AsObject = {
        hash: string,
    }
}

export class Tag extends jspb.Message { 
    getName(): string;
    setName(value: string): Tag;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Tag.AsObject;
    static toObject(includeInstance: boolean, msg: Tag): Tag.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Tag, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Tag;
    static deserializeBinaryFromReader(message: Tag, reader: jspb.BinaryReader): Tag;
}

export namespace Tag {
    export type AsObject = {
        name: string,
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

    getTombstone(): boolean;
    setTombstone(value: boolean): RepoInfo;


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
        tombstone: boolean,
    }
}

export class RepoAuthInfo extends jspb.Message { 
    getAccessLevel(): client_auth_auth_pb.Scope;
    setAccessLevel(value: client_auth_auth_pb.Scope): RepoAuthInfo;


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
        accessLevel: client_auth_auth_pb.Scope,
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


    hasTree(): boolean;
    clearTree(): void;
    getTree(): Object | undefined;
    setTree(value?: Object): CommitInfo;

    clearTreesList(): void;
    getTreesList(): Array<Object>;
    setTreesList(value: Array<Object>): CommitInfo;
    addTrees(value?: Object, index?: number): Object;


    hasDatums(): boolean;
    clearDatums(): void;
    getDatums(): Object | undefined;
    setDatums(value?: Object): CommitInfo;

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
        tree?: Object.AsObject,
        treesList: Array<Object.AsObject>,
        datums?: Object.AsObject,
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

    clearChildrenList(): void;
    getChildrenList(): Array<string>;
    setChildrenList(value: Array<string>): FileInfo;
    addChildren(value: string, index?: number): string;

    clearObjectsList(): void;
    getObjectsList(): Array<Object>;
    setObjectsList(value: Array<Object>): FileInfo;
    addObjects(value?: Object, index?: number): Object;

    clearBlockrefsList(): void;
    getBlockrefsList(): Array<BlockRef>;
    setBlockrefsList(value: Array<BlockRef>): FileInfo;
    addBlockrefs(value?: BlockRef, index?: number): BlockRef;

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
        childrenList: Array<string>,
        objectsList: Array<Object.AsObject>,
        blockrefsList: Array<BlockRef.AsObject>,
        hash: Uint8Array | string,
    }
}

export class ByteRange extends jspb.Message { 
    getLower(): number;
    setLower(value: number): ByteRange;

    getUpper(): number;
    setUpper(value: number): ByteRange;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ByteRange.AsObject;
    static toObject(includeInstance: boolean, msg: ByteRange): ByteRange.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ByteRange, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ByteRange;
    static deserializeBinaryFromReader(message: ByteRange, reader: jspb.BinaryReader): ByteRange;
}

export namespace ByteRange {
    export type AsObject = {
        lower: number,
        upper: number,
    }
}

export class BlockRef extends jspb.Message { 

    hasBlock(): boolean;
    clearBlock(): void;
    getBlock(): Block | undefined;
    setBlock(value?: Block): BlockRef;


    hasRange(): boolean;
    clearRange(): void;
    getRange(): ByteRange | undefined;
    setRange(value?: ByteRange): BlockRef;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BlockRef.AsObject;
    static toObject(includeInstance: boolean, msg: BlockRef): BlockRef.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BlockRef, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BlockRef;
    static deserializeBinaryFromReader(message: BlockRef, reader: jspb.BinaryReader): BlockRef;
}

export namespace BlockRef {
    export type AsObject = {
        block?: Block.AsObject,
        range?: ByteRange.AsObject,
    }
}

export class ObjectInfo extends jspb.Message { 

    hasObject(): boolean;
    clearObject(): void;
    getObject(): Object | undefined;
    setObject(value?: Object): ObjectInfo;


    hasBlockRef(): boolean;
    clearBlockRef(): void;
    getBlockRef(): BlockRef | undefined;
    setBlockRef(value?: BlockRef): ObjectInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ObjectInfo.AsObject;
    static toObject(includeInstance: boolean, msg: ObjectInfo): ObjectInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ObjectInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ObjectInfo;
    static deserializeBinaryFromReader(message: ObjectInfo, reader: jspb.BinaryReader): ObjectInfo;
}

export namespace ObjectInfo {
    export type AsObject = {
        object?: Object.AsObject,
        blockRef?: BlockRef.AsObject,
    }
}

export class Compaction extends jspb.Message { 
    clearInputPrefixesList(): void;
    getInputPrefixesList(): Array<string>;
    setInputPrefixesList(value: Array<string>): Compaction;
    addInputPrefixes(value: string, index?: number): string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Compaction.AsObject;
    static toObject(includeInstance: boolean, msg: Compaction): Compaction.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Compaction, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Compaction;
    static deserializeBinaryFromReader(message: Compaction, reader: jspb.BinaryReader): Compaction;
}

export namespace Compaction {
    export type AsObject = {
        inputPrefixesList: Array<string>,
    }
}

export class Shard extends jspb.Message { 

    hasCompaction(): boolean;
    clearCompaction(): void;
    getCompaction(): Compaction | undefined;
    setCompaction(value?: Compaction): Shard;


    hasRange(): boolean;
    clearRange(): void;
    getRange(): PathRange | undefined;
    setRange(value?: PathRange): Shard;

    getOutputPath(): string;
    setOutputPath(value: string): Shard;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Shard.AsObject;
    static toObject(includeInstance: boolean, msg: Shard): Shard.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Shard, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Shard;
    static deserializeBinaryFromReader(message: Shard, reader: jspb.BinaryReader): Shard;
}

export namespace Shard {
    export type AsObject = {
        compaction?: Compaction.AsObject,
        range?: PathRange.AsObject,
        outputPath: string,
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

    getSplitTransaction(): boolean;
    setSplitTransaction(value: boolean): DeleteRepoRequest;


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
        splitTransaction: boolean,
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

export class BuildCommitRequest extends jspb.Message { 

    hasParent(): boolean;
    clearParent(): void;
    getParent(): Commit | undefined;
    setParent(value?: Commit): BuildCommitRequest;

    getBranch(): string;
    setBranch(value: string): BuildCommitRequest;


    hasOrigin(): boolean;
    clearOrigin(): void;
    getOrigin(): CommitOrigin | undefined;
    setOrigin(value?: CommitOrigin): BuildCommitRequest;

    clearProvenanceList(): void;
    getProvenanceList(): Array<CommitProvenance>;
    setProvenanceList(value: Array<CommitProvenance>): BuildCommitRequest;
    addProvenance(value?: CommitProvenance, index?: number): CommitProvenance;


    hasTree(): boolean;
    clearTree(): void;
    getTree(): Object | undefined;
    setTree(value?: Object): BuildCommitRequest;

    clearTreesList(): void;
    getTreesList(): Array<Object>;
    setTreesList(value: Array<Object>): BuildCommitRequest;
    addTrees(value?: Object, index?: number): Object;


    hasDatums(): boolean;
    clearDatums(): void;
    getDatums(): Object | undefined;
    setDatums(value?: Object): BuildCommitRequest;

    getId(): string;
    setId(value: string): BuildCommitRequest;

    getSizeBytes(): number;
    setSizeBytes(value: number): BuildCommitRequest;


    hasStarted(): boolean;
    clearStarted(): void;
    getStarted(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setStarted(value?: google_protobuf_timestamp_pb.Timestamp): BuildCommitRequest;


    hasFinished(): boolean;
    clearFinished(): void;
    getFinished(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setFinished(value?: google_protobuf_timestamp_pb.Timestamp): BuildCommitRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BuildCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: BuildCommitRequest): BuildCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BuildCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BuildCommitRequest;
    static deserializeBinaryFromReader(message: BuildCommitRequest, reader: jspb.BinaryReader): BuildCommitRequest;
}

export namespace BuildCommitRequest {
    export type AsObject = {
        parent?: Commit.AsObject,
        branch: string,
        origin?: CommitOrigin.AsObject,
        provenanceList: Array<CommitProvenance.AsObject>,
        tree?: Object.AsObject,
        treesList: Array<Object.AsObject>,
        datums?: Object.AsObject,
        id: string,
        sizeBytes: number,
        started?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        finished?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class FinishCommitRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): FinishCommitRequest;

    getDescription(): string;
    setDescription(value: string): FinishCommitRequest;


    hasTree(): boolean;
    clearTree(): void;
    getTree(): Object | undefined;
    setTree(value?: Object): FinishCommitRequest;

    clearTreesList(): void;
    getTreesList(): Array<Object>;
    setTreesList(value: Array<Object>): FinishCommitRequest;
    addTrees(value?: Object, index?: number): Object;


    hasDatums(): boolean;
    clearDatums(): void;
    getDatums(): Object | undefined;
    setDatums(value?: Object): FinishCommitRequest;

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
        tree?: Object.AsObject,
        treesList: Array<Object.AsObject>,
        datums?: Object.AsObject,
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

export class DeleteCommitRequest extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): DeleteCommitRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteCommitRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteCommitRequest): DeleteCommitRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteCommitRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteCommitRequest;
    static deserializeBinaryFromReader(message: DeleteCommitRequest, reader: jspb.BinaryReader): DeleteCommitRequest;
}

export namespace DeleteCommitRequest {
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

export class GetFileRequest extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): GetFileRequest;

    getOffsetBytes(): number;
    setOffsetBytes(value: number): GetFileRequest;

    getSizeBytes(): number;
    setSizeBytes(value: number): GetFileRequest;


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
        offsetBytes: number,
        sizeBytes: number,
    }
}

export class OverwriteIndex extends jspb.Message { 
    getIndex(): number;
    setIndex(value: number): OverwriteIndex;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): OverwriteIndex.AsObject;
    static toObject(includeInstance: boolean, msg: OverwriteIndex): OverwriteIndex.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: OverwriteIndex, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): OverwriteIndex;
    static deserializeBinaryFromReader(message: OverwriteIndex, reader: jspb.BinaryReader): OverwriteIndex;
}

export namespace OverwriteIndex {
    export type AsObject = {
        index: number,
    }
}

export class PutFileRequest extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): PutFileRequest;

    getValue(): Uint8Array | string;
    getValue_asU8(): Uint8Array;
    getValue_asB64(): string;
    setValue(value: Uint8Array | string): PutFileRequest;

    getUrl(): string;
    setUrl(value: string): PutFileRequest;

    getRecursive(): boolean;
    setRecursive(value: boolean): PutFileRequest;

    getDelimiter(): Delimiter;
    setDelimiter(value: Delimiter): PutFileRequest;

    getTargetFileDatums(): number;
    setTargetFileDatums(value: number): PutFileRequest;

    getTargetFileBytes(): number;
    setTargetFileBytes(value: number): PutFileRequest;

    getHeaderRecords(): number;
    setHeaderRecords(value: number): PutFileRequest;


    hasOverwriteIndex(): boolean;
    clearOverwriteIndex(): void;
    getOverwriteIndex(): OverwriteIndex | undefined;
    setOverwriteIndex(value?: OverwriteIndex): PutFileRequest;

    getDelete(): boolean;
    setDelete(value: boolean): PutFileRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PutFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PutFileRequest): PutFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PutFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PutFileRequest;
    static deserializeBinaryFromReader(message: PutFileRequest, reader: jspb.BinaryReader): PutFileRequest;
}

export namespace PutFileRequest {
    export type AsObject = {
        file?: File.AsObject,
        value: Uint8Array | string,
        url: string,
        recursive: boolean,
        delimiter: Delimiter,
        targetFileDatums: number,
        targetFileBytes: number,
        headerRecords: number,
        overwriteIndex?: OverwriteIndex.AsObject,
        pb_delete: boolean,
    }
}

export class PutFileRecord extends jspb.Message { 
    getSizeBytes(): number;
    setSizeBytes(value: number): PutFileRecord;

    getObjectHash(): string;
    setObjectHash(value: string): PutFileRecord;


    hasOverwriteIndex(): boolean;
    clearOverwriteIndex(): void;
    getOverwriteIndex(): OverwriteIndex | undefined;
    setOverwriteIndex(value?: OverwriteIndex): PutFileRecord;


    hasBlockRef(): boolean;
    clearBlockRef(): void;
    getBlockRef(): BlockRef | undefined;
    setBlockRef(value?: BlockRef): PutFileRecord;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PutFileRecord.AsObject;
    static toObject(includeInstance: boolean, msg: PutFileRecord): PutFileRecord.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PutFileRecord, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PutFileRecord;
    static deserializeBinaryFromReader(message: PutFileRecord, reader: jspb.BinaryReader): PutFileRecord;
}

export namespace PutFileRecord {
    export type AsObject = {
        sizeBytes: number,
        objectHash: string,
        overwriteIndex?: OverwriteIndex.AsObject,
        blockRef?: BlockRef.AsObject,
    }
}

export class PutFileRecords extends jspb.Message { 
    getSplit(): boolean;
    setSplit(value: boolean): PutFileRecords;

    clearRecordsList(): void;
    getRecordsList(): Array<PutFileRecord>;
    setRecordsList(value: Array<PutFileRecord>): PutFileRecords;
    addRecords(value?: PutFileRecord, index?: number): PutFileRecord;

    getTombstone(): boolean;
    setTombstone(value: boolean): PutFileRecords;


    hasHeader(): boolean;
    clearHeader(): void;
    getHeader(): PutFileRecord | undefined;
    setHeader(value?: PutFileRecord): PutFileRecords;


    hasFooter(): boolean;
    clearFooter(): void;
    getFooter(): PutFileRecord | undefined;
    setFooter(value?: PutFileRecord): PutFileRecords;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PutFileRecords.AsObject;
    static toObject(includeInstance: boolean, msg: PutFileRecords): PutFileRecords.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PutFileRecords, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PutFileRecords;
    static deserializeBinaryFromReader(message: PutFileRecords, reader: jspb.BinaryReader): PutFileRecords;
}

export namespace PutFileRecords {
    export type AsObject = {
        split: boolean,
        recordsList: Array<PutFileRecord.AsObject>,
        tombstone: boolean,
        header?: PutFileRecord.AsObject,
        footer?: PutFileRecord.AsObject,
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

    getHistory(): number;
    setHistory(value: number): ListFileRequest;


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
        history: number,
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

export class FileInfos extends jspb.Message { 
    clearFileInfoList(): void;
    getFileInfoList(): Array<FileInfo>;
    setFileInfoList(value: Array<FileInfo>): FileInfos;
    addFileInfo(value?: FileInfo, index?: number): FileInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FileInfos.AsObject;
    static toObject(includeInstance: boolean, msg: FileInfos): FileInfos.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FileInfos, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FileInfos;
    static deserializeBinaryFromReader(message: FileInfos, reader: jspb.BinaryReader): FileInfos;
}

export namespace FileInfos {
    export type AsObject = {
        fileInfoList: Array<FileInfo.AsObject>,
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
    clearNewFilesList(): void;
    getNewFilesList(): Array<FileInfo>;
    setNewFilesList(value: Array<FileInfo>): DiffFileResponse;
    addNewFiles(value?: FileInfo, index?: number): FileInfo;

    clearOldFilesList(): void;
    getOldFilesList(): Array<FileInfo>;
    setOldFilesList(value: Array<FileInfo>): DiffFileResponse;
    addOldFiles(value?: FileInfo, index?: number): FileInfo;


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
        newFilesList: Array<FileInfo.AsObject>,
        oldFilesList: Array<FileInfo.AsObject>,
    }
}

export class DeleteFileRequest extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): DeleteFileRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteFileRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteFileRequest): DeleteFileRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteFileRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteFileRequest;
    static deserializeBinaryFromReader(message: DeleteFileRequest, reader: jspb.BinaryReader): DeleteFileRequest;
}

export namespace DeleteFileRequest {
    export type AsObject = {
        file?: File.AsObject,
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

export class FileOperationRequestV2 extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): FileOperationRequestV2;


    hasPutTar(): boolean;
    clearPutTar(): void;
    getPutTar(): PutTarRequestV2 | undefined;
    setPutTar(value?: PutTarRequestV2): FileOperationRequestV2;


    hasDeleteFiles(): boolean;
    clearDeleteFiles(): void;
    getDeleteFiles(): DeleteFilesRequestV2 | undefined;
    setDeleteFiles(value?: DeleteFilesRequestV2): FileOperationRequestV2;


    getOperationCase(): FileOperationRequestV2.OperationCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FileOperationRequestV2.AsObject;
    static toObject(includeInstance: boolean, msg: FileOperationRequestV2): FileOperationRequestV2.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FileOperationRequestV2, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FileOperationRequestV2;
    static deserializeBinaryFromReader(message: FileOperationRequestV2, reader: jspb.BinaryReader): FileOperationRequestV2;
}

export namespace FileOperationRequestV2 {
    export type AsObject = {
        commit?: Commit.AsObject,
        putTar?: PutTarRequestV2.AsObject,
        deleteFiles?: DeleteFilesRequestV2.AsObject,
    }

    export enum OperationCase {
        OPERATION_NOT_SET = 0,
    
    PUT_TAR = 2,

    DELETE_FILES = 3,

    }

}

export class PutTarRequestV2 extends jspb.Message { 
    getOverwrite(): boolean;
    setOverwrite(value: boolean): PutTarRequestV2;

    getTag(): string;
    setTag(value: string): PutTarRequestV2;

    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): PutTarRequestV2;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PutTarRequestV2.AsObject;
    static toObject(includeInstance: boolean, msg: PutTarRequestV2): PutTarRequestV2.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PutTarRequestV2, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PutTarRequestV2;
    static deserializeBinaryFromReader(message: PutTarRequestV2, reader: jspb.BinaryReader): PutTarRequestV2;
}

export namespace PutTarRequestV2 {
    export type AsObject = {
        overwrite: boolean,
        tag: string,
        data: Uint8Array | string,
    }
}

export class DeleteFilesRequestV2 extends jspb.Message { 
    clearFilesList(): void;
    getFilesList(): Array<string>;
    setFilesList(value: Array<string>): DeleteFilesRequestV2;
    addFiles(value: string, index?: number): string;

    getTag(): string;
    setTag(value: string): DeleteFilesRequestV2;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteFilesRequestV2.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteFilesRequestV2): DeleteFilesRequestV2.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteFilesRequestV2, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteFilesRequestV2;
    static deserializeBinaryFromReader(message: DeleteFilesRequestV2, reader: jspb.BinaryReader): DeleteFilesRequestV2;
}

export namespace DeleteFilesRequestV2 {
    export type AsObject = {
        filesList: Array<string>,
        tag: string,
    }
}

export class GetTarRequestV2 extends jspb.Message { 

    hasFile(): boolean;
    clearFile(): void;
    getFile(): File | undefined;
    setFile(value?: File): GetTarRequestV2;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetTarRequestV2.AsObject;
    static toObject(includeInstance: boolean, msg: GetTarRequestV2): GetTarRequestV2.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetTarRequestV2, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetTarRequestV2;
    static deserializeBinaryFromReader(message: GetTarRequestV2, reader: jspb.BinaryReader): GetTarRequestV2;
}

export namespace GetTarRequestV2 {
    export type AsObject = {
        file?: File.AsObject,
    }
}

export class DiffFileResponseV2 extends jspb.Message { 

    hasOldFile(): boolean;
    clearOldFile(): void;
    getOldFile(): FileInfo | undefined;
    setOldFile(value?: FileInfo): DiffFileResponseV2;


    hasNewFile(): boolean;
    clearNewFile(): void;
    getNewFile(): FileInfo | undefined;
    setNewFile(value?: FileInfo): DiffFileResponseV2;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DiffFileResponseV2.AsObject;
    static toObject(includeInstance: boolean, msg: DiffFileResponseV2): DiffFileResponseV2.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DiffFileResponseV2, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DiffFileResponseV2;
    static deserializeBinaryFromReader(message: DiffFileResponseV2, reader: jspb.BinaryReader): DiffFileResponseV2;
}

export namespace DiffFileResponseV2 {
    export type AsObject = {
        oldFile?: FileInfo.AsObject,
        newFile?: FileInfo.AsObject,
    }
}

export class CreateTmpFileSetResponse extends jspb.Message { 
    getFilesetId(): string;
    setFilesetId(value: string): CreateTmpFileSetResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateTmpFileSetResponse.AsObject;
    static toObject(includeInstance: boolean, msg: CreateTmpFileSetResponse): CreateTmpFileSetResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateTmpFileSetResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateTmpFileSetResponse;
    static deserializeBinaryFromReader(message: CreateTmpFileSetResponse, reader: jspb.BinaryReader): CreateTmpFileSetResponse;
}

export namespace CreateTmpFileSetResponse {
    export type AsObject = {
        filesetId: string,
    }
}

export class RenewTmpFileSetRequest extends jspb.Message { 
    getFilesetId(): string;
    setFilesetId(value: string): RenewTmpFileSetRequest;

    getTtlSeconds(): number;
    setTtlSeconds(value: number): RenewTmpFileSetRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RenewTmpFileSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RenewTmpFileSetRequest): RenewTmpFileSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RenewTmpFileSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RenewTmpFileSetRequest;
    static deserializeBinaryFromReader(message: RenewTmpFileSetRequest, reader: jspb.BinaryReader): RenewTmpFileSetRequest;
}

export namespace RenewTmpFileSetRequest {
    export type AsObject = {
        filesetId: string,
        ttlSeconds: number,
    }
}

export class ClearCommitRequestV2 extends jspb.Message { 

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): Commit | undefined;
    setCommit(value?: Commit): ClearCommitRequestV2;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ClearCommitRequestV2.AsObject;
    static toObject(includeInstance: boolean, msg: ClearCommitRequestV2): ClearCommitRequestV2.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ClearCommitRequestV2, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ClearCommitRequestV2;
    static deserializeBinaryFromReader(message: ClearCommitRequestV2, reader: jspb.BinaryReader): ClearCommitRequestV2;
}

export namespace ClearCommitRequestV2 {
    export type AsObject = {
        commit?: Commit.AsObject,
    }
}

export class PutObjectRequest extends jspb.Message { 
    getValue(): Uint8Array | string;
    getValue_asU8(): Uint8Array;
    getValue_asB64(): string;
    setValue(value: Uint8Array | string): PutObjectRequest;

    clearTagsList(): void;
    getTagsList(): Array<Tag>;
    setTagsList(value: Array<Tag>): PutObjectRequest;
    addTags(value?: Tag, index?: number): Tag;


    hasBlock(): boolean;
    clearBlock(): void;
    getBlock(): Block | undefined;
    setBlock(value?: Block): PutObjectRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PutObjectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PutObjectRequest): PutObjectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PutObjectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PutObjectRequest;
    static deserializeBinaryFromReader(message: PutObjectRequest, reader: jspb.BinaryReader): PutObjectRequest;
}

export namespace PutObjectRequest {
    export type AsObject = {
        value: Uint8Array | string,
        tagsList: Array<Tag.AsObject>,
        block?: Block.AsObject,
    }
}

export class CreateObjectRequest extends jspb.Message { 

    hasObject(): boolean;
    clearObject(): void;
    getObject(): Object | undefined;
    setObject(value?: Object): CreateObjectRequest;


    hasBlockRef(): boolean;
    clearBlockRef(): void;
    getBlockRef(): BlockRef | undefined;
    setBlockRef(value?: BlockRef): CreateObjectRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateObjectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateObjectRequest): CreateObjectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateObjectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateObjectRequest;
    static deserializeBinaryFromReader(message: CreateObjectRequest, reader: jspb.BinaryReader): CreateObjectRequest;
}

export namespace CreateObjectRequest {
    export type AsObject = {
        object?: Object.AsObject,
        blockRef?: BlockRef.AsObject,
    }
}

export class GetObjectsRequest extends jspb.Message { 
    clearObjectsList(): void;
    getObjectsList(): Array<Object>;
    setObjectsList(value: Array<Object>): GetObjectsRequest;
    addObjects(value?: Object, index?: number): Object;

    getOffsetBytes(): number;
    setOffsetBytes(value: number): GetObjectsRequest;

    getSizeBytes(): number;
    setSizeBytes(value: number): GetObjectsRequest;

    getTotalSize(): number;
    setTotalSize(value: number): GetObjectsRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetObjectsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetObjectsRequest): GetObjectsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetObjectsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetObjectsRequest;
    static deserializeBinaryFromReader(message: GetObjectsRequest, reader: jspb.BinaryReader): GetObjectsRequest;
}

export namespace GetObjectsRequest {
    export type AsObject = {
        objectsList: Array<Object.AsObject>,
        offsetBytes: number,
        sizeBytes: number,
        totalSize: number,
    }
}

export class PutBlockRequest extends jspb.Message { 

    hasBlock(): boolean;
    clearBlock(): void;
    getBlock(): Block | undefined;
    setBlock(value?: Block): PutBlockRequest;

    getValue(): Uint8Array | string;
    getValue_asU8(): Uint8Array;
    getValue_asB64(): string;
    setValue(value: Uint8Array | string): PutBlockRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PutBlockRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PutBlockRequest): PutBlockRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PutBlockRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PutBlockRequest;
    static deserializeBinaryFromReader(message: PutBlockRequest, reader: jspb.BinaryReader): PutBlockRequest;
}

export namespace PutBlockRequest {
    export type AsObject = {
        block?: Block.AsObject,
        value: Uint8Array | string,
    }
}

export class GetBlockRequest extends jspb.Message { 

    hasBlock(): boolean;
    clearBlock(): void;
    getBlock(): Block | undefined;
    setBlock(value?: Block): GetBlockRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetBlockRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetBlockRequest): GetBlockRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetBlockRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetBlockRequest;
    static deserializeBinaryFromReader(message: GetBlockRequest, reader: jspb.BinaryReader): GetBlockRequest;
}

export namespace GetBlockRequest {
    export type AsObject = {
        block?: Block.AsObject,
    }
}

export class GetBlocksRequest extends jspb.Message { 
    clearBlockrefsList(): void;
    getBlockrefsList(): Array<BlockRef>;
    setBlockrefsList(value: Array<BlockRef>): GetBlocksRequest;
    addBlockrefs(value?: BlockRef, index?: number): BlockRef;

    getOffsetBytes(): number;
    setOffsetBytes(value: number): GetBlocksRequest;

    getSizeBytes(): number;
    setSizeBytes(value: number): GetBlocksRequest;

    getTotalSize(): number;
    setTotalSize(value: number): GetBlocksRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetBlocksRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetBlocksRequest): GetBlocksRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetBlocksRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetBlocksRequest;
    static deserializeBinaryFromReader(message: GetBlocksRequest, reader: jspb.BinaryReader): GetBlocksRequest;
}

export namespace GetBlocksRequest {
    export type AsObject = {
        blockrefsList: Array<BlockRef.AsObject>,
        offsetBytes: number,
        sizeBytes: number,
        totalSize: number,
    }
}

export class ListBlockRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListBlockRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListBlockRequest): ListBlockRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListBlockRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListBlockRequest;
    static deserializeBinaryFromReader(message: ListBlockRequest, reader: jspb.BinaryReader): ListBlockRequest;
}

export namespace ListBlockRequest {
    export type AsObject = {
    }
}

export class TagObjectRequest extends jspb.Message { 

    hasObject(): boolean;
    clearObject(): void;
    getObject(): Object | undefined;
    setObject(value?: Object): TagObjectRequest;

    clearTagsList(): void;
    getTagsList(): Array<Tag>;
    setTagsList(value: Array<Tag>): TagObjectRequest;
    addTags(value?: Tag, index?: number): Tag;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TagObjectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: TagObjectRequest): TagObjectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TagObjectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TagObjectRequest;
    static deserializeBinaryFromReader(message: TagObjectRequest, reader: jspb.BinaryReader): TagObjectRequest;
}

export namespace TagObjectRequest {
    export type AsObject = {
        object?: Object.AsObject,
        tagsList: Array<Tag.AsObject>,
    }
}

export class ListObjectsRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListObjectsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListObjectsRequest): ListObjectsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListObjectsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListObjectsRequest;
    static deserializeBinaryFromReader(message: ListObjectsRequest, reader: jspb.BinaryReader): ListObjectsRequest;
}

export namespace ListObjectsRequest {
    export type AsObject = {
    }
}

export class ListTagsRequest extends jspb.Message { 
    getPrefix(): string;
    setPrefix(value: string): ListTagsRequest;

    getIncludeObject(): boolean;
    setIncludeObject(value: boolean): ListTagsRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListTagsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListTagsRequest): ListTagsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListTagsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListTagsRequest;
    static deserializeBinaryFromReader(message: ListTagsRequest, reader: jspb.BinaryReader): ListTagsRequest;
}

export namespace ListTagsRequest {
    export type AsObject = {
        prefix: string,
        includeObject: boolean,
    }
}

export class ListTagsResponse extends jspb.Message { 

    hasTag(): boolean;
    clearTag(): void;
    getTag(): Tag | undefined;
    setTag(value?: Tag): ListTagsResponse;


    hasObject(): boolean;
    clearObject(): void;
    getObject(): Object | undefined;
    setObject(value?: Object): ListTagsResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListTagsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListTagsResponse): ListTagsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListTagsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListTagsResponse;
    static deserializeBinaryFromReader(message: ListTagsResponse, reader: jspb.BinaryReader): ListTagsResponse;
}

export namespace ListTagsResponse {
    export type AsObject = {
        tag?: Tag.AsObject,
        object?: Object.AsObject,
    }
}

export class DeleteObjectsRequest extends jspb.Message { 
    clearObjectsList(): void;
    getObjectsList(): Array<Object>;
    setObjectsList(value: Array<Object>): DeleteObjectsRequest;
    addObjects(value?: Object, index?: number): Object;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteObjectsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteObjectsRequest): DeleteObjectsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteObjectsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteObjectsRequest;
    static deserializeBinaryFromReader(message: DeleteObjectsRequest, reader: jspb.BinaryReader): DeleteObjectsRequest;
}

export namespace DeleteObjectsRequest {
    export type AsObject = {
        objectsList: Array<Object.AsObject>,
    }
}

export class DeleteObjectsResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteObjectsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteObjectsResponse): DeleteObjectsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteObjectsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteObjectsResponse;
    static deserializeBinaryFromReader(message: DeleteObjectsResponse, reader: jspb.BinaryReader): DeleteObjectsResponse;
}

export namespace DeleteObjectsResponse {
    export type AsObject = {
    }
}

export class DeleteTagsRequest extends jspb.Message { 
    clearTagsList(): void;
    getTagsList(): Array<Tag>;
    setTagsList(value: Array<Tag>): DeleteTagsRequest;
    addTags(value?: Tag, index?: number): Tag;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteTagsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteTagsRequest): DeleteTagsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteTagsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteTagsRequest;
    static deserializeBinaryFromReader(message: DeleteTagsRequest, reader: jspb.BinaryReader): DeleteTagsRequest;
}

export namespace DeleteTagsRequest {
    export type AsObject = {
        tagsList: Array<Tag.AsObject>,
    }
}

export class DeleteTagsResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteTagsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteTagsResponse): DeleteTagsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteTagsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteTagsResponse;
    static deserializeBinaryFromReader(message: DeleteTagsResponse, reader: jspb.BinaryReader): DeleteTagsResponse;
}

export namespace DeleteTagsResponse {
    export type AsObject = {
    }
}

export class CheckObjectRequest extends jspb.Message { 

    hasObject(): boolean;
    clearObject(): void;
    getObject(): Object | undefined;
    setObject(value?: Object): CheckObjectRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CheckObjectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CheckObjectRequest): CheckObjectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CheckObjectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CheckObjectRequest;
    static deserializeBinaryFromReader(message: CheckObjectRequest, reader: jspb.BinaryReader): CheckObjectRequest;
}

export namespace CheckObjectRequest {
    export type AsObject = {
        object?: Object.AsObject,
    }
}

export class CheckObjectResponse extends jspb.Message { 
    getExists(): boolean;
    setExists(value: boolean): CheckObjectResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CheckObjectResponse.AsObject;
    static toObject(includeInstance: boolean, msg: CheckObjectResponse): CheckObjectResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CheckObjectResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CheckObjectResponse;
    static deserializeBinaryFromReader(message: CheckObjectResponse, reader: jspb.BinaryReader): CheckObjectResponse;
}

export namespace CheckObjectResponse {
    export type AsObject = {
        exists: boolean,
    }
}

export class Objects extends jspb.Message { 
    clearObjectsList(): void;
    getObjectsList(): Array<Object>;
    setObjectsList(value: Array<Object>): Objects;
    addObjects(value?: Object, index?: number): Object;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Objects.AsObject;
    static toObject(includeInstance: boolean, msg: Objects): Objects.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Objects, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Objects;
    static deserializeBinaryFromReader(message: Objects, reader: jspb.BinaryReader): Objects;
}

export namespace Objects {
    export type AsObject = {
        objectsList: Array<Object.AsObject>,
    }
}

export class PutObjDirectRequest extends jspb.Message { 
    getObj(): string;
    setObj(value: string): PutObjDirectRequest;

    getValue(): Uint8Array | string;
    getValue_asU8(): Uint8Array;
    getValue_asB64(): string;
    setValue(value: Uint8Array | string): PutObjDirectRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PutObjDirectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PutObjDirectRequest): PutObjDirectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PutObjDirectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PutObjDirectRequest;
    static deserializeBinaryFromReader(message: PutObjDirectRequest, reader: jspb.BinaryReader): PutObjDirectRequest;
}

export namespace PutObjDirectRequest {
    export type AsObject = {
        obj: string,
        value: Uint8Array | string,
    }
}

export class GetObjDirectRequest extends jspb.Message { 
    getObj(): string;
    setObj(value: string): GetObjDirectRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetObjDirectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetObjDirectRequest): GetObjDirectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetObjDirectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetObjDirectRequest;
    static deserializeBinaryFromReader(message: GetObjDirectRequest, reader: jspb.BinaryReader): GetObjDirectRequest;
}

export namespace GetObjDirectRequest {
    export type AsObject = {
        obj: string,
    }
}

export class DeleteObjDirectRequest extends jspb.Message { 
    getObject(): string;
    setObject(value: string): DeleteObjDirectRequest;

    getPrefix(): string;
    setPrefix(value: string): DeleteObjDirectRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteObjDirectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteObjDirectRequest): DeleteObjDirectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteObjDirectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteObjDirectRequest;
    static deserializeBinaryFromReader(message: DeleteObjDirectRequest, reader: jspb.BinaryReader): DeleteObjDirectRequest;
}

export namespace DeleteObjDirectRequest {
    export type AsObject = {
        object: string,
        prefix: string,
    }
}

export class ObjectIndex extends jspb.Message { 

    getObjectsMap(): jspb.Map<string, BlockRef>;
    clearObjectsMap(): void;


    getTagsMap(): jspb.Map<string, Object>;
    clearTagsMap(): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ObjectIndex.AsObject;
    static toObject(includeInstance: boolean, msg: ObjectIndex): ObjectIndex.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ObjectIndex, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ObjectIndex;
    static deserializeBinaryFromReader(message: ObjectIndex, reader: jspb.BinaryReader): ObjectIndex;
}

export namespace ObjectIndex {
    export type AsObject = {

        objectsMap: Array<[string, BlockRef.AsObject]>,

        tagsMap: Array<[string, Object.AsObject]>,
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
