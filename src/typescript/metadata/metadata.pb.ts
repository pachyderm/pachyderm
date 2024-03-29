/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb"
import * as Pfs_v2Pfs from "../pfs/pfs.pb"

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (
    keyof T extends infer K ?
      (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
        : never)
    : never);
export type EditReplace = {
  replacement?: {[key: string]: string}
}

export type EditAddKey = {
  key?: string
  value?: string
}

export type EditEditKey = {
  key?: string
  value?: string
}

export type EditDeleteKey = {
  key?: string
}


type BaseEdit = {
}

export type Edit = BaseEdit
  & OneOf<{ project: Pfs_v2Pfs.ProjectPicker; commit: Pfs_v2Pfs.CommitPicker; branch: Pfs_v2Pfs.BranchPicker; repo: Pfs_v2Pfs.RepoPicker }>
  & OneOf<{ replace: EditReplace; addKey: EditAddKey; editKey: EditEditKey; deleteKey: EditDeleteKey }>

export type EditMetadataRequest = {
  edits?: Edit[]
}

export type EditMetadataResponse = {
}

export class API {
  static EditMetadata(req: EditMetadataRequest, initReq?: fm.InitReq): Promise<EditMetadataResponse> {
    return fm.fetchReq<EditMetadataRequest, EditMetadataResponse>(`/metadata.API/EditMetadata`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}