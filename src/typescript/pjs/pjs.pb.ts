/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb"
import * as GoogleProtobufAny from "../google/protobuf/any.pb"

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (
    keyof T extends infer K ?
      (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
        : never)
    : never);

export enum JobState {
  JOB_STATE_UNSPECIFIED = "JOB_STATE_UNSPECIFIED",
  QUEUED = "QUEUED",
  PROCESSING = "PROCESSING",
  DONE = "DONE",
}

export enum JobErrorCode {
  JOB_ERROR_CODE_UNSPECIFIED = "JOB_ERROR_CODE_UNSPECIFIED",
  FAILED = "FAILED",
  DISCONNECTED = "DISCONNECTED",
  CANCELED = "CANCELED",
}

export type Job = {
  id?: string
}


type BaseJobInfo = {
  job?: Job
  parentJob?: Job
  state?: JobState
  spec?: GoogleProtobufAny.Any
  input?: QueueElement
}

export type JobInfo = BaseJobInfo
  & OneOf<{ output: QueueElement; error: JobErrorCode }>

export type JobInfoDetails = {
  jobInfo?: JobInfo
}

export type Queue = {
  id?: Uint8Array
}

export type QueueInfo = {
  queue?: Queue
  spec?: GoogleProtobufAny.Any
}

export type QueueInfoDetails = {
  queueInfo?: QueueInfo
  size?: string
}

export type QueueElement = {
  data?: Uint8Array
  filesets?: string[]
}

export type CreateJobRequest = {
  context?: string
  spec?: GoogleProtobufAny.Any
  input?: QueueElement
  cacheRead?: boolean
  cacheWrite?: boolean
}

export type CreateJobResponse = {
  id?: Job
}

export type CancelJobRequest = {
  context?: string
  job?: Job
}

export type CancelJobResponse = {
}

export type DeleteJobRequest = {
  context?: string
  job?: Job
}

export type DeleteJobResponse = {
}

export type ListJobRequest = {
  context?: string
  job?: Job
}

export type WalkJobRequest = {
  context?: string
  job?: Job
}

export type InspectJobRequest = {
  context?: string
  job?: Job
}


type BaseProcessQueueRequest = {
  queue?: Queue
}

export type ProcessQueueRequest = BaseProcessQueueRequest
  & OneOf<{ output: QueueElement; failed: boolean }>

export type ProcessQueueResponse = {
  context?: string
  input?: QueueElement
}

export type ListQueueRequest = {
}

export type InspectQueueRequest = {
  queue?: Queue
}

export class API {
  static CreateJob(req: CreateJobRequest, initReq?: fm.InitReq): Promise<CreateJobResponse> {
    return fm.fetchReq<CreateJobRequest, CreateJobResponse>(`/pjs.API/CreateJob`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static CancelJob(req: CancelJobRequest, initReq?: fm.InitReq): Promise<CancelJobResponse> {
    return fm.fetchReq<CancelJobRequest, CancelJobResponse>(`/pjs.API/CancelJob`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static DeleteJob(req: DeleteJobRequest, initReq?: fm.InitReq): Promise<DeleteJobResponse> {
    return fm.fetchReq<DeleteJobRequest, DeleteJobResponse>(`/pjs.API/DeleteJob`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListJob(req: ListJobRequest, entityNotifier?: fm.NotifyStreamEntityArrival<JobInfo>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ListJobRequest, JobInfo>(`/pjs.API/ListJob`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static WalkJob(req: WalkJobRequest, entityNotifier?: fm.NotifyStreamEntityArrival<JobInfo>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<WalkJobRequest, JobInfo>(`/pjs.API/WalkJob`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static InspectJob(req: InspectJobRequest, initReq?: fm.InitReq): Promise<JobInfoDetails> {
    return fm.fetchReq<InspectJobRequest, JobInfoDetails>(`/pjs.API/InspectJob`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListQueue(req: ListQueueRequest, entityNotifier?: fm.NotifyStreamEntityArrival<QueueInfo>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ListQueueRequest, QueueInfo>(`/pjs.API/ListQueue`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static InspectQueue(req: InspectQueueRequest, initReq?: fm.InitReq): Promise<QueueInfoDetails> {
    return fm.fetchReq<InspectQueueRequest, QueueInfoDetails>(`/pjs.API/InspectQueue`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}