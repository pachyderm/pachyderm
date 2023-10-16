/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as Pfs_v2Pfs from "../../../../pfs/pfs.pb"
import * as Pps_v2Pps from "../../../../pps/pps.pb"
import * as DatumDatum from "../../datum/datum.pb"
export type CreateParallelDatumsTask = {
  job?: Pps_v2Pps.Job
  salt?: string
  fileSetId?: string
  baseFileSetId?: string
  pathRange?: Pfs_v2Pfs.PathRange
}

export type CreateParallelDatumsTaskResult = {
  fileSetId?: string
  stats?: DatumDatum.Stats
}

export type CreateSerialDatumsTask = {
  job?: Pps_v2Pps.Job
  salt?: string
  fileSetId?: string
  baseMetaCommit?: Pfs_v2Pfs.Commit
  noSkip?: boolean
  pathRange?: Pfs_v2Pfs.PathRange
}

export type CreateSerialDatumsTaskResult = {
  fileSetId?: string
  outputDeleteFileSetId?: string
  metaDeleteFileSetId?: string
  stats?: DatumDatum.Stats
}

export type CreateDatumSetsTask = {
  fileSetId?: string
  pathRange?: Pfs_v2Pfs.PathRange
  setSpec?: DatumDatum.SetSpec
}

export type CreateDatumSetsTaskResult = {
  datumSets?: Pfs_v2Pfs.PathRange[]
}

export type DatumSetTask = {
  job?: Pps_v2Pps.Job
  fileSetId?: string
  pathRange?: Pfs_v2Pfs.PathRange
  outputCommit?: Pfs_v2Pfs.Commit
}

export type DatumSetTaskResult = {
  outputFileSetId?: string
  metaFileSetId?: string
  stats?: DatumDatum.Stats
}