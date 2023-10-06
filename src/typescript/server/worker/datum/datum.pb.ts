/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as Pfs_v2Pfs from "../../../pfs/pfs.pb"
import * as Pps_v2Pps from "../../../pps/pps.pb"
import * as CommonCommon from "../common/common.pb"

export enum State {
  PROCESSED = "PROCESSED",
  FAILED = "FAILED",
  RECOVERED = "RECOVERED",
}

export enum KeyTaskType {
  JOIN = "JOIN",
  GROUP = "GROUP",
}

export enum MergeTaskType {
  JOIN = "JOIN",
  GROUP = "GROUP",
}

export type Meta = {
  job?: Pps_v2Pps.Job
  inputs?: CommonCommon.Input[]
  hash?: string
  state?: State
  reason?: string
  stats?: Pps_v2Pps.ProcessStats
  index?: string
  imageId?: string
}

export type Stats = {
  processStats?: Pps_v2Pps.ProcessStats
  processed?: string
  skipped?: string
  total?: string
  failed?: string
  recovered?: string
  failedId?: string
}

export type PFSTask = {
  input?: Pps_v2Pps.PFSInput
  pathRange?: Pfs_v2Pfs.PathRange
  baseIndex?: string
  authToken?: string
}

export type PFSTaskResult = {
  fileSetId?: string
}

export type CrossTask = {
  fileSetIds?: string[]
  baseFileSetIndex?: string
  baseFileSetPathRange?: Pfs_v2Pfs.PathRange
  baseIndex?: string
  authToken?: string
}

export type CrossTaskResult = {
  fileSetId?: string
}

export type KeyTask = {
  fileSetId?: string
  pathRange?: Pfs_v2Pfs.PathRange
  type?: KeyTaskType
  authToken?: string
}

export type KeyTaskResult = {
  fileSetId?: string
}

export type MergeTask = {
  fileSetIds?: string[]
  pathRange?: Pfs_v2Pfs.PathRange
  type?: MergeTaskType
  authToken?: string
}

export type MergeTaskResult = {
  fileSetId?: string
}

export type ComposeTask = {
  fileSetIds?: string[]
  authToken?: string
}

export type ComposeTaskResult = {
  fileSetId?: string
}

export type SetSpec = {
  number?: string
  sizeBytes?: string
}