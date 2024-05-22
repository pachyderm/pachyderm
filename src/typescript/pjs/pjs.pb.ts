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
  JobState_UNSPECIFIED = "JobState_UNSPECIFIED",
  QUEUED = "QUEUED",
  PROCESSING = "PROCESSING",
  DONE = "DONE",
}

export enum JobErrorCode {
  JobErrorCode_UNSPECIFIED = "JobErrorCode_UNSPECIFIED",
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

export type ListJobResponse = {
  id?: Job
  info?: JobInfo
  details?: JobInfoDetails
}

export type WalkJobRequest = {
  context?: string
  job?: Job
}

export type InspectJobRequest = {
  context?: string
  job?: Job
}

export type InspectJobResponse = {
  details?: JobInfoDetails
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

export type ListQueueResponse = {
  id?: Queue
  info?: QueueInfo
  details?: QueueInfoDetails
}

export type InspectQueueRequest = {
  queue?: Queue
}

export type InspectQueueResponse = {
  details?: QueueInfoDetails
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
  static ListJob(req: ListJobRequest, entityNotifier?: fm.NotifyStreamEntityArrival<ListJobResponse>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ListJobRequest, ListJobResponse>(`/pjs.API/ListJob`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static WalkJob(req: WalkJobRequest, entityNotifier?: fm.NotifyStreamEntityArrival<ListJobResponse>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<WalkJobRequest, ListJobResponse>(`/pjs.API/WalkJob`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static InspectJob(req: InspectJobRequest, initReq?: fm.InitReq): Promise<InspectJobResponse> {
    return fm.fetchReq<InspectJobRequest, InspectJobResponse>(`/pjs.API/InspectJob`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListQueue(req: ListQueueRequest, entityNotifier?: fm.NotifyStreamEntityArrival<ListQueueResponse>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ListQueueRequest, ListQueueResponse>(`/pjs.API/ListQueue`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static InspectQueue(req: InspectQueueRequest, initReq?: fm.InitReq): Promise<InspectQueueResponse> {
    return fm.fetchReq<InspectQueueRequest, InspectQueueResponse>(`/pjs.API/InspectQueue`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}