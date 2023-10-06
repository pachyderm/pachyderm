/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as Pfs_v2Pfs from "../../pfs/pfs.pb"
export type CommitSpec = {
  count?: string
  modifications?: ModificationSpec[]
  fileSources?: FileSourceSpec[]
  validator?: ValidatorSpec
}

export type ModificationSpec = {
  count?: string
  putFile?: PutFileSpec
}

export type PutFileSpec = {
  count?: string
  source?: string
}

export type PutFileTask = {
  count?: string
  fileSource?: FileSourceSpec
  seed?: string
  authToken?: string
}

export type PutFileTaskResult = {
  fileSetId?: string
  hash?: Uint8Array
}

export type FileSourceSpec = {
  name?: string
  random?: RandomFileSourceSpec
}

export type RandomFileSourceSpec = {
  directory?: RandomDirectorySpec
  sizes?: SizeSpec[]
  incrementPath?: boolean
}

export type RandomDirectorySpec = {
  depth?: SizeSpec
  run?: string
}

export type SizeSpec = {
  minSize?: string
  maxSize?: string
  prob?: string
}

export type ValidatorSpec = {
  frequency?: FrequencySpec
}

export type FrequencySpec = {
  count?: string
  prob?: string
}

export type StateCommit = {
  commit?: Pfs_v2Pfs.Commit
  hash?: Uint8Array
}

export type State = {
  commits?: StateCommit[]
}