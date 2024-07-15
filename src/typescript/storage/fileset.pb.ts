/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as CdrCdr from "../cdr/cdr.pb"
import * as fm from "../fetch.pb"
import * as GoogleProtobufEmpty from "../google/protobuf/empty.pb"
import * as GoogleProtobufWrappers from "../google/protobuf/wrappers.pb"

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (
    keyof T extends infer K ?
      (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
        : never)
    : never);
export type AppendFile = {
  path?: string
  data?: GoogleProtobufWrappers.BytesValue
}

export type DeleteFile = {
  path?: string
}

export type CopyFile = {
  filesetId?: string
  src?: string
  dst?: string
}


type BaseCreateFilesetRequest = {
}

export type CreateFilesetRequest = BaseCreateFilesetRequest
  & OneOf<{ appendFile: AppendFile; deleteFile: DeleteFile; copyFile: CopyFile }>

export type CreateFilesetResponse = {
  filesetId?: string
}


type BaseFileFilter = {
}

export type FileFilter = BaseFileFilter
  & OneOf<{ pathRange: PathRange; pathRegex: string }>

export type ReadFilesetRequest = {
  filesetId?: string
  filters?: FileFilter[]
  emptyFiles?: boolean
}

export type ReadFilesetResponse = {
  path?: string
  data?: GoogleProtobufWrappers.BytesValue
}

export type ReadFilesetCDRResponse = {
  path?: string
  ref?: CdrCdr.Ref
}

export type RenewFilesetRequest = {
  filesetId?: string
  ttlSeconds?: string
}

export type ComposeFilesetRequest = {
  filesetIds?: string[]
  ttlSeconds?: string
}

export type ComposeFilesetResponse = {
  filesetId?: string
}

export type ShardFilesetRequest = {
  filesetId?: string
  numFiles?: string
  sizeBytes?: string
}

export type PathRange = {
  lower?: string
  upper?: string
}

export type ShardFilesetResponse = {
  shards?: PathRange[]
}

export class Fileset {
  static ReadFileset(req: ReadFilesetRequest, entityNotifier?: fm.NotifyStreamEntityArrival<ReadFilesetResponse>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ReadFilesetRequest, ReadFilesetResponse>(`/storage.Fileset/ReadFileset`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ReadFilesetCDR(req: ReadFilesetRequest, entityNotifier?: fm.NotifyStreamEntityArrival<ReadFilesetCDRResponse>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ReadFilesetRequest, ReadFilesetCDRResponse>(`/storage.Fileset/ReadFilesetCDR`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static RenewFileset(req: RenewFilesetRequest, initReq?: fm.InitReq): Promise<GoogleProtobufEmpty.Empty> {
    return fm.fetchReq<RenewFilesetRequest, GoogleProtobufEmpty.Empty>(`/storage.Fileset/RenewFileset`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ComposeFileset(req: ComposeFilesetRequest, initReq?: fm.InitReq): Promise<ComposeFilesetResponse> {
    return fm.fetchReq<ComposeFilesetRequest, ComposeFilesetResponse>(`/storage.Fileset/ComposeFileset`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ShardFileset(req: ShardFilesetRequest, initReq?: fm.InitReq): Promise<ShardFilesetResponse> {
    return fm.fetchReq<ShardFilesetRequest, ShardFilesetResponse>(`/storage.Fileset/ShardFileset`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}