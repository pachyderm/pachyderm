/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb"
import * as GoogleProtobufStruct from "../google/protobuf/struct.pb"
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb"
import * as Pps_v2Pps from "../pps/pps.pb"

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (
    keyof T extends infer K ?
      (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
        : never)
    : never);

export enum LogLevel {
  LOG_LEVEL_DEBUG = "LOG_LEVEL_DEBUG",
  LOG_LEVEL_INFO = "LOG_LEVEL_INFO",
  LOG_LEVEL_ERROR = "LOG_LEVEL_ERROR",
}

export enum LogFormat {
  LOG_FORMAT_UNKNOWN = "LOG_FORMAT_UNKNOWN",
  LOG_FORMAT_VERBATIM_WITH_TIMESTAMP = "LOG_FORMAT_VERBATIM_WITH_TIMESTAMP",
  LOG_FORMAT_PARSED_JSON = "LOG_FORMAT_PARSED_JSON",
  LOG_FORMAT_PPS_LOGMESSAGE = "LOG_FORMAT_PPS_LOGMESSAGE",
}


type BaseLogQuery = {
}

export type LogQuery = BaseLogQuery
  & OneOf<{ user: UserLogQuery; admin: AdminLogQuery }>


type BaseAdminLogQuery = {
}

export type AdminLogQuery = BaseAdminLogQuery
  & OneOf<{ logql: string; pod: string; podContainer: PodContainer; app: string; master: PipelineLogQuery; storage: PipelineLogQuery; user: UserLogQuery }>

export type PodContainer = {
  pod?: string
  container?: string
}


type BaseUserLogQuery = {
}

export type UserLogQuery = BaseUserLogQuery
  & OneOf<{ project: string; pipeline: PipelineLogQuery; datum: string; job: string; pipelineJob: PipelineJobLogQuery }>

export type PipelineLogQuery = {
  project?: string
  pipeline?: string
}

export type PipelineJobLogQuery = {
  pipeline?: PipelineLogQuery
  job?: string
}

export type LogFilter = {
  timeRange?: TimeRangeLogFilter
  limit?: string
  regex?: RegexLogFilter
  level?: LogLevel
}

export type TimeRangeLogFilter = {
  from?: GoogleProtobufTimestamp.Timestamp
  until?: GoogleProtobufTimestamp.Timestamp
}

export type RegexLogFilter = {
  pattern?: string
  negate?: boolean
}

export type GetLogsRequest = {
  query?: LogQuery
  filter?: LogFilter
  tail?: boolean
  wantPagingHint?: boolean
  logFormat?: LogFormat
}


type BaseGetLogsResponse = {
}

export type GetLogsResponse = BaseGetLogsResponse
  & OneOf<{ pagingHint: PagingHint; log: LogMessage }>

export type PagingHint = {
  older?: GetLogsRequest
  newer?: GetLogsRequest
}


type BaseLogMessage = {
}

export type LogMessage = BaseLogMessage
  & OneOf<{ verbatim: VerbatimLogMessage; json: ParsedJSONLogMessage; ppsLogMessage: Pps_v2Pps.LogMessage }>

export type VerbatimLogMessage = {
  line?: Uint8Array
  timestamp?: GoogleProtobufTimestamp.Timestamp
}

export type ParsedJSONLogMessage = {
  verbatim?: VerbatimLogMessage
  object?: GoogleProtobufStruct.Struct
  nativeTimestamp?: GoogleProtobufTimestamp.Timestamp
  ppsLogMessage?: Pps_v2Pps.LogMessage
}

export class API {
  static GetLogs(req: GetLogsRequest, entityNotifier?: fm.NotifyStreamEntityArrival<GetLogsResponse>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<GetLogsRequest, GetLogsResponse>(`/logs.API/GetLogs`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}