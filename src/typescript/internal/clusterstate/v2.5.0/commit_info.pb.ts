/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as GoogleProtobufTimestamp from "../../../google/protobuf/timestamp.pb"
import * as Pfs_v2Pfs from "../../../pfs/pfs.pb"
export type CommitInfo = {
  commit?: Pfs_v2Pfs.Commit
  origin?: Pfs_v2Pfs.CommitOrigin
  description?: string
  parentCommit?: Pfs_v2Pfs.Commit
  childCommits?: Pfs_v2Pfs.Commit[]
  started?: GoogleProtobufTimestamp.Timestamp
  finishing?: GoogleProtobufTimestamp.Timestamp
  finished?: GoogleProtobufTimestamp.Timestamp
  directProvenance?: Pfs_v2Pfs.Branch[]
  error?: string
  sizeBytesUpperBound?: string
  details?: Pfs_v2Pfs.CommitInfoDetails
}