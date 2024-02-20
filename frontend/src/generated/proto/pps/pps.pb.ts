/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb";
import * as GoogleProtobufDuration from "../google/protobuf/duration.pb";
import * as GoogleProtobufEmpty from "../google/protobuf/empty.pb";
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb";
import * as GoogleProtobufWrappers from "../google/protobuf/wrappers.pb";
import * as Pfs_v2Pfs from "../pfs/pfs.pb";
import * as TaskapiTask from "../task/task.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
    | { [k in keyof T]?: undefined }
    | (
        keyof T extends infer K ?
        (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
            : never)
        : never);

export enum JobState {
    JOB_STATE_UNKNOWN = "JOB_STATE_UNKNOWN",
    JOB_CREATED = "JOB_CREATED",
    JOB_STARTING = "JOB_STARTING",
    JOB_RUNNING = "JOB_RUNNING",
    JOB_FAILURE = "JOB_FAILURE",
    JOB_SUCCESS = "JOB_SUCCESS",
    JOB_KILLED = "JOB_KILLED",
    JOB_EGRESSING = "JOB_EGRESSING",
    JOB_FINISHING = "JOB_FINISHING",
    JOB_UNRUNNABLE = "JOB_UNRUNNABLE",
}

export enum DatumState {
    UNKNOWN = "UNKNOWN",
    FAILED = "FAILED",
    SUCCESS = "SUCCESS",
    SKIPPED = "SKIPPED",
    STARTING = "STARTING",
    RECOVERED = "RECOVERED",
}

export enum WorkerState {
    WORKER_STATE_UNKNOWN = "WORKER_STATE_UNKNOWN",
    POD_RUNNING = "POD_RUNNING",
    POD_SUCCESS = "POD_SUCCESS",
    POD_FAILED = "POD_FAILED",
}

export enum PipelineState {
    PIPELINE_STATE_UNKNOWN = "PIPELINE_STATE_UNKNOWN",
    PIPELINE_STARTING = "PIPELINE_STARTING",
    PIPELINE_RUNNING = "PIPELINE_RUNNING",
    PIPELINE_RESTARTING = "PIPELINE_RESTARTING",
    PIPELINE_FAILURE = "PIPELINE_FAILURE",
    PIPELINE_PAUSED = "PIPELINE_PAUSED",
    PIPELINE_STANDBY = "PIPELINE_STANDBY",
    PIPELINE_CRASHING = "PIPELINE_CRASHING",
}

export enum TolerationOperator {
    EMPTY = "EMPTY",
    EXISTS = "EXISTS",
    EQUAL = "EQUAL",
}

export enum TaintEffect {
    ALL_EFFECTS = "ALL_EFFECTS",
    NO_SCHEDULE = "NO_SCHEDULE",
    PREFER_NO_SCHEDULE = "PREFER_NO_SCHEDULE",
    NO_EXECUTE = "NO_EXECUTE",
}

export enum PipelineInfoPipelineType {
    PIPELINT_TYPE_UNKNOWN = "PIPELINT_TYPE_UNKNOWN",
    PIPELINE_TYPE_TRANSFORM = "PIPELINE_TYPE_TRANSFORM",
    PIPELINE_TYPE_SPOUT = "PIPELINE_TYPE_SPOUT",
    PIPELINE_TYPE_SERVICE = "PIPELINE_TYPE_SERVICE",
}

export type SecretMount = {
    __typename?: "SecretMount";
    name?: string;
    key?: string;
    mountPath?: string;
    envVar?: string;
};

export type Transform = {
    __typename?: "Transform";
    image?: string;
    cmd?: string[];
    errCmd?: string[];
    env?: { [key: string]: string; };
    secrets?: SecretMount[];
    imagePullSecrets?: string[];
    stdin?: string[];
    errStdin?: string[];
    acceptReturnCode?: string[];
    debug?: boolean;
    user?: string;
    workingDir?: string;
    dockerfile?: string;
    memoryVolume?: boolean;
    datumBatching?: boolean;
};

export type TFJob = {
    __typename?: "TFJob";
    tfJob?: string;
};


type BaseEgress = {
    __typename?: "BaseEgress";
    uRL?: string;
};

export type Egress = BaseEgress
    & OneOf<{ objectStorage: Pfs_v2Pfs.ObjectStorageEgress; sqlDatabase: Pfs_v2Pfs.SQLDatabaseEgress; }>;

export type Determined = {
    __typename?: "Determined";
    workspaces?: string[];
};

export type Job = {
    __typename?: "Job";
    pipeline?: Pipeline;
    id?: string;
};

export type Metadata = {
    __typename?: "Metadata";
    annotations?: { [key: string]: string; };
    labels?: { [key: string]: string; };
};

export type Service = {
    __typename?: "Service";
    internalPort?: number;
    externalPort?: number;
    ip?: string;
    type?: string;
};

export type Spout = {
    __typename?: "Spout";
    service?: Service;
};

export type PFSInput = {
    __typename?: "PFSInput";
    project?: string;
    name?: string;
    repo?: string;
    repoType?: string;
    branch?: string;
    commit?: string;
    glob?: string;
    joinOn?: string;
    outerJoin?: boolean;
    groupBy?: string;
    lazy?: boolean;
    emptyFiles?: boolean;
    s3?: boolean;
    trigger?: Pfs_v2Pfs.Trigger;
};

export type CronInput = {
    __typename?: "CronInput";
    name?: string;
    project?: string;
    repo?: string;
    commit?: string;
    spec?: string;
    overwrite?: boolean;
    start?: GoogleProtobufTimestamp.Timestamp;
};

export type Input = {
    __typename?: "Input";
    pfs?: PFSInput;
    join?: Input[];
    group?: Input[];
    cross?: Input[];
    union?: Input[];
    cron?: CronInput;
};

export type JobInput = {
    __typename?: "JobInput";
    name?: string;
    commit?: Pfs_v2Pfs.Commit;
    glob?: string;
    lazy?: boolean;
};

export type ParallelismSpec = {
    __typename?: "ParallelismSpec";
    constant?: string;
};

export type InputFile = {
    __typename?: "InputFile";
    path?: string;
    hash?: Uint8Array;
};

export type Datum = {
    __typename?: "Datum";
    job?: Job;
    id?: string;
};

export type DatumInfo = {
    __typename?: "DatumInfo";
    datum?: Datum;
    state?: DatumState;
    stats?: ProcessStats;
    pfsState?: Pfs_v2Pfs.File;
    data?: Pfs_v2Pfs.FileInfo[];
    imageId?: string;
};

export type Aggregate = {
    __typename?: "Aggregate";
    count?: string;
    mean?: number;
    stddev?: number;
    fifthPercentile?: number;
    ninetyFifthPercentile?: number;
};

export type ProcessStats = {
    __typename?: "ProcessStats";
    downloadTime?: GoogleProtobufDuration.Duration;
    processTime?: GoogleProtobufDuration.Duration;
    uploadTime?: GoogleProtobufDuration.Duration;
    downloadBytes?: string;
    uploadBytes?: string;
};

export type AggregateProcessStats = {
    __typename?: "AggregateProcessStats";
    downloadTime?: Aggregate;
    processTime?: Aggregate;
    uploadTime?: Aggregate;
    downloadBytes?: Aggregate;
    uploadBytes?: Aggregate;
};

export type WorkerStatus = {
    __typename?: "WorkerStatus";
    workerId?: string;
    jobId?: string;
    datumStatus?: DatumStatus;
};

export type DatumStatus = {
    __typename?: "DatumStatus";
    started?: GoogleProtobufTimestamp.Timestamp;
    data?: InputFile[];
};

export type ResourceSpec = {
    __typename?: "ResourceSpec";
    cpu?: number;
    memory?: string;
    gpu?: GPUSpec;
    disk?: string;
};

export type GPUSpec = {
    __typename?: "GPUSpec";
    type?: string;
    number?: string;
};

export type JobSetInfo = {
    __typename?: "JobSetInfo";
    jobSet?: JobSet;
    jobs?: JobInfo[];
};

export type JobInfoDetails = {
    __typename?: "JobInfoDetails";
    transform?: Transform;
    parallelismSpec?: ParallelismSpec;
    egress?: Egress;
    service?: Service;
    spout?: Spout;
    workerStatus?: WorkerStatus[];
    resourceRequests?: ResourceSpec;
    resourceLimits?: ResourceSpec;
    sidecarResourceLimits?: ResourceSpec;
    input?: Input;
    salt?: string;
    datumSetSpec?: DatumSetSpec;
    datumTimeout?: GoogleProtobufDuration.Duration;
    jobTimeout?: GoogleProtobufDuration.Duration;
    datumTries?: string;
    schedulingSpec?: SchedulingSpec;
    podSpec?: string;
    podPatch?: string;
    sidecarResourceRequests?: ResourceSpec;
};

export type JobInfo = {
    __typename?: "JobInfo";
    job?: Job;
    pipelineVersion?: string;
    outputCommit?: Pfs_v2Pfs.Commit;
    restart?: string;
    dataProcessed?: string;
    dataSkipped?: string;
    dataTotal?: string;
    dataFailed?: string;
    dataRecovered?: string;
    stats?: ProcessStats;
    state?: JobState;
    reason?: string;
    created?: GoogleProtobufTimestamp.Timestamp;
    started?: GoogleProtobufTimestamp.Timestamp;
    finished?: GoogleProtobufTimestamp.Timestamp;
    details?: JobInfoDetails;
    authToken?: string;
};

export type Worker = {
    __typename?: "Worker";
    name?: string;
    state?: WorkerState;
};

export type Pipeline = {
    __typename?: "Pipeline";
    project?: Pfs_v2Pfs.Project;
    name?: string;
};

export type Toleration = {
    __typename?: "Toleration";
    key?: string;
    operator?: TolerationOperator;
    value?: string;
    effect?: TaintEffect;
    tolerationSeconds?: GoogleProtobufWrappers.Int64Value;
};

export type PipelineInfoDetails = {
    __typename?: "PipelineInfoDetails";
    transform?: Transform;
    tfJob?: TFJob;
    parallelismSpec?: ParallelismSpec;
    egress?: Egress;
    createdAt?: GoogleProtobufTimestamp.Timestamp;
    recentError?: string;
    workersRequested?: string;
    workersAvailable?: string;
    outputBranch?: string;
    resourceRequests?: ResourceSpec;
    resourceLimits?: ResourceSpec;
    sidecarResourceLimits?: ResourceSpec;
    input?: Input;
    description?: string;
    salt?: string;
    reason?: string;
    service?: Service;
    spout?: Spout;
    datumSetSpec?: DatumSetSpec;
    datumTimeout?: GoogleProtobufDuration.Duration;
    jobTimeout?: GoogleProtobufDuration.Duration;
    datumTries?: string;
    schedulingSpec?: SchedulingSpec;
    podSpec?: string;
    podPatch?: string;
    s3Out?: boolean;
    metadata?: Metadata;
    reprocessSpec?: string;
    unclaimedTasks?: string;
    workerRc?: string;
    autoscaling?: boolean;
    tolerations?: Toleration[];
    sidecarResourceRequests?: ResourceSpec;
    determined?: Determined;
    maximumExpectedUptime?: GoogleProtobufDuration.Duration;
    workersStartedAt?: GoogleProtobufTimestamp.Timestamp;
};

export type PipelineInfo = {
    __typename?: "PipelineInfo";
    pipeline?: Pipeline;
    version?: string;
    specCommit?: Pfs_v2Pfs.Commit;
    stopped?: boolean;
    state?: PipelineState;
    reason?: string;
    lastJobState?: JobState;
    parallelism?: string;
    type?: PipelineInfoPipelineType;
    authToken?: string;
    details?: PipelineInfoDetails;
    userSpecJson?: string;
    effectiveSpecJson?: string;
};

export type PipelineInfos = {
    __typename?: "PipelineInfos";
    pipelineInfo?: PipelineInfo[];
};

export type JobSet = {
    __typename?: "JobSet";
    id?: string;
};

export type InspectJobSetRequest = {
    __typename?: "InspectJobSetRequest";
    jobSet?: JobSet;
    wait?: boolean;
    details?: boolean;
};

export type ListJobSetRequest = {
    __typename?: "ListJobSetRequest";
    details?: boolean;
    projects?: Pfs_v2Pfs.Project[];
    paginationMarker?: GoogleProtobufTimestamp.Timestamp;
    number?: string;
    reverse?: boolean;
    jqFilter?: string;
};

export type InspectJobRequest = {
    __typename?: "InspectJobRequest";
    job?: Job;
    wait?: boolean;
    details?: boolean;
};

export type ListJobRequest = {
    __typename?: "ListJobRequest";
    projects?: Pfs_v2Pfs.Project[];
    pipeline?: Pipeline;
    inputCommit?: Pfs_v2Pfs.Commit[];
    history?: string;
    details?: boolean;
    jqFilter?: string;
    paginationMarker?: GoogleProtobufTimestamp.Timestamp;
    number?: string;
    reverse?: boolean;
};

export type SubscribeJobRequest = {
    __typename?: "SubscribeJobRequest";
    pipeline?: Pipeline;
    details?: boolean;
};

export type DeleteJobRequest = {
    __typename?: "DeleteJobRequest";
    job?: Job;
};

export type StopJobRequest = {
    __typename?: "StopJobRequest";
    job?: Job;
    reason?: string;
};

export type UpdateJobStateRequest = {
    __typename?: "UpdateJobStateRequest";
    job?: Job;
    state?: JobState;
    reason?: string;
    restart?: string;
    dataProcessed?: string;
    dataSkipped?: string;
    dataFailed?: string;
    dataRecovered?: string;
    dataTotal?: string;
    stats?: ProcessStats;
};

export type GetLogsRequest = {
    __typename?: "GetLogsRequest";
    pipeline?: Pipeline;
    job?: Job;
    dataFilters?: string[];
    datum?: Datum;
    master?: boolean;
    follow?: boolean;
    tail?: string;
    useLokiBackend?: boolean;
    since?: GoogleProtobufDuration.Duration;
};

export type LogMessage = {
    __typename?: "LogMessage";
    projectName?: string;
    pipelineName?: string;
    jobId?: string;
    workerId?: string;
    datumId?: string;
    master?: boolean;
    data?: InputFile[];
    user?: boolean;
    ts?: GoogleProtobufTimestamp.Timestamp;
    message?: string;
};

export type RestartDatumRequest = {
    __typename?: "RestartDatumRequest";
    job?: Job;
    dataFilters?: string[];
};

export type InspectDatumRequest = {
    __typename?: "InspectDatumRequest";
    datum?: Datum;
};

export type ListDatumRequestFilter = {
    __typename?: "ListDatumRequestFilter";
    state?: DatumState[];
};

export type ListDatumRequest = {
    __typename?: "ListDatumRequest";
    job?: Job;
    input?: Input;
    filter?: ListDatumRequestFilter;
    paginationMarker?: string;
    number?: string;
    reverse?: boolean;
};

export type CreateDatumRequest = {
    __typename?: "CreateDatumRequest";
    input?: Input;
    number?: string;
};

export type DatumSetSpec = {
    __typename?: "DatumSetSpec";
    number?: string;
    sizeBytes?: string;
    perWorker?: string;
};

export type SchedulingSpec = {
    __typename?: "SchedulingSpec";
    nodeSelector?: { [key: string]: string; };
    priorityClassName?: string;
};

export type RerunPipelineRequest = {
    __typename?: "RerunPipelineRequest";
    pipeline?: Pipeline;
    reprocess?: boolean;
};

export type CreatePipelineRequest = {
    __typename?: "CreatePipelineRequest";
    pipeline?: Pipeline;
    tfJob?: TFJob;
    transform?: Transform;
    parallelismSpec?: ParallelismSpec;
    egress?: Egress;
    update?: boolean;
    outputBranch?: string;
    s3Out?: boolean;
    resourceRequests?: ResourceSpec;
    resourceLimits?: ResourceSpec;
    sidecarResourceLimits?: ResourceSpec;
    input?: Input;
    description?: string;
    reprocess?: boolean;
    service?: Service;
    spout?: Spout;
    datumSetSpec?: DatumSetSpec;
    datumTimeout?: GoogleProtobufDuration.Duration;
    jobTimeout?: GoogleProtobufDuration.Duration;
    salt?: string;
    datumTries?: string;
    schedulingSpec?: SchedulingSpec;
    podSpec?: string;
    podPatch?: string;
    specCommit?: Pfs_v2Pfs.Commit;
    metadata?: Metadata;
    reprocessSpec?: string;
    autoscaling?: boolean;
    tolerations?: Toleration[];
    sidecarResourceRequests?: ResourceSpec;
    dryRun?: boolean;
    determined?: Determined;
    maximumExpectedUptime?: GoogleProtobufDuration.Duration;
};

export type CreatePipelineV2Request = {
    __typename?: "CreatePipelineV2Request";
    createPipelineRequestJson?: string;
    dryRun?: boolean;
    update?: boolean;
    reprocess?: boolean;
};

export type CreatePipelineV2Response = {
    __typename?: "CreatePipelineV2Response";
    effectiveCreatePipelineRequestJson?: string;
};

export type InspectPipelineRequest = {
    __typename?: "InspectPipelineRequest";
    pipeline?: Pipeline;
    details?: boolean;
};

export type ListPipelineRequest = {
    __typename?: "ListPipelineRequest";
    pipeline?: Pipeline;
    history?: string;
    details?: boolean;
    jqFilter?: string;
    commitSet?: Pfs_v2Pfs.CommitSet;
    projects?: Pfs_v2Pfs.Project[];
};

export type DeletePipelineRequest = {
    __typename?: "DeletePipelineRequest";
    pipeline?: Pipeline;
    all?: boolean;
    force?: boolean;
    keepRepo?: boolean;
    mustExist?: boolean;
};

export type DeletePipelinesRequest = {
    __typename?: "DeletePipelinesRequest";
    projects?: Pfs_v2Pfs.Project[];
    force?: boolean;
    keepRepo?: boolean;
    all?: boolean;
};

export type DeletePipelinesResponse = {
    __typename?: "DeletePipelinesResponse";
    pipelines?: Pipeline[];
};

export type StartPipelineRequest = {
    __typename?: "StartPipelineRequest";
    pipeline?: Pipeline;
};

export type StopPipelineRequest = {
    __typename?: "StopPipelineRequest";
    pipeline?: Pipeline;
    mustExist?: boolean;
};

export type RunPipelineRequest = {
    __typename?: "RunPipelineRequest";
    pipeline?: Pipeline;
    provenance?: Pfs_v2Pfs.Commit[];
    jobId?: string;
};

export type RunCronRequest = {
    __typename?: "RunCronRequest";
    pipeline?: Pipeline;
};


type BaseCheckStatusRequest = {
    __typename?: "BaseCheckStatusRequest";
};

export type CheckStatusRequest = BaseCheckStatusRequest
    & OneOf<{ all: boolean; project: Pfs_v2Pfs.Project; }>;

export type CheckStatusResponse = {
    __typename?: "CheckStatusResponse";
    project?: Pfs_v2Pfs.Project;
    pipeline?: Pipeline;
    alerts?: string[];
};

export type CreateSecretRequest = {
    __typename?: "CreateSecretRequest";
    file?: Uint8Array;
};

export type DeleteSecretRequest = {
    __typename?: "DeleteSecretRequest";
    secret?: Secret;
};

export type InspectSecretRequest = {
    __typename?: "InspectSecretRequest";
    secret?: Secret;
};

export type Secret = {
    __typename?: "Secret";
    name?: string;
};

export type SecretInfo = {
    __typename?: "SecretInfo";
    secret?: Secret;
    type?: string;
    creationTimestamp?: GoogleProtobufTimestamp.Timestamp;
};

export type SecretInfos = {
    __typename?: "SecretInfos";
    secretInfo?: SecretInfo[];
};

export type ActivateAuthRequest = {
    __typename?: "ActivateAuthRequest";
};

export type ActivateAuthResponse = {
    __typename?: "ActivateAuthResponse";
};

export type RunLoadTestRequest = {
    __typename?: "RunLoadTestRequest";
    dagSpec?: string;
    loadSpec?: string;
    seed?: string;
    parallelism?: string;
    podPatch?: string;
    stateId?: string;
};

export type RunLoadTestResponse = {
    __typename?: "RunLoadTestResponse";
    error?: string;
    stateId?: string;
};

export type RenderTemplateRequest = {
    __typename?: "RenderTemplateRequest";
    template?: string;
    args?: { [key: string]: string; };
};

export type RenderTemplateResponse = {
    __typename?: "RenderTemplateResponse";
    json?: string;
    specs?: CreatePipelineRequest[];
};

export type LokiRequest = {
    __typename?: "LokiRequest";
    since?: GoogleProtobufDuration.Duration;
    query?: string;
};

export type LokiLogMessage = {
    __typename?: "LokiLogMessage";
    message?: string;
};

export type ClusterDefaults = {
    __typename?: "ClusterDefaults";
    createPipelineRequest?: CreatePipelineRequest;
};

export type GetClusterDefaultsRequest = {
    __typename?: "GetClusterDefaultsRequest";
};

export type GetClusterDefaultsResponse = {
    __typename?: "GetClusterDefaultsResponse";
    clusterDefaultsJson?: string;
};

export type SetClusterDefaultsRequest = {
    __typename?: "SetClusterDefaultsRequest";
    regenerate?: boolean;
    reprocess?: boolean;
    dryRun?: boolean;
    clusterDefaultsJson?: string;
};

export type SetClusterDefaultsResponse = {
    __typename?: "SetClusterDefaultsResponse";
    affectedPipelines?: Pipeline[];
};

export type CreatePipelineTransaction = {
    __typename?: "CreatePipelineTransaction";
    createPipelineRequest?: CreatePipelineRequest;
    userJson?: string;
    effectiveJson?: string;
};

export type ProjectDefaults = {
    __typename?: "ProjectDefaults";
    createPipelineRequest?: CreatePipelineRequest;
};

export type GetProjectDefaultsRequest = {
    __typename?: "GetProjectDefaultsRequest";
    project?: Pfs_v2Pfs.Project;
};

export type GetProjectDefaultsResponse = {
    __typename?: "GetProjectDefaultsResponse";
    projectDefaultsJson?: string;
};

export type SetProjectDefaultsRequest = {
    __typename?: "SetProjectDefaultsRequest";
    project?: Pfs_v2Pfs.Project;
    regenerate?: boolean;
    reprocess?: boolean;
    dryRun?: boolean;
    projectDefaultsJson?: string;
};

export type SetProjectDefaultsResponse = {
    __typename?: "SetProjectDefaultsResponse";
    affectedPipelines?: Pipeline[];
};

export class API {
    static InspectJob(req: InspectJobRequest, initReq?: fm.InitReq): Promise<JobInfo> {
        return fm.fetchReq<InspectJobRequest, JobInfo>(`/pps_v2.API/InspectJob`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectJobSet(req: InspectJobSetRequest, entityNotifier?: fm.NotifyStreamEntityArrival<JobInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<InspectJobSetRequest, JobInfo>(`/pps_v2.API/InspectJobSet`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListJob(req: ListJobRequest, entityNotifier?: fm.NotifyStreamEntityArrival<JobInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ListJobRequest, JobInfo>(`/pps_v2.API/ListJob`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListJobSet(req: ListJobSetRequest, entityNotifier?: fm.NotifyStreamEntityArrival<JobSetInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ListJobSetRequest, JobSetInfo>(`/pps_v2.API/ListJobSet`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static SubscribeJob(req: SubscribeJobRequest, entityNotifier?: fm.NotifyStreamEntityArrival<JobInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<SubscribeJobRequest, JobInfo>(`/pps_v2.API/SubscribeJob`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteJob(req: DeleteJobRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<DeleteJobRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/DeleteJob`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static StopJob(req: StopJobRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<StopJobRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/StopJob`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectDatum(req: InspectDatumRequest, initReq?: fm.InitReq): Promise<DatumInfo> {
        return fm.fetchReq<InspectDatumRequest, DatumInfo>(`/pps_v2.API/InspectDatum`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListDatum(req: ListDatumRequest, entityNotifier?: fm.NotifyStreamEntityArrival<DatumInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ListDatumRequest, DatumInfo>(`/pps_v2.API/ListDatum`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RestartDatum(req: RestartDatumRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<RestartDatumRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/RestartDatum`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RerunPipeline(req: RerunPipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<RerunPipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/RerunPipeline`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static CreatePipeline(req: CreatePipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<CreatePipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/CreatePipeline`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static CreatePipelineV2(req: CreatePipelineV2Request, initReq?: fm.InitReq): Promise<CreatePipelineV2Response> {
        return fm.fetchReq<CreatePipelineV2Request, CreatePipelineV2Response>(`/pps_v2.API/CreatePipelineV2`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectPipeline(req: InspectPipelineRequest, initReq?: fm.InitReq): Promise<PipelineInfo> {
        return fm.fetchReq<InspectPipelineRequest, PipelineInfo>(`/pps_v2.API/InspectPipeline`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListPipeline(req: ListPipelineRequest, entityNotifier?: fm.NotifyStreamEntityArrival<PipelineInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<ListPipelineRequest, PipelineInfo>(`/pps_v2.API/ListPipeline`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeletePipeline(req: DeletePipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<DeletePipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/DeletePipeline`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeletePipelines(req: DeletePipelinesRequest, initReq?: fm.InitReq): Promise<DeletePipelinesResponse> {
        return fm.fetchReq<DeletePipelinesRequest, DeletePipelinesResponse>(`/pps_v2.API/DeletePipelines`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static StartPipeline(req: StartPipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<StartPipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/StartPipeline`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static StopPipeline(req: StopPipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<StopPipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/StopPipeline`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RunPipeline(req: RunPipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<RunPipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/RunPipeline`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RunCron(req: RunCronRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<RunCronRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/RunCron`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static CheckStatus(req: CheckStatusRequest, entityNotifier?: fm.NotifyStreamEntityArrival<CheckStatusResponse>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<CheckStatusRequest, CheckStatusResponse>(`/pps_v2.API/CheckStatus`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static CreateSecret(req: CreateSecretRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<CreateSecretRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/CreateSecret`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteSecret(req: DeleteSecretRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<DeleteSecretRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/DeleteSecret`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListSecret(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<SecretInfos> {
        return fm.fetchReq<GoogleProtobufEmpty.Empty, SecretInfos>(`/pps_v2.API/ListSecret`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static InspectSecret(req: InspectSecretRequest, initReq?: fm.InitReq): Promise<SecretInfo> {
        return fm.fetchReq<InspectSecretRequest, SecretInfo>(`/pps_v2.API/InspectSecret`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteAll(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<GoogleProtobufEmpty.Empty, GoogleProtobufEmpty.Empty>(`/pps_v2.API/DeleteAll`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetLogs(req: GetLogsRequest, entityNotifier?: fm.NotifyStreamEntityArrival<LogMessage>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<GetLogsRequest, LogMessage>(`/pps_v2.API/GetLogs`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ActivateAuth(req: ActivateAuthRequest, initReq?: fm.InitReq): Promise<ActivateAuthResponse> {
        return fm.fetchReq<ActivateAuthRequest, ActivateAuthResponse>(`/pps_v2.API/ActivateAuth`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static UpdateJobState(req: UpdateJobStateRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
        return fm.fetchReq<UpdateJobStateRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/UpdateJobState`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RunLoadTest(req: RunLoadTestRequest, initReq?: fm.InitReq): Promise<RunLoadTestResponse> {
        return fm.fetchReq<RunLoadTestRequest, RunLoadTestResponse>(`/pps_v2.API/RunLoadTest`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RunLoadTestDefault(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<RunLoadTestResponse> {
        return fm.fetchReq<GoogleProtobufEmpty.Empty, RunLoadTestResponse>(`/pps_v2.API/RunLoadTestDefault`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RenderTemplate(req: RenderTemplateRequest, initReq?: fm.InitReq): Promise<RenderTemplateResponse> {
        return fm.fetchReq<RenderTemplateRequest, RenderTemplateResponse>(`/pps_v2.API/RenderTemplate`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListTask(req: TaskapiTask.ListTaskRequest, entityNotifier?: fm.NotifyStreamEntityArrival<TaskapiTask.TaskInfo>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<TaskapiTask.ListTaskRequest, TaskapiTask.TaskInfo>(`/pps_v2.API/ListTask`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetKubeEvents(req: LokiRequest, entityNotifier?: fm.NotifyStreamEntityArrival<LokiLogMessage>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<LokiRequest, LokiLogMessage>(`/pps_v2.API/GetKubeEvents`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static QueryLoki(req: LokiRequest, entityNotifier?: fm.NotifyStreamEntityArrival<LokiLogMessage>, initReq?: fm.InitReq): Promise<void> {
        return fm.fetchStreamingRequest<LokiRequest, LokiLogMessage>(`/pps_v2.API/QueryLoki`, entityNotifier, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetClusterDefaults(req: GetClusterDefaultsRequest, initReq?: fm.InitReq): Promise<GetClusterDefaultsResponse> {
        return fm.fetchReq<GetClusterDefaultsRequest, GetClusterDefaultsResponse>(`/pps_v2.API/GetClusterDefaults`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static SetClusterDefaults(req: SetClusterDefaultsRequest, initReq?: fm.InitReq): Promise<SetClusterDefaultsResponse> {
        return fm.fetchReq<SetClusterDefaultsRequest, SetClusterDefaultsResponse>(`/pps_v2.API/SetClusterDefaults`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetProjectDefaults(req: GetProjectDefaultsRequest, initReq?: fm.InitReq): Promise<GetProjectDefaultsResponse> {
        return fm.fetchReq<GetProjectDefaultsRequest, GetProjectDefaultsResponse>(`/pps_v2.API/GetProjectDefaults`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static SetProjectDefaults(req: SetProjectDefaultsRequest, initReq?: fm.InitReq): Promise<SetProjectDefaultsResponse> {
        return fm.fetchReq<SetProjectDefaultsRequest, SetProjectDefaultsResponse>(`/pps_v2.API/SetProjectDefaults`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
}
