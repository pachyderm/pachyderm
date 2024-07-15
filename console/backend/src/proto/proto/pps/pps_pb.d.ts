// package: pps_v2
// file: pps/pps.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as google_protobuf_duration_pb from "google-protobuf/google/protobuf/duration_pb";
import * as google_protobuf_wrappers_pb from "google-protobuf/google/protobuf/wrappers_pb";
import * as pfs_pfs_pb from "../pfs/pfs_pb";
import * as task_task_pb from "../task/task_pb";
import * as protoextensions_json_schema_options_pb from "../protoextensions/json-schema-options_pb";
import * as protoextensions_validate_pb from "../protoextensions/validate_pb";

export class SecretMount extends jspb.Message { 
    getName(): string;
    setName(value: string): SecretMount;
    getKey(): string;
    setKey(value: string): SecretMount;
    getMountPath(): string;
    setMountPath(value: string): SecretMount;
    getEnvVar(): string;
    setEnvVar(value: string): SecretMount;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SecretMount.AsObject;
    static toObject(includeInstance: boolean, msg: SecretMount): SecretMount.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SecretMount, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SecretMount;
    static deserializeBinaryFromReader(message: SecretMount, reader: jspb.BinaryReader): SecretMount;
}

export namespace SecretMount {
    export type AsObject = {
        name: string,
        key: string,
        mountPath: string,
        envVar: string,
    }
}

export class Transform extends jspb.Message { 
    getImage(): string;
    setImage(value: string): Transform;
    clearCmdList(): void;
    getCmdList(): Array<string>;
    setCmdList(value: Array<string>): Transform;
    addCmd(value: string, index?: number): string;
    clearErrCmdList(): void;
    getErrCmdList(): Array<string>;
    setErrCmdList(value: Array<string>): Transform;
    addErrCmd(value: string, index?: number): string;

    getEnvMap(): jspb.Map<string, string>;
    clearEnvMap(): void;
    clearSecretsList(): void;
    getSecretsList(): Array<SecretMount>;
    setSecretsList(value: Array<SecretMount>): Transform;
    addSecrets(value?: SecretMount, index?: number): SecretMount;
    clearImagePullSecretsList(): void;
    getImagePullSecretsList(): Array<string>;
    setImagePullSecretsList(value: Array<string>): Transform;
    addImagePullSecrets(value: string, index?: number): string;
    clearStdinList(): void;
    getStdinList(): Array<string>;
    setStdinList(value: Array<string>): Transform;
    addStdin(value: string, index?: number): string;
    clearErrStdinList(): void;
    getErrStdinList(): Array<string>;
    setErrStdinList(value: Array<string>): Transform;
    addErrStdin(value: string, index?: number): string;
    clearAcceptReturnCodeList(): void;
    getAcceptReturnCodeList(): Array<number>;
    setAcceptReturnCodeList(value: Array<number>): Transform;
    addAcceptReturnCode(value: number, index?: number): number;
    getDebug(): boolean;
    setDebug(value: boolean): Transform;
    getUser(): string;
    setUser(value: string): Transform;
    getWorkingDir(): string;
    setWorkingDir(value: string): Transform;
    getDockerfile(): string;
    setDockerfile(value: string): Transform;
    getMemoryVolume(): boolean;
    setMemoryVolume(value: boolean): Transform;
    getDatumBatching(): boolean;
    setDatumBatching(value: boolean): Transform;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Transform.AsObject;
    static toObject(includeInstance: boolean, msg: Transform): Transform.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Transform, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Transform;
    static deserializeBinaryFromReader(message: Transform, reader: jspb.BinaryReader): Transform;
}

export namespace Transform {
    export type AsObject = {
        image: string,
        cmdList: Array<string>,
        errCmdList: Array<string>,

        envMap: Array<[string, string]>,
        secretsList: Array<SecretMount.AsObject>,
        imagePullSecretsList: Array<string>,
        stdinList: Array<string>,
        errStdinList: Array<string>,
        acceptReturnCodeList: Array<number>,
        debug: boolean,
        user: string,
        workingDir: string,
        dockerfile: string,
        memoryVolume: boolean,
        datumBatching: boolean,
    }
}

export class TFJob extends jspb.Message { 
    getTfJob(): string;
    setTfJob(value: string): TFJob;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TFJob.AsObject;
    static toObject(includeInstance: boolean, msg: TFJob): TFJob.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TFJob, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TFJob;
    static deserializeBinaryFromReader(message: TFJob, reader: jspb.BinaryReader): TFJob;
}

export namespace TFJob {
    export type AsObject = {
        tfJob: string,
    }
}

export class Egress extends jspb.Message { 
    getUrl(): string;
    setUrl(value: string): Egress;

    hasObjectStorage(): boolean;
    clearObjectStorage(): void;
    getObjectStorage(): pfs_pfs_pb.ObjectStorageEgress | undefined;
    setObjectStorage(value?: pfs_pfs_pb.ObjectStorageEgress): Egress;

    hasSqlDatabase(): boolean;
    clearSqlDatabase(): void;
    getSqlDatabase(): pfs_pfs_pb.SQLDatabaseEgress | undefined;
    setSqlDatabase(value?: pfs_pfs_pb.SQLDatabaseEgress): Egress;

    getTargetCase(): Egress.TargetCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Egress.AsObject;
    static toObject(includeInstance: boolean, msg: Egress): Egress.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Egress, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Egress;
    static deserializeBinaryFromReader(message: Egress, reader: jspb.BinaryReader): Egress;
}

export namespace Egress {
    export type AsObject = {
        url: string,
        objectStorage?: pfs_pfs_pb.ObjectStorageEgress.AsObject,
        sqlDatabase?: pfs_pfs_pb.SQLDatabaseEgress.AsObject,
    }

    export enum TargetCase {
        TARGET_NOT_SET = 0,
        OBJECT_STORAGE = 2,
        SQL_DATABASE = 3,
    }

}

export class Determined extends jspb.Message { 
    clearWorkspacesList(): void;
    getWorkspacesList(): Array<string>;
    setWorkspacesList(value: Array<string>): Determined;
    addWorkspaces(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Determined.AsObject;
    static toObject(includeInstance: boolean, msg: Determined): Determined.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Determined, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Determined;
    static deserializeBinaryFromReader(message: Determined, reader: jspb.BinaryReader): Determined;
}

export namespace Determined {
    export type AsObject = {
        workspacesList: Array<string>,
    }
}

export class Job extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): Job;
    getId(): string;
    setId(value: string): Job;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Job.AsObject;
    static toObject(includeInstance: boolean, msg: Job): Job.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Job, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Job;
    static deserializeBinaryFromReader(message: Job, reader: jspb.BinaryReader): Job;
}

export namespace Job {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        id: string,
    }
}

export class Metadata extends jspb.Message { 

    getAnnotationsMap(): jspb.Map<string, string>;
    clearAnnotationsMap(): void;

    getLabelsMap(): jspb.Map<string, string>;
    clearLabelsMap(): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Metadata.AsObject;
    static toObject(includeInstance: boolean, msg: Metadata): Metadata.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Metadata, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Metadata;
    static deserializeBinaryFromReader(message: Metadata, reader: jspb.BinaryReader): Metadata;
}

export namespace Metadata {
    export type AsObject = {

        annotationsMap: Array<[string, string]>,

        labelsMap: Array<[string, string]>,
    }
}

export class Service extends jspb.Message { 
    getInternalPort(): number;
    setInternalPort(value: number): Service;
    getExternalPort(): number;
    setExternalPort(value: number): Service;
    getIp(): string;
    setIp(value: string): Service;
    getType(): string;
    setType(value: string): Service;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Service.AsObject;
    static toObject(includeInstance: boolean, msg: Service): Service.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Service, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Service;
    static deserializeBinaryFromReader(message: Service, reader: jspb.BinaryReader): Service;
}

export namespace Service {
    export type AsObject = {
        internalPort: number,
        externalPort: number,
        ip: string,
        type: string,
    }
}

export class Spout extends jspb.Message { 

    hasService(): boolean;
    clearService(): void;
    getService(): Service | undefined;
    setService(value?: Service): Spout;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Spout.AsObject;
    static toObject(includeInstance: boolean, msg: Spout): Spout.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Spout, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Spout;
    static deserializeBinaryFromReader(message: Spout, reader: jspb.BinaryReader): Spout;
}

export namespace Spout {
    export type AsObject = {
        service?: Service.AsObject,
    }
}

export class PFSInput extends jspb.Message { 
    getProject(): string;
    setProject(value: string): PFSInput;
    getName(): string;
    setName(value: string): PFSInput;
    getRepo(): string;
    setRepo(value: string): PFSInput;
    getRepoType(): string;
    setRepoType(value: string): PFSInput;
    getBranch(): string;
    setBranch(value: string): PFSInput;
    getCommit(): string;
    setCommit(value: string): PFSInput;
    getGlob(): string;
    setGlob(value: string): PFSInput;
    getJoinOn(): string;
    setJoinOn(value: string): PFSInput;
    getOuterJoin(): boolean;
    setOuterJoin(value: boolean): PFSInput;
    getGroupBy(): string;
    setGroupBy(value: string): PFSInput;
    getLazy(): boolean;
    setLazy(value: boolean): PFSInput;
    getEmptyFiles(): boolean;
    setEmptyFiles(value: boolean): PFSInput;
    getS3(): boolean;
    setS3(value: boolean): PFSInput;

    hasTrigger(): boolean;
    clearTrigger(): void;
    getTrigger(): pfs_pfs_pb.Trigger | undefined;
    setTrigger(value?: pfs_pfs_pb.Trigger): PFSInput;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PFSInput.AsObject;
    static toObject(includeInstance: boolean, msg: PFSInput): PFSInput.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PFSInput, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PFSInput;
    static deserializeBinaryFromReader(message: PFSInput, reader: jspb.BinaryReader): PFSInput;
}

export namespace PFSInput {
    export type AsObject = {
        project: string,
        name: string,
        repo: string,
        repoType: string,
        branch: string,
        commit: string,
        glob: string,
        joinOn: string,
        outerJoin: boolean,
        groupBy: string,
        lazy: boolean,
        emptyFiles: boolean,
        s3: boolean,
        trigger?: pfs_pfs_pb.Trigger.AsObject,
    }
}

export class CronInput extends jspb.Message { 
    getName(): string;
    setName(value: string): CronInput;
    getProject(): string;
    setProject(value: string): CronInput;
    getRepo(): string;
    setRepo(value: string): CronInput;
    getCommit(): string;
    setCommit(value: string): CronInput;
    getSpec(): string;
    setSpec(value: string): CronInput;
    getOverwrite(): boolean;
    setOverwrite(value: boolean): CronInput;

    hasStart(): boolean;
    clearStart(): void;
    getStart(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setStart(value?: google_protobuf_timestamp_pb.Timestamp): CronInput;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CronInput.AsObject;
    static toObject(includeInstance: boolean, msg: CronInput): CronInput.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CronInput, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CronInput;
    static deserializeBinaryFromReader(message: CronInput, reader: jspb.BinaryReader): CronInput;
}

export namespace CronInput {
    export type AsObject = {
        name: string,
        project: string,
        repo: string,
        commit: string,
        spec: string,
        overwrite: boolean,
        start?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class Input extends jspb.Message { 

    hasPfs(): boolean;
    clearPfs(): void;
    getPfs(): PFSInput | undefined;
    setPfs(value?: PFSInput): Input;
    clearJoinList(): void;
    getJoinList(): Array<Input>;
    setJoinList(value: Array<Input>): Input;
    addJoin(value?: Input, index?: number): Input;
    clearGroupList(): void;
    getGroupList(): Array<Input>;
    setGroupList(value: Array<Input>): Input;
    addGroup(value?: Input, index?: number): Input;
    clearCrossList(): void;
    getCrossList(): Array<Input>;
    setCrossList(value: Array<Input>): Input;
    addCross(value?: Input, index?: number): Input;
    clearUnionList(): void;
    getUnionList(): Array<Input>;
    setUnionList(value: Array<Input>): Input;
    addUnion(value?: Input, index?: number): Input;

    hasCron(): boolean;
    clearCron(): void;
    getCron(): CronInput | undefined;
    setCron(value?: CronInput): Input;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Input.AsObject;
    static toObject(includeInstance: boolean, msg: Input): Input.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Input, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Input;
    static deserializeBinaryFromReader(message: Input, reader: jspb.BinaryReader): Input;
}

export namespace Input {
    export type AsObject = {
        pfs?: PFSInput.AsObject,
        joinList: Array<Input.AsObject>,
        groupList: Array<Input.AsObject>,
        crossList: Array<Input.AsObject>,
        unionList: Array<Input.AsObject>,
        cron?: CronInput.AsObject,
    }
}

export class JobInput extends jspb.Message { 
    getName(): string;
    setName(value: string): JobInput;

    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): pfs_pfs_pb.Commit | undefined;
    setCommit(value?: pfs_pfs_pb.Commit): JobInput;
    getGlob(): string;
    setGlob(value: string): JobInput;
    getLazy(): boolean;
    setLazy(value: boolean): JobInput;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): JobInput.AsObject;
    static toObject(includeInstance: boolean, msg: JobInput): JobInput.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: JobInput, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): JobInput;
    static deserializeBinaryFromReader(message: JobInput, reader: jspb.BinaryReader): JobInput;
}

export namespace JobInput {
    export type AsObject = {
        name: string,
        commit?: pfs_pfs_pb.Commit.AsObject,
        glob: string,
        lazy: boolean,
    }
}

export class ParallelismSpec extends jspb.Message { 
    getConstant(): number;
    setConstant(value: number): ParallelismSpec;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ParallelismSpec.AsObject;
    static toObject(includeInstance: boolean, msg: ParallelismSpec): ParallelismSpec.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ParallelismSpec, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ParallelismSpec;
    static deserializeBinaryFromReader(message: ParallelismSpec, reader: jspb.BinaryReader): ParallelismSpec;
}

export namespace ParallelismSpec {
    export type AsObject = {
        constant: number,
    }
}

export class InputFile extends jspb.Message { 
    getPath(): string;
    setPath(value: string): InputFile;
    getHash(): Uint8Array | string;
    getHash_asU8(): Uint8Array;
    getHash_asB64(): string;
    setHash(value: Uint8Array | string): InputFile;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InputFile.AsObject;
    static toObject(includeInstance: boolean, msg: InputFile): InputFile.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InputFile, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InputFile;
    static deserializeBinaryFromReader(message: InputFile, reader: jspb.BinaryReader): InputFile;
}

export namespace InputFile {
    export type AsObject = {
        path: string,
        hash: Uint8Array | string,
    }
}

export class Datum extends jspb.Message { 

    hasJob(): boolean;
    clearJob(): void;
    getJob(): Job | undefined;
    setJob(value?: Job): Datum;
    getId(): string;
    setId(value: string): Datum;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Datum.AsObject;
    static toObject(includeInstance: boolean, msg: Datum): Datum.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Datum, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Datum;
    static deserializeBinaryFromReader(message: Datum, reader: jspb.BinaryReader): Datum;
}

export namespace Datum {
    export type AsObject = {
        job?: Job.AsObject,
        id: string,
    }
}

export class DatumInfo extends jspb.Message { 

    hasDatum(): boolean;
    clearDatum(): void;
    getDatum(): Datum | undefined;
    setDatum(value?: Datum): DatumInfo;
    getState(): DatumState;
    setState(value: DatumState): DatumInfo;

    hasStats(): boolean;
    clearStats(): void;
    getStats(): ProcessStats | undefined;
    setStats(value?: ProcessStats): DatumInfo;

    hasPfsState(): boolean;
    clearPfsState(): void;
    getPfsState(): pfs_pfs_pb.File | undefined;
    setPfsState(value?: pfs_pfs_pb.File): DatumInfo;
    clearDataList(): void;
    getDataList(): Array<pfs_pfs_pb.FileInfo>;
    setDataList(value: Array<pfs_pfs_pb.FileInfo>): DatumInfo;
    addData(value?: pfs_pfs_pb.FileInfo, index?: number): pfs_pfs_pb.FileInfo;
    getImageId(): string;
    setImageId(value: string): DatumInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DatumInfo.AsObject;
    static toObject(includeInstance: boolean, msg: DatumInfo): DatumInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DatumInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DatumInfo;
    static deserializeBinaryFromReader(message: DatumInfo, reader: jspb.BinaryReader): DatumInfo;
}

export namespace DatumInfo {
    export type AsObject = {
        datum?: Datum.AsObject,
        state: DatumState,
        stats?: ProcessStats.AsObject,
        pfsState?: pfs_pfs_pb.File.AsObject,
        dataList: Array<pfs_pfs_pb.FileInfo.AsObject>,
        imageId: string,
    }
}

export class Aggregate extends jspb.Message { 
    getCount(): number;
    setCount(value: number): Aggregate;
    getMean(): number;
    setMean(value: number): Aggregate;
    getStddev(): number;
    setStddev(value: number): Aggregate;
    getFifthPercentile(): number;
    setFifthPercentile(value: number): Aggregate;
    getNinetyFifthPercentile(): number;
    setNinetyFifthPercentile(value: number): Aggregate;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Aggregate.AsObject;
    static toObject(includeInstance: boolean, msg: Aggregate): Aggregate.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Aggregate, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Aggregate;
    static deserializeBinaryFromReader(message: Aggregate, reader: jspb.BinaryReader): Aggregate;
}

export namespace Aggregate {
    export type AsObject = {
        count: number,
        mean: number,
        stddev: number,
        fifthPercentile: number,
        ninetyFifthPercentile: number,
    }
}

export class ProcessStats extends jspb.Message { 

    hasDownloadTime(): boolean;
    clearDownloadTime(): void;
    getDownloadTime(): google_protobuf_duration_pb.Duration | undefined;
    setDownloadTime(value?: google_protobuf_duration_pb.Duration): ProcessStats;

    hasProcessTime(): boolean;
    clearProcessTime(): void;
    getProcessTime(): google_protobuf_duration_pb.Duration | undefined;
    setProcessTime(value?: google_protobuf_duration_pb.Duration): ProcessStats;

    hasUploadTime(): boolean;
    clearUploadTime(): void;
    getUploadTime(): google_protobuf_duration_pb.Duration | undefined;
    setUploadTime(value?: google_protobuf_duration_pb.Duration): ProcessStats;
    getDownloadBytes(): number;
    setDownloadBytes(value: number): ProcessStats;
    getUploadBytes(): number;
    setUploadBytes(value: number): ProcessStats;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProcessStats.AsObject;
    static toObject(includeInstance: boolean, msg: ProcessStats): ProcessStats.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProcessStats, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProcessStats;
    static deserializeBinaryFromReader(message: ProcessStats, reader: jspb.BinaryReader): ProcessStats;
}

export namespace ProcessStats {
    export type AsObject = {
        downloadTime?: google_protobuf_duration_pb.Duration.AsObject,
        processTime?: google_protobuf_duration_pb.Duration.AsObject,
        uploadTime?: google_protobuf_duration_pb.Duration.AsObject,
        downloadBytes: number,
        uploadBytes: number,
    }
}

export class AggregateProcessStats extends jspb.Message { 

    hasDownloadTime(): boolean;
    clearDownloadTime(): void;
    getDownloadTime(): Aggregate | undefined;
    setDownloadTime(value?: Aggregate): AggregateProcessStats;

    hasProcessTime(): boolean;
    clearProcessTime(): void;
    getProcessTime(): Aggregate | undefined;
    setProcessTime(value?: Aggregate): AggregateProcessStats;

    hasUploadTime(): boolean;
    clearUploadTime(): void;
    getUploadTime(): Aggregate | undefined;
    setUploadTime(value?: Aggregate): AggregateProcessStats;

    hasDownloadBytes(): boolean;
    clearDownloadBytes(): void;
    getDownloadBytes(): Aggregate | undefined;
    setDownloadBytes(value?: Aggregate): AggregateProcessStats;

    hasUploadBytes(): boolean;
    clearUploadBytes(): void;
    getUploadBytes(): Aggregate | undefined;
    setUploadBytes(value?: Aggregate): AggregateProcessStats;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AggregateProcessStats.AsObject;
    static toObject(includeInstance: boolean, msg: AggregateProcessStats): AggregateProcessStats.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AggregateProcessStats, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AggregateProcessStats;
    static deserializeBinaryFromReader(message: AggregateProcessStats, reader: jspb.BinaryReader): AggregateProcessStats;
}

export namespace AggregateProcessStats {
    export type AsObject = {
        downloadTime?: Aggregate.AsObject,
        processTime?: Aggregate.AsObject,
        uploadTime?: Aggregate.AsObject,
        downloadBytes?: Aggregate.AsObject,
        uploadBytes?: Aggregate.AsObject,
    }
}

export class WorkerStatus extends jspb.Message { 
    getWorkerId(): string;
    setWorkerId(value: string): WorkerStatus;
    getJobId(): string;
    setJobId(value: string): WorkerStatus;

    hasDatumStatus(): boolean;
    clearDatumStatus(): void;
    getDatumStatus(): DatumStatus | undefined;
    setDatumStatus(value?: DatumStatus): WorkerStatus;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WorkerStatus.AsObject;
    static toObject(includeInstance: boolean, msg: WorkerStatus): WorkerStatus.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WorkerStatus, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WorkerStatus;
    static deserializeBinaryFromReader(message: WorkerStatus, reader: jspb.BinaryReader): WorkerStatus;
}

export namespace WorkerStatus {
    export type AsObject = {
        workerId: string,
        jobId: string,
        datumStatus?: DatumStatus.AsObject,
    }
}

export class DatumStatus extends jspb.Message { 

    hasStarted(): boolean;
    clearStarted(): void;
    getStarted(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setStarted(value?: google_protobuf_timestamp_pb.Timestamp): DatumStatus;
    clearDataList(): void;
    getDataList(): Array<InputFile>;
    setDataList(value: Array<InputFile>): DatumStatus;
    addData(value?: InputFile, index?: number): InputFile;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DatumStatus.AsObject;
    static toObject(includeInstance: boolean, msg: DatumStatus): DatumStatus.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DatumStatus, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DatumStatus;
    static deserializeBinaryFromReader(message: DatumStatus, reader: jspb.BinaryReader): DatumStatus;
}

export namespace DatumStatus {
    export type AsObject = {
        started?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        dataList: Array<InputFile.AsObject>,
    }
}

export class ResourceSpec extends jspb.Message { 
    getCpu(): number;
    setCpu(value: number): ResourceSpec;
    getMemory(): string;
    setMemory(value: string): ResourceSpec;

    hasGpu(): boolean;
    clearGpu(): void;
    getGpu(): GPUSpec | undefined;
    setGpu(value?: GPUSpec): ResourceSpec;
    getDisk(): string;
    setDisk(value: string): ResourceSpec;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResourceSpec.AsObject;
    static toObject(includeInstance: boolean, msg: ResourceSpec): ResourceSpec.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResourceSpec, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResourceSpec;
    static deserializeBinaryFromReader(message: ResourceSpec, reader: jspb.BinaryReader): ResourceSpec;
}

export namespace ResourceSpec {
    export type AsObject = {
        cpu: number,
        memory: string,
        gpu?: GPUSpec.AsObject,
        disk: string,
    }
}

export class GPUSpec extends jspb.Message { 
    getType(): string;
    setType(value: string): GPUSpec;
    getNumber(): number;
    setNumber(value: number): GPUSpec;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GPUSpec.AsObject;
    static toObject(includeInstance: boolean, msg: GPUSpec): GPUSpec.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GPUSpec, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GPUSpec;
    static deserializeBinaryFromReader(message: GPUSpec, reader: jspb.BinaryReader): GPUSpec;
}

export namespace GPUSpec {
    export type AsObject = {
        type: string,
        number: number,
    }
}

export class JobSetInfo extends jspb.Message { 

    hasJobSet(): boolean;
    clearJobSet(): void;
    getJobSet(): JobSet | undefined;
    setJobSet(value?: JobSet): JobSetInfo;
    clearJobsList(): void;
    getJobsList(): Array<JobInfo>;
    setJobsList(value: Array<JobInfo>): JobSetInfo;
    addJobs(value?: JobInfo, index?: number): JobInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): JobSetInfo.AsObject;
    static toObject(includeInstance: boolean, msg: JobSetInfo): JobSetInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: JobSetInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): JobSetInfo;
    static deserializeBinaryFromReader(message: JobSetInfo, reader: jspb.BinaryReader): JobSetInfo;
}

export namespace JobSetInfo {
    export type AsObject = {
        jobSet?: JobSet.AsObject,
        jobsList: Array<JobInfo.AsObject>,
    }
}

export class JobInfo extends jspb.Message { 

    hasJob(): boolean;
    clearJob(): void;
    getJob(): Job | undefined;
    setJob(value?: Job): JobInfo;
    getPipelineVersion(): number;
    setPipelineVersion(value: number): JobInfo;

    hasOutputCommit(): boolean;
    clearOutputCommit(): void;
    getOutputCommit(): pfs_pfs_pb.Commit | undefined;
    setOutputCommit(value?: pfs_pfs_pb.Commit): JobInfo;
    getRestart(): number;
    setRestart(value: number): JobInfo;
    getDataProcessed(): number;
    setDataProcessed(value: number): JobInfo;
    getDataSkipped(): number;
    setDataSkipped(value: number): JobInfo;
    getDataTotal(): number;
    setDataTotal(value: number): JobInfo;
    getDataFailed(): number;
    setDataFailed(value: number): JobInfo;
    getDataRecovered(): number;
    setDataRecovered(value: number): JobInfo;

    hasStats(): boolean;
    clearStats(): void;
    getStats(): ProcessStats | undefined;
    setStats(value?: ProcessStats): JobInfo;
    getState(): JobState;
    setState(value: JobState): JobInfo;
    getReason(): string;
    setReason(value: string): JobInfo;

    hasCreated(): boolean;
    clearCreated(): void;
    getCreated(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setCreated(value?: google_protobuf_timestamp_pb.Timestamp): JobInfo;

    hasStarted(): boolean;
    clearStarted(): void;
    getStarted(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setStarted(value?: google_protobuf_timestamp_pb.Timestamp): JobInfo;

    hasFinished(): boolean;
    clearFinished(): void;
    getFinished(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setFinished(value?: google_protobuf_timestamp_pb.Timestamp): JobInfo;

    hasDetails(): boolean;
    clearDetails(): void;
    getDetails(): JobInfo.Details | undefined;
    setDetails(value?: JobInfo.Details): JobInfo;
    getAuthToken(): string;
    setAuthToken(value: string): JobInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): JobInfo.AsObject;
    static toObject(includeInstance: boolean, msg: JobInfo): JobInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: JobInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): JobInfo;
    static deserializeBinaryFromReader(message: JobInfo, reader: jspb.BinaryReader): JobInfo;
}

export namespace JobInfo {
    export type AsObject = {
        job?: Job.AsObject,
        pipelineVersion: number,
        outputCommit?: pfs_pfs_pb.Commit.AsObject,
        restart: number,
        dataProcessed: number,
        dataSkipped: number,
        dataTotal: number,
        dataFailed: number,
        dataRecovered: number,
        stats?: ProcessStats.AsObject,
        state: JobState,
        reason: string,
        created?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        started?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        finished?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        details?: JobInfo.Details.AsObject,
        authToken: string,
    }


    export class Details extends jspb.Message { 

        hasTransform(): boolean;
        clearTransform(): void;
        getTransform(): Transform | undefined;
        setTransform(value?: Transform): Details;

        hasParallelismSpec(): boolean;
        clearParallelismSpec(): void;
        getParallelismSpec(): ParallelismSpec | undefined;
        setParallelismSpec(value?: ParallelismSpec): Details;

        hasEgress(): boolean;
        clearEgress(): void;
        getEgress(): Egress | undefined;
        setEgress(value?: Egress): Details;

        hasService(): boolean;
        clearService(): void;
        getService(): Service | undefined;
        setService(value?: Service): Details;

        hasSpout(): boolean;
        clearSpout(): void;
        getSpout(): Spout | undefined;
        setSpout(value?: Spout): Details;
        clearWorkerStatusList(): void;
        getWorkerStatusList(): Array<WorkerStatus>;
        setWorkerStatusList(value: Array<WorkerStatus>): Details;
        addWorkerStatus(value?: WorkerStatus, index?: number): WorkerStatus;

        hasResourceRequests(): boolean;
        clearResourceRequests(): void;
        getResourceRequests(): ResourceSpec | undefined;
        setResourceRequests(value?: ResourceSpec): Details;

        hasResourceLimits(): boolean;
        clearResourceLimits(): void;
        getResourceLimits(): ResourceSpec | undefined;
        setResourceLimits(value?: ResourceSpec): Details;

        hasSidecarResourceLimits(): boolean;
        clearSidecarResourceLimits(): void;
        getSidecarResourceLimits(): ResourceSpec | undefined;
        setSidecarResourceLimits(value?: ResourceSpec): Details;

        hasInput(): boolean;
        clearInput(): void;
        getInput(): Input | undefined;
        setInput(value?: Input): Details;
        getSalt(): string;
        setSalt(value: string): Details;

        hasDatumSetSpec(): boolean;
        clearDatumSetSpec(): void;
        getDatumSetSpec(): DatumSetSpec | undefined;
        setDatumSetSpec(value?: DatumSetSpec): Details;

        hasDatumTimeout(): boolean;
        clearDatumTimeout(): void;
        getDatumTimeout(): google_protobuf_duration_pb.Duration | undefined;
        setDatumTimeout(value?: google_protobuf_duration_pb.Duration): Details;

        hasJobTimeout(): boolean;
        clearJobTimeout(): void;
        getJobTimeout(): google_protobuf_duration_pb.Duration | undefined;
        setJobTimeout(value?: google_protobuf_duration_pb.Duration): Details;
        getDatumTries(): number;
        setDatumTries(value: number): Details;

        hasSchedulingSpec(): boolean;
        clearSchedulingSpec(): void;
        getSchedulingSpec(): SchedulingSpec | undefined;
        setSchedulingSpec(value?: SchedulingSpec): Details;
        getPodSpec(): string;
        setPodSpec(value: string): Details;
        getPodPatch(): string;
        setPodPatch(value: string): Details;

        hasSidecarResourceRequests(): boolean;
        clearSidecarResourceRequests(): void;
        getSidecarResourceRequests(): ResourceSpec | undefined;
        setSidecarResourceRequests(value?: ResourceSpec): Details;

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
            transform?: Transform.AsObject,
            parallelismSpec?: ParallelismSpec.AsObject,
            egress?: Egress.AsObject,
            service?: Service.AsObject,
            spout?: Spout.AsObject,
            workerStatusList: Array<WorkerStatus.AsObject>,
            resourceRequests?: ResourceSpec.AsObject,
            resourceLimits?: ResourceSpec.AsObject,
            sidecarResourceLimits?: ResourceSpec.AsObject,
            input?: Input.AsObject,
            salt: string,
            datumSetSpec?: DatumSetSpec.AsObject,
            datumTimeout?: google_protobuf_duration_pb.Duration.AsObject,
            jobTimeout?: google_protobuf_duration_pb.Duration.AsObject,
            datumTries: number,
            schedulingSpec?: SchedulingSpec.AsObject,
            podSpec: string,
            podPatch: string,
            sidecarResourceRequests?: ResourceSpec.AsObject,
        }
    }

}

export class Worker extends jspb.Message { 
    getName(): string;
    setName(value: string): Worker;
    getState(): WorkerState;
    setState(value: WorkerState): Worker;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Worker.AsObject;
    static toObject(includeInstance: boolean, msg: Worker): Worker.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Worker, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Worker;
    static deserializeBinaryFromReader(message: Worker, reader: jspb.BinaryReader): Worker;
}

export namespace Worker {
    export type AsObject = {
        name: string,
        state: WorkerState,
    }
}

export class Pipeline extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): pfs_pfs_pb.Project | undefined;
    setProject(value?: pfs_pfs_pb.Project): Pipeline;
    getName(): string;
    setName(value: string): Pipeline;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Pipeline.AsObject;
    static toObject(includeInstance: boolean, msg: Pipeline): Pipeline.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Pipeline, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Pipeline;
    static deserializeBinaryFromReader(message: Pipeline, reader: jspb.BinaryReader): Pipeline;
}

export namespace Pipeline {
    export type AsObject = {
        project?: pfs_pfs_pb.Project.AsObject,
        name: string,
    }
}

export class Toleration extends jspb.Message { 
    getKey(): string;
    setKey(value: string): Toleration;
    getOperator(): TolerationOperator;
    setOperator(value: TolerationOperator): Toleration;
    getValue(): string;
    setValue(value: string): Toleration;
    getEffect(): TaintEffect;
    setEffect(value: TaintEffect): Toleration;

    hasTolerationSeconds(): boolean;
    clearTolerationSeconds(): void;
    getTolerationSeconds(): google_protobuf_wrappers_pb.Int64Value | undefined;
    setTolerationSeconds(value?: google_protobuf_wrappers_pb.Int64Value): Toleration;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Toleration.AsObject;
    static toObject(includeInstance: boolean, msg: Toleration): Toleration.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Toleration, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Toleration;
    static deserializeBinaryFromReader(message: Toleration, reader: jspb.BinaryReader): Toleration;
}

export namespace Toleration {
    export type AsObject = {
        key: string,
        operator: TolerationOperator,
        value: string,
        effect: TaintEffect,
        tolerationSeconds?: google_protobuf_wrappers_pb.Int64Value.AsObject,
    }
}

export class PipelineInfo extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): PipelineInfo;
    getVersion(): number;
    setVersion(value: number): PipelineInfo;

    hasSpecCommit(): boolean;
    clearSpecCommit(): void;
    getSpecCommit(): pfs_pfs_pb.Commit | undefined;
    setSpecCommit(value?: pfs_pfs_pb.Commit): PipelineInfo;
    getStopped(): boolean;
    setStopped(value: boolean): PipelineInfo;
    getState(): PipelineState;
    setState(value: PipelineState): PipelineInfo;
    getReason(): string;
    setReason(value: string): PipelineInfo;
    getLastJobState(): JobState;
    setLastJobState(value: JobState): PipelineInfo;
    getParallelism(): number;
    setParallelism(value: number): PipelineInfo;
    getType(): PipelineInfo.PipelineType;
    setType(value: PipelineInfo.PipelineType): PipelineInfo;
    getAuthToken(): string;
    setAuthToken(value: string): PipelineInfo;

    hasDetails(): boolean;
    clearDetails(): void;
    getDetails(): PipelineInfo.Details | undefined;
    setDetails(value?: PipelineInfo.Details): PipelineInfo;
    getUserSpecJson(): string;
    setUserSpecJson(value: string): PipelineInfo;
    getEffectiveSpecJson(): string;
    setEffectiveSpecJson(value: string): PipelineInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PipelineInfo.AsObject;
    static toObject(includeInstance: boolean, msg: PipelineInfo): PipelineInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PipelineInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PipelineInfo;
    static deserializeBinaryFromReader(message: PipelineInfo, reader: jspb.BinaryReader): PipelineInfo;
}

export namespace PipelineInfo {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        version: number,
        specCommit?: pfs_pfs_pb.Commit.AsObject,
        stopped: boolean,
        state: PipelineState,
        reason: string,
        lastJobState: JobState,
        parallelism: number,
        type: PipelineInfo.PipelineType,
        authToken: string,
        details?: PipelineInfo.Details.AsObject,
        userSpecJson: string,
        effectiveSpecJson: string,
    }


    export class Details extends jspb.Message { 

        hasTransform(): boolean;
        clearTransform(): void;
        getTransform(): Transform | undefined;
        setTransform(value?: Transform): Details;

        hasTfJob(): boolean;
        clearTfJob(): void;
        getTfJob(): TFJob | undefined;
        setTfJob(value?: TFJob): Details;

        hasParallelismSpec(): boolean;
        clearParallelismSpec(): void;
        getParallelismSpec(): ParallelismSpec | undefined;
        setParallelismSpec(value?: ParallelismSpec): Details;

        hasEgress(): boolean;
        clearEgress(): void;
        getEgress(): Egress | undefined;
        setEgress(value?: Egress): Details;

        hasCreatedAt(): boolean;
        clearCreatedAt(): void;
        getCreatedAt(): google_protobuf_timestamp_pb.Timestamp | undefined;
        setCreatedAt(value?: google_protobuf_timestamp_pb.Timestamp): Details;
        getRecentError(): string;
        setRecentError(value: string): Details;
        getWorkersRequested(): number;
        setWorkersRequested(value: number): Details;
        getWorkersAvailable(): number;
        setWorkersAvailable(value: number): Details;
        getOutputBranch(): string;
        setOutputBranch(value: string): Details;

        hasResourceRequests(): boolean;
        clearResourceRequests(): void;
        getResourceRequests(): ResourceSpec | undefined;
        setResourceRequests(value?: ResourceSpec): Details;

        hasResourceLimits(): boolean;
        clearResourceLimits(): void;
        getResourceLimits(): ResourceSpec | undefined;
        setResourceLimits(value?: ResourceSpec): Details;

        hasSidecarResourceLimits(): boolean;
        clearSidecarResourceLimits(): void;
        getSidecarResourceLimits(): ResourceSpec | undefined;
        setSidecarResourceLimits(value?: ResourceSpec): Details;

        hasInput(): boolean;
        clearInput(): void;
        getInput(): Input | undefined;
        setInput(value?: Input): Details;
        getDescription(): string;
        setDescription(value: string): Details;
        getSalt(): string;
        setSalt(value: string): Details;
        getReason(): string;
        setReason(value: string): Details;

        hasService(): boolean;
        clearService(): void;
        getService(): Service | undefined;
        setService(value?: Service): Details;

        hasSpout(): boolean;
        clearSpout(): void;
        getSpout(): Spout | undefined;
        setSpout(value?: Spout): Details;

        hasDatumSetSpec(): boolean;
        clearDatumSetSpec(): void;
        getDatumSetSpec(): DatumSetSpec | undefined;
        setDatumSetSpec(value?: DatumSetSpec): Details;

        hasDatumTimeout(): boolean;
        clearDatumTimeout(): void;
        getDatumTimeout(): google_protobuf_duration_pb.Duration | undefined;
        setDatumTimeout(value?: google_protobuf_duration_pb.Duration): Details;

        hasJobTimeout(): boolean;
        clearJobTimeout(): void;
        getJobTimeout(): google_protobuf_duration_pb.Duration | undefined;
        setJobTimeout(value?: google_protobuf_duration_pb.Duration): Details;
        getDatumTries(): number;
        setDatumTries(value: number): Details;

        hasSchedulingSpec(): boolean;
        clearSchedulingSpec(): void;
        getSchedulingSpec(): SchedulingSpec | undefined;
        setSchedulingSpec(value?: SchedulingSpec): Details;
        getPodSpec(): string;
        setPodSpec(value: string): Details;
        getPodPatch(): string;
        setPodPatch(value: string): Details;
        getS3Out(): boolean;
        setS3Out(value: boolean): Details;

        hasMetadata(): boolean;
        clearMetadata(): void;
        getMetadata(): Metadata | undefined;
        setMetadata(value?: Metadata): Details;
        getReprocessSpec(): string;
        setReprocessSpec(value: string): Details;
        getUnclaimedTasks(): number;
        setUnclaimedTasks(value: number): Details;
        getWorkerRc(): string;
        setWorkerRc(value: string): Details;
        getAutoscaling(): boolean;
        setAutoscaling(value: boolean): Details;
        clearTolerationsList(): void;
        getTolerationsList(): Array<Toleration>;
        setTolerationsList(value: Array<Toleration>): Details;
        addTolerations(value?: Toleration, index?: number): Toleration;

        hasSidecarResourceRequests(): boolean;
        clearSidecarResourceRequests(): void;
        getSidecarResourceRequests(): ResourceSpec | undefined;
        setSidecarResourceRequests(value?: ResourceSpec): Details;

        hasDetermined(): boolean;
        clearDetermined(): void;
        getDetermined(): Determined | undefined;
        setDetermined(value?: Determined): Details;

        hasMaximumExpectedUptime(): boolean;
        clearMaximumExpectedUptime(): void;
        getMaximumExpectedUptime(): google_protobuf_duration_pb.Duration | undefined;
        setMaximumExpectedUptime(value?: google_protobuf_duration_pb.Duration): Details;

        hasWorkersStartedAt(): boolean;
        clearWorkersStartedAt(): void;
        getWorkersStartedAt(): google_protobuf_timestamp_pb.Timestamp | undefined;
        setWorkersStartedAt(value?: google_protobuf_timestamp_pb.Timestamp): Details;

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
            transform?: Transform.AsObject,
            tfJob?: TFJob.AsObject,
            parallelismSpec?: ParallelismSpec.AsObject,
            egress?: Egress.AsObject,
            createdAt?: google_protobuf_timestamp_pb.Timestamp.AsObject,
            recentError: string,
            workersRequested: number,
            workersAvailable: number,
            outputBranch: string,
            resourceRequests?: ResourceSpec.AsObject,
            resourceLimits?: ResourceSpec.AsObject,
            sidecarResourceLimits?: ResourceSpec.AsObject,
            input?: Input.AsObject,
            description: string,
            salt: string,
            reason: string,
            service?: Service.AsObject,
            spout?: Spout.AsObject,
            datumSetSpec?: DatumSetSpec.AsObject,
            datumTimeout?: google_protobuf_duration_pb.Duration.AsObject,
            jobTimeout?: google_protobuf_duration_pb.Duration.AsObject,
            datumTries: number,
            schedulingSpec?: SchedulingSpec.AsObject,
            podSpec: string,
            podPatch: string,
            s3Out: boolean,
            metadata?: Metadata.AsObject,
            reprocessSpec: string,
            unclaimedTasks: number,
            workerRc: string,
            autoscaling: boolean,
            tolerationsList: Array<Toleration.AsObject>,
            sidecarResourceRequests?: ResourceSpec.AsObject,
            determined?: Determined.AsObject,
            maximumExpectedUptime?: google_protobuf_duration_pb.Duration.AsObject,
            workersStartedAt?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        }
    }


    export enum PipelineType {
    PIPELINT_TYPE_UNKNOWN = 0,
    PIPELINE_TYPE_TRANSFORM = 1,
    PIPELINE_TYPE_SPOUT = 2,
    PIPELINE_TYPE_SERVICE = 3,
    }

}

export class PipelineInfos extends jspb.Message { 
    clearPipelineInfoList(): void;
    getPipelineInfoList(): Array<PipelineInfo>;
    setPipelineInfoList(value: Array<PipelineInfo>): PipelineInfos;
    addPipelineInfo(value?: PipelineInfo, index?: number): PipelineInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PipelineInfos.AsObject;
    static toObject(includeInstance: boolean, msg: PipelineInfos): PipelineInfos.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PipelineInfos, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PipelineInfos;
    static deserializeBinaryFromReader(message: PipelineInfos, reader: jspb.BinaryReader): PipelineInfos;
}

export namespace PipelineInfos {
    export type AsObject = {
        pipelineInfoList: Array<PipelineInfo.AsObject>,
    }
}

export class JobSet extends jspb.Message { 
    getId(): string;
    setId(value: string): JobSet;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): JobSet.AsObject;
    static toObject(includeInstance: boolean, msg: JobSet): JobSet.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: JobSet, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): JobSet;
    static deserializeBinaryFromReader(message: JobSet, reader: jspb.BinaryReader): JobSet;
}

export namespace JobSet {
    export type AsObject = {
        id: string,
    }
}

export class InspectJobSetRequest extends jspb.Message { 

    hasJobSet(): boolean;
    clearJobSet(): void;
    getJobSet(): JobSet | undefined;
    setJobSet(value?: JobSet): InspectJobSetRequest;
    getWait(): boolean;
    setWait(value: boolean): InspectJobSetRequest;
    getDetails(): boolean;
    setDetails(value: boolean): InspectJobSetRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectJobSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectJobSetRequest): InspectJobSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectJobSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectJobSetRequest;
    static deserializeBinaryFromReader(message: InspectJobSetRequest, reader: jspb.BinaryReader): InspectJobSetRequest;
}

export namespace InspectJobSetRequest {
    export type AsObject = {
        jobSet?: JobSet.AsObject,
        wait: boolean,
        details: boolean,
    }
}

export class ListJobSetRequest extends jspb.Message { 
    getDetails(): boolean;
    setDetails(value: boolean): ListJobSetRequest;
    clearProjectsList(): void;
    getProjectsList(): Array<pfs_pfs_pb.Project>;
    setProjectsList(value: Array<pfs_pfs_pb.Project>): ListJobSetRequest;
    addProjects(value?: pfs_pfs_pb.Project, index?: number): pfs_pfs_pb.Project;

    hasPaginationmarker(): boolean;
    clearPaginationmarker(): void;
    getPaginationmarker(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setPaginationmarker(value?: google_protobuf_timestamp_pb.Timestamp): ListJobSetRequest;
    getNumber(): number;
    setNumber(value: number): ListJobSetRequest;
    getReverse(): boolean;
    setReverse(value: boolean): ListJobSetRequest;
    getJqfilter(): string;
    setJqfilter(value: string): ListJobSetRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListJobSetRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListJobSetRequest): ListJobSetRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListJobSetRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListJobSetRequest;
    static deserializeBinaryFromReader(message: ListJobSetRequest, reader: jspb.BinaryReader): ListJobSetRequest;
}

export namespace ListJobSetRequest {
    export type AsObject = {
        details: boolean,
        projectsList: Array<pfs_pfs_pb.Project.AsObject>,
        paginationmarker?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        number: number,
        reverse: boolean,
        jqfilter: string,
    }
}

export class InspectJobRequest extends jspb.Message { 

    hasJob(): boolean;
    clearJob(): void;
    getJob(): Job | undefined;
    setJob(value?: Job): InspectJobRequest;
    getWait(): boolean;
    setWait(value: boolean): InspectJobRequest;
    getDetails(): boolean;
    setDetails(value: boolean): InspectJobRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectJobRequest): InspectJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectJobRequest;
    static deserializeBinaryFromReader(message: InspectJobRequest, reader: jspb.BinaryReader): InspectJobRequest;
}

export namespace InspectJobRequest {
    export type AsObject = {
        job?: Job.AsObject,
        wait: boolean,
        details: boolean,
    }
}

export class ListJobRequest extends jspb.Message { 
    clearProjectsList(): void;
    getProjectsList(): Array<pfs_pfs_pb.Project>;
    setProjectsList(value: Array<pfs_pfs_pb.Project>): ListJobRequest;
    addProjects(value?: pfs_pfs_pb.Project, index?: number): pfs_pfs_pb.Project;

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): ListJobRequest;
    clearInputCommitList(): void;
    getInputCommitList(): Array<pfs_pfs_pb.Commit>;
    setInputCommitList(value: Array<pfs_pfs_pb.Commit>): ListJobRequest;
    addInputCommit(value?: pfs_pfs_pb.Commit, index?: number): pfs_pfs_pb.Commit;
    getHistory(): number;
    setHistory(value: number): ListJobRequest;
    getDetails(): boolean;
    setDetails(value: boolean): ListJobRequest;
    getJqfilter(): string;
    setJqfilter(value: string): ListJobRequest;

    hasPaginationmarker(): boolean;
    clearPaginationmarker(): void;
    getPaginationmarker(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setPaginationmarker(value?: google_protobuf_timestamp_pb.Timestamp): ListJobRequest;
    getNumber(): number;
    setNumber(value: number): ListJobRequest;
    getReverse(): boolean;
    setReverse(value: boolean): ListJobRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListJobRequest): ListJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListJobRequest;
    static deserializeBinaryFromReader(message: ListJobRequest, reader: jspb.BinaryReader): ListJobRequest;
}

export namespace ListJobRequest {
    export type AsObject = {
        projectsList: Array<pfs_pfs_pb.Project.AsObject>,
        pipeline?: Pipeline.AsObject,
        inputCommitList: Array<pfs_pfs_pb.Commit.AsObject>,
        history: number,
        details: boolean,
        jqfilter: string,
        paginationmarker?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        number: number,
        reverse: boolean,
    }
}

export class SubscribeJobRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): SubscribeJobRequest;
    getDetails(): boolean;
    setDetails(value: boolean): SubscribeJobRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SubscribeJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SubscribeJobRequest): SubscribeJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SubscribeJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SubscribeJobRequest;
    static deserializeBinaryFromReader(message: SubscribeJobRequest, reader: jspb.BinaryReader): SubscribeJobRequest;
}

export namespace SubscribeJobRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        details: boolean,
    }
}

export class DeleteJobRequest extends jspb.Message { 

    hasJob(): boolean;
    clearJob(): void;
    getJob(): Job | undefined;
    setJob(value?: Job): DeleteJobRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteJobRequest): DeleteJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteJobRequest;
    static deserializeBinaryFromReader(message: DeleteJobRequest, reader: jspb.BinaryReader): DeleteJobRequest;
}

export namespace DeleteJobRequest {
    export type AsObject = {
        job?: Job.AsObject,
    }
}

export class StopJobRequest extends jspb.Message { 

    hasJob(): boolean;
    clearJob(): void;
    getJob(): Job | undefined;
    setJob(value?: Job): StopJobRequest;
    getReason(): string;
    setReason(value: string): StopJobRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StopJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: StopJobRequest): StopJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StopJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StopJobRequest;
    static deserializeBinaryFromReader(message: StopJobRequest, reader: jspb.BinaryReader): StopJobRequest;
}

export namespace StopJobRequest {
    export type AsObject = {
        job?: Job.AsObject,
        reason: string,
    }
}

export class UpdateJobStateRequest extends jspb.Message { 

    hasJob(): boolean;
    clearJob(): void;
    getJob(): Job | undefined;
    setJob(value?: Job): UpdateJobStateRequest;
    getState(): JobState;
    setState(value: JobState): UpdateJobStateRequest;
    getReason(): string;
    setReason(value: string): UpdateJobStateRequest;
    getRestart(): number;
    setRestart(value: number): UpdateJobStateRequest;
    getDataProcessed(): number;
    setDataProcessed(value: number): UpdateJobStateRequest;
    getDataSkipped(): number;
    setDataSkipped(value: number): UpdateJobStateRequest;
    getDataFailed(): number;
    setDataFailed(value: number): UpdateJobStateRequest;
    getDataRecovered(): number;
    setDataRecovered(value: number): UpdateJobStateRequest;
    getDataTotal(): number;
    setDataTotal(value: number): UpdateJobStateRequest;

    hasStats(): boolean;
    clearStats(): void;
    getStats(): ProcessStats | undefined;
    setStats(value?: ProcessStats): UpdateJobStateRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UpdateJobStateRequest.AsObject;
    static toObject(includeInstance: boolean, msg: UpdateJobStateRequest): UpdateJobStateRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UpdateJobStateRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UpdateJobStateRequest;
    static deserializeBinaryFromReader(message: UpdateJobStateRequest, reader: jspb.BinaryReader): UpdateJobStateRequest;
}

export namespace UpdateJobStateRequest {
    export type AsObject = {
        job?: Job.AsObject,
        state: JobState,
        reason: string,
        restart: number,
        dataProcessed: number,
        dataSkipped: number,
        dataFailed: number,
        dataRecovered: number,
        dataTotal: number,
        stats?: ProcessStats.AsObject,
    }
}

export class GetLogsRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): GetLogsRequest;

    hasJob(): boolean;
    clearJob(): void;
    getJob(): Job | undefined;
    setJob(value?: Job): GetLogsRequest;
    clearDataFiltersList(): void;
    getDataFiltersList(): Array<string>;
    setDataFiltersList(value: Array<string>): GetLogsRequest;
    addDataFilters(value: string, index?: number): string;

    hasDatum(): boolean;
    clearDatum(): void;
    getDatum(): Datum | undefined;
    setDatum(value?: Datum): GetLogsRequest;
    getMaster(): boolean;
    setMaster(value: boolean): GetLogsRequest;
    getFollow(): boolean;
    setFollow(value: boolean): GetLogsRequest;
    getTail(): number;
    setTail(value: number): GetLogsRequest;
    getUseLokiBackend(): boolean;
    setUseLokiBackend(value: boolean): GetLogsRequest;

    hasSince(): boolean;
    clearSince(): void;
    getSince(): google_protobuf_duration_pb.Duration | undefined;
    setSince(value?: google_protobuf_duration_pb.Duration): GetLogsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetLogsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetLogsRequest): GetLogsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetLogsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetLogsRequest;
    static deserializeBinaryFromReader(message: GetLogsRequest, reader: jspb.BinaryReader): GetLogsRequest;
}

export namespace GetLogsRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        job?: Job.AsObject,
        dataFiltersList: Array<string>,
        datum?: Datum.AsObject,
        master: boolean,
        follow: boolean,
        tail: number,
        useLokiBackend: boolean,
        since?: google_protobuf_duration_pb.Duration.AsObject,
    }
}

export class LogMessage extends jspb.Message { 
    getProjectName(): string;
    setProjectName(value: string): LogMessage;
    getPipelineName(): string;
    setPipelineName(value: string): LogMessage;
    getJobId(): string;
    setJobId(value: string): LogMessage;
    getWorkerId(): string;
    setWorkerId(value: string): LogMessage;
    getDatumId(): string;
    setDatumId(value: string): LogMessage;
    getMaster(): boolean;
    setMaster(value: boolean): LogMessage;
    clearDataList(): void;
    getDataList(): Array<InputFile>;
    setDataList(value: Array<InputFile>): LogMessage;
    addData(value?: InputFile, index?: number): InputFile;
    getUser(): boolean;
    setUser(value: boolean): LogMessage;

    hasTs(): boolean;
    clearTs(): void;
    getTs(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTs(value?: google_protobuf_timestamp_pb.Timestamp): LogMessage;
    getMessage(): string;
    setMessage(value: string): LogMessage;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): LogMessage.AsObject;
    static toObject(includeInstance: boolean, msg: LogMessage): LogMessage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: LogMessage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): LogMessage;
    static deserializeBinaryFromReader(message: LogMessage, reader: jspb.BinaryReader): LogMessage;
}

export namespace LogMessage {
    export type AsObject = {
        projectName: string,
        pipelineName: string,
        jobId: string,
        workerId: string,
        datumId: string,
        master: boolean,
        dataList: Array<InputFile.AsObject>,
        user: boolean,
        ts?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        message: string,
    }
}

export class RestartDatumRequest extends jspb.Message { 

    hasJob(): boolean;
    clearJob(): void;
    getJob(): Job | undefined;
    setJob(value?: Job): RestartDatumRequest;
    clearDataFiltersList(): void;
    getDataFiltersList(): Array<string>;
    setDataFiltersList(value: Array<string>): RestartDatumRequest;
    addDataFilters(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RestartDatumRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RestartDatumRequest): RestartDatumRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RestartDatumRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RestartDatumRequest;
    static deserializeBinaryFromReader(message: RestartDatumRequest, reader: jspb.BinaryReader): RestartDatumRequest;
}

export namespace RestartDatumRequest {
    export type AsObject = {
        job?: Job.AsObject,
        dataFiltersList: Array<string>,
    }
}

export class InspectDatumRequest extends jspb.Message { 

    hasDatum(): boolean;
    clearDatum(): void;
    getDatum(): Datum | undefined;
    setDatum(value?: Datum): InspectDatumRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectDatumRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectDatumRequest): InspectDatumRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectDatumRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectDatumRequest;
    static deserializeBinaryFromReader(message: InspectDatumRequest, reader: jspb.BinaryReader): InspectDatumRequest;
}

export namespace InspectDatumRequest {
    export type AsObject = {
        datum?: Datum.AsObject,
    }
}

export class ListDatumRequest extends jspb.Message { 

    hasJob(): boolean;
    clearJob(): void;
    getJob(): Job | undefined;
    setJob(value?: Job): ListDatumRequest;

    hasInput(): boolean;
    clearInput(): void;
    getInput(): Input | undefined;
    setInput(value?: Input): ListDatumRequest;

    hasFilter(): boolean;
    clearFilter(): void;
    getFilter(): ListDatumRequest.Filter | undefined;
    setFilter(value?: ListDatumRequest.Filter): ListDatumRequest;
    getPaginationmarker(): string;
    setPaginationmarker(value: string): ListDatumRequest;
    getNumber(): number;
    setNumber(value: number): ListDatumRequest;
    getReverse(): boolean;
    setReverse(value: boolean): ListDatumRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListDatumRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListDatumRequest): ListDatumRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListDatumRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListDatumRequest;
    static deserializeBinaryFromReader(message: ListDatumRequest, reader: jspb.BinaryReader): ListDatumRequest;
}

export namespace ListDatumRequest {
    export type AsObject = {
        job?: Job.AsObject,
        input?: Input.AsObject,
        filter?: ListDatumRequest.Filter.AsObject,
        paginationmarker: string,
        number: number,
        reverse: boolean,
    }


    export class Filter extends jspb.Message { 
        clearStateList(): void;
        getStateList(): Array<DatumState>;
        setStateList(value: Array<DatumState>): Filter;
        addState(value: DatumState, index?: number): DatumState;

        serializeBinary(): Uint8Array;
        toObject(includeInstance?: boolean): Filter.AsObject;
        static toObject(includeInstance: boolean, msg: Filter): Filter.AsObject;
        static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
        static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
        static serializeBinaryToWriter(message: Filter, writer: jspb.BinaryWriter): void;
        static deserializeBinary(bytes: Uint8Array): Filter;
        static deserializeBinaryFromReader(message: Filter, reader: jspb.BinaryReader): Filter;
    }

    export namespace Filter {
        export type AsObject = {
            stateList: Array<DatumState>,
        }
    }

}

export class StartCreateDatumRequest extends jspb.Message { 

    hasInput(): boolean;
    clearInput(): void;
    getInput(): Input | undefined;
    setInput(value?: Input): StartCreateDatumRequest;
    getNumber(): number;
    setNumber(value: number): StartCreateDatumRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StartCreateDatumRequest.AsObject;
    static toObject(includeInstance: boolean, msg: StartCreateDatumRequest): StartCreateDatumRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StartCreateDatumRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StartCreateDatumRequest;
    static deserializeBinaryFromReader(message: StartCreateDatumRequest, reader: jspb.BinaryReader): StartCreateDatumRequest;
}

export namespace StartCreateDatumRequest {
    export type AsObject = {
        input?: Input.AsObject,
        number: number,
    }
}

export class ContinueCreateDatumRequest extends jspb.Message { 
    getNumber(): number;
    setNumber(value: number): ContinueCreateDatumRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ContinueCreateDatumRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ContinueCreateDatumRequest): ContinueCreateDatumRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ContinueCreateDatumRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ContinueCreateDatumRequest;
    static deserializeBinaryFromReader(message: ContinueCreateDatumRequest, reader: jspb.BinaryReader): ContinueCreateDatumRequest;
}

export namespace ContinueCreateDatumRequest {
    export type AsObject = {
        number: number,
    }
}

export class CreateDatumRequest extends jspb.Message { 

    hasStart(): boolean;
    clearStart(): void;
    getStart(): StartCreateDatumRequest | undefined;
    setStart(value?: StartCreateDatumRequest): CreateDatumRequest;

    hasContinue(): boolean;
    clearContinue(): void;
    getContinue(): ContinueCreateDatumRequest | undefined;
    setContinue(value?: ContinueCreateDatumRequest): CreateDatumRequest;

    getBodyCase(): CreateDatumRequest.BodyCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateDatumRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateDatumRequest): CreateDatumRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateDatumRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateDatumRequest;
    static deserializeBinaryFromReader(message: CreateDatumRequest, reader: jspb.BinaryReader): CreateDatumRequest;
}

export namespace CreateDatumRequest {
    export type AsObject = {
        start?: StartCreateDatumRequest.AsObject,
        pb_continue?: ContinueCreateDatumRequest.AsObject,
    }

    export enum BodyCase {
        BODY_NOT_SET = 0,
        START = 1,
        CONTINUE = 2,
    }

}

export class DatumSetSpec extends jspb.Message { 
    getNumber(): number;
    setNumber(value: number): DatumSetSpec;
    getSizeBytes(): number;
    setSizeBytes(value: number): DatumSetSpec;
    getPerWorker(): number;
    setPerWorker(value: number): DatumSetSpec;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DatumSetSpec.AsObject;
    static toObject(includeInstance: boolean, msg: DatumSetSpec): DatumSetSpec.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DatumSetSpec, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DatumSetSpec;
    static deserializeBinaryFromReader(message: DatumSetSpec, reader: jspb.BinaryReader): DatumSetSpec;
}

export namespace DatumSetSpec {
    export type AsObject = {
        number: number,
        sizeBytes: number,
        perWorker: number,
    }
}

export class SchedulingSpec extends jspb.Message { 

    getNodeSelectorMap(): jspb.Map<string, string>;
    clearNodeSelectorMap(): void;
    getPriorityClassName(): string;
    setPriorityClassName(value: string): SchedulingSpec;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SchedulingSpec.AsObject;
    static toObject(includeInstance: boolean, msg: SchedulingSpec): SchedulingSpec.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SchedulingSpec, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SchedulingSpec;
    static deserializeBinaryFromReader(message: SchedulingSpec, reader: jspb.BinaryReader): SchedulingSpec;
}

export namespace SchedulingSpec {
    export type AsObject = {

        nodeSelectorMap: Array<[string, string]>,
        priorityClassName: string,
    }
}

export class RerunPipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): RerunPipelineRequest;
    getReprocess(): boolean;
    setReprocess(value: boolean): RerunPipelineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RerunPipelineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RerunPipelineRequest): RerunPipelineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RerunPipelineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RerunPipelineRequest;
    static deserializeBinaryFromReader(message: RerunPipelineRequest, reader: jspb.BinaryReader): RerunPipelineRequest;
}

export namespace RerunPipelineRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        reprocess: boolean,
    }
}

export class CreatePipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): CreatePipelineRequest;

    hasTfJob(): boolean;
    clearTfJob(): void;
    getTfJob(): TFJob | undefined;
    setTfJob(value?: TFJob): CreatePipelineRequest;

    hasTransform(): boolean;
    clearTransform(): void;
    getTransform(): Transform | undefined;
    setTransform(value?: Transform): CreatePipelineRequest;

    hasParallelismSpec(): boolean;
    clearParallelismSpec(): void;
    getParallelismSpec(): ParallelismSpec | undefined;
    setParallelismSpec(value?: ParallelismSpec): CreatePipelineRequest;

    hasEgress(): boolean;
    clearEgress(): void;
    getEgress(): Egress | undefined;
    setEgress(value?: Egress): CreatePipelineRequest;
    getUpdate(): boolean;
    setUpdate(value: boolean): CreatePipelineRequest;
    getOutputBranch(): string;
    setOutputBranch(value: string): CreatePipelineRequest;
    getS3Out(): boolean;
    setS3Out(value: boolean): CreatePipelineRequest;

    hasResourceRequests(): boolean;
    clearResourceRequests(): void;
    getResourceRequests(): ResourceSpec | undefined;
    setResourceRequests(value?: ResourceSpec): CreatePipelineRequest;

    hasResourceLimits(): boolean;
    clearResourceLimits(): void;
    getResourceLimits(): ResourceSpec | undefined;
    setResourceLimits(value?: ResourceSpec): CreatePipelineRequest;

    hasSidecarResourceLimits(): boolean;
    clearSidecarResourceLimits(): void;
    getSidecarResourceLimits(): ResourceSpec | undefined;
    setSidecarResourceLimits(value?: ResourceSpec): CreatePipelineRequest;

    hasInput(): boolean;
    clearInput(): void;
    getInput(): Input | undefined;
    setInput(value?: Input): CreatePipelineRequest;
    getDescription(): string;
    setDescription(value: string): CreatePipelineRequest;
    getReprocess(): boolean;
    setReprocess(value: boolean): CreatePipelineRequest;

    hasService(): boolean;
    clearService(): void;
    getService(): Service | undefined;
    setService(value?: Service): CreatePipelineRequest;

    hasSpout(): boolean;
    clearSpout(): void;
    getSpout(): Spout | undefined;
    setSpout(value?: Spout): CreatePipelineRequest;

    hasDatumSetSpec(): boolean;
    clearDatumSetSpec(): void;
    getDatumSetSpec(): DatumSetSpec | undefined;
    setDatumSetSpec(value?: DatumSetSpec): CreatePipelineRequest;

    hasDatumTimeout(): boolean;
    clearDatumTimeout(): void;
    getDatumTimeout(): google_protobuf_duration_pb.Duration | undefined;
    setDatumTimeout(value?: google_protobuf_duration_pb.Duration): CreatePipelineRequest;

    hasJobTimeout(): boolean;
    clearJobTimeout(): void;
    getJobTimeout(): google_protobuf_duration_pb.Duration | undefined;
    setJobTimeout(value?: google_protobuf_duration_pb.Duration): CreatePipelineRequest;
    getSalt(): string;
    setSalt(value: string): CreatePipelineRequest;
    getDatumTries(): number;
    setDatumTries(value: number): CreatePipelineRequest;

    hasSchedulingSpec(): boolean;
    clearSchedulingSpec(): void;
    getSchedulingSpec(): SchedulingSpec | undefined;
    setSchedulingSpec(value?: SchedulingSpec): CreatePipelineRequest;
    getPodSpec(): string;
    setPodSpec(value: string): CreatePipelineRequest;
    getPodPatch(): string;
    setPodPatch(value: string): CreatePipelineRequest;

    hasSpecCommit(): boolean;
    clearSpecCommit(): void;
    getSpecCommit(): pfs_pfs_pb.Commit | undefined;
    setSpecCommit(value?: pfs_pfs_pb.Commit): CreatePipelineRequest;

    hasMetadata(): boolean;
    clearMetadata(): void;
    getMetadata(): Metadata | undefined;
    setMetadata(value?: Metadata): CreatePipelineRequest;
    getReprocessSpec(): string;
    setReprocessSpec(value: string): CreatePipelineRequest;
    getAutoscaling(): boolean;
    setAutoscaling(value: boolean): CreatePipelineRequest;
    clearTolerationsList(): void;
    getTolerationsList(): Array<Toleration>;
    setTolerationsList(value: Array<Toleration>): CreatePipelineRequest;
    addTolerations(value?: Toleration, index?: number): Toleration;

    hasSidecarResourceRequests(): boolean;
    clearSidecarResourceRequests(): void;
    getSidecarResourceRequests(): ResourceSpec | undefined;
    setSidecarResourceRequests(value?: ResourceSpec): CreatePipelineRequest;
    getDryRun(): boolean;
    setDryRun(value: boolean): CreatePipelineRequest;

    hasDetermined(): boolean;
    clearDetermined(): void;
    getDetermined(): Determined | undefined;
    setDetermined(value?: Determined): CreatePipelineRequest;

    hasMaximumExpectedUptime(): boolean;
    clearMaximumExpectedUptime(): void;
    getMaximumExpectedUptime(): google_protobuf_duration_pb.Duration | undefined;
    setMaximumExpectedUptime(value?: google_protobuf_duration_pb.Duration): CreatePipelineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreatePipelineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreatePipelineRequest): CreatePipelineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreatePipelineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreatePipelineRequest;
    static deserializeBinaryFromReader(message: CreatePipelineRequest, reader: jspb.BinaryReader): CreatePipelineRequest;
}

export namespace CreatePipelineRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        tfJob?: TFJob.AsObject,
        transform?: Transform.AsObject,
        parallelismSpec?: ParallelismSpec.AsObject,
        egress?: Egress.AsObject,
        update: boolean,
        outputBranch: string,
        s3Out: boolean,
        resourceRequests?: ResourceSpec.AsObject,
        resourceLimits?: ResourceSpec.AsObject,
        sidecarResourceLimits?: ResourceSpec.AsObject,
        input?: Input.AsObject,
        description: string,
        reprocess: boolean,
        service?: Service.AsObject,
        spout?: Spout.AsObject,
        datumSetSpec?: DatumSetSpec.AsObject,
        datumTimeout?: google_protobuf_duration_pb.Duration.AsObject,
        jobTimeout?: google_protobuf_duration_pb.Duration.AsObject,
        salt: string,
        datumTries: number,
        schedulingSpec?: SchedulingSpec.AsObject,
        podSpec: string,
        podPatch: string,
        specCommit?: pfs_pfs_pb.Commit.AsObject,
        metadata?: Metadata.AsObject,
        reprocessSpec: string,
        autoscaling: boolean,
        tolerationsList: Array<Toleration.AsObject>,
        sidecarResourceRequests?: ResourceSpec.AsObject,
        dryRun: boolean,
        determined?: Determined.AsObject,
        maximumExpectedUptime?: google_protobuf_duration_pb.Duration.AsObject,
    }
}

export class CreatePipelineV2Request extends jspb.Message { 
    getCreatePipelineRequestJson(): string;
    setCreatePipelineRequestJson(value: string): CreatePipelineV2Request;
    getDryRun(): boolean;
    setDryRun(value: boolean): CreatePipelineV2Request;
    getUpdate(): boolean;
    setUpdate(value: boolean): CreatePipelineV2Request;
    getReprocess(): boolean;
    setReprocess(value: boolean): CreatePipelineV2Request;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreatePipelineV2Request.AsObject;
    static toObject(includeInstance: boolean, msg: CreatePipelineV2Request): CreatePipelineV2Request.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreatePipelineV2Request, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreatePipelineV2Request;
    static deserializeBinaryFromReader(message: CreatePipelineV2Request, reader: jspb.BinaryReader): CreatePipelineV2Request;
}

export namespace CreatePipelineV2Request {
    export type AsObject = {
        createPipelineRequestJson: string,
        dryRun: boolean,
        update: boolean,
        reprocess: boolean,
    }
}

export class CreatePipelineV2Response extends jspb.Message { 
    getEffectiveCreatePipelineRequestJson(): string;
    setEffectiveCreatePipelineRequestJson(value: string): CreatePipelineV2Response;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreatePipelineV2Response.AsObject;
    static toObject(includeInstance: boolean, msg: CreatePipelineV2Response): CreatePipelineV2Response.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreatePipelineV2Response, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreatePipelineV2Response;
    static deserializeBinaryFromReader(message: CreatePipelineV2Response, reader: jspb.BinaryReader): CreatePipelineV2Response;
}

export namespace CreatePipelineV2Response {
    export type AsObject = {
        effectiveCreatePipelineRequestJson: string,
    }
}

export class InspectPipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): InspectPipelineRequest;
    getDetails(): boolean;
    setDetails(value: boolean): InspectPipelineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectPipelineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectPipelineRequest): InspectPipelineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectPipelineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectPipelineRequest;
    static deserializeBinaryFromReader(message: InspectPipelineRequest, reader: jspb.BinaryReader): InspectPipelineRequest;
}

export namespace InspectPipelineRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        details: boolean,
    }
}

export class ListPipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): ListPipelineRequest;
    getHistory(): number;
    setHistory(value: number): ListPipelineRequest;
    getDetails(): boolean;
    setDetails(value: boolean): ListPipelineRequest;
    getJqfilter(): string;
    setJqfilter(value: string): ListPipelineRequest;

    hasCommitSet(): boolean;
    clearCommitSet(): void;
    getCommitSet(): pfs_pfs_pb.CommitSet | undefined;
    setCommitSet(value?: pfs_pfs_pb.CommitSet): ListPipelineRequest;
    clearProjectsList(): void;
    getProjectsList(): Array<pfs_pfs_pb.Project>;
    setProjectsList(value: Array<pfs_pfs_pb.Project>): ListPipelineRequest;
    addProjects(value?: pfs_pfs_pb.Project, index?: number): pfs_pfs_pb.Project;

    hasPage(): boolean;
    clearPage(): void;
    getPage(): PipelinePage | undefined;
    setPage(value?: PipelinePage): ListPipelineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListPipelineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListPipelineRequest): ListPipelineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListPipelineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListPipelineRequest;
    static deserializeBinaryFromReader(message: ListPipelineRequest, reader: jspb.BinaryReader): ListPipelineRequest;
}

export namespace ListPipelineRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        history: number,
        details: boolean,
        jqfilter: string,
        commitSet?: pfs_pfs_pb.CommitSet.AsObject,
        projectsList: Array<pfs_pfs_pb.Project.AsObject>,
        page?: PipelinePage.AsObject,
    }
}

export class PipelinePage extends jspb.Message { 
    getOrder(): PipelinePage.Ordering;
    setOrder(value: PipelinePage.Ordering): PipelinePage;
    getPageSize(): number;
    setPageSize(value: number): PipelinePage;
    getPageIndex(): number;
    setPageIndex(value: number): PipelinePage;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PipelinePage.AsObject;
    static toObject(includeInstance: boolean, msg: PipelinePage): PipelinePage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PipelinePage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PipelinePage;
    static deserializeBinaryFromReader(message: PipelinePage, reader: jspb.BinaryReader): PipelinePage;
}

export namespace PipelinePage {
    export type AsObject = {
        order: PipelinePage.Ordering,
        pageSize: number,
        pageIndex: number,
    }

    export enum Ordering {
    RECENT = 0,
    }

}

export class DeletePipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): DeletePipelineRequest;
    getAll(): boolean;
    setAll(value: boolean): DeletePipelineRequest;
    getForce(): boolean;
    setForce(value: boolean): DeletePipelineRequest;
    getKeepRepo(): boolean;
    setKeepRepo(value: boolean): DeletePipelineRequest;
    getMustExist(): boolean;
    setMustExist(value: boolean): DeletePipelineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeletePipelineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeletePipelineRequest): DeletePipelineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeletePipelineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeletePipelineRequest;
    static deserializeBinaryFromReader(message: DeletePipelineRequest, reader: jspb.BinaryReader): DeletePipelineRequest;
}

export namespace DeletePipelineRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        all: boolean,
        force: boolean,
        keepRepo: boolean,
        mustExist: boolean,
    }
}

export class DeletePipelinesRequest extends jspb.Message { 
    clearProjectsList(): void;
    getProjectsList(): Array<pfs_pfs_pb.Project>;
    setProjectsList(value: Array<pfs_pfs_pb.Project>): DeletePipelinesRequest;
    addProjects(value?: pfs_pfs_pb.Project, index?: number): pfs_pfs_pb.Project;
    getForce(): boolean;
    setForce(value: boolean): DeletePipelinesRequest;
    getKeepRepo(): boolean;
    setKeepRepo(value: boolean): DeletePipelinesRequest;
    getAll(): boolean;
    setAll(value: boolean): DeletePipelinesRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeletePipelinesRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeletePipelinesRequest): DeletePipelinesRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeletePipelinesRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeletePipelinesRequest;
    static deserializeBinaryFromReader(message: DeletePipelinesRequest, reader: jspb.BinaryReader): DeletePipelinesRequest;
}

export namespace DeletePipelinesRequest {
    export type AsObject = {
        projectsList: Array<pfs_pfs_pb.Project.AsObject>,
        force: boolean,
        keepRepo: boolean,
        all: boolean,
    }
}

export class DeletePipelinesResponse extends jspb.Message { 
    clearPipelinesList(): void;
    getPipelinesList(): Array<Pipeline>;
    setPipelinesList(value: Array<Pipeline>): DeletePipelinesResponse;
    addPipelines(value?: Pipeline, index?: number): Pipeline;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeletePipelinesResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeletePipelinesResponse): DeletePipelinesResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeletePipelinesResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeletePipelinesResponse;
    static deserializeBinaryFromReader(message: DeletePipelinesResponse, reader: jspb.BinaryReader): DeletePipelinesResponse;
}

export namespace DeletePipelinesResponse {
    export type AsObject = {
        pipelinesList: Array<Pipeline.AsObject>,
    }
}

export class StartPipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): StartPipelineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StartPipelineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: StartPipelineRequest): StartPipelineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StartPipelineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StartPipelineRequest;
    static deserializeBinaryFromReader(message: StartPipelineRequest, reader: jspb.BinaryReader): StartPipelineRequest;
}

export namespace StartPipelineRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
    }
}

export class StopPipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): StopPipelineRequest;
    getMustExist(): boolean;
    setMustExist(value: boolean): StopPipelineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StopPipelineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: StopPipelineRequest): StopPipelineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StopPipelineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StopPipelineRequest;
    static deserializeBinaryFromReader(message: StopPipelineRequest, reader: jspb.BinaryReader): StopPipelineRequest;
}

export namespace StopPipelineRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        mustExist: boolean,
    }
}

export class RunPipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): RunPipelineRequest;
    clearProvenanceList(): void;
    getProvenanceList(): Array<pfs_pfs_pb.Commit>;
    setProvenanceList(value: Array<pfs_pfs_pb.Commit>): RunPipelineRequest;
    addProvenance(value?: pfs_pfs_pb.Commit, index?: number): pfs_pfs_pb.Commit;
    getJobId(): string;
    setJobId(value: string): RunPipelineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RunPipelineRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RunPipelineRequest): RunPipelineRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RunPipelineRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RunPipelineRequest;
    static deserializeBinaryFromReader(message: RunPipelineRequest, reader: jspb.BinaryReader): RunPipelineRequest;
}

export namespace RunPipelineRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        provenanceList: Array<pfs_pfs_pb.Commit.AsObject>,
        jobId: string,
    }
}

export class RunCronRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): RunCronRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RunCronRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RunCronRequest): RunCronRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RunCronRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RunCronRequest;
    static deserializeBinaryFromReader(message: RunCronRequest, reader: jspb.BinaryReader): RunCronRequest;
}

export namespace RunCronRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
    }
}

export class CheckStatusRequest extends jspb.Message { 

    hasAll(): boolean;
    clearAll(): void;
    getAll(): boolean;
    setAll(value: boolean): CheckStatusRequest;

    hasProject(): boolean;
    clearProject(): void;
    getProject(): pfs_pfs_pb.Project | undefined;
    setProject(value?: pfs_pfs_pb.Project): CheckStatusRequest;

    getContextCase(): CheckStatusRequest.ContextCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CheckStatusRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CheckStatusRequest): CheckStatusRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CheckStatusRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CheckStatusRequest;
    static deserializeBinaryFromReader(message: CheckStatusRequest, reader: jspb.BinaryReader): CheckStatusRequest;
}

export namespace CheckStatusRequest {
    export type AsObject = {
        all: boolean,
        project?: pfs_pfs_pb.Project.AsObject,
    }

    export enum ContextCase {
        CONTEXT_NOT_SET = 0,
        ALL = 1,
        PROJECT = 2,
    }

}

export class CheckStatusResponse extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): pfs_pfs_pb.Project | undefined;
    setProject(value?: pfs_pfs_pb.Project): CheckStatusResponse;

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): CheckStatusResponse;
    clearAlertsList(): void;
    getAlertsList(): Array<string>;
    setAlertsList(value: Array<string>): CheckStatusResponse;
    addAlerts(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CheckStatusResponse.AsObject;
    static toObject(includeInstance: boolean, msg: CheckStatusResponse): CheckStatusResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CheckStatusResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CheckStatusResponse;
    static deserializeBinaryFromReader(message: CheckStatusResponse, reader: jspb.BinaryReader): CheckStatusResponse;
}

export namespace CheckStatusResponse {
    export type AsObject = {
        project?: pfs_pfs_pb.Project.AsObject,
        pipeline?: Pipeline.AsObject,
        alertsList: Array<string>,
    }
}

export class CreateSecretRequest extends jspb.Message { 
    getFile(): Uint8Array | string;
    getFile_asU8(): Uint8Array;
    getFile_asB64(): string;
    setFile(value: Uint8Array | string): CreateSecretRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateSecretRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateSecretRequest): CreateSecretRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateSecretRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateSecretRequest;
    static deserializeBinaryFromReader(message: CreateSecretRequest, reader: jspb.BinaryReader): CreateSecretRequest;
}

export namespace CreateSecretRequest {
    export type AsObject = {
        file: Uint8Array | string,
    }
}

export class DeleteSecretRequest extends jspb.Message { 

    hasSecret(): boolean;
    clearSecret(): void;
    getSecret(): Secret | undefined;
    setSecret(value?: Secret): DeleteSecretRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteSecretRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteSecretRequest): DeleteSecretRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteSecretRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteSecretRequest;
    static deserializeBinaryFromReader(message: DeleteSecretRequest, reader: jspb.BinaryReader): DeleteSecretRequest;
}

export namespace DeleteSecretRequest {
    export type AsObject = {
        secret?: Secret.AsObject,
    }
}

export class InspectSecretRequest extends jspb.Message { 

    hasSecret(): boolean;
    clearSecret(): void;
    getSecret(): Secret | undefined;
    setSecret(value?: Secret): InspectSecretRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectSecretRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectSecretRequest): InspectSecretRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectSecretRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectSecretRequest;
    static deserializeBinaryFromReader(message: InspectSecretRequest, reader: jspb.BinaryReader): InspectSecretRequest;
}

export namespace InspectSecretRequest {
    export type AsObject = {
        secret?: Secret.AsObject,
    }
}

export class Secret extends jspb.Message { 
    getName(): string;
    setName(value: string): Secret;

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
    }
}

export class SecretInfo extends jspb.Message { 

    hasSecret(): boolean;
    clearSecret(): void;
    getSecret(): Secret | undefined;
    setSecret(value?: Secret): SecretInfo;
    getType(): string;
    setType(value: string): SecretInfo;

    hasCreationTimestamp(): boolean;
    clearCreationTimestamp(): void;
    getCreationTimestamp(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setCreationTimestamp(value?: google_protobuf_timestamp_pb.Timestamp): SecretInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SecretInfo.AsObject;
    static toObject(includeInstance: boolean, msg: SecretInfo): SecretInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SecretInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SecretInfo;
    static deserializeBinaryFromReader(message: SecretInfo, reader: jspb.BinaryReader): SecretInfo;
}

export namespace SecretInfo {
    export type AsObject = {
        secret?: Secret.AsObject,
        type: string,
        creationTimestamp?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class SecretInfos extends jspb.Message { 
    clearSecretInfoList(): void;
    getSecretInfoList(): Array<SecretInfo>;
    setSecretInfoList(value: Array<SecretInfo>): SecretInfos;
    addSecretInfo(value?: SecretInfo, index?: number): SecretInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SecretInfos.AsObject;
    static toObject(includeInstance: boolean, msg: SecretInfos): SecretInfos.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SecretInfos, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SecretInfos;
    static deserializeBinaryFromReader(message: SecretInfos, reader: jspb.BinaryReader): SecretInfos;
}

export namespace SecretInfos {
    export type AsObject = {
        secretInfoList: Array<SecretInfo.AsObject>,
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
    getDagSpec(): string;
    setDagSpec(value: string): RunLoadTestRequest;
    getLoadSpec(): string;
    setLoadSpec(value: string): RunLoadTestRequest;
    getSeed(): number;
    setSeed(value: number): RunLoadTestRequest;
    getParallelism(): number;
    setParallelism(value: number): RunLoadTestRequest;
    getPodPatch(): string;
    setPodPatch(value: string): RunLoadTestRequest;
    getStateId(): string;
    setStateId(value: string): RunLoadTestRequest;

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
        dagSpec: string,
        loadSpec: string,
        seed: number,
        parallelism: number,
        podPatch: string,
        stateId: string,
    }
}

export class RunLoadTestResponse extends jspb.Message { 
    getError(): string;
    setError(value: string): RunLoadTestResponse;
    getStateId(): string;
    setStateId(value: string): RunLoadTestResponse;

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
        error: string,
        stateId: string,
    }
}

export class RenderTemplateRequest extends jspb.Message { 
    getTemplate(): string;
    setTemplate(value: string): RenderTemplateRequest;

    getArgsMap(): jspb.Map<string, string>;
    clearArgsMap(): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RenderTemplateRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RenderTemplateRequest): RenderTemplateRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RenderTemplateRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RenderTemplateRequest;
    static deserializeBinaryFromReader(message: RenderTemplateRequest, reader: jspb.BinaryReader): RenderTemplateRequest;
}

export namespace RenderTemplateRequest {
    export type AsObject = {
        template: string,

        argsMap: Array<[string, string]>,
    }
}

export class RenderTemplateResponse extends jspb.Message { 
    getJson(): string;
    setJson(value: string): RenderTemplateResponse;
    clearSpecsList(): void;
    getSpecsList(): Array<CreatePipelineRequest>;
    setSpecsList(value: Array<CreatePipelineRequest>): RenderTemplateResponse;
    addSpecs(value?: CreatePipelineRequest, index?: number): CreatePipelineRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RenderTemplateResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RenderTemplateResponse): RenderTemplateResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RenderTemplateResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RenderTemplateResponse;
    static deserializeBinaryFromReader(message: RenderTemplateResponse, reader: jspb.BinaryReader): RenderTemplateResponse;
}

export namespace RenderTemplateResponse {
    export type AsObject = {
        json: string,
        specsList: Array<CreatePipelineRequest.AsObject>,
    }
}

export class LokiRequest extends jspb.Message { 

    hasSince(): boolean;
    clearSince(): void;
    getSince(): google_protobuf_duration_pb.Duration | undefined;
    setSince(value?: google_protobuf_duration_pb.Duration): LokiRequest;
    getQuery(): string;
    setQuery(value: string): LokiRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): LokiRequest.AsObject;
    static toObject(includeInstance: boolean, msg: LokiRequest): LokiRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: LokiRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): LokiRequest;
    static deserializeBinaryFromReader(message: LokiRequest, reader: jspb.BinaryReader): LokiRequest;
}

export namespace LokiRequest {
    export type AsObject = {
        since?: google_protobuf_duration_pb.Duration.AsObject,
        query: string,
    }
}

export class LokiLogMessage extends jspb.Message { 
    getMessage(): string;
    setMessage(value: string): LokiLogMessage;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): LokiLogMessage.AsObject;
    static toObject(includeInstance: boolean, msg: LokiLogMessage): LokiLogMessage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: LokiLogMessage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): LokiLogMessage;
    static deserializeBinaryFromReader(message: LokiLogMessage, reader: jspb.BinaryReader): LokiLogMessage;
}

export namespace LokiLogMessage {
    export type AsObject = {
        message: string,
    }
}

export class ClusterDefaults extends jspb.Message { 

    hasCreatePipelineRequest(): boolean;
    clearCreatePipelineRequest(): void;
    getCreatePipelineRequest(): CreatePipelineRequest | undefined;
    setCreatePipelineRequest(value?: CreatePipelineRequest): ClusterDefaults;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ClusterDefaults.AsObject;
    static toObject(includeInstance: boolean, msg: ClusterDefaults): ClusterDefaults.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ClusterDefaults, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ClusterDefaults;
    static deserializeBinaryFromReader(message: ClusterDefaults, reader: jspb.BinaryReader): ClusterDefaults;
}

export namespace ClusterDefaults {
    export type AsObject = {
        createPipelineRequest?: CreatePipelineRequest.AsObject,
    }
}

export class GetClusterDefaultsRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetClusterDefaultsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetClusterDefaultsRequest): GetClusterDefaultsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetClusterDefaultsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetClusterDefaultsRequest;
    static deserializeBinaryFromReader(message: GetClusterDefaultsRequest, reader: jspb.BinaryReader): GetClusterDefaultsRequest;
}

export namespace GetClusterDefaultsRequest {
    export type AsObject = {
    }
}

export class GetClusterDefaultsResponse extends jspb.Message { 
    getClusterDefaultsJson(): string;
    setClusterDefaultsJson(value: string): GetClusterDefaultsResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetClusterDefaultsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetClusterDefaultsResponse): GetClusterDefaultsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetClusterDefaultsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetClusterDefaultsResponse;
    static deserializeBinaryFromReader(message: GetClusterDefaultsResponse, reader: jspb.BinaryReader): GetClusterDefaultsResponse;
}

export namespace GetClusterDefaultsResponse {
    export type AsObject = {
        clusterDefaultsJson: string,
    }
}

export class SetClusterDefaultsRequest extends jspb.Message { 
    getRegenerate(): boolean;
    setRegenerate(value: boolean): SetClusterDefaultsRequest;
    getReprocess(): boolean;
    setReprocess(value: boolean): SetClusterDefaultsRequest;
    getDryRun(): boolean;
    setDryRun(value: boolean): SetClusterDefaultsRequest;
    getClusterDefaultsJson(): string;
    setClusterDefaultsJson(value: string): SetClusterDefaultsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetClusterDefaultsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SetClusterDefaultsRequest): SetClusterDefaultsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetClusterDefaultsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetClusterDefaultsRequest;
    static deserializeBinaryFromReader(message: SetClusterDefaultsRequest, reader: jspb.BinaryReader): SetClusterDefaultsRequest;
}

export namespace SetClusterDefaultsRequest {
    export type AsObject = {
        regenerate: boolean,
        reprocess: boolean,
        dryRun: boolean,
        clusterDefaultsJson: string,
    }
}

export class SetClusterDefaultsResponse extends jspb.Message { 
    clearAffectedPipelinesList(): void;
    getAffectedPipelinesList(): Array<Pipeline>;
    setAffectedPipelinesList(value: Array<Pipeline>): SetClusterDefaultsResponse;
    addAffectedPipelines(value?: Pipeline, index?: number): Pipeline;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetClusterDefaultsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SetClusterDefaultsResponse): SetClusterDefaultsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetClusterDefaultsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetClusterDefaultsResponse;
    static deserializeBinaryFromReader(message: SetClusterDefaultsResponse, reader: jspb.BinaryReader): SetClusterDefaultsResponse;
}

export namespace SetClusterDefaultsResponse {
    export type AsObject = {
        affectedPipelinesList: Array<Pipeline.AsObject>,
    }
}

export class CreatePipelineTransaction extends jspb.Message { 

    hasCreatePipelineRequest(): boolean;
    clearCreatePipelineRequest(): void;
    getCreatePipelineRequest(): CreatePipelineRequest | undefined;
    setCreatePipelineRequest(value?: CreatePipelineRequest): CreatePipelineTransaction;
    getUserJson(): string;
    setUserJson(value: string): CreatePipelineTransaction;
    getEffectiveJson(): string;
    setEffectiveJson(value: string): CreatePipelineTransaction;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreatePipelineTransaction.AsObject;
    static toObject(includeInstance: boolean, msg: CreatePipelineTransaction): CreatePipelineTransaction.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreatePipelineTransaction, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreatePipelineTransaction;
    static deserializeBinaryFromReader(message: CreatePipelineTransaction, reader: jspb.BinaryReader): CreatePipelineTransaction;
}

export namespace CreatePipelineTransaction {
    export type AsObject = {
        createPipelineRequest?: CreatePipelineRequest.AsObject,
        userJson: string,
        effectiveJson: string,
    }
}

export class ProjectDefaults extends jspb.Message { 

    hasCreatePipelineRequest(): boolean;
    clearCreatePipelineRequest(): void;
    getCreatePipelineRequest(): CreatePipelineRequest | undefined;
    setCreatePipelineRequest(value?: CreatePipelineRequest): ProjectDefaults;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProjectDefaults.AsObject;
    static toObject(includeInstance: boolean, msg: ProjectDefaults): ProjectDefaults.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProjectDefaults, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProjectDefaults;
    static deserializeBinaryFromReader(message: ProjectDefaults, reader: jspb.BinaryReader): ProjectDefaults;
}

export namespace ProjectDefaults {
    export type AsObject = {
        createPipelineRequest?: CreatePipelineRequest.AsObject,
    }
}

export class GetProjectDefaultsRequest extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): pfs_pfs_pb.Project | undefined;
    setProject(value?: pfs_pfs_pb.Project): GetProjectDefaultsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetProjectDefaultsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetProjectDefaultsRequest): GetProjectDefaultsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetProjectDefaultsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetProjectDefaultsRequest;
    static deserializeBinaryFromReader(message: GetProjectDefaultsRequest, reader: jspb.BinaryReader): GetProjectDefaultsRequest;
}

export namespace GetProjectDefaultsRequest {
    export type AsObject = {
        project?: pfs_pfs_pb.Project.AsObject,
    }
}

export class GetProjectDefaultsResponse extends jspb.Message { 
    getProjectDefaultsJson(): string;
    setProjectDefaultsJson(value: string): GetProjectDefaultsResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetProjectDefaultsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetProjectDefaultsResponse): GetProjectDefaultsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetProjectDefaultsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetProjectDefaultsResponse;
    static deserializeBinaryFromReader(message: GetProjectDefaultsResponse, reader: jspb.BinaryReader): GetProjectDefaultsResponse;
}

export namespace GetProjectDefaultsResponse {
    export type AsObject = {
        projectDefaultsJson: string,
    }
}

export class SetProjectDefaultsRequest extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): pfs_pfs_pb.Project | undefined;
    setProject(value?: pfs_pfs_pb.Project): SetProjectDefaultsRequest;
    getRegenerate(): boolean;
    setRegenerate(value: boolean): SetProjectDefaultsRequest;
    getReprocess(): boolean;
    setReprocess(value: boolean): SetProjectDefaultsRequest;
    getDryRun(): boolean;
    setDryRun(value: boolean): SetProjectDefaultsRequest;
    getProjectDefaultsJson(): string;
    setProjectDefaultsJson(value: string): SetProjectDefaultsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetProjectDefaultsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SetProjectDefaultsRequest): SetProjectDefaultsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetProjectDefaultsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetProjectDefaultsRequest;
    static deserializeBinaryFromReader(message: SetProjectDefaultsRequest, reader: jspb.BinaryReader): SetProjectDefaultsRequest;
}

export namespace SetProjectDefaultsRequest {
    export type AsObject = {
        project?: pfs_pfs_pb.Project.AsObject,
        regenerate: boolean,
        reprocess: boolean,
        dryRun: boolean,
        projectDefaultsJson: string,
    }
}

export class SetProjectDefaultsResponse extends jspb.Message { 
    clearAffectedPipelinesList(): void;
    getAffectedPipelinesList(): Array<Pipeline>;
    setAffectedPipelinesList(value: Array<Pipeline>): SetProjectDefaultsResponse;
    addAffectedPipelines(value?: Pipeline, index?: number): Pipeline;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetProjectDefaultsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SetProjectDefaultsResponse): SetProjectDefaultsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetProjectDefaultsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetProjectDefaultsResponse;
    static deserializeBinaryFromReader(message: SetProjectDefaultsResponse, reader: jspb.BinaryReader): SetProjectDefaultsResponse;
}

export namespace SetProjectDefaultsResponse {
    export type AsObject = {
        affectedPipelinesList: Array<Pipeline.AsObject>,
    }
}

export class PipelinesSummaryRequest extends jspb.Message { 
    clearProjectsList(): void;
    getProjectsList(): Array<pfs_pfs_pb.ProjectPicker>;
    setProjectsList(value: Array<pfs_pfs_pb.ProjectPicker>): PipelinesSummaryRequest;
    addProjects(value?: pfs_pfs_pb.ProjectPicker, index?: number): pfs_pfs_pb.ProjectPicker;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PipelinesSummaryRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PipelinesSummaryRequest): PipelinesSummaryRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PipelinesSummaryRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PipelinesSummaryRequest;
    static deserializeBinaryFromReader(message: PipelinesSummaryRequest, reader: jspb.BinaryReader): PipelinesSummaryRequest;
}

export namespace PipelinesSummaryRequest {
    export type AsObject = {
        projectsList: Array<pfs_pfs_pb.ProjectPicker.AsObject>,
    }
}

export class PipelinesSummaryResponse extends jspb.Message { 
    clearSummariesList(): void;
    getSummariesList(): Array<PipelinesSummary>;
    setSummariesList(value: Array<PipelinesSummary>): PipelinesSummaryResponse;
    addSummaries(value?: PipelinesSummary, index?: number): PipelinesSummary;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PipelinesSummaryResponse.AsObject;
    static toObject(includeInstance: boolean, msg: PipelinesSummaryResponse): PipelinesSummaryResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PipelinesSummaryResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PipelinesSummaryResponse;
    static deserializeBinaryFromReader(message: PipelinesSummaryResponse, reader: jspb.BinaryReader): PipelinesSummaryResponse;
}

export namespace PipelinesSummaryResponse {
    export type AsObject = {
        summariesList: Array<PipelinesSummary.AsObject>,
    }
}

export class PipelinesSummary extends jspb.Message { 

    hasProject(): boolean;
    clearProject(): void;
    getProject(): pfs_pfs_pb.Project | undefined;
    setProject(value?: pfs_pfs_pb.Project): PipelinesSummary;
    getActivePipelines(): number;
    setActivePipelines(value: number): PipelinesSummary;
    getPausedPipelines(): number;
    setPausedPipelines(value: number): PipelinesSummary;
    getFailedPipelines(): number;
    setFailedPipelines(value: number): PipelinesSummary;
    getUnhealthyPipelines(): number;
    setUnhealthyPipelines(value: number): PipelinesSummary;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PipelinesSummary.AsObject;
    static toObject(includeInstance: boolean, msg: PipelinesSummary): PipelinesSummary.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PipelinesSummary, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PipelinesSummary;
    static deserializeBinaryFromReader(message: PipelinesSummary, reader: jspb.BinaryReader): PipelinesSummary;
}

export namespace PipelinesSummary {
    export type AsObject = {
        project?: pfs_pfs_pb.Project.AsObject,
        activePipelines: number,
        pausedPipelines: number,
        failedPipelines: number,
        unhealthyPipelines: number,
    }
}

export enum JobState {
    JOB_STATE_UNKNOWN = 0,
    JOB_CREATED = 1,
    JOB_STARTING = 2,
    JOB_RUNNING = 3,
    JOB_FAILURE = 4,
    JOB_SUCCESS = 5,
    JOB_KILLED = 6,
    JOB_EGRESSING = 7,
    JOB_FINISHING = 8,
    JOB_UNRUNNABLE = 9,
}

export enum DatumState {
    UNKNOWN = 0,
    FAILED = 1,
    SUCCESS = 2,
    SKIPPED = 3,
    STARTING = 4,
    RECOVERED = 5,
}

export enum WorkerState {
    WORKER_STATE_UNKNOWN = 0,
    POD_RUNNING = 1,
    POD_SUCCESS = 2,
    POD_FAILED = 3,
}

export enum PipelineState {
    PIPELINE_STATE_UNKNOWN = 0,
    PIPELINE_STARTING = 1,
    PIPELINE_RUNNING = 2,
    PIPELINE_RESTARTING = 3,
    PIPELINE_FAILURE = 4,
    PIPELINE_PAUSED = 5,
    PIPELINE_STANDBY = 6,
    PIPELINE_CRASHING = 7,
}

export enum TolerationOperator {
    EMPTY = 0,
    EXISTS = 1,
    EQUAL = 2,
}

export enum TaintEffect {
    ALL_EFFECTS = 0,
    NO_SCHEDULE = 1,
    PREFER_NO_SCHEDULE = 2,
    NO_EXECUTE = 3,
}
