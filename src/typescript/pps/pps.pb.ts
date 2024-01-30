/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb"
import * as GoogleProtobufDuration from "../google/protobuf/duration.pb"
import * as GoogleProtobufEmpty from "../google/protobuf/empty.pb"
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb"
import * as GoogleProtobufWrappers from "../google/protobuf/wrappers.pb"
import * as Pfs_v2Pfs from "../pfs/pfs.pb"
import * as TaskapiTask from "../task/task.pb"

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
  name?: string
  key?: string
  mountPath?: string
  envVar?: string
}

export type Transform = {
  image?: string
  cmd?: string[]
  errCmd?: string[]
  env?: {[key: string]: string}
  secrets?: SecretMount[]
  imagePullSecrets?: string[]
  stdin?: string[]
  errStdin?: string[]
  acceptReturnCode?: string[]
  debug?: boolean
  user?: string
  workingDir?: string
  dockerfile?: string
  memoryVolume?: boolean
  datumBatching?: boolean
}

export type TFJob = {
  tfJob?: string
}


type BaseEgress = {
  uRL?: string
}

export type Egress = BaseEgress
  & OneOf<{ objectStorage: Pfs_v2Pfs.ObjectStorageEgress; sqlDatabase: Pfs_v2Pfs.SQLDatabaseEgress }>

export type Determined = {
  workspaces?: string[]
}

export type Job = {
  pipeline?: Pipeline
  id?: string
}

export type Metadata = {
  annotations?: {[key: string]: string}
  labels?: {[key: string]: string}
}

export type Service = {
  internalPort?: number
  externalPort?: number
  ip?: string
  type?: string
}

export type Spout = {
  service?: Service
}

export type PFSInput = {
  project?: string
  name?: string
  repo?: string
  repoType?: string
  branch?: string
  commit?: string
  glob?: string
  joinOn?: string
  outerJoin?: boolean
  groupBy?: string
  lazy?: boolean
  emptyFiles?: boolean
  s3?: boolean
  trigger?: Pfs_v2Pfs.Trigger
}

export type CronInput = {
  name?: string
  project?: string
  repo?: string
  commit?: string
  spec?: string
  overwrite?: boolean
  start?: GoogleProtobufTimestamp.Timestamp
}

export type Input = {
  pfs?: PFSInput
  join?: Input[]
  group?: Input[]
  cross?: Input[]
  union?: Input[]
  cron?: CronInput
}

export type JobInput = {
  name?: string
  commit?: Pfs_v2Pfs.Commit
  glob?: string
  lazy?: boolean
}

export type ParallelismSpec = {
  constant?: string
}

export type InputFile = {
  path?: string
  hash?: Uint8Array
}

export type Datum = {
  job?: Job
  id?: string
}

export type DatumInfo = {
  datum?: Datum
  state?: DatumState
  stats?: ProcessStats
  pfsState?: Pfs_v2Pfs.File
  data?: Pfs_v2Pfs.FileInfo[]
  imageId?: string
}

export type Aggregate = {
  count?: string
  mean?: number
  stddev?: number
  fifthPercentile?: number
  ninetyFifthPercentile?: number
}

export type ProcessStats = {
  downloadTime?: GoogleProtobufDuration.Duration
  processTime?: GoogleProtobufDuration.Duration
  uploadTime?: GoogleProtobufDuration.Duration
  downloadBytes?: string
  uploadBytes?: string
}

export type AggregateProcessStats = {
  downloadTime?: Aggregate
  processTime?: Aggregate
  uploadTime?: Aggregate
  downloadBytes?: Aggregate
  uploadBytes?: Aggregate
}

export type WorkerStatus = {
  workerId?: string
  jobId?: string
  datumStatus?: DatumStatus
}

export type DatumStatus = {
  started?: GoogleProtobufTimestamp.Timestamp
  data?: InputFile[]
}

export type ResourceSpec = {
  cpu?: number
  memory?: string
  gpu?: GPUSpec
  disk?: string
}

export type GPUSpec = {
  type?: string
  number?: string
}

export type JobSetInfo = {
  jobSet?: JobSet
  jobs?: JobInfo[]
}

export type JobInfoDetails = {
  transform?: Transform
  parallelismSpec?: ParallelismSpec
  egress?: Egress
  service?: Service
  spout?: Spout
  workerStatus?: WorkerStatus[]
  resourceRequests?: ResourceSpec
  resourceLimits?: ResourceSpec
  sidecarResourceLimits?: ResourceSpec
  input?: Input
  salt?: string
  datumSetSpec?: DatumSetSpec
  datumTimeout?: GoogleProtobufDuration.Duration
  jobTimeout?: GoogleProtobufDuration.Duration
  datumTries?: string
  schedulingSpec?: SchedulingSpec
  podSpec?: string
  podPatch?: string
  sidecarResourceRequests?: ResourceSpec
}

export type JobInfo = {
  job?: Job
  pipelineVersion?: string
  outputCommit?: Pfs_v2Pfs.Commit
  restart?: string
  dataProcessed?: string
  dataSkipped?: string
  dataTotal?: string
  dataFailed?: string
  dataRecovered?: string
  stats?: ProcessStats
  state?: JobState
  reason?: string
  created?: GoogleProtobufTimestamp.Timestamp
  started?: GoogleProtobufTimestamp.Timestamp
  finished?: GoogleProtobufTimestamp.Timestamp
  details?: JobInfoDetails
  authToken?: string
}

export type Worker = {
  name?: string
  state?: WorkerState
}

export type Pipeline = {
  project?: Pfs_v2Pfs.Project
  name?: string
}

export type Toleration = {
  key?: string
  operator?: TolerationOperator
  value?: string
  effect?: TaintEffect
  tolerationSeconds?: GoogleProtobufWrappers.Int64Value
}

export type PipelineInfoDetails = {
  transform?: Transform
  tfJob?: TFJob
  parallelismSpec?: ParallelismSpec
  egress?: Egress
  createdAt?: GoogleProtobufTimestamp.Timestamp
  recentError?: string
  workersRequested?: string
  workersAvailable?: string
  outputBranch?: string
  resourceRequests?: ResourceSpec
  resourceLimits?: ResourceSpec
  sidecarResourceLimits?: ResourceSpec
  input?: Input
  description?: string
  salt?: string
  reason?: string
  service?: Service
  spout?: Spout
  datumSetSpec?: DatumSetSpec
  datumTimeout?: GoogleProtobufDuration.Duration
  jobTimeout?: GoogleProtobufDuration.Duration
  datumTries?: string
  schedulingSpec?: SchedulingSpec
  podSpec?: string
  podPatch?: string
  s3Out?: boolean
  metadata?: Metadata
  reprocessSpec?: string
  unclaimedTasks?: string
  workerRc?: string
  autoscaling?: boolean
  tolerations?: Toleration[]
  sidecarResourceRequests?: ResourceSpec
  determined?: Determined
  maximumExpectedUptime?: GoogleProtobufDuration.Duration
  workersStartedAt?: GoogleProtobufTimestamp.Timestamp
}

export type PipelineInfo = {
  pipeline?: Pipeline
  version?: string
  specCommit?: Pfs_v2Pfs.Commit
  stopped?: boolean
  state?: PipelineState
  reason?: string
  lastJobState?: JobState
  parallelism?: string
  type?: PipelineInfoPipelineType
  authToken?: string
  details?: PipelineInfoDetails
  userSpecJson?: string
  effectiveSpecJson?: string
}

export type PipelineInfos = {
  pipelineInfo?: PipelineInfo[]
}

export type JobSet = {
  id?: string
}

export type InspectJobSetRequest = {
  jobSet?: JobSet
  wait?: boolean
  details?: boolean
}

export type ListJobSetRequest = {
  details?: boolean
  projects?: Pfs_v2Pfs.Project[]
  paginationMarker?: GoogleProtobufTimestamp.Timestamp
  number?: string
  reverse?: boolean
  jqFilter?: string
}

export type InspectJobRequest = {
  job?: Job
  wait?: boolean
  details?: boolean
}

export type ListJobRequest = {
  projects?: Pfs_v2Pfs.Project[]
  pipeline?: Pipeline
  inputCommit?: Pfs_v2Pfs.Commit[]
  history?: string
  details?: boolean
  jqFilter?: string
  paginationMarker?: GoogleProtobufTimestamp.Timestamp
  number?: string
  reverse?: boolean
}

export type SubscribeJobRequest = {
  pipeline?: Pipeline
  details?: boolean
}

export type DeleteJobRequest = {
  job?: Job
}

export type StopJobRequest = {
  job?: Job
  reason?: string
}

export type UpdateJobStateRequest = {
  job?: Job
  state?: JobState
  reason?: string
  restart?: string
  dataProcessed?: string
  dataSkipped?: string
  dataFailed?: string
  dataRecovered?: string
  dataTotal?: string
  stats?: ProcessStats
}

export type GetLogsRequest = {
  pipeline?: Pipeline
  job?: Job
  dataFilters?: string[]
  datum?: Datum
  master?: boolean
  follow?: boolean
  tail?: string
  useLokiBackend?: boolean
  since?: GoogleProtobufDuration.Duration
}

export type LogMessage = {
  projectName?: string
  pipelineName?: string
  jobId?: string
  workerId?: string
  datumId?: string
  master?: boolean
  data?: InputFile[]
  user?: boolean
  ts?: GoogleProtobufTimestamp.Timestamp
  message?: string
}

export type RestartDatumRequest = {
  job?: Job
  dataFilters?: string[]
}

export type InspectDatumRequest = {
  datum?: Datum
}

export type ListDatumRequestFilter = {
  state?: DatumState[]
}

export type ListDatumRequest = {
  job?: Job
  input?: Input
  filter?: ListDatumRequestFilter
  paginationMarker?: string
  number?: string
  reverse?: boolean
}

export type CreateDatumRequest = {
  input?: Input
  number?: string
}

export type DatumSetSpec = {
  number?: string
  sizeBytes?: string
  perWorker?: string
}

export type SchedulingSpec = {
  nodeSelector?: {[key: string]: string}
  priorityClassName?: string
}

export type RerunPipelineRequest = {
  pipeline?: Pipeline
  reprocess?: boolean
}

export type CreatePipelineRequest = {
  pipeline?: Pipeline
  tfJob?: TFJob
  transform?: Transform
  parallelismSpec?: ParallelismSpec
  egress?: Egress
  update?: boolean
  outputBranch?: string
  s3Out?: boolean
  resourceRequests?: ResourceSpec
  resourceLimits?: ResourceSpec
  sidecarResourceLimits?: ResourceSpec
  input?: Input
  description?: string
  reprocess?: boolean
  service?: Service
  spout?: Spout
  datumSetSpec?: DatumSetSpec
  datumTimeout?: GoogleProtobufDuration.Duration
  jobTimeout?: GoogleProtobufDuration.Duration
  salt?: string
  datumTries?: string
  schedulingSpec?: SchedulingSpec
  podSpec?: string
  podPatch?: string
  specCommit?: Pfs_v2Pfs.Commit
  metadata?: Metadata
  reprocessSpec?: string
  autoscaling?: boolean
  tolerations?: Toleration[]
  sidecarResourceRequests?: ResourceSpec
  dryRun?: boolean
  determined?: Determined
  maximumExpectedUptime?: GoogleProtobufDuration.Duration
}

export type CreatePipelineV2Request = {
  createPipelineRequestJson?: string
  dryRun?: boolean
  update?: boolean
  reprocess?: boolean
}

export type CreatePipelineV2Response = {
  effectiveCreatePipelineRequestJson?: string
}

export type InspectPipelineRequest = {
  pipeline?: Pipeline
  details?: boolean
}

export type ListPipelineRequest = {
  pipeline?: Pipeline
  history?: string
  details?: boolean
  jqFilter?: string
  commitSet?: Pfs_v2Pfs.CommitSet
  projects?: Pfs_v2Pfs.Project[]
}

export type DeletePipelineRequest = {
  pipeline?: Pipeline
  all?: boolean
  force?: boolean
  keepRepo?: boolean
  mustExist?: boolean
}

export type DeletePipelinesRequest = {
  projects?: Pfs_v2Pfs.Project[]
  force?: boolean
  keepRepo?: boolean
  all?: boolean
}

export type DeletePipelinesResponse = {
  pipelines?: Pipeline[]
}

export type StartPipelineRequest = {
  pipeline?: Pipeline
}

export type StopPipelineRequest = {
  pipeline?: Pipeline
  mustExist?: boolean
}

export type RunPipelineRequest = {
  pipeline?: Pipeline
  provenance?: Pfs_v2Pfs.Commit[]
  jobId?: string
}

export type RunCronRequest = {
  pipeline?: Pipeline
}


type BaseCheckStatusRequest = {
}

export type CheckStatusRequest = BaseCheckStatusRequest
  & OneOf<{ all: boolean; project: Pfs_v2Pfs.Project }>

export type CheckStatusResponse = {
  project?: Pfs_v2Pfs.Project
  pipeline?: Pipeline
  alerts?: string[]
}

export type CreateSecretRequest = {
  file?: Uint8Array
}

export type DeleteSecretRequest = {
  secret?: Secret
}

export type InspectSecretRequest = {
  secret?: Secret
}

export type Secret = {
  name?: string
}

export type SecretInfo = {
  secret?: Secret
  type?: string
  creationTimestamp?: GoogleProtobufTimestamp.Timestamp
}

export type SecretInfos = {
  secretInfo?: SecretInfo[]
}

export type ActivateAuthRequest = {
}

export type ActivateAuthResponse = {
}

export type RunLoadTestRequest = {
  dagSpec?: string
  loadSpec?: string
  seed?: string
  parallelism?: string
  podPatch?: string
  stateId?: string
}

export type RunLoadTestResponse = {
  error?: string
  stateId?: string
}

export type RenderTemplateRequest = {
  template?: string
  args?: {[key: string]: string}
}

export type RenderTemplateResponse = {
  json?: string
  specs?: CreatePipelineRequest[]
}

export type LokiRequest = {
  since?: GoogleProtobufDuration.Duration
  query?: string
}

export type LokiLogMessage = {
  message?: string
}

export type ClusterDefaults = {
  createPipelineRequest?: CreatePipelineRequest
}

export type GetClusterDefaultsRequest = {
}

export type GetClusterDefaultsResponse = {
  clusterDefaultsJson?: string
}

export type SetClusterDefaultsRequest = {
  regenerate?: boolean
  reprocess?: boolean
  dryRun?: boolean
  clusterDefaultsJson?: string
}

export type SetClusterDefaultsResponse = {
  affectedPipelines?: Pipeline[]
}

export type CreatePipelineTransaction = {
  createPipelineRequest?: CreatePipelineRequest
  userJson?: string
  effectiveJson?: string
}

export type ProjectDefaults = {
  createPipelineRequest?: CreatePipelineRequest
}

export type GetProjectDefaultsRequest = {
  project?: Pfs_v2Pfs.Project
}

export type GetProjectDefaultsResponse = {
  projectDefaultsJson?: string
}

export type SetProjectDefaultsRequest = {
  project?: Pfs_v2Pfs.Project
  regenerate?: boolean
  reprocess?: boolean
  dryRun?: boolean
  projectDefaultsJson?: string
}

export type SetProjectDefaultsResponse = {
  affectedPipelines?: Pipeline[]
}

export class API {
  static InspectJob(req: InspectJobRequest, initReq?: fm.InitReq): Promise<JobInfo> {
    return fm.fetchReq<InspectJobRequest, JobInfo>(`/pps_v2.API/InspectJob`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static InspectJobSet(req: InspectJobSetRequest, entityNotifier?: fm.NotifyStreamEntityArrival<JobInfo>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<InspectJobSetRequest, JobInfo>(`/pps_v2.API/InspectJobSet`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListJob(req: ListJobRequest, entityNotifier?: fm.NotifyStreamEntityArrival<JobInfo>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ListJobRequest, JobInfo>(`/pps_v2.API/ListJob`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListJobSet(req: ListJobSetRequest, entityNotifier?: fm.NotifyStreamEntityArrival<JobSetInfo>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ListJobSetRequest, JobSetInfo>(`/pps_v2.API/ListJobSet`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static SubscribeJob(req: SubscribeJobRequest, entityNotifier?: fm.NotifyStreamEntityArrival<JobInfo>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<SubscribeJobRequest, JobInfo>(`/pps_v2.API/SubscribeJob`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static DeleteJob(req: DeleteJobRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<DeleteJobRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/DeleteJob`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static StopJob(req: StopJobRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<StopJobRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/StopJob`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static InspectDatum(req: InspectDatumRequest, initReq?: fm.InitReq): Promise<DatumInfo> {
    return fm.fetchReq<InspectDatumRequest, DatumInfo>(`/pps_v2.API/InspectDatum`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListDatum(req: ListDatumRequest, entityNotifier?: fm.NotifyStreamEntityArrival<DatumInfo>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ListDatumRequest, DatumInfo>(`/pps_v2.API/ListDatum`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static RestartDatum(req: RestartDatumRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<RestartDatumRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/RestartDatum`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static RerunPipeline(req: RerunPipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<RerunPipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/RerunPipeline`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static CreatePipeline(req: CreatePipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<CreatePipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/CreatePipeline`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static CreatePipelineV2(req: CreatePipelineV2Request, initReq?: fm.InitReq): Promise<CreatePipelineV2Response> {
    return fm.fetchReq<CreatePipelineV2Request, CreatePipelineV2Response>(`/pps_v2.API/CreatePipelineV2`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static InspectPipeline(req: InspectPipelineRequest, initReq?: fm.InitReq): Promise<PipelineInfo> {
    return fm.fetchReq<InspectPipelineRequest, PipelineInfo>(`/pps_v2.API/InspectPipeline`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListPipeline(req: ListPipelineRequest, entityNotifier?: fm.NotifyStreamEntityArrival<PipelineInfo>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ListPipelineRequest, PipelineInfo>(`/pps_v2.API/ListPipeline`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static DeletePipeline(req: DeletePipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<DeletePipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/DeletePipeline`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static DeletePipelines(req: DeletePipelinesRequest, initReq?: fm.InitReq): Promise<DeletePipelinesResponse> {
    return fm.fetchReq<DeletePipelinesRequest, DeletePipelinesResponse>(`/pps_v2.API/DeletePipelines`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static StartPipeline(req: StartPipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<StartPipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/StartPipeline`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static StopPipeline(req: StopPipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<StopPipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/StopPipeline`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static RunPipeline(req: RunPipelineRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<RunPipelineRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/RunPipeline`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static RunCron(req: RunCronRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<RunCronRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/RunCron`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static CheckStatus(req: CheckStatusRequest, entityNotifier?: fm.NotifyStreamEntityArrival<CheckStatusResponse>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<CheckStatusRequest, CheckStatusResponse>(`/pps_v2.API/CheckStatus`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static CreateSecret(req: CreateSecretRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<CreateSecretRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/CreateSecret`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static DeleteSecret(req: DeleteSecretRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<DeleteSecretRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/DeleteSecret`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListSecret(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<SecretInfos> {
    return fm.fetchReq<GoogleProtobufEmpty.Empty, SecretInfos>(`/pps_v2.API/ListSecret`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static InspectSecret(req: InspectSecretRequest, initReq?: fm.InitReq): Promise<SecretInfo> {
    return fm.fetchReq<InspectSecretRequest, SecretInfo>(`/pps_v2.API/InspectSecret`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static DeleteAll(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<GoogleProtobufEmpty.Empty, GoogleProtobufEmpty.Empty>(`/pps_v2.API/DeleteAll`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static GetLogs(req: GetLogsRequest, entityNotifier?: fm.NotifyStreamEntityArrival<LogMessage>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<GetLogsRequest, LogMessage>(`/pps_v2.API/GetLogs`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ActivateAuth(req: ActivateAuthRequest, initReq?: fm.InitReq): Promise<ActivateAuthResponse> {
    return fm.fetchReq<ActivateAuthRequest, ActivateAuthResponse>(`/pps_v2.API/ActivateAuth`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static UpdateJobState(req: UpdateJobStateRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<UpdateJobStateRequest, GoogleProtobufEmpty.Empty>(`/pps_v2.API/UpdateJobState`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static RunLoadTest(req: RunLoadTestRequest, initReq?: fm.InitReq): Promise<RunLoadTestResponse> {
    return fm.fetchReq<RunLoadTestRequest, RunLoadTestResponse>(`/pps_v2.API/RunLoadTest`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static RunLoadTestDefault(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<RunLoadTestResponse> {
    return fm.fetchReq<GoogleProtobufEmpty.Empty, RunLoadTestResponse>(`/pps_v2.API/RunLoadTestDefault`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static RenderTemplate(req: RenderTemplateRequest, initReq?: fm.InitReq): Promise<RenderTemplateResponse> {
    return fm.fetchReq<RenderTemplateRequest, RenderTemplateResponse>(`/pps_v2.API/RenderTemplate`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListTask(req: TaskapiTask.ListTaskRequest, entityNotifier?: fm.NotifyStreamEntityArrival<TaskapiTask.TaskInfo>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<TaskapiTask.ListTaskRequest, TaskapiTask.TaskInfo>(`/pps_v2.API/ListTask`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static GetKubeEvents(req: LokiRequest, entityNotifier?: fm.NotifyStreamEntityArrival<LokiLogMessage>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<LokiRequest, LokiLogMessage>(`/pps_v2.API/GetKubeEvents`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static QueryLoki(req: LokiRequest, entityNotifier?: fm.NotifyStreamEntityArrival<LokiLogMessage>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<LokiRequest, LokiLogMessage>(`/pps_v2.API/QueryLoki`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static GetClusterDefaults(req: GetClusterDefaultsRequest, initReq?: fm.InitReq): Promise<GetClusterDefaultsResponse> {
    return fm.fetchReq<GetClusterDefaultsRequest, GetClusterDefaultsResponse>(`/pps_v2.API/GetClusterDefaults`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static SetClusterDefaults(req: SetClusterDefaultsRequest, initReq?: fm.InitReq): Promise<SetClusterDefaultsResponse> {
    return fm.fetchReq<SetClusterDefaultsRequest, SetClusterDefaultsResponse>(`/pps_v2.API/SetClusterDefaults`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static GetProjectDefaults(req: GetProjectDefaultsRequest, initReq?: fm.InitReq): Promise<GetProjectDefaultsResponse> {
    return fm.fetchReq<GetProjectDefaultsRequest, GetProjectDefaultsResponse>(`/pps_v2.API/GetProjectDefaults`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static SetProjectDefaults(req: SetProjectDefaultsRequest, initReq?: fm.InitReq): Promise<SetProjectDefaultsResponse> {
    return fm.fetchReq<SetProjectDefaultsRequest, SetProjectDefaultsResponse>(`/pps_v2.API/SetProjectDefaults`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}