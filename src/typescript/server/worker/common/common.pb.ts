/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as Pfs_v2Pfs from "../../../pfs/pfs.pb"
export type Input = {
  fileInfo?: Pfs_v2Pfs.FileInfo
  name?: string
  joinOn?: string
  outerJoin?: boolean
  groupBy?: string
  lazy?: boolean
  branch?: string
  emptyFiles?: boolean
  s3?: boolean
}