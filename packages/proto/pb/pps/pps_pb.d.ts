// package: pps
// file: pps/pps.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as google_protobuf_duration_pb from "google-protobuf/google/protobuf/duration_pb";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";
import * as pfs_pfs_pb from "../pfs/pfs_pb";

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


    hasBuild(): boolean;
    clearBuild(): void;
    getBuild(): BuildSpec | undefined;
    setBuild(value?: BuildSpec): Transform;


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
        build?: BuildSpec.AsObject,
    }
}

export class BuildSpec extends jspb.Message { 
    getPath(): string;
    setPath(value: string): BuildSpec;

    getLanguage(): string;
    setLanguage(value: string): BuildSpec;

    getImage(): string;
    setImage(value: string): BuildSpec;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BuildSpec.AsObject;
    static toObject(includeInstance: boolean, msg: BuildSpec): BuildSpec.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BuildSpec, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BuildSpec;
    static deserializeBinaryFromReader(message: BuildSpec, reader: jspb.BinaryReader): BuildSpec;
}

export namespace BuildSpec {
    export type AsObject = {
        path: string,
        language: string,
        image: string,
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
    }
}

export class PipelineJob extends jspb.Message { 
    getId(): string;
    setId(value: string): PipelineJob;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PipelineJob.AsObject;
    static toObject(includeInstance: boolean, msg: PipelineJob): PipelineJob.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PipelineJob, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PipelineJob;
    static deserializeBinaryFromReader(message: PipelineJob, reader: jspb.BinaryReader): PipelineJob;
}

export namespace PipelineJob {
    export type AsObject = {
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

    getRepo(): string;
    setRepo(value: string): CronInput;

    getRepoType(): string;
    setRepoType(value: string): CronInput;

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
        repo: string,
        repoType: string,
        commit: string,
        spec: string,
        overwrite: boolean,
        start?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class GitInput extends jspb.Message { 
    getName(): string;
    setName(value: string): GitInput;

    getUrl(): string;
    setUrl(value: string): GitInput;

    getBranch(): string;
    setBranch(value: string): GitInput;

    getCommit(): string;
    setCommit(value: string): GitInput;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GitInput.AsObject;
    static toObject(includeInstance: boolean, msg: GitInput): GitInput.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GitInput, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GitInput;
    static deserializeBinaryFromReader(message: GitInput, reader: jspb.BinaryReader): GitInput;
}

export namespace GitInput {
    export type AsObject = {
        name: string,
        url: string,
        branch: string,
        commit: string,
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


    hasGit(): boolean;
    clearGit(): void;
    getGit(): GitInput | undefined;
    setGit(value?: GitInput): Input;


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
        git?: GitInput.AsObject,
    }
}

export class PipelineJobInput extends jspb.Message { 
    getName(): string;
    setName(value: string): PipelineJobInput;


    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): pfs_pfs_pb.Commit | undefined;
    setCommit(value?: pfs_pfs_pb.Commit): PipelineJobInput;

    getGlob(): string;
    setGlob(value: string): PipelineJobInput;

    getLazy(): boolean;
    setLazy(value: boolean): PipelineJobInput;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PipelineJobInput.AsObject;
    static toObject(includeInstance: boolean, msg: PipelineJobInput): PipelineJobInput.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PipelineJobInput, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PipelineJobInput;
    static deserializeBinaryFromReader(message: PipelineJobInput, reader: jspb.BinaryReader): PipelineJobInput;
}

export namespace PipelineJobInput {
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

    getCoefficient(): number;
    setCoefficient(value: number): ParallelismSpec;


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
        coefficient: number,
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
    getId(): string;
    setId(value: string): Datum;


    hasPipelineJob(): boolean;
    clearPipelineJob(): void;
    getPipelineJob(): PipelineJob | undefined;
    setPipelineJob(value?: PipelineJob): Datum;


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
        id: string,
        pipelineJob?: PipelineJob.AsObject,
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

    getPipelineJobId(): string;
    setPipelineJobId(value: string): WorkerStatus;

    clearDataList(): void;
    getDataList(): Array<InputFile>;
    setDataList(value: Array<InputFile>): WorkerStatus;
    addData(value?: InputFile, index?: number): InputFile;


    hasStarted(): boolean;
    clearStarted(): void;
    getStarted(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setStarted(value?: google_protobuf_timestamp_pb.Timestamp): WorkerStatus;


    hasStats(): boolean;
    clearStats(): void;
    getStats(): ProcessStats | undefined;
    setStats(value?: ProcessStats): WorkerStatus;

    getQueueSize(): number;
    setQueueSize(value: number): WorkerStatus;

    getDataProcessed(): number;
    setDataProcessed(value: number): WorkerStatus;

    getDataRecovered(): number;
    setDataRecovered(value: number): WorkerStatus;


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
        pipelineJobId: string,
        dataList: Array<InputFile.AsObject>,
        started?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        stats?: ProcessStats.AsObject,
        queueSize: number,
        dataProcessed: number,
        dataRecovered: number,
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

export class StoredPipelineJobInfo extends jspb.Message { 

    hasPipelineJob(): boolean;
    clearPipelineJob(): void;
    getPipelineJob(): PipelineJob | undefined;
    setPipelineJob(value?: PipelineJob): StoredPipelineJobInfo;


    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): StoredPipelineJobInfo;


    hasOutputCommit(): boolean;
    clearOutputCommit(): void;
    getOutputCommit(): pfs_pfs_pb.Commit | undefined;
    setOutputCommit(value?: pfs_pfs_pb.Commit): StoredPipelineJobInfo;

    getRestart(): number;
    setRestart(value: number): StoredPipelineJobInfo;

    getDataProcessed(): number;
    setDataProcessed(value: number): StoredPipelineJobInfo;

    getDataSkipped(): number;
    setDataSkipped(value: number): StoredPipelineJobInfo;

    getDataTotal(): number;
    setDataTotal(value: number): StoredPipelineJobInfo;

    getDataFailed(): number;
    setDataFailed(value: number): StoredPipelineJobInfo;

    getDataRecovered(): number;
    setDataRecovered(value: number): StoredPipelineJobInfo;


    hasStats(): boolean;
    clearStats(): void;
    getStats(): ProcessStats | undefined;
    setStats(value?: ProcessStats): StoredPipelineJobInfo;


    hasStatsCommit(): boolean;
    clearStatsCommit(): void;
    getStatsCommit(): pfs_pfs_pb.Commit | undefined;
    setStatsCommit(value?: pfs_pfs_pb.Commit): StoredPipelineJobInfo;

    getState(): PipelineJobState;
    setState(value: PipelineJobState): StoredPipelineJobInfo;

    getReason(): string;
    setReason(value: string): StoredPipelineJobInfo;


    hasStarted(): boolean;
    clearStarted(): void;
    getStarted(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setStarted(value?: google_protobuf_timestamp_pb.Timestamp): StoredPipelineJobInfo;


    hasFinished(): boolean;
    clearFinished(): void;
    getFinished(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setFinished(value?: google_protobuf_timestamp_pb.Timestamp): StoredPipelineJobInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StoredPipelineJobInfo.AsObject;
    static toObject(includeInstance: boolean, msg: StoredPipelineJobInfo): StoredPipelineJobInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StoredPipelineJobInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StoredPipelineJobInfo;
    static deserializeBinaryFromReader(message: StoredPipelineJobInfo, reader: jspb.BinaryReader): StoredPipelineJobInfo;
}

export namespace StoredPipelineJobInfo {
    export type AsObject = {
        pipelineJob?: PipelineJob.AsObject,
        pipeline?: Pipeline.AsObject,
        outputCommit?: pfs_pfs_pb.Commit.AsObject,
        restart: number,
        dataProcessed: number,
        dataSkipped: number,
        dataTotal: number,
        dataFailed: number,
        dataRecovered: number,
        stats?: ProcessStats.AsObject,
        statsCommit?: pfs_pfs_pb.Commit.AsObject,
        state: PipelineJobState,
        reason: string,
        started?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        finished?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class PipelineJobInfo extends jspb.Message { 

    hasPipelineJob(): boolean;
    clearPipelineJob(): void;
    getPipelineJob(): PipelineJob | undefined;
    setPipelineJob(value?: PipelineJob): PipelineJobInfo;


    hasTransform(): boolean;
    clearTransform(): void;
    getTransform(): Transform | undefined;
    setTransform(value?: Transform): PipelineJobInfo;


    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): PipelineJobInfo;

    getPipelineVersion(): number;
    setPipelineVersion(value: number): PipelineJobInfo;


    hasSpecCommit(): boolean;
    clearSpecCommit(): void;
    getSpecCommit(): pfs_pfs_pb.Commit | undefined;
    setSpecCommit(value?: pfs_pfs_pb.Commit): PipelineJobInfo;


    hasParallelismSpec(): boolean;
    clearParallelismSpec(): void;
    getParallelismSpec(): ParallelismSpec | undefined;
    setParallelismSpec(value?: ParallelismSpec): PipelineJobInfo;


    hasEgress(): boolean;
    clearEgress(): void;
    getEgress(): Egress | undefined;
    setEgress(value?: Egress): PipelineJobInfo;


    hasParentJob(): boolean;
    clearParentJob(): void;
    getParentJob(): PipelineJob | undefined;
    setParentJob(value?: PipelineJob): PipelineJobInfo;


    hasStarted(): boolean;
    clearStarted(): void;
    getStarted(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setStarted(value?: google_protobuf_timestamp_pb.Timestamp): PipelineJobInfo;


    hasFinished(): boolean;
    clearFinished(): void;
    getFinished(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setFinished(value?: google_protobuf_timestamp_pb.Timestamp): PipelineJobInfo;


    hasOutputCommit(): boolean;
    clearOutputCommit(): void;
    getOutputCommit(): pfs_pfs_pb.Commit | undefined;
    setOutputCommit(value?: pfs_pfs_pb.Commit): PipelineJobInfo;

    getState(): PipelineJobState;
    setState(value: PipelineJobState): PipelineJobInfo;

    getReason(): string;
    setReason(value: string): PipelineJobInfo;


    hasService(): boolean;
    clearService(): void;
    getService(): Service | undefined;
    setService(value?: Service): PipelineJobInfo;


    hasSpout(): boolean;
    clearSpout(): void;
    getSpout(): Spout | undefined;
    setSpout(value?: Spout): PipelineJobInfo;


    hasOutputRepo(): boolean;
    clearOutputRepo(): void;
    getOutputRepo(): pfs_pfs_pb.Repo | undefined;
    setOutputRepo(value?: pfs_pfs_pb.Repo): PipelineJobInfo;

    getOutputBranch(): string;
    setOutputBranch(value: string): PipelineJobInfo;

    getRestart(): number;
    setRestart(value: number): PipelineJobInfo;

    getDataProcessed(): number;
    setDataProcessed(value: number): PipelineJobInfo;

    getDataSkipped(): number;
    setDataSkipped(value: number): PipelineJobInfo;

    getDataFailed(): number;
    setDataFailed(value: number): PipelineJobInfo;

    getDataRecovered(): number;
    setDataRecovered(value: number): PipelineJobInfo;

    getDataTotal(): number;
    setDataTotal(value: number): PipelineJobInfo;


    hasStats(): boolean;
    clearStats(): void;
    getStats(): ProcessStats | undefined;
    setStats(value?: ProcessStats): PipelineJobInfo;

    clearWorkerStatusList(): void;
    getWorkerStatusList(): Array<WorkerStatus>;
    setWorkerStatusList(value: Array<WorkerStatus>): PipelineJobInfo;
    addWorkerStatus(value?: WorkerStatus, index?: number): WorkerStatus;


    hasResourceRequests(): boolean;
    clearResourceRequests(): void;
    getResourceRequests(): ResourceSpec | undefined;
    setResourceRequests(value?: ResourceSpec): PipelineJobInfo;


    hasResourceLimits(): boolean;
    clearResourceLimits(): void;
    getResourceLimits(): ResourceSpec | undefined;
    setResourceLimits(value?: ResourceSpec): PipelineJobInfo;


    hasSidecarResourceLimits(): boolean;
    clearSidecarResourceLimits(): void;
    getSidecarResourceLimits(): ResourceSpec | undefined;
    setSidecarResourceLimits(value?: ResourceSpec): PipelineJobInfo;


    hasInput(): boolean;
    clearInput(): void;
    getInput(): Input | undefined;
    setInput(value?: Input): PipelineJobInfo;


    hasNewBranch(): boolean;
    clearNewBranch(): void;
    getNewBranch(): pfs_pfs_pb.BranchInfo | undefined;
    setNewBranch(value?: pfs_pfs_pb.BranchInfo): PipelineJobInfo;


    hasStatsCommit(): boolean;
    clearStatsCommit(): void;
    getStatsCommit(): pfs_pfs_pb.Commit | undefined;
    setStatsCommit(value?: pfs_pfs_pb.Commit): PipelineJobInfo;

    getEnableStats(): boolean;
    setEnableStats(value: boolean): PipelineJobInfo;

    getSalt(): string;
    setSalt(value: string): PipelineJobInfo;


    hasChunkSpec(): boolean;
    clearChunkSpec(): void;
    getChunkSpec(): ChunkSpec | undefined;
    setChunkSpec(value?: ChunkSpec): PipelineJobInfo;


    hasDatumTimeout(): boolean;
    clearDatumTimeout(): void;
    getDatumTimeout(): google_protobuf_duration_pb.Duration | undefined;
    setDatumTimeout(value?: google_protobuf_duration_pb.Duration): PipelineJobInfo;


    hasJobTimeout(): boolean;
    clearJobTimeout(): void;
    getJobTimeout(): google_protobuf_duration_pb.Duration | undefined;
    setJobTimeout(value?: google_protobuf_duration_pb.Duration): PipelineJobInfo;

    getDatumTries(): number;
    setDatumTries(value: number): PipelineJobInfo;


    hasSchedulingSpec(): boolean;
    clearSchedulingSpec(): void;
    getSchedulingSpec(): SchedulingSpec | undefined;
    setSchedulingSpec(value?: SchedulingSpec): PipelineJobInfo;

    getPodSpec(): string;
    setPodSpec(value: string): PipelineJobInfo;

    getPodPatch(): string;
    setPodPatch(value: string): PipelineJobInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PipelineJobInfo.AsObject;
    static toObject(includeInstance: boolean, msg: PipelineJobInfo): PipelineJobInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PipelineJobInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PipelineJobInfo;
    static deserializeBinaryFromReader(message: PipelineJobInfo, reader: jspb.BinaryReader): PipelineJobInfo;
}

export namespace PipelineJobInfo {
    export type AsObject = {
        pipelineJob?: PipelineJob.AsObject,
        transform?: Transform.AsObject,
        pipeline?: Pipeline.AsObject,
        pipelineVersion: number,
        specCommit?: pfs_pfs_pb.Commit.AsObject,
        parallelismSpec?: ParallelismSpec.AsObject,
        egress?: Egress.AsObject,
        parentJob?: PipelineJob.AsObject,
        started?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        finished?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        outputCommit?: pfs_pfs_pb.Commit.AsObject,
        state: PipelineJobState,
        reason: string,
        service?: Service.AsObject,
        spout?: Spout.AsObject,
        outputRepo?: pfs_pfs_pb.Repo.AsObject,
        outputBranch: string,
        restart: number,
        dataProcessed: number,
        dataSkipped: number,
        dataFailed: number,
        dataRecovered: number,
        dataTotal: number,
        stats?: ProcessStats.AsObject,
        workerStatusList: Array<WorkerStatus.AsObject>,
        resourceRequests?: ResourceSpec.AsObject,
        resourceLimits?: ResourceSpec.AsObject,
        sidecarResourceLimits?: ResourceSpec.AsObject,
        input?: Input.AsObject,
        newBranch?: pfs_pfs_pb.BranchInfo.AsObject,
        statsCommit?: pfs_pfs_pb.Commit.AsObject,
        enableStats: boolean,
        salt: string,
        chunkSpec?: ChunkSpec.AsObject,
        datumTimeout?: google_protobuf_duration_pb.Duration.AsObject,
        jobTimeout?: google_protobuf_duration_pb.Duration.AsObject,
        datumTries: number,
        schedulingSpec?: SchedulingSpec.AsObject,
        podSpec: string,
        podPatch: string,
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
        name: string,
    }
}

export class StoredPipelineInfo extends jspb.Message { 
    getState(): PipelineState;
    setState(value: PipelineState): StoredPipelineInfo;

    getReason(): string;
    setReason(value: string): StoredPipelineInfo;


    hasSpecCommit(): boolean;
    clearSpecCommit(): void;
    getSpecCommit(): pfs_pfs_pb.Commit | undefined;
    setSpecCommit(value?: pfs_pfs_pb.Commit): StoredPipelineInfo;


    getJobCountsMap(): jspb.Map<number, number>;
    clearJobCountsMap(): void;

    getAuthToken(): string;
    setAuthToken(value: string): StoredPipelineInfo;

    getLastJobState(): PipelineJobState;
    setLastJobState(value: PipelineJobState): StoredPipelineInfo;

    getParallelism(): number;
    setParallelism(value: number): StoredPipelineInfo;


    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): StoredPipelineInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StoredPipelineInfo.AsObject;
    static toObject(includeInstance: boolean, msg: StoredPipelineInfo): StoredPipelineInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StoredPipelineInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StoredPipelineInfo;
    static deserializeBinaryFromReader(message: StoredPipelineInfo, reader: jspb.BinaryReader): StoredPipelineInfo;
}

export namespace StoredPipelineInfo {
    export type AsObject = {
        state: PipelineState,
        reason: string,
        specCommit?: pfs_pfs_pb.Commit.AsObject,

        jobCountsMap: Array<[number, number]>,
        authToken: string,
        lastJobState: PipelineJobState,
        parallelism: number,
        pipeline?: Pipeline.AsObject,
    }
}

export class PipelineInfo extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): PipelineInfo;

    getVersion(): number;
    setVersion(value: number): PipelineInfo;


    hasTransform(): boolean;
    clearTransform(): void;
    getTransform(): Transform | undefined;
    setTransform(value?: Transform): PipelineInfo;


    hasTfJob(): boolean;
    clearTfJob(): void;
    getTfJob(): TFJob | undefined;
    setTfJob(value?: TFJob): PipelineInfo;


    hasParallelismSpec(): boolean;
    clearParallelismSpec(): void;
    getParallelismSpec(): ParallelismSpec | undefined;
    setParallelismSpec(value?: ParallelismSpec): PipelineInfo;


    hasEgress(): boolean;
    clearEgress(): void;
    getEgress(): Egress | undefined;
    setEgress(value?: Egress): PipelineInfo;


    hasCreatedAt(): boolean;
    clearCreatedAt(): void;
    getCreatedAt(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setCreatedAt(value?: google_protobuf_timestamp_pb.Timestamp): PipelineInfo;

    getState(): PipelineState;
    setState(value: PipelineState): PipelineInfo;

    getStopped(): boolean;
    setStopped(value: boolean): PipelineInfo;

    getRecentError(): string;
    setRecentError(value: string): PipelineInfo;

    getWorkersRequested(): number;
    setWorkersRequested(value: number): PipelineInfo;

    getWorkersAvailable(): number;
    setWorkersAvailable(value: number): PipelineInfo;


    getJobCountsMap(): jspb.Map<number, number>;
    clearJobCountsMap(): void;

    getLastJobState(): PipelineJobState;
    setLastJobState(value: PipelineJobState): PipelineInfo;

    getOutputBranch(): string;
    setOutputBranch(value: string): PipelineInfo;


    hasResourceRequests(): boolean;
    clearResourceRequests(): void;
    getResourceRequests(): ResourceSpec | undefined;
    setResourceRequests(value?: ResourceSpec): PipelineInfo;


    hasResourceLimits(): boolean;
    clearResourceLimits(): void;
    getResourceLimits(): ResourceSpec | undefined;
    setResourceLimits(value?: ResourceSpec): PipelineInfo;


    hasSidecarResourceLimits(): boolean;
    clearSidecarResourceLimits(): void;
    getSidecarResourceLimits(): ResourceSpec | undefined;
    setSidecarResourceLimits(value?: ResourceSpec): PipelineInfo;


    hasInput(): boolean;
    clearInput(): void;
    getInput(): Input | undefined;
    setInput(value?: Input): PipelineInfo;

    getDescription(): string;
    setDescription(value: string): PipelineInfo;

    getCacheSize(): string;
    setCacheSize(value: string): PipelineInfo;

    getEnableStats(): boolean;
    setEnableStats(value: boolean): PipelineInfo;

    getSalt(): string;
    setSalt(value: string): PipelineInfo;

    getReason(): string;
    setReason(value: string): PipelineInfo;

    getMaxQueueSize(): number;
    setMaxQueueSize(value: number): PipelineInfo;


    hasService(): boolean;
    clearService(): void;
    getService(): Service | undefined;
    setService(value?: Service): PipelineInfo;


    hasSpout(): boolean;
    clearSpout(): void;
    getSpout(): Spout | undefined;
    setSpout(value?: Spout): PipelineInfo;


    hasChunkSpec(): boolean;
    clearChunkSpec(): void;
    getChunkSpec(): ChunkSpec | undefined;
    setChunkSpec(value?: ChunkSpec): PipelineInfo;


    hasDatumTimeout(): boolean;
    clearDatumTimeout(): void;
    getDatumTimeout(): google_protobuf_duration_pb.Duration | undefined;
    setDatumTimeout(value?: google_protobuf_duration_pb.Duration): PipelineInfo;


    hasJobTimeout(): boolean;
    clearJobTimeout(): void;
    getJobTimeout(): google_protobuf_duration_pb.Duration | undefined;
    setJobTimeout(value?: google_protobuf_duration_pb.Duration): PipelineInfo;

    getGithookUrl(): string;
    setGithookUrl(value: string): PipelineInfo;


    hasSpecCommit(): boolean;
    clearSpecCommit(): void;
    getSpecCommit(): pfs_pfs_pb.Commit | undefined;
    setSpecCommit(value?: pfs_pfs_pb.Commit): PipelineInfo;

    getStandby(): boolean;
    setStandby(value: boolean): PipelineInfo;

    getDatumTries(): number;
    setDatumTries(value: number): PipelineInfo;


    hasSchedulingSpec(): boolean;
    clearSchedulingSpec(): void;
    getSchedulingSpec(): SchedulingSpec | undefined;
    setSchedulingSpec(value?: SchedulingSpec): PipelineInfo;

    getPodSpec(): string;
    setPodSpec(value: string): PipelineInfo;

    getPodPatch(): string;
    setPodPatch(value: string): PipelineInfo;

    getS3Out(): boolean;
    setS3Out(value: boolean): PipelineInfo;


    hasMetadata(): boolean;
    clearMetadata(): void;
    getMetadata(): Metadata | undefined;
    setMetadata(value?: Metadata): PipelineInfo;

    getReprocessSpec(): string;
    setReprocessSpec(value: string): PipelineInfo;


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
        transform?: Transform.AsObject,
        tfJob?: TFJob.AsObject,
        parallelismSpec?: ParallelismSpec.AsObject,
        egress?: Egress.AsObject,
        createdAt?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        state: PipelineState,
        stopped: boolean,
        recentError: string,
        workersRequested: number,
        workersAvailable: number,

        jobCountsMap: Array<[number, number]>,
        lastJobState: PipelineJobState,
        outputBranch: string,
        resourceRequests?: ResourceSpec.AsObject,
        resourceLimits?: ResourceSpec.AsObject,
        sidecarResourceLimits?: ResourceSpec.AsObject,
        input?: Input.AsObject,
        description: string,
        cacheSize: string,
        enableStats: boolean,
        salt: string,
        reason: string,
        maxQueueSize: number,
        service?: Service.AsObject,
        spout?: Spout.AsObject,
        chunkSpec?: ChunkSpec.AsObject,
        datumTimeout?: google_protobuf_duration_pb.Duration.AsObject,
        jobTimeout?: google_protobuf_duration_pb.Duration.AsObject,
        githookUrl: string,
        specCommit?: pfs_pfs_pb.Commit.AsObject,
        standby: boolean,
        datumTries: number,
        schedulingSpec?: SchedulingSpec.AsObject,
        podSpec: string,
        podPatch: string,
        s3Out: boolean,
        metadata?: Metadata.AsObject,
        reprocessSpec: string,
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

export class CreatePipelineJobRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): CreatePipelineJobRequest;


    hasOutputCommit(): boolean;
    clearOutputCommit(): void;
    getOutputCommit(): pfs_pfs_pb.Commit | undefined;
    setOutputCommit(value?: pfs_pfs_pb.Commit): CreatePipelineJobRequest;

    getRestart(): number;
    setRestart(value: number): CreatePipelineJobRequest;

    getDataProcessed(): number;
    setDataProcessed(value: number): CreatePipelineJobRequest;

    getDataSkipped(): number;
    setDataSkipped(value: number): CreatePipelineJobRequest;

    getDataTotal(): number;
    setDataTotal(value: number): CreatePipelineJobRequest;

    getDataFailed(): number;
    setDataFailed(value: number): CreatePipelineJobRequest;

    getDataRecovered(): number;
    setDataRecovered(value: number): CreatePipelineJobRequest;


    hasStats(): boolean;
    clearStats(): void;
    getStats(): ProcessStats | undefined;
    setStats(value?: ProcessStats): CreatePipelineJobRequest;


    hasStatsCommit(): boolean;
    clearStatsCommit(): void;
    getStatsCommit(): pfs_pfs_pb.Commit | undefined;
    setStatsCommit(value?: pfs_pfs_pb.Commit): CreatePipelineJobRequest;

    getState(): PipelineJobState;
    setState(value: PipelineJobState): CreatePipelineJobRequest;

    getReason(): string;
    setReason(value: string): CreatePipelineJobRequest;


    hasStarted(): boolean;
    clearStarted(): void;
    getStarted(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setStarted(value?: google_protobuf_timestamp_pb.Timestamp): CreatePipelineJobRequest;


    hasFinished(): boolean;
    clearFinished(): void;
    getFinished(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setFinished(value?: google_protobuf_timestamp_pb.Timestamp): CreatePipelineJobRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreatePipelineJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreatePipelineJobRequest): CreatePipelineJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreatePipelineJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreatePipelineJobRequest;
    static deserializeBinaryFromReader(message: CreatePipelineJobRequest, reader: jspb.BinaryReader): CreatePipelineJobRequest;
}

export namespace CreatePipelineJobRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        outputCommit?: pfs_pfs_pb.Commit.AsObject,
        restart: number,
        dataProcessed: number,
        dataSkipped: number,
        dataTotal: number,
        dataFailed: number,
        dataRecovered: number,
        stats?: ProcessStats.AsObject,
        statsCommit?: pfs_pfs_pb.Commit.AsObject,
        state: PipelineJobState,
        reason: string,
        started?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        finished?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class InspectPipelineJobRequest extends jspb.Message { 

    hasPipelineJob(): boolean;
    clearPipelineJob(): void;
    getPipelineJob(): PipelineJob | undefined;
    setPipelineJob(value?: PipelineJob): InspectPipelineJobRequest;


    hasOutputCommit(): boolean;
    clearOutputCommit(): void;
    getOutputCommit(): pfs_pfs_pb.Commit | undefined;
    setOutputCommit(value?: pfs_pfs_pb.Commit): InspectPipelineJobRequest;

    getBlockState(): boolean;
    setBlockState(value: boolean): InspectPipelineJobRequest;

    getFull(): boolean;
    setFull(value: boolean): InspectPipelineJobRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectPipelineJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectPipelineJobRequest): InspectPipelineJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectPipelineJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectPipelineJobRequest;
    static deserializeBinaryFromReader(message: InspectPipelineJobRequest, reader: jspb.BinaryReader): InspectPipelineJobRequest;
}

export namespace InspectPipelineJobRequest {
    export type AsObject = {
        pipelineJob?: PipelineJob.AsObject,
        outputCommit?: pfs_pfs_pb.Commit.AsObject,
        blockState: boolean,
        full: boolean,
    }
}

export class ListPipelineJobRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): ListPipelineJobRequest;

    clearInputCommitList(): void;
    getInputCommitList(): Array<pfs_pfs_pb.Commit>;
    setInputCommitList(value: Array<pfs_pfs_pb.Commit>): ListPipelineJobRequest;
    addInputCommit(value?: pfs_pfs_pb.Commit, index?: number): pfs_pfs_pb.Commit;


    hasOutputCommit(): boolean;
    clearOutputCommit(): void;
    getOutputCommit(): pfs_pfs_pb.Commit | undefined;
    setOutputCommit(value?: pfs_pfs_pb.Commit): ListPipelineJobRequest;

    getHistory(): number;
    setHistory(value: number): ListPipelineJobRequest;

    getFull(): boolean;
    setFull(value: boolean): ListPipelineJobRequest;

    getJqfilter(): string;
    setJqfilter(value: string): ListPipelineJobRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListPipelineJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListPipelineJobRequest): ListPipelineJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListPipelineJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListPipelineJobRequest;
    static deserializeBinaryFromReader(message: ListPipelineJobRequest, reader: jspb.BinaryReader): ListPipelineJobRequest;
}

export namespace ListPipelineJobRequest {
    export type AsObject = {
        pipeline?: Pipeline.AsObject,
        inputCommitList: Array<pfs_pfs_pb.Commit.AsObject>,
        outputCommit?: pfs_pfs_pb.Commit.AsObject,
        history: number,
        full: boolean,
        jqfilter: string,
    }
}

export class FlushPipelineJobRequest extends jspb.Message { 
    clearCommitsList(): void;
    getCommitsList(): Array<pfs_pfs_pb.Commit>;
    setCommitsList(value: Array<pfs_pfs_pb.Commit>): FlushPipelineJobRequest;
    addCommits(value?: pfs_pfs_pb.Commit, index?: number): pfs_pfs_pb.Commit;

    clearToPipelinesList(): void;
    getToPipelinesList(): Array<Pipeline>;
    setToPipelinesList(value: Array<Pipeline>): FlushPipelineJobRequest;
    addToPipelines(value?: Pipeline, index?: number): Pipeline;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FlushPipelineJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: FlushPipelineJobRequest): FlushPipelineJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FlushPipelineJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FlushPipelineJobRequest;
    static deserializeBinaryFromReader(message: FlushPipelineJobRequest, reader: jspb.BinaryReader): FlushPipelineJobRequest;
}

export namespace FlushPipelineJobRequest {
    export type AsObject = {
        commitsList: Array<pfs_pfs_pb.Commit.AsObject>,
        toPipelinesList: Array<Pipeline.AsObject>,
    }
}

export class DeletePipelineJobRequest extends jspb.Message { 

    hasPipelineJob(): boolean;
    clearPipelineJob(): void;
    getPipelineJob(): PipelineJob | undefined;
    setPipelineJob(value?: PipelineJob): DeletePipelineJobRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeletePipelineJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeletePipelineJobRequest): DeletePipelineJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeletePipelineJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeletePipelineJobRequest;
    static deserializeBinaryFromReader(message: DeletePipelineJobRequest, reader: jspb.BinaryReader): DeletePipelineJobRequest;
}

export namespace DeletePipelineJobRequest {
    export type AsObject = {
        pipelineJob?: PipelineJob.AsObject,
    }
}

export class StopPipelineJobRequest extends jspb.Message { 

    hasPipelineJob(): boolean;
    clearPipelineJob(): void;
    getPipelineJob(): PipelineJob | undefined;
    setPipelineJob(value?: PipelineJob): StopPipelineJobRequest;


    hasOutputCommit(): boolean;
    clearOutputCommit(): void;
    getOutputCommit(): pfs_pfs_pb.Commit | undefined;
    setOutputCommit(value?: pfs_pfs_pb.Commit): StopPipelineJobRequest;

    getReason(): string;
    setReason(value: string): StopPipelineJobRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StopPipelineJobRequest.AsObject;
    static toObject(includeInstance: boolean, msg: StopPipelineJobRequest): StopPipelineJobRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StopPipelineJobRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StopPipelineJobRequest;
    static deserializeBinaryFromReader(message: StopPipelineJobRequest, reader: jspb.BinaryReader): StopPipelineJobRequest;
}

export namespace StopPipelineJobRequest {
    export type AsObject = {
        pipelineJob?: PipelineJob.AsObject,
        outputCommit?: pfs_pfs_pb.Commit.AsObject,
        reason: string,
    }
}

export class UpdatePipelineJobStateRequest extends jspb.Message { 

    hasPipelineJob(): boolean;
    clearPipelineJob(): void;
    getPipelineJob(): PipelineJob | undefined;
    setPipelineJob(value?: PipelineJob): UpdatePipelineJobStateRequest;

    getState(): PipelineJobState;
    setState(value: PipelineJobState): UpdatePipelineJobStateRequest;

    getReason(): string;
    setReason(value: string): UpdatePipelineJobStateRequest;

    getRestart(): number;
    setRestart(value: number): UpdatePipelineJobStateRequest;

    getDataProcessed(): number;
    setDataProcessed(value: number): UpdatePipelineJobStateRequest;

    getDataSkipped(): number;
    setDataSkipped(value: number): UpdatePipelineJobStateRequest;

    getDataFailed(): number;
    setDataFailed(value: number): UpdatePipelineJobStateRequest;

    getDataRecovered(): number;
    setDataRecovered(value: number): UpdatePipelineJobStateRequest;

    getDataTotal(): number;
    setDataTotal(value: number): UpdatePipelineJobStateRequest;


    hasStats(): boolean;
    clearStats(): void;
    getStats(): ProcessStats | undefined;
    setStats(value?: ProcessStats): UpdatePipelineJobStateRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UpdatePipelineJobStateRequest.AsObject;
    static toObject(includeInstance: boolean, msg: UpdatePipelineJobStateRequest): UpdatePipelineJobStateRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UpdatePipelineJobStateRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UpdatePipelineJobStateRequest;
    static deserializeBinaryFromReader(message: UpdatePipelineJobStateRequest, reader: jspb.BinaryReader): UpdatePipelineJobStateRequest;
}

export namespace UpdatePipelineJobStateRequest {
    export type AsObject = {
        pipelineJob?: PipelineJob.AsObject,
        state: PipelineJobState,
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


    hasPipelineJob(): boolean;
    clearPipelineJob(): void;
    getPipelineJob(): PipelineJob | undefined;
    setPipelineJob(value?: PipelineJob): GetLogsRequest;

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
        pipelineJob?: PipelineJob.AsObject,
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
    getPipelineName(): string;
    setPipelineName(value: string): LogMessage;

    getPipelineJobId(): string;
    setPipelineJobId(value: string): LogMessage;

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
        pipelineName: string,
        pipelineJobId: string,
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

    hasPipelineJob(): boolean;
    clearPipelineJob(): void;
    getPipelineJob(): PipelineJob | undefined;
    setPipelineJob(value?: PipelineJob): RestartDatumRequest;

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
        pipelineJob?: PipelineJob.AsObject,
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

    hasPipelineJob(): boolean;
    clearPipelineJob(): void;
    getPipelineJob(): PipelineJob | undefined;
    setPipelineJob(value?: PipelineJob): ListDatumRequest;


    hasInput(): boolean;
    clearInput(): void;
    getInput(): Input | undefined;
    setInput(value?: Input): ListDatumRequest;


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
        pipelineJob?: PipelineJob.AsObject,
        input?: Input.AsObject,
    }
}

export class ChunkSpec extends jspb.Message { 
    getNumber(): number;
    setNumber(value: number): ChunkSpec;

    getSizeBytes(): number;
    setSizeBytes(value: number): ChunkSpec;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ChunkSpec.AsObject;
    static toObject(includeInstance: boolean, msg: ChunkSpec): ChunkSpec.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ChunkSpec, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ChunkSpec;
    static deserializeBinaryFromReader(message: ChunkSpec, reader: jspb.BinaryReader): ChunkSpec;
}

export namespace ChunkSpec {
    export type AsObject = {
        number: number,
        sizeBytes: number,
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

    getCacheSize(): string;
    setCacheSize(value: string): CreatePipelineRequest;

    getEnableStats(): boolean;
    setEnableStats(value: boolean): CreatePipelineRequest;

    getReprocess(): boolean;
    setReprocess(value: boolean): CreatePipelineRequest;

    getMaxQueueSize(): number;
    setMaxQueueSize(value: number): CreatePipelineRequest;


    hasService(): boolean;
    clearService(): void;
    getService(): Service | undefined;
    setService(value?: Service): CreatePipelineRequest;


    hasSpout(): boolean;
    clearSpout(): void;
    getSpout(): Spout | undefined;
    setSpout(value?: Spout): CreatePipelineRequest;


    hasChunkSpec(): boolean;
    clearChunkSpec(): void;
    getChunkSpec(): ChunkSpec | undefined;
    setChunkSpec(value?: ChunkSpec): CreatePipelineRequest;


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

    getStandby(): boolean;
    setStandby(value: boolean): CreatePipelineRequest;

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
        cacheSize: string,
        enableStats: boolean,
        reprocess: boolean,
        maxQueueSize: number,
        service?: Service.AsObject,
        spout?: Spout.AsObject,
        chunkSpec?: ChunkSpec.AsObject,
        datumTimeout?: google_protobuf_duration_pb.Duration.AsObject,
        jobTimeout?: google_protobuf_duration_pb.Duration.AsObject,
        salt: string,
        standby: boolean,
        datumTries: number,
        schedulingSpec?: SchedulingSpec.AsObject,
        podSpec: string,
        podPatch: string,
        specCommit?: pfs_pfs_pb.Commit.AsObject,
        metadata?: Metadata.AsObject,
        reprocessSpec: string,
    }
}

export class InspectPipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): InspectPipelineRequest;


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
    }
}

export class ListPipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): ListPipelineRequest;

    getHistory(): number;
    setHistory(value: number): ListPipelineRequest;

    getAllowIncomplete(): boolean;
    setAllowIncomplete(value: boolean): ListPipelineRequest;

    getJqfilter(): string;
    setJqfilter(value: string): ListPipelineRequest;


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
        allowIncomplete: boolean,
        jqfilter: string,
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
    }
}

export class RunPipelineRequest extends jspb.Message { 

    hasPipeline(): boolean;
    clearPipeline(): void;
    getPipeline(): Pipeline | undefined;
    setPipeline(value?: Pipeline): RunPipelineRequest;

    clearProvenanceList(): void;
    getProvenanceList(): Array<pfs_pfs_pb.CommitProvenance>;
    setProvenanceList(value: Array<pfs_pfs_pb.CommitProvenance>): RunPipelineRequest;
    addProvenance(value?: pfs_pfs_pb.CommitProvenance, index?: number): pfs_pfs_pb.CommitProvenance;

    getPipelineJobId(): string;
    setPipelineJobId(value: string): RunPipelineRequest;


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
        provenanceList: Array<pfs_pfs_pb.CommitProvenance.AsObject>,
        pipelineJobId: string,
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

export enum PipelineJobState {
    JOB_STARTING = 0,
    JOB_RUNNING = 1,
    JOB_FAILURE = 2,
    JOB_SUCCESS = 3,
    JOB_KILLED = 4,
    JOB_EGRESSING = 5,
}

export enum DatumState {
    FAILED = 0,
    SUCCESS = 1,
    SKIPPED = 2,
    STARTING = 3,
    RECOVERED = 4,
}

export enum WorkerState {
    POD_RUNNING = 0,
    POD_SUCCESS = 1,
    POD_FAILED = 2,
}

export enum PipelineState {
    PIPELINE_STARTING = 0,
    PIPELINE_RUNNING = 1,
    PIPELINE_RESTARTING = 2,
    PIPELINE_FAILURE = 3,
    PIPELINE_PAUSED = 4,
    PIPELINE_STANDBY = 5,
    PIPELINE_CRASHING = 6,
}
