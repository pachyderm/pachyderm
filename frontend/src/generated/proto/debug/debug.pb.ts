/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb";
import * as GoogleProtobufDuration from "../google/protobuf/duration.pb";
import * as GoogleProtobufEmpty from "../google/protobuf/empty.pb";
import * as GoogleProtobufWrappers from "../google/protobuf/wrappers.pb";
import * as Pfs_v2Pfs from "../pfs/pfs.pb";
import * as Pps_v2Pps from "../pps/pps.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
    | { [k in keyof T]?: undefined }
    | (
        keyof T extends infer K ?
        (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
            : never)
        : never);

export enum SetLogLevelRequestLogLevel {
    UNKNOWN = "UNKNOWN",
    DEBUG = "DEBUG",
    INFO = "INFO",
    ERROR = "ERROR",
    OFF = "OFF",
}

export type ProfileRequest = {
    __typename?: "ProfileRequest";
    profile?: Profile;
    filter?: Filter;
};

export type Profile = {
    __typename?: "Profile";
    name?: string;
    duration?: GoogleProtobufDuration.Duration;
};


type BaseFilter = {
    __typename?: "BaseFilter";
};

export type Filter = BaseFilter
    & OneOf<{ pachd: boolean; pipeline: Pps_v2Pps.Pipeline; worker: Worker; database: boolean; }>;

export type Worker = {
    __typename?: "Worker";
    pod?: string;
    redirected?: boolean;
};

export type BinaryRequest = {
    __typename?: "BinaryRequest";
    filter?: Filter;
};

export type DumpRequest = {
    __typename?: "DumpRequest";
    filter?: Filter;
    limit?: string;
};


type BaseSetLogLevelRequest = {
    __typename?: "BaseSetLogLevelRequest";
    duration?: GoogleProtobufDuration.Duration;
    recurse?: boolean;
};

export type SetLogLevelRequest = BaseSetLogLevelRequest
    & OneOf<{ pachyderm: SetLogLevelRequestLogLevel; grpc: SetLogLevelRequestLogLevel; }>;

export type SetLogLevelResponse = {
    __typename?: "SetLogLevelResponse";
    affectedPods?: string[];
    erroredPods?: string[];
};

export type GetDumpV2TemplateRequest = {
    __typename?: "GetDumpV2TemplateRequest";
    filters?: string[];
};

export type GetDumpV2TemplateResponse = {
    __typename?: "GetDumpV2TemplateResponse";
    request?: DumpV2Request;
};

export type Pipeline = {
    __typename?: "Pipeline";
    project?: string;
    name?: string;
};

export type Pod = {
    __typename?: "Pod";
    name?: string;
    ip?: string;
    containers?: string[];
};

export type App = {
    __typename?: "App";
    name?: string;
    pods?: Pod[];
    timeout?: GoogleProtobufDuration.Duration;
    pipeline?: Pipeline;
};

export type System = {
    __typename?: "System";
    helm?: boolean;
    database?: boolean;
    version?: boolean;
    describes?: App[];
    logs?: App[];
    lokiLogs?: App[];
    binaries?: App[];
    profiles?: App[];
};

export type StarlarkLiteral = {
    __typename?: "StarlarkLiteral";
    name?: string;
    programText?: string;
};


type BaseStarlark = {
    __typename?: "BaseStarlark";
    timeout?: GoogleProtobufDuration.Duration;
};

export type Starlark = BaseStarlark
    & OneOf<{ builtin: string; literal: StarlarkLiteral; }>;

export type DumpV2RequestDefaults = {
    __typename?: "DumpV2RequestDefaults";
    clusterDefaults?: boolean;
};

export type DumpV2Request = {
    __typename?: "DumpV2Request";
    system?: System;
    pipelines?: Pipeline[];
    inputRepos?: boolean;
    timeout?: GoogleProtobufDuration.Duration;
    defaults?: DumpV2RequestDefaults;
    starlarkScripts?: Starlark[];
};

export type DumpContent = {
    __typename?: "DumpContent";
    content?: Uint8Array;
};

export type DumpProgress = {
    __typename?: "DumpProgress";
    task?: string;
    total?: string;
    progress?: string;
};


type BaseDumpChunk = {
    __typename?: "BaseDumpChunk";
};

export type DumpChunk = BaseDumpChunk
    & OneOf<{ content: DumpContent; progress: DumpProgress; }>;

export type RunPFSLoadTestRequest = {
    __typename?: "RunPFSLoadTestRequest";
    spec?: string;
    branch?: Pfs_v2Pfs.Branch;
    seed?: string;
    stateId?: string;
};

export type RunPFSLoadTestResponse = {
    __typename?: "RunPFSLoadTestResponse";
    spec?: string;
    branch?: Pfs_v2Pfs.Branch;
    seed?: string;
    error?: string;
    duration?: GoogleProtobufDuration.Duration;
    stateId?: string;
};

export class Debug {
    static Profile(req: ProfileRequest, entityNotifier?: fm.NotifyStreamEntityArrival<GoogleProtobufWrappers.BytesValue>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ProfileRequest, GoogleProtobufWrappers.BytesValue>(`/debug_v2.Debug/Profile`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Binary(req: BinaryRequest, entityNotifier?: fm.NotifyStreamEntityArrival<GoogleProtobufWrappers.BytesValue>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<BinaryRequest, GoogleProtobufWrappers.BytesValue>(`/debug_v2.Debug/Binary`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Dump(req: DumpRequest, entityNotifier?: fm.NotifyStreamEntityArrival<GoogleProtobufWrappers.BytesValue>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<DumpRequest, GoogleProtobufWrappers.BytesValue>(`/debug_v2.Debug/Dump`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static SetLogLevel(req: SetLogLevelRequest, initReq?: fm.InitReq): Promise<SetLogLevelResponse> {
        return fm.fetchReq<SetLogLevelRequest, SetLogLevelResponse>(`/debug_v2.Debug/SetLogLevel`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetDumpV2Template(req: GetDumpV2TemplateRequest, initReq?: fm.InitReq): Promise<GetDumpV2TemplateResponse> {
        return fm.fetchReq<GetDumpV2TemplateRequest, GetDumpV2TemplateResponse>(`/debug_v2.Debug/GetDumpV2Template`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DumpV2(req: DumpV2Request, entityNotifier?: fm.NotifyStreamEntityArrival<DumpChunk>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<DumpV2Request, DumpChunk>(`/debug_v2.Debug/DumpV2`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RunPFSLoadTest(req: RunPFSLoadTestRequest, initReq?: fm.InitReq): Promise<RunPFSLoadTestResponse> {
        return fm.fetchReq<RunPFSLoadTestRequest, RunPFSLoadTestResponse>(`/debug_v2.Debug/RunPFSLoadTest`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RunPFSLoadTestDefault(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<RunPFSLoadTestResponse> {
        return fm.fetchReq<GoogleProtobufEmpty.Empty, RunPFSLoadTestResponse>(`/debug_v2.Debug/RunPFSLoadTestDefault`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
}
