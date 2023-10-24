/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as IndexIndex from "../../../internal/storage/fileset/index/index.pb"
import * as Pfs_v2Pfs from "../../../pfs/pfs.pb"
export type ShardTask = {
  inputs?: string[]
  pathRange?: PathRange
}

export type ShardTaskResult = {
  compactTasks?: CompactTask[]
}

export type PathRange = {
  lower?: string
  upper?: string
}

export type CompactTask = {
  inputs?: string[]
  pathRange?: PathRange
}

export type CompactTaskResult = {
  id?: string
}

export type ConcatTask = {
  inputs?: string[]
}

export type ConcatTaskResult = {
  id?: string
}

export type ValidateTask = {
  id?: string
  pathRange?: PathRange
}

export type ValidateTaskResult = {
  first?: IndexIndex.Index
  last?: IndexIndex.Index
  error?: string
  sizeBytes?: string
}

export type PutFileURLTask = {
  dst?: string
  datum?: string
  uRL?: string
  paths?: string[]
  startOffset?: string
  endOffset?: string
}

export type PutFileURLTaskResult = {
  id?: string
}

export type GetFileURLTask = {
  uRL?: string
  fileset?: string
  pathRange?: Pfs_v2Pfs.PathRange
}

export type GetFileURLTaskResult = {
}