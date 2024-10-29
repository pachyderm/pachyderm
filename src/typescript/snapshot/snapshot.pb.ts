/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb"
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb"
export type CreateSnapshotRequest = {
  metadata?: {[key: string]: string}
}

export type CreateSnapshotResponse = {
  id?: string
}

export type DeleteSnapshotRequest = {
  id?: string
}

export type DeleteSnapshotResponse = {
}

export type SnapshotInfo = {
  id?: string
  metadata?: {[key: string]: string}
  chunksetId?: string
  pachydermVersion?: string
  createdAt?: GoogleProtobufTimestamp.Timestamp
}

export type InspectSnapshotRequest = {
  id?: string
}

export type InspectSnapshotResponse = {
  info?: SnapshotInfo
  fileset?: string
}

export type ListSnapshotRequest = {
  since?: GoogleProtobufTimestamp.Timestamp
  limit?: number
}

export type ListSnapshotResponse = {
  info?: SnapshotInfo
}

export class API {
  static CreateSnapshot(req: CreateSnapshotRequest, initReq?: fm.InitReq): Promise<CreateSnapshotResponse> {
    return fm.fetchReq<CreateSnapshotRequest, CreateSnapshotResponse>(`/snapshot.API/CreateSnapshot`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static DeleteSnapshot(req: DeleteSnapshotRequest, initReq?: fm.InitReq): Promise<DeleteSnapshotResponse> {
    return fm.fetchReq<DeleteSnapshotRequest, DeleteSnapshotResponse>(`/snapshot.API/DeleteSnapshot`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static InspectSnapshot(req: InspectSnapshotRequest, initReq?: fm.InitReq): Promise<InspectSnapshotResponse> {
    return fm.fetchReq<InspectSnapshotRequest, InspectSnapshotResponse>(`/snapshot.API/InspectSnapshot`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
  static ListSnapshot(req: ListSnapshotRequest, entityNotifier?: fm.NotifyStreamEntityArrival<ListSnapshotResponse>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ListSnapshotRequest, ListSnapshotResponse>(`/snapshot.API/ListSnapshot`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}